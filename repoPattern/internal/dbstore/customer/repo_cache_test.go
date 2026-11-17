package customer

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"repoPattern/pkg/kvstore"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var _ kvstore.Store = (*kvstore.MockStore)(nil)

func TestRepoCacheDoesNotSerializeOperations(t *testing.T) {
	for _, operation := range []string{"read", "update name", "update anonymous status"} {
		t.Run(operation, func(t *testing.T) {
			ctx := context.Background()
			repo := NewMockRepository(gomock.NewController(t))
			store := kvstore.NewMockStore(gomock.NewController(t))
			cache := NewRepoWithCache(repo, store)
			started := make(chan struct{})
			release := make(chan struct{})
			firstDone := make(chan struct{})

			store.EXPECT().Get(ctx, "cust_1").Return("", kvstore.ErrNotFound).Times(1)
			repo.EXPECT().GetCustomer(ctx, 1).DoAndReturn(func(context.Context, int) Customer {
				close(started)
				<-release
				return Customer{}
			})
			go func() {
				defer close(firstDone)
				cache.GetCustomer(ctx, 1)
			}()
			defer func() {
				close(release)
				<-firstDone
			}()
			select {
			case <-started:
			case <-time.After(5 * time.Second):
				t.Fatal("first read did not reach the repository")
			}

			want := Customer{CustomerId: 1, CustomerName: "Alice"}
			var run func() error
			switch operation {
			case "read":
				data, err := json.Marshal(want)
				require.NoError(t, err)
				store.EXPECT().Get(ctx, "cust_1").Return(string(data), nil).Times(1)
				run = func() error {
					if got := cache.GetCustomer(ctx, 1); got != want {
						return errors.New("unexpected cached customer")
					}
					return nil
				}
			case "update name":
				repo.EXPECT().UpdateCustomerName(ctx, 1, "Alice").Return(&want, nil)
				run = func() error {
					_, err := cache.UpdateCustomerName(ctx, 1, "Alice")
					return err
				}
			case "update anonymous status":
				repo.EXPECT().UpdateCustomerAnonymousStatus(ctx, 1, true).Return(nil)
				run = func() error { return cache.UpdateCustomerAnonymousStatus(ctx, 1, true) }
			}
			done := make(chan error, 1)
			go func() { done <- run() }()
			select {
			case err := <-done:
				require.NoError(t, err)
			case <-time.After(5 * time.Second):
				t.Fatal("operation blocked behind the in-flight read")
			}
		})
	}
}

func TestRepoCacheWithMemoryStore(t *testing.T) {
	ctx := context.Background()
	store, err := kvstore.NewMemStore(kvstore.Config{}, kvstore.WithCacheKeyPrefix("test:"))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, store.Close()) })
	repo := NewMockRepository(gomock.NewController(t))
	cachedRepo := NewRepoWithCache(repo, store)
	refreshingRepo := NewRepoWithCache(repo, store, WithRefreshCache())

	original := Customer{CustomerId: 1, CustomerName: "Alice"}
	renamed := Customer{CustomerId: 1, CustomerName: "Bob"}
	anonymous := Customer{CustomerId: 1, CustomerName: "Bob", IsAnonymous: true}

	repo.EXPECT().GetCustomer(ctx, 1).Return(original).Times(1)
	require.Equal(t, original, cachedRepo.GetCustomer(ctx, 1))
	require.Equal(t, original, cachedRepo.GetCustomer(ctx, 1))
	raw, err := store.Get(ctx, "cust_1")
	require.NoError(t, err)
	var document map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(raw), &document))
	require.Len(t, document, 3)
	require.Equal(t, "Alice", document["customerName"])
	require.Equal(t, false, document["IsAnonymous"])
	require.Equal(t, float64(1), document["customerId"])
	ttl, err := store.TTL(ctx, "cust_1")
	require.NoError(t, err)
	require.Positive(t, ttl)
	require.LessOrEqual(t, ttl, customerCacheTTL)

	repo.EXPECT().UpdateCustomerName(ctx, 1, "Bob").Return(&renamed, nil).Times(1)
	got, err := refreshingRepo.UpdateCustomerName(ctx, 1, "Bob")
	require.NoError(t, err)
	require.Equal(t, &renamed, got)
	_, err = store.Get(ctx, "cust_1")
	require.ErrorIs(t, err, kvstore.ErrNotFound)
	repo.EXPECT().GetCustomer(ctx, 1).Return(renamed).Times(1)
	require.Equal(t, renamed, cachedRepo.GetCustomer(ctx, 1))

	repo.EXPECT().UpdateCustomerAnonymousStatus(ctx, 1, true).Return(nil).Times(1)
	require.NoError(t, refreshingRepo.UpdateCustomerAnonymousStatus(ctx, 1, true))
	repo.EXPECT().GetCustomer(ctx, 1).Return(anonymous).Times(1)
	require.Equal(t, anonymous, cachedRepo.GetCustomer(ctx, 1))
}

func TestRepoCacheReadFallbacks(t *testing.T) {
	ctx := context.Background()
	want := Customer{CustomerId: 1, CustomerName: "Alice"}
	for _, tt := range []struct {
		name  string
		value string
		err   error
	}{
		{name: "missing", err: kvstore.ErrNotFound},
		{name: "empty"},
		{name: "unavailable", err: errors.New("cache unavailable")},
		{name: "invalid JSON", value: "invalid"},
		{name: "wrong customer", value: `{"customerId":2,"customerName":"Bob"}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockRepository(gomock.NewController(t))
			store := kvstore.NewMockStore(gomock.NewController(t))
			store.EXPECT().Get(ctx, "cust_1").Return(tt.value, tt.err).Times(1)
			repo.EXPECT().GetCustomer(ctx, 1).Return(want).Times(1)
			data, err := json.Marshal(want)
			require.NoError(t, err)
			store.EXPECT().Set(ctx, "cust_1", string(data), customerCacheTTL).Return(nil).Times(1)
			require.Equal(t, want, NewRepoWithCache(repo, store).GetCustomer(ctx, 1))
		})
	}
}

func TestRepoCacheRefreshAndMissing(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(gomock.NewController(t))
	store := kvstore.NewMockStore(gomock.NewController(t))
	repo.EXPECT().GetCustomer(ctx, 1).Return(Customer{}).Times(2)
	cache := NewRepoWithCache(repo, store, WithRefreshCache())
	require.Equal(t, Customer{}, cache.GetCustomer(ctx, 0))
	require.Equal(t, Customer{}, cache.GetCustomer(ctx, 1))
	require.Equal(t, Customer{}, cache.GetCustomer(ctx, 1))
}

func TestRepoCacheWriteFailures(t *testing.T) {
	ctx := context.Background()
	failure := errors.New("write failed")
	t.Run("repository failures leave cache alone", func(t *testing.T) {
		repo := NewMockRepository(gomock.NewController(t))
		store := kvstore.NewMockStore(gomock.NewController(t))
		repo.EXPECT().UpdateCustomerName(ctx, 1, "Bob").Return(nil, failure).Times(1)
		repo.EXPECT().UpdateCustomerAnonymousStatus(ctx, 1, true).Return(failure).Times(1)
		cache := NewRepoWithCache(repo, store, WithRefreshCache())
		got, err := cache.UpdateCustomerName(ctx, 1, "Bob")
		require.Nil(t, got)
		require.ErrorIs(t, err, failure)
		require.ErrorIs(t, cache.UpdateCustomerAnonymousStatus(ctx, 1, true), failure)
	})
	t.Run("invalidation failures propagate", func(t *testing.T) {
		repo := NewMockRepository(gomock.NewController(t))
		store := kvstore.NewMockStore(gomock.NewController(t))
		want := Customer{CustomerId: 1, CustomerName: "Bob"}
		repo.EXPECT().UpdateCustomerName(ctx, 1, "Bob").Return(&want, nil).Times(1)
		repo.EXPECT().UpdateCustomerAnonymousStatus(ctx, 1, true).Return(nil).Times(1)
		store.EXPECT().Del(ctx, "cust_1").Return(int64(0), failure).Times(2)
		cache := NewRepoWithCache(repo, store, WithRefreshCache())
		got, err := cache.UpdateCustomerName(ctx, 1, "Bob")
		require.Equal(t, &want, got)
		require.ErrorIs(t, err, failure)
		require.ErrorIs(t, cache.UpdateCustomerAnonymousStatus(ctx, 1, true), failure)
	})
	t.Run("cache fill failure still returns database record", func(t *testing.T) {
		repo := NewMockRepository(gomock.NewController(t))
		store := kvstore.NewMockStore(gomock.NewController(t))
		want := Customer{CustomerId: 1, CustomerName: "Bob"}
		data, err := json.Marshal(want)
		require.NoError(t, err)
		repo.EXPECT().GetCustomer(ctx, 1).Return(want).Times(1)
		store.EXPECT().Set(ctx, "cust_1", string(data), customerCacheTTL).Return(failure).Times(1)
		require.Equal(t, want, NewRepoWithCache(repo, store, WithRefreshCache()).GetCustomer(ctx, 1))
	})
}

func TestRepoCacheUpdatesWithoutRefreshLeaveCacheUntouched(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(gomock.NewController(t))
	store := kvstore.NewMockStore(gomock.NewController(t))
	cache := NewRepoWithCache(repo, store)
	original := Customer{CustomerId: 1, CustomerName: "Alice"}
	updated := Customer{CustomerId: 1, CustomerName: "Bob"}
	data, err := json.Marshal(original)
	require.NoError(t, err)

	repo.EXPECT().UpdateCustomerName(ctx, 1, "Bob").Return(&updated, nil)
	repo.EXPECT().UpdateCustomerAnonymousStatus(ctx, 1, true).Return(nil)
	store.EXPECT().Get(ctx, "cust_1").Return(string(data), nil).Times(2)

	got, err := cache.UpdateCustomerName(ctx, 1, "Bob")
	require.NoError(t, err)
	require.Equal(t, &updated, got)
	require.Equal(t, original, cache.GetCustomer(ctx, 1))
	require.NoError(t, cache.UpdateCustomerAnonymousStatus(ctx, 1, true))
	require.Equal(t, original, cache.GetCustomer(ctx, 1))
}
