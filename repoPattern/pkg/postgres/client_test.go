package postgres

import (
	"context"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewClientInvalidConfiguration(t *testing.T) {
	for _, uri := range []string{"", "://invalid", "postgres://localhost/sample?statement_cache_capacity=0"} {
		t.Run(uri, func(t *testing.T) {
			client, err := NewClient(context.Background(), uri)
			require.Error(t, err)
			require.Nil(t, client)
		})
	}

}

func TestPreparedStatementReuse(t *testing.T) {
	uri := os.Getenv("POSTGRES_TEST_URI")
	if uri == "" {
		t.Skip("set POSTGRES_TEST_URI to verify prepared statements against PostgreSQL")
	}
	config, err := url.Parse(uri)
	require.NoError(t, err)
	params := config.Query()
	params.Set("pool_max_conns", "1")
	params.Set("default_query_exec_mode", "simple_protocol")
	config.RawQuery = params.Encode()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := NewClient(ctx, config.String())
	require.NoError(t, err)
	defer client.Close()

	const query = "SELECT $1::bigint"
	const inspect = `SELECT name, generic_plans + custom_plans
			FROM pg_prepared_statements WHERE statement = $1`
	var firstName string
	for i := 1; i <= 2; i++ {
		var got int
		require.NoError(t, client.QueryRow(ctx, query, i).Scan(&got))
		require.Equal(t, i, got)
		var name string
		var executions int64
		require.NoError(t, client.QueryRow(ctx, inspect, query).Scan(&name, &executions))
		require.NotEmpty(t, name)
		require.Equal(t, int64(i), executions)
		if i == 1 {
			firstName = name
		} else {
			require.Equal(t, firstName, name)
		}
	}
}

func TestNewClientCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client, err := NewClient(ctx, "postgres://postgres@localhost:5432/sample?sslmode=disable")
	require.ErrorIs(t, err, context.Canceled)
	require.Nil(t, client)
}
