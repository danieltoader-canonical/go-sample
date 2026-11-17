package main

import (
	"cmdPattern/pkg/sys"
	"context"
	"os"
	"os/signal"
	"syscall"

	"cmdPattern/internal/cmd/cli"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	commands := cli.Setup(ctx)
	sys.CheckErrorWithExit(ctx, commands.ExecuteContext(ctx))
}
