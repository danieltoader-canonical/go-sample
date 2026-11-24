// Command cli stores a sample invoice in Valkey as JSON, Avro, Protobuf or MessagePack
// and compares payload size and speed.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"dataSerializationForCache/pkg/kvstore"
)

// defaultAddr is the Valkey address published by compose.yaml (port 6380 avoids clashing with repoPattern).
const defaultAddr = "127.0.0.1:6380"

// connectFunc opens a kvstore.Store. Tests pass one that returns kvstore.MemStore or a mock.
type connectFunc func(addr string) (kvstore.Store, error)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := newRootCmd(connectValkey).ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func connectValkey(addr string) (kvstore.Store, error) {
	return kvstore.NewValKeyStore(
		kvstore.Config{Addrs: []string{addr}},
		kvstore.WithTimeouts(time.Second, time.Second, time.Second),
	)
}

// newRootCmd builds `cli` with the store, fetch and compare subcommands.
func newRootCmd(connect connectFunc) *cobra.Command {
	var addr string
	root := &cobra.Command{
		Use:           "cli",
		Short:         "Serialize invoices with JSON, Avro, Protobuf or MessagePack and cache them in Valkey",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.CompletionOptions.DisableDefaultCmd = true
	root.PersistentFlags().StringVar(&addr, "addr", defaultAddr, "Valkey address")

	open := func() (kvstore.Store, error) {
		store, err := connect(addr)
		if err != nil {
			return nil, fmt.Errorf("connect to valkey: %w", err)
		}
		return store, nil
	}
	root.AddCommand(storeCmd(open), fetchCmd(open), compareCmd(open))
	return root
}
