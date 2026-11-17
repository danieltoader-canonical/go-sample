package customer

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"repoPattern/pkg/postgres"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRepo(t *testing.T) {
	want := Customer{CustomerId: 1, IsAnonymous: true, CustomerName: "Alice"}
	failure := errors.New("database unavailable")
	for _, operation := range []string{"read", "update name", "update anonymous status"} {
		for _, result := range []struct {
			name string
			err  error
		}{
			{name: "success"},
			{name: "not found", err: postgres.ErrNoRows},
			{name: "failure", err: failure},
		} {
			t.Run(operation+"/"+result.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				client := postgres.NewMockClient(ctrl)
				row := postgres.NewMockRow(ctrl)
				client.EXPECT().GetOptions().Return(&postgres.Options{
					ReadTimeoutMedium: time.Second, WriteTimeoutMedium: time.Second,
				})
				repo := NewRepo(client)
				var sql string
				var args []any
				client.EXPECT().QueryRow(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, query string, params ...any) postgres.Row {
						deadline, ok := ctx.Deadline()
						require.True(t, ok)
						require.WithinDuration(t, time.Now(), deadline, time.Second)
						sql, args = query, params
						return row
					})
				row.EXPECT().Scan(gomock.Any()).DoAndReturn(func(dest ...any) error {
					if result.err != nil {
						return result.err
					}
					*dest[0].(*int) = want.CustomerId
					if len(dest) == 3 {
						*dest[1].(*bool) = want.IsAnonymous
						*dest[2].(*string) = want.CustomerName
					}
					return nil
				})
				ctx := context.Background()
				switch operation {
				case "read":
					got := repo.GetCustomer(ctx, 1)
					if result.err != nil {
						require.Equal(t, Customer{}, got)
					} else {
						require.Equal(t, want, got)
					}
					require.Equal(t, []any{1}, args)
					require.Contains(t, sql, "WHERE customer_id = $1")
				case "update name":
					got, err := repo.UpdateCustomerName(ctx, 1, "Alice")
					require.ErrorIs(t, err, result.err)
					if result.err != nil {
						require.Nil(t, got)
					} else {
						require.Equal(t, &want, got)
					}
					require.Equal(t, []any{1, "Alice"}, args)
					require.Contains(t, sql, "DO UPDATE SET customer_name = EXCLUDED.customer_name")
					require.NotContains(t, strings.Split(sql, "DO UPDATE")[1], "is_anonymous =")
				case "update anonymous status":
					require.ErrorIs(t, repo.UpdateCustomerAnonymousStatus(ctx, 1, true), result.err)
					require.Equal(t, []any{1, true}, args)
					require.Contains(t, sql, "DO UPDATE SET is_anonymous = EXCLUDED.is_anonymous")
					require.NotContains(t, sql, "customer_name")
				}
			})
		}
	}
}

func TestRepoMock(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(gomock.NewController(t))
	want := Customer{CustomerId: 7, CustomerName: "New name"}
	repo.EXPECT().GetCustomer(ctx, 7).Return(want)
	repo.EXPECT().UpdateCustomerName(ctx, 7, "New name").Return(&want, nil)
	repo.EXPECT().UpdateCustomerName(ctx, 8, "Missing").Return(nil, postgres.ErrNoRows)
	repo.EXPECT().UpdateCustomerAnonymousStatus(ctx, 7, true).Return(errors.New("write failed"))

	require.Equal(t, want, repo.GetCustomer(ctx, 7))
	got, err := repo.UpdateCustomerName(ctx, 7, "New name")
	require.NoError(t, err)
	require.Equal(t, &want, got)
	got, err = repo.UpdateCustomerName(ctx, 8, "Missing")
	require.ErrorIs(t, err, postgres.ErrNoRows)
	require.Nil(t, got)
	require.EqualError(t, repo.UpdateCustomerAnonymousStatus(ctx, 7, true), "write failed")
}
