package kvstore

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

// MemStore implements Store using an in-memory ristretto cache.
// It is safe for concurrent use by multiple goroutines.
type MemStore struct {
	client *ristretto.Cache[string, interface{}]
	mu     sync.RWMutex
	// expirations stores absolute UnixNano deadlines for keys with TTL; 0 means no expiration.
	expirations map[string]int64

	opts KVOptions
}

// NewMemStore creates a MemStore with ristretto using the shared Config options.
func NewMemStore(_ Config, optFns ...Option) (*MemStore, error) {
	// Sensible defaults for ristretto
	numCounters := int64(1e5)
	maxCost := int64(1 << 28) // 256MB
	bufferItems := int64(64)

	c, err := ristretto.NewCache[string, interface{}](&ristretto.Config[string, interface{}]{
		NumCounters: numCounters,
		MaxCost:     maxCost,
		BufferItems: bufferItems,
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
	return &MemStore{client: c, expirations: make(map[string]int64), opts: opts}, nil
}

// Get fetches the string value for the given key or returns ErrNotFound if missing.
func (m *MemStore) Get(ctx context.Context, key string) (string, error) {
	var cancel context.CancelFunc
	ctx, cancel = withTimeout(ctx, m.opts.ReadTimeout)
	defer cancel()
	pk := m.PrefixedKey(key)
	if m.isExpired(pk) {
		if _, err := m.Del(context.Background(), key); err != nil {
			return "", err
		}
		return "", ErrNotFound
	}
	if v, ok := m.client.Get(pk); ok {
		if s, ok2 := v.(string); ok2 {
			return s, nil
		}
		return "", nil
	}
	return "", ErrNotFound
}

// Set sets the value for key with an optional expiration (0 means no TTL).
func (m *MemStore) Set(_ context.Context, key string, value interface{}, expiration time.Duration) error {
	pk := m.PrefixedKey(key)
	deadline := int64(0)
	if expiration > 0 {
		deadline = time.Now().Add(expiration).UnixNano()
	}
	m.mu.Lock()
	m.expirations[pk] = deadline
	m.mu.Unlock()
	if expiration > 0 {
		m.client.SetWithTTL(pk, value, 1, expiration)
	} else {
		m.client.Set(pk, value, 1)
	}
	m.client.Wait()
	return nil
}

// MGet fetches multiple keys in a single call; missing keys yield nil values.
func (m *MemStore) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	var cancel context.CancelFunc
	ctx, cancel = withTimeout(ctx, m.opts.ReadTimeout)
	defer cancel()
	vals := make([]interface{}, len(keys))
	now := time.Now().UnixNano()
	for i, k := range keys {
		pk := m.PrefixedKey(k)
		m.mu.RLock()
		dl := m.expirations[pk]
		m.mu.RUnlock()
		if dl > 0 && dl <= now {
			if _, err := m.Del(context.Background(), k); err != nil {
				return nil, err
			}
			vals[i] = nil
			continue
		}
		if v, ok := m.client.Get(pk); ok {
			vals[i] = v
		} else {
			vals[i] = nil
		}
	}
	return vals, nil
}

// SetNX sets key only if it does not exist and returns true when the value was set.
func (m *MemStore) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	var cancel context.CancelFunc
	ctx, cancel = withTimeout(ctx, m.opts.WriteTimeout)
	defer cancel()
	_, err := m.Get(ctx, key)
	if err == nil {
		return false, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return false, err
	}
	return true, m.Set(ctx, key, value, expiration)
}

// Incr atomically increments the integer value at key, creating it if missing.
func (m *MemStore) Incr(ctx context.Context, key string) (int64, error) {
	var cancel context.CancelFunc
	ctx, cancel = withTimeout(ctx, m.opts.WriteTimeout)
	defer cancel()
	// Simple, safe approach for demo: get/parse/set loop under lock.
	m.mu.Lock()
	defer m.mu.Unlock()
	pk := m.PrefixedKey(key)
	var cur int64
	if v, ok := m.client.Get(pk); ok {
		switch t := v.(type) {
		case int64:
			cur = t
		case string:
			if parsed, perr := parseInt64(t); perr == nil {
				cur = parsed
			}
		}
	}
	cur++
	m.client.Set(pk, cur, 1)
	m.client.Wait()
	return cur, nil
}

// Del removes one or more keys and returns the number of keys removed.
func (m *MemStore) Del(ctx context.Context, keys ...string) (int64, error) {
	var cancel context.CancelFunc
	ctx, cancel = withTimeout(ctx, m.opts.DeleteTimeout)
	defer cancel()
	var n int64
	for _, k := range keys {
		pk := m.PrefixedKey(k)
		m.client.Del(pk)
		m.mu.Lock()
		delete(m.expirations, pk)
		m.mu.Unlock()
		n++
	}
	return n, nil
}

// Expire updates the key's time-to-live; returns true if the TTL was set.
func (m *MemStore) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	var cancel context.CancelFunc
	ctx, cancel = withTimeout(ctx, m.opts.WriteTimeout)
	defer cancel()
	pk := m.PrefixedKey(key)
	if expiration <= 0 {
		m.mu.Lock()
		m.expirations[pk] = 0
		m.mu.Unlock()
		return true, nil
	}
	m.mu.Lock()
	m.expirations[pk] = time.Now().Add(expiration).UnixNano()
	m.mu.Unlock()
	// Re-set with TTL if exists
	if v, ok := m.client.Get(pk); ok {
		m.client.SetWithTTL(pk, v, 1, expiration)
		m.client.Wait()
		return true, nil
	}
	return false, ErrNotFound
}

// TTL returns the remaining TTL for key, ErrNotFound if the key does not exist.
func (m *MemStore) TTL(ctx context.Context, key string) (time.Duration, error) {
	var cancel context.CancelFunc
	ctx, cancel = withTimeout(ctx, m.opts.ReadTimeout)
	defer cancel()
	pk := m.PrefixedKey(key)
	if m.isExpired(pk) {
		if _, err := m.Del(context.Background(), key); err != nil {
			return 0, err
		}
		return 0, ErrNotFound
	}
	m.mu.RLock()
	dl := m.expirations[pk]
	m.mu.RUnlock()
	if dl == 0 {
		return -1, nil // no TTL
	}
	rem := time.Until(time.Unix(0, dl))
	if rem < 0 {
		return 0, ErrNotFound
	}
	return rem, nil
}

// Ping is a no-op for the in-memory store and always returns nil.
func (m *MemStore) Ping(_ context.Context) error { return nil }

// Close frees internal resources used by the in-memory client.
func (m *MemStore) Close() error {
	m.client.Close()
	return nil
}

// CloseWithCheck calls Close and logs any error.
func (m *MemStore) CloseWithCheck(ctx context.Context) {
	err := m.Close()
	if err != nil {
		slog.ErrorContext(ctx, "mem: failed to close connection", "error", err)
	}
}

func (m *MemStore) isExpired(key string) bool {
	now := time.Now().UnixNano()
	m.mu.RLock()
	dl := m.expirations[key]
	m.mu.RUnlock()
	return dl > 0 && dl <= now
}

func parseInt64(s string) (int64, error) {
	var n int64
	var sign int64 = 1
	for i, b := range []byte(s) {
		if i == 0 && (b == '+' || b == '-') {
			if b == '-' {
				sign = -1
			}
			continue
		}
		if b < '0' || b > '9' {
			return 0, ErrNotFound
		}
		n = n*10 + int64(b-'0')
	}
	return sign * n, nil
}

// prefixedKey returns the prefixed key if a prefix is configured.
func (m *MemStore) PrefixedKey(key string) string {
	if p := m.opts.CacheKeyPrefix; p != "" {
		return p + key
	}
	return key
}

// pkAll applies the prefix in-place to all keys and returns the slice for chaining.
func (m *MemStore) pkAll(keys []string) []string {
	if m.opts.CacheKeyPrefix == "" {
		return keys
	}
	for i := range keys {
		keys[i] = m.opts.CacheKeyPrefix + keys[i]
	}
	return keys
}
