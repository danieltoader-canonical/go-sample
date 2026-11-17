package kvstore

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore implements Store using go-redis UniversalClient, supporting cluster and single-node setups.
type RedisStore struct {
	client redis.UniversalClient
	opts   KVOptions
}

// NewRedisStore creates a new RedisStore using go-redis UniversalClient.
// If cfg.Addrs has more than one element, a cluster setup is assumed.
func NewRedisStore(cfg Config, optFns ...Option) (*RedisStore, error) {
	ropts := &redis.UniversalOptions{
		Addrs:    cfg.Addrs,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
		ReadOnly: cfg.ReadOnly,
	}
	if cfg.TLS {
		ropts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	c := redis.NewUniversalClient(ropts)
	// apply functional options
	var opts KVOptions
	for _, fn := range optFns {
		if fn != nil {
			fn(&opts)
		}
	}
	return &RedisStore{client: c, opts: opts}, nil
}

// Get fetches the string value for the given key or returns ErrNotFound if missing.
func (s *RedisStore) Get(ctx context.Context, key string) (string, error) {
	var cancel context.CancelFunc
	ctx, cancel = withTimeout(ctx, s.opts.ReadTimeout)
	defer cancel()
	key = s.PrefixedKey(key)
	val, err := s.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrNotFound
	}
	return val, err
}

// Set sets the value for key with an optional expiration (0 means no TTL).
func (s *RedisStore) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var cancel context.CancelFunc
	ctx, cancel = withTimeout(ctx, s.opts.WriteTimeout)
	defer cancel()
	key = s.PrefixedKey(key)
	return s.client.Set(ctx, key, value, expiration).Err()
}

// MGet fetches multiple keys in a single call; missing keys yield nil values.
func (s *RedisStore) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	var cancel context.CancelFunc
	ctx, cancel = withTimeout(ctx, s.opts.ReadTimeout)
	defer cancel()
	keys = s.pkAll(keys)
	return s.client.MGet(ctx, keys...).Result()
}

// SetNX sets key only if it does not exist and returns true when the value was set.
func (s *RedisStore) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	var cancel context.CancelFunc
	ctx, cancel = withTimeout(ctx, s.opts.WriteTimeout)
	defer cancel()
	key = s.PrefixedKey(key)
	return s.client.SetNX(ctx, key, value, expiration).Result()
}

// Incr atomically increments the integer value at key, creating it if missing.
func (s *RedisStore) Incr(ctx context.Context, key string) (int64, error) {
	var cancel context.CancelFunc
	ctx, cancel = withTimeout(ctx, s.opts.WriteTimeout)
	defer cancel()
	key = s.PrefixedKey(key)
	return s.client.Incr(ctx, key).Result()
}

// Del removes one or more keys and returns the number of keys removed.
func (s *RedisStore) Del(ctx context.Context, keys ...string) (int64, error) {
	var cancel context.CancelFunc
	ctx, cancel = withTimeout(ctx, s.opts.DeleteTimeout)
	defer cancel()
	keys = s.pkAll(keys)
	return s.client.Del(ctx, keys...).Result()
}

// Expire updates the key's time-to-live; returns true if the TTL was set.
func (s *RedisStore) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	var cancel context.CancelFunc
	ctx, cancel = withTimeout(ctx, s.opts.WriteTimeout)
	defer cancel()
	key = s.PrefixedKey(key)
	return s.client.Expire(ctx, key, expiration).Result()
}

// TTL returns the remaining TTL for key, ErrNotFound if the key does not exist.
func (s *RedisStore) TTL(ctx context.Context, key string) (time.Duration, error) {
	var cancel context.CancelFunc
	ctx, cancel = withTimeout(ctx, s.opts.ReadTimeout)
	defer cancel()
	key = s.PrefixedKey(key)
	ttl, err := s.client.TTL(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return 0, ErrNotFound
	}
	return ttl, err
}

// Ping checks connectivity to the Redis server.
func (s *RedisStore) Ping(ctx context.Context) error {
	var cancel context.CancelFunc
	ctx, cancel = withTimeout(ctx, s.opts.ReadTimeout)
	defer cancel()
	return s.client.Ping(ctx).Err()
}

// Close closes the underlying client and frees resources.
func (s *RedisStore) Close() error {
	return s.client.Close()
}

// CloseWithCheck calls Close and logs any error.
func (s *RedisStore) CloseWithCheck(ctx context.Context) {
	err := s.Close()
	if err != nil {
		slog.ErrorContext(ctx, "redis: failed to close connection", "error", err)
	}
}

// Underlying exposes the underlying go-redis UniversalClient.
func (s *RedisStore) Underlying() redis.UniversalClient { return s.client }

// prefixedKey returns the prefixed key if a prefix is configured.
func (s *RedisStore) PrefixedKey(key string) string {
	if p := s.opts.CacheKeyPrefix; p != "" {
		return p + key
	}
	return key
}

// pkAll applies the prefix in-place to all keys and returns the slice for chaining.
func (s *RedisStore) pkAll(keys []string) []string {
	if s.opts.CacheKeyPrefix == "" {
		return keys
	}
	for i := range keys {
		keys[i] = s.opts.CacheKeyPrefix + keys[i]
	}
	return keys
}
