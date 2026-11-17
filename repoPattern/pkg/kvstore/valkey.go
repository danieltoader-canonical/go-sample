package kvstore

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/valkey-io/valkey-go"
)

// ValKeyStore implements Store using valkey-go client, supporting cluster and single-node setups.
// We attempt to mimic go-redis UniversalClient behavior: if multiple Addrs are provided, use cluster client.
type ValKeyStore struct {
	client valkey.Client
	opts   KVOptions
}

// NewValKeyStore creates a new ValKeyStore.
func NewValKeyStore(cfg Config, optFns ...Option) (*ValKeyStore, error) {
	// Build options. valkey-go exposes options per client type; select based on number of addrs.
	var c valkey.Client
	var err error

	// TLS config
	var tlsConf *tls.Config
	if cfg.TLS {
		tlsConf = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	// Build valkey ClientOption and create client
	c, err = valkey.NewClient(valkey.ClientOption{
		InitAddress: cfg.Addrs,
		TLSConfig:   tlsConf,
		Username:    cfg.Username,
		Password:    cfg.Password,
		SelectDB:    cfg.DB,
		ReplicaOnly: cfg.ReadOnly,
	})
	if err != nil {
		return nil, err
	}
	// apply functional options
	var opts KVOptions
	for _, fn := range optFns {
		if fn != nil {
			fn(&opts)
		}
	}
	return &ValKeyStore{client: c, opts: opts}, nil
}

// Get fetches the string value for the given key or returns ErrNotFound if missing.
func (s *ValKeyStore) Get(ctx context.Context, key string) (string, error) {
	var cancel context.CancelFunc
	ctx, cancel = withTimeout(ctx, s.opts.ReadTimeout)
	defer cancel()
	key = s.PrefixedKey(key)
	resp := s.client.Do(ctx, s.client.B().Get().Key(key).Build())
	val, err := resp.ToString()
	if err != nil {
		if valkey.IsValkeyNil(err) || errors.Is(err, valkey.Nil) {
			return "", ErrNotFound
		}
		return "", err
	}
	return val, nil
}

// Set sets the value for key with an optional expiration (0 means no TTL).
func (s *ValKeyStore) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var cancel context.CancelFunc
	ctx, cancel = withTimeout(ctx, s.opts.WriteTimeout)
	defer cancel()
	key = s.PrefixedKey(key)
	b := s.client.B().Set().Key(key).Value(fmt.Sprint(value))
	if expiration > 0 {
		return s.client.Do(ctx, b.Ex(expiration).Build()).Error()
	}
	return s.client.Do(ctx, b.Build()).Error()
}

// MGet fetches multiple keys in a single call; missing keys yield nil values.
func (s *ValKeyStore) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	var cancel context.CancelFunc
	ctx, cancel = withTimeout(ctx, s.opts.ReadTimeout)
	defer cancel()
	keys = s.pkAll(keys)
	arr, err := s.client.Do(ctx, s.client.B().Mget().Key(keys...).Build()).ToArray()
	if err != nil {
		return nil, err
	}
	vals := make([]interface{}, len(arr))
	for i, m := range arr {
		s, e := m.ToString()
		if e == nil {
			vals[i] = s
			continue
		}
		if valkey.IsValkeyNil(e) || errors.Is(e, valkey.Nil) {
			vals[i] = nil
			continue
		}
		return nil, e
	}
	return vals, nil
}

// SetNX sets key only if it does not exist and returns true when the value was set.
func (s *ValKeyStore) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	var cancel context.CancelFunc
	ctx, cancel = withTimeout(ctx, s.opts.WriteTimeout)
	defer cancel()
	key = s.PrefixedKey(key)
	b := s.client.B().Set().Key(key).Value(fmt.Sprint(value))
	var err error
	cond := b.Nx()
	if expiration > 0 {
		err = s.client.Do(ctx, cond.Ex(expiration).Build()).Error()
	} else {
		err = s.client.Do(ctx, cond.Build()).Error()
	}
	if err == nil {
		return true, nil
	}
	if valkey.IsValkeyNil(err) || errors.Is(err, valkey.Nil) {
		return false, nil
	}
	return false, err
}

// Incr atomically increments the integer value at key, creating it if missing.
func (s *ValKeyStore) Incr(ctx context.Context, key string) (int64, error) {
	var cancel context.CancelFunc
	ctx, cancel = withTimeout(ctx, s.opts.WriteTimeout)
	defer cancel()
	key = s.PrefixedKey(key)
	n, err := s.client.Do(ctx, s.client.B().Incr().Key(key).Build()).ToInt64()
	return n, err
}

// Del removes one or more keys and returns the number of keys removed.
func (s *ValKeyStore) Del(ctx context.Context, keys ...string) (int64, error) {
	var cancel context.CancelFunc
	ctx, cancel = withTimeout(ctx, s.opts.DeleteTimeout)
	defer cancel()
	keys = s.pkAll(keys)
	n, err := s.client.Do(ctx, s.client.B().Del().Key(keys...).Build()).ToInt64()
	return n, err
}

// Expire updates the key's time-to-live; returns true if the TTL was set.
func (s *ValKeyStore) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	var cancel context.CancelFunc
	ctx, cancel = withTimeout(ctx, s.opts.WriteTimeout)
	defer cancel()
	key = s.PrefixedKey(key)
	secs := int64(expiration / time.Second)
	ok, err := s.client.Do(ctx, s.client.B().Expire().Key(key).Seconds(secs).Build()).ToBool()
	return ok, err
}

// TTL returns the remaining TTL for key, ErrNotFound if the key does not exist.
func (s *ValKeyStore) TTL(ctx context.Context, key string) (time.Duration, error) {
	var cancel context.CancelFunc
	ctx, cancel = withTimeout(ctx, s.opts.ReadTimeout)
	defer cancel()
	key = s.PrefixedKey(key)
	secs, err := s.client.Do(ctx, s.client.B().Ttl().Key(key).Build()).ToInt64()
	if err != nil {
		return 0, err
	}
	if secs == -2 { // key does not exist
		return 0, ErrNotFound
	}
	if secs == -1 { // no TTL
		return -1, nil
	}
	return time.Duration(secs) * time.Second, nil
}

// Ping checks connectivity to the Valkey server.
func (s *ValKeyStore) Ping(ctx context.Context) error {
	var cancel context.CancelFunc
	ctx, cancel = withTimeout(ctx, s.opts.ReadTimeout)
	defer cancel()
	return s.client.Do(ctx, s.client.B().Ping().Build()).Error()
}

// Close closes the underlying client and frees resources.
func (s *ValKeyStore) Close() error {
	s.client.Close()
	return nil
}

// CloseWithCheck calls Close and logs any error.
func (s *ValKeyStore) CloseWithCheck(ctx context.Context) {
	err := s.Close()
	if err != nil {
		slog.ErrorContext(ctx, "valkey: failed to close connection", "error", err)
	}
}

// Underlying returns the underlying valkey-go client for advanced use-cases.
func (s *ValKeyStore) Underlying() valkey.Client { return s.client }

// prefixedKey returns the prefixed key if a prefix is configured.
func (s *ValKeyStore) PrefixedKey(key string) string {
	if p := s.opts.CacheKeyPrefix; p != "" {
		return p + key
	}
	return key
}

// pkAll applies the prefix in-place to all keys and returns the slice for chaining.
func (s *ValKeyStore) pkAll(keys []string) []string {
	if s.opts.CacheKeyPrefix == "" {
		return keys
	}
	for i := range keys {
		keys[i] = s.opts.CacheKeyPrefix + keys[i]
	}
	return keys
}
