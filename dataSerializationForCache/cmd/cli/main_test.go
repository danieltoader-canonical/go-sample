package main

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"dataSerializationForCache/pkg/kvstore"
)

// keepOpen ignores Close so data written by one command is still there for the next,
// as it would be in Valkey. Closing a MemStore discards its contents.
type keepOpen struct{ kvstore.Store }

func (keepOpen) Close() error                   { return nil }
func (keepOpen) CloseWithCheck(context.Context) {}

func newMemStore(t *testing.T) kvstore.Store {
	t.Helper()
	store, err := kvstore.NewMemStore(kvstore.Config{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	return keepOpen{store}
}

func connectTo(store kvstore.Store) connectFunc {
	return func(string) (kvstore.Store, error) { return store, nil }
}

func run(t *testing.T, connect connectFunc, args ...string) (string, error) {
	t.Helper()
	cmd := newRootCmd(connect)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.ExecuteContext(context.Background())
	return out.String(), err
}

func TestStoreThenFetch(t *testing.T) {
	for _, format := range []string{"json", "avro", "protobuf", "msgpack"} {
		t.Run(format, func(t *testing.T) {
			store := newMemStore(t)

			out, err := run(t, connectTo(store), "store", "--format", format, "--id", "INV-7", "--ttl", "1m")
			require.NoError(t, err)
			key := "invoice:" + format + ":INV-7"
			assert.Contains(t, out, "stored "+key)
			assert.Contains(t, out, "ttl 1m0s")
			ttl, err := store.TTL(context.Background(), key)
			require.NoError(t, err)
			assert.InDelta(t, time.Minute, ttl, float64(time.Second))

			out, err = run(t, connectTo(store), "fetch", "--format", format, "--id", "INV-7")
			require.NoError(t, err)
			assert.Regexp(t, `fetched `+key+` \(\d+ bytes\) in \S+ \(get \S+, decode \S+\)`, out)
			assert.Contains(t, out, "invoice INV-7 for CUST-42 issued 2026-01-15: 40 lines, total 4154.70 EUR")
		})
	}
}

func TestStoreWithoutTTL(t *testing.T) {
	out, err := run(t, connectTo(newMemStore(t)), "store", "--ttl", "0")
	require.NoError(t, err)
	assert.Contains(t, out, "ttl none")
}

func TestFetchMissingKey(t *testing.T) {
	_, err := run(t, connectTo(newMemStore(t)), "fetch", "--format", "avro", "--id", "INV-404")
	assert.ErrorContains(t, err, "invoice:avro:INV-404 not found")
}

func TestUnknownFormat(t *testing.T) {
	_, err := run(t, connectTo(newMemStore(t)), "store", "--format", "xml")
	assert.ErrorContains(t, err, `unknown format "xml"`)
}

func TestConnectError(t *testing.T) {
	failing := func(string) (kvstore.Store, error) { return nil, errors.New("refused") }
	_, err := run(t, failing, "store")
	assert.ErrorContains(t, err, "connect to valkey: refused")
}

func TestStoreError(t *testing.T) {
	store := kvstore.NewMockStore(gomock.NewController(t))
	store.EXPECT().Set(gomock.Any(), "invoice:json:INV-1001", gomock.Any(), 10*time.Minute).Return(errors.New("readonly"))
	store.EXPECT().CloseWithCheck(gomock.Any())
	_, err := run(t, connectTo(store), "store")
	assert.ErrorContains(t, err, "store invoice:json:INV-1001: readonly")
}

func TestCompare(t *testing.T) {
	store := newMemStore(t)
	out, err := run(t, connectTo(store), "compare", "--id", "INV-9", "--runs", "3")
	require.NoError(t, err)

	assert.Contains(t, out, "invoice INV-9 with 40 lines, averages over 3 runs")
	assert.Regexp(t, `FORMAT\s+BYTES\s+ENCODE\s+SET\s+SET TOTAL\s+GET\s+DECODE\s+GET TOTAL\s+TOTAL\s+ROUND-TRIP`, out)
	for _, format := range []string{"json", "avro", "protobuf", "msgpack"} {
		assert.Regexp(t, format+`\s+\d+(\s+\S+){7}\s+ok`, out)
		_, getErr := store.Get(context.Background(), "invoice:"+format+":INV-9")
		assert.NoError(t, getErr)
	}
	assert.NotContains(t, out, "mismatch")
}

func TestCompareRejectsInvalidRuns(t *testing.T) {
	_, err := run(t, connectTo(newMemStore(t)), "compare", "--runs", "0")
	assert.ErrorContains(t, err, "--runs must be at least 1, got 0")
}

func TestFormatDuration(t *testing.T) {
	assert.Equal(t, "12.3µs", formatDuration(12_345*time.Nanosecond))
	assert.Equal(t, "1.23ms", formatDuration(1_234_567*time.Nanosecond))
	assert.Equal(t, "0s", formatDuration(0))
}

func TestAddrFlagIsPassedToConnector(t *testing.T) {
	var gotAddr string
	connect := func(addr string) (kvstore.Store, error) {
		gotAddr = addr
		return newMemStore(t), nil
	}
	_, err := run(t, connect, "store", "--addr", "valkey:6380")
	require.NoError(t, err)
	assert.Equal(t, "valkey:6380", gotAddr)
}
