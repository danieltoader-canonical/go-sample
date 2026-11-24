package kvstore

import (
	"context"
	"time"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=interface.go -destination=store_mock.go -package=kvstore

// Store is a minimal key/value storage abstraction used across the project.
// Implementations: redis (go-redis), valkey (valkey-go), mem (hypermodeinc/ristretto).
// Methods are intentionally aligned with common Redis semantics.
// All methods should be safe for concurrent use.
//
// Notes:
//   - TTL of 0 means no expiration.
//   - For MGet, returns values in the same order as keys; missing keys yield nil entries.
//   - For SetNX, returns true if the key was set.
//   - Close should release any resources; it is safe to call multiple times.
//   - Ping should validate connectivity where applicable (no-op for memory store).
//   - Incr creates the key with value 1 if it does not exist.
//   - Del returns the number of removed keys.
//   - TTL returns remaining time to live or a negative value if no TTL is set; implementations may return (0, ErrNotFound) if the key does not exist.
//     We avoid importing redis.Nil here and define ErrNotFound locally to keep the interface generic.
type Store interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	MGet(ctx context.Context, keys ...string) ([]interface{}, error)
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	Incr(ctx context.Context, key string) (int64, error)
	Del(ctx context.Context, keys ...string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) (bool, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
	Ping(ctx context.Context) error
	Close() error
	CloseWithCheck(ctx context.Context)

	// Internal helpers used by in-package implementations; included so types are recognized.
	PrefixedKey(key string) string
	pkAll(keys []string) []string
}

// Config contains generic connection options used by Redis/Valkey implementations.
// For cluster deployments, provide multiple Addrs.
// For single node deployments, provide a single Addr.
// TLS controls whether to use TLS when connecting.
// DB is relevant for single-node setups; cluster deployments typically ignore DB index.
type Config struct {
	Addrs    []string
	Username string
	Password string
	DB       int
	TLS      bool
	// ReadOnly enables read-only routing to replicas for cluster-aware clients where supported.
	ReadOnly bool
}

// KVOptions contains cross-cutting client options that apply to all Store implementations.
// These should NOT be set via Config.
type KVOptions struct {
	// CacheKeyPrefix will be prepended to every key by the store implementation.
	CacheKeyPrefix string

	// Operation timeouts: if zero, no deadline is applied by the store implementation.
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	DeleteTimeout time.Duration
}

// Option represents a functional option that mutates KVOptions.
// Use With* helpers to configure the store behavior in constructors.
// Example:
//
//	NewRedisStore(cfg, WithCacheKeyPrefix("app:"), WithReadTimeout(50*time.Millisecond))
//
// Zero values are respected: passing a zero duration means no timeout.
// If the same option is applied multiple times, the last one wins.
//
// Note: Constructors accept a variadic list of Option.
type Option func(*KVOptions)

// WithCacheKeyPrefix sets the client key prefix to be prepended to all keys.
func WithCacheKeyPrefix(prefix string) Option {
	return func(o *KVOptions) { o.CacheKeyPrefix = prefix }
}

// WithReadTimeout sets the read timeout applied via context.WithTimeout for read operations.
func WithReadTimeout(d time.Duration) Option {
	return func(o *KVOptions) { o.ReadTimeout = d }
}

// WithWriteTimeout sets the write timeout applied via context.WithTimeout for write operations.
func WithWriteTimeout(d time.Duration) Option {
	return func(o *KVOptions) { o.WriteTimeout = d }
}

// WithDeleteTimeout sets the delete timeout applied via context.WithTimeout for delete operations.
func WithDeleteTimeout(d time.Duration) Option {
	return func(o *KVOptions) { o.DeleteTimeout = d }
}

// WithTimeouts sets all three operation timeouts at once.
func WithTimeouts(readTO, writeTO, deleteTO time.Duration) Option {
	return func(o *KVOptions) {
		o.ReadTimeout = readTO
		o.WriteTimeout = writeTO
		o.DeleteTimeout = deleteTO
	}
}

// WithOptions applies all option fields at once from the provided KVOptions struct.
// If combined with other Option functions, the usual "last one wins" rule applies per field
// because this replaces the entire struct state at the time it is applied.
func WithOptions(all KVOptions) Option {
	return func(o *KVOptions) { *o = all }
}

// ErrNotFound is returned when a key is missing.
// Implementations may map underlying client-specific sentinel errors to this value.
type notFoundError struct{}

func (notFoundError) Error() string { return "kvstore: key not found" }

// ErrNotFound is the sentinel error returned when a key is not found.
var ErrNotFound error = notFoundError{}

// withTimeout wraps context.WithTimeout and returns a no-op cancel when d <= 0.
// This helps reduce duplicated timeout handling code across implementations.
func withTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if d <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, d)
}
