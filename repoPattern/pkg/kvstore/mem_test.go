package kvstore

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemStore_BasicOperations_Table(t *testing.T) {
	ctx := context.Background()
	store, err := NewMemStore(Config{}, WithCacheKeyPrefix("app:"))
	if err != nil {
		t.Fatalf("NewMemStore error: %v", err)
	}
	// ensure cleanup
	defer func() { _ = store.Close() }()

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "Set/Get without TTL",
			run: func(t *testing.T) {
				if err := store.Set(ctx, "k1", "v1", 0); err != nil {
					t.Fatalf("set: %v", err)
				}
				got, err := store.Get(ctx, "k1")
				if err != nil {
					t.Fatalf("get: %v", err)
				}
				if got != "v1" {
					t.Fatalf("want v1 got %s", got)
				}
			},
		},
		{
			name: "Get not found returns ErrNotFound",
			run: func(t *testing.T) {
				_, err := store.Get(ctx, "missing")
				if !errors.Is(err, ErrNotFound) {
					t.Fatalf("want ErrNotFound got %v", err)
				}
			},
		},
		{
			name: "SetNX sets first time only",
			run: func(t *testing.T) {
				ok, err := store.SetNX(ctx, "nx", "1", 0)
				if err != nil || !ok {
					t.Fatalf("setnx #1: ok %v err %v", ok, err)
				}
				ok, err = store.SetNX(ctx, "nx", "2", 0)
				if err != nil || ok {
					t.Fatalf("setnx #2 should be false, got ok %v err %v", ok, err)
				}
				v, _ := store.Get(ctx, "nx")
				if v != "1" {
					t.Fatalf("value changed unexpectedly: %s", v)
				}
			},
		},
		{
			name: "Incr creates and increments",
			run: func(t *testing.T) {
				v, err := store.Incr(ctx, "ctr")
				if err != nil || v != 1 {
					t.Fatalf("incr #1: v %d err %v", v, err)
				}
				v, err = store.Incr(ctx, "ctr")
				if err != nil || v != 2 {
					t.Fatalf("incr #2: v %d err %v", v, err)
				}
			},
		},
		{
			name: "MGet returns values and nils in order",
			run: func(t *testing.T) {
				_ = store.Set(ctx, "a", "A", 0)
				vals, err := store.MGet(ctx, "a", "b", "k1")
				if err != nil {
					t.Fatalf("mget: %v", err)
				}
				if len(vals) != 3 {
					t.Fatalf("mget len: %d", len(vals))
				}
				if vals[0] != "A" {
					t.Fatalf("want A got %v", vals[0])
				}
				if vals[1] != nil {
					t.Fatalf("want nil for missing got %v", vals[1])
				}
				if vals[2] != "v1" {
					t.Fatalf("want v1 got %v", vals[2])
				}
			},
		},
		{
			name: "Expire and TTL work",
			run: func(t *testing.T) {
				_ = store.Set(ctx, "ttlkey", "x", 0)
				ok, err := store.Expire(ctx, "ttlkey", 50*time.Millisecond)
				if err != nil || !ok {
					t.Fatalf("expire: ok %v err %v", ok, err)
				}
				d, err := store.TTL(ctx, "ttlkey")
				if err != nil || d <= 0 {
					t.Fatalf("ttl should be >0, got %v err %v", d, err)
				}
				time.Sleep(60 * time.Millisecond)
				_, err = store.Get(ctx, "ttlkey")
				if !errors.Is(err, ErrNotFound) {
					t.Fatalf("expired key should be missing, got %v", err)
				}
			},
		},
		{
			name: "Del removes multiple keys",
			run: func(t *testing.T) {
				_ = store.Set(ctx, "d1", "1", 0)
				_ = store.Set(ctx, "d2", "2", 0)
				n, err := store.Del(ctx, "d1", "d2", "missing")
				if err != nil || n != 3 {
					t.Fatalf("del n %d err %v", n, err)
				}
			},
		},
		{
			name: "PrefixedKey and pkAll apply prefix",
			run: func(t *testing.T) {
				if got := store.PrefixedKey("raw"); got != "app:raw" {
					t.Fatalf("prefixed got %s", got)
				}
				keys := store.pkAll([]string{"k1", "k2"})
				if keys[0] != "app:k1" || keys[1] != "app:k2" {
					t.Fatalf("pkAll wrong: %v", keys)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
