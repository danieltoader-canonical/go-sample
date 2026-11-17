package customer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"repoPattern/pkg/kvstore"
)

const customerCacheTTL = 24 * time.Hour

type RepoCache struct {
	Repository
	storage      kvstore.Store
	refreshCache bool
}

type RepoCacheOption func(*RepoCache)

var _ Repository = (*RepoCache)(nil)

func NewRepoWithCache(repo Repository, storage kvstore.Store, opts ...RepoCacheOption) *RepoCache {
	r := &RepoCache{Repository: repo, storage: storage}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func WithRefreshCache() RepoCacheOption {
	return func(r *RepoCache) { r.refreshCache = true }
}

func (r *RepoCache) GetCustomer(ctx context.Context, customerId int) Customer {
	if customerId == 0 {
		return Customer{}
	}
	key := getCustomerCacheKey(customerId)
	if !r.refreshCache {
		value, err := r.storage.Get(ctx, key)
		if err == nil && value != "" {
			var customer Customer
			if decodeErr := json.Unmarshal([]byte(value), &customer); decodeErr == nil && customer.CustomerId == customerId {
				return customer
			} else {
				slog.WarnContext(ctx, "invalid customer cache entry", "customerId", customerId, "error", decodeErr)
			}
		} else if err != nil && !errors.Is(err, kvstore.ErrNotFound) {
			slog.WarnContext(ctx, "unable to read customer cache", "customerId", customerId, "error", err)
		}
	}

	customer := r.Repository.GetCustomer(ctx, customerId)
	if customer.CustomerId != 0 {
		data, err := json.Marshal(customer)
		if err == nil {
			// Use a string so Redis, Valkey and memory stores share the same encoding.
			err = r.storage.Set(ctx, key, string(data), customerCacheTTL)
		}
		if err != nil {
			slog.WarnContext(ctx, "unable to cache customer", "customerId", customerId, "error", err)
		}
	}
	return customer
}

func (r *RepoCache) UpdateCustomerName(ctx context.Context, customerId int, name string) (*Customer, error) {
	customer, err := r.Repository.UpdateCustomerName(ctx, customerId, name)
	if err != nil {
		return nil, err
	}
	return customer, r.invalidate(ctx, customerId)
}

func (r *RepoCache) UpdateCustomerAnonymousStatus(ctx context.Context, customerId int, isAnonymous bool) error {
	if err := r.Repository.UpdateCustomerAnonymousStatus(ctx, customerId, isAnonymous); err != nil {
		return err
	}
	return r.invalidate(ctx, customerId)
}

func (r *RepoCache) invalidate(ctx context.Context, customerId int) error {
	if !r.refreshCache {
		return nil
	}
	if _, err := r.storage.Del(ctx, getCustomerCacheKey(customerId)); err != nil {
		return fmt.Errorf("customer updated but cache invalidation failed: %w", err)
	}
	return nil
}

func getCustomerCacheKey(customerId int) string {
	return fmt.Sprintf("cust_%d", customerId)
}
