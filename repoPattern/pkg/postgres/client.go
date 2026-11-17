package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=client.go -destination=client_mock.go -package=postgres

var ErrNoRows = pgx.ErrNoRows

type Row interface {
	Scan(...any) error
}

type Client interface {
	QueryRow(context.Context, string, ...any) Row
	GetOptions() *Options
	Ping(context.Context) error
	Close()
}

type Options struct {
	ReadTimeoutMedium  time.Duration
	WriteTimeoutMedium time.Duration
}

type PostgresClient struct {
	pool    *pgxpool.Pool
	options Options
}

func NewClient(ctx context.Context, uri string) (*PostgresClient, error) {
	if uri == "" {
		return nil, fmt.Errorf("PostgreSQL URI is required")
	}
	config, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return nil, fmt.Errorf("parse PostgreSQL configuration: %w", err)
	}
	if config.ConnConfig.StatementCacheCapacity <= 0 {
		return nil, fmt.Errorf("prepared statements require a positive statement_cache_capacity")
	}
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheStatement
	config.ConnConfig.ConnectTimeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create PostgreSQL pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping PostgreSQL: %w", err)
	}
	return &PostgresClient{
		pool: pool,
		options: Options{
			ReadTimeoutMedium: 5 * time.Second, WriteTimeoutMedium: 5 * time.Second,
		},
	}, nil
}

func (c *PostgresClient) QueryRow(ctx context.Context, sql string, args ...any) Row {
	return c.pool.QueryRow(ctx, sql, args...)
}

func (c *PostgresClient) GetOptions() *Options { return &c.options }

func (c *PostgresClient) Ping(ctx context.Context) error { return c.pool.Ping(ctx) }

func (c *PostgresClient) Close() { c.pool.Close() }
