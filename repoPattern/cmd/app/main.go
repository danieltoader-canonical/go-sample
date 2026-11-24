package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"repoPattern/internal/cmd/root"
	"repoPattern/internal/dbstore/customer"
	"repoPattern/pkg/kvstore"
	"repoPattern/pkg/postgres"
)

func main() {
	if err := root.NewRootCommand(newRepository).Execute(); err != nil {
		slog.Error("customer example failed", "error", err)
		os.Exit(1)
	}
}

func newRepository(ctx context.Context) (customer.Repository, func(), error) {
	client, err := postgres.NewClient(ctx, envOrDefault("POSTGRES_URI", "postgres://postgres@localhost:"+envOrDefault("POSTGRES_PORT", "5432")+"/sample?sslmode=disable"))
	if err != nil {
		return nil, nil, err
	}
	store, err := kvstore.NewValKeyStore(
		kvstore.Config{Addrs: []string{envOrDefault("VALKEY_ADDR", "localhost:"+envOrDefault("VALKEY_PORT", "6379"))}},
		kvstore.WithCacheKeyPrefix("repoPattern:"),
	)
	if err != nil {
		client.Close()
		return nil, nil, err
	}
	cleanup := func() {
		store.CloseWithCheck(context.Background())
		client.Close()
	}
	if err := store.Ping(ctx); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("ping cache: %w", err)
	}

	repo := customer.NewRepoWithCache(customer.NewRepo(client), store, customer.WithRefreshCache())
	return repo, cleanup, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
