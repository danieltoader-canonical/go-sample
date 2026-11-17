# Command Pattern

A standalone Cobra CLI demonstrating command groups, version reporting, and a
minimal sample consumer without a dependency container or private application
packages. Requires Go (see `go.mod`); no external services are needed.

## Run

From this directory:

```sh
make start
make run ARGS='command get-version'
make run ARGS='--help'
```

`make start` and `make run` default to `consumer sample`; neither starts containers.
Use `make help` to list targets. Like `repoPattern`, the Makefile supports `GO`,
`ARGS`, `BENCH`, and `BENCHTIME` overrides, plus `test`, `bench`, and `generate`.
There are currently no benchmarks or generation directives in this sample.
There are no `up` or `down` targets because no services are required.

From the repository root, select this sample with
`make PROJECT=cmdPattern run ARGS='consumer sample'`.

Equivalent Go commands:

```sh
go run ./cmd/cli --help
go run ./cmd/cli --version
go run ./cmd/cli command get-version
go run ./cmd/cli command get-version --short
go run ./cmd/cli command patch list
go run ./cmd/cli command patch info
go run ./cmd/cli command patch promote
go run ./cmd/cli command patch remove
go run ./cmd/cli consumer sample
```

The `command patch` group demonstrates registering related Cobra subcommands.
These sample commands take no arguments and print a message: `list` prints
`Listing patches`, `info` prints `Patch information`, `promote` prints
`Patch promoted`, and `remove` prints `Patch removed`. They do not modify patch
data.

Version information comes from `pkg/version`, with defaults of application
`default` and version `1.0.0`. The full version command also prints the commit,
build time, and Go version. Set metadata through linker flags, for example:

```sh
go build -o cli -ldflags "-X cmdPattern/pkg/version.app=cli -X cmdPattern/pkg/version.version=1.2.3 -X cmdPattern/pkg/version.hash=abcdef123456 -X cmdPattern/pkg/version.buildTime=2026-11-17T20:00:00Z" ./cmd/cli
```

Short output includes the version, seven-character commit hash, and Unix build
timestamp. An empty or invalid build time logs a warning and uses timestamp zero.
Add `--verbose` to log command start and successful completion to stderr.

Build a binary with:

```sh
go build -o cli ./cmd/cli
./cli --help
```

## Sample consumer

`consumer sample` is a minimal Cobra command that prints `Sample consumer` and
exits. It needs no configuration, broker, or other dependencies.

## Structure

- `cmd/cli/main.go`: entrypoint, signal handling, and error reporting.
- `internal/cmd/cli/cli.go`: root command and command-group registration.
- `internal/cmd/cli/command`: general commands, including `get-version` and the
  `patch` command group (`list`, `info`, `promote`, and `remove`).
- `internal/patch`: sample patch operations used by the patch CLI commands.
- `internal/cmd/cli/consumer`: minimal sample consumer command.
- `pkg/sys`: logs command errors and exits with status 1.
- `pkg/version`: linker-set build metadata and version formatting.

Add a subcommand to its group's `Register` method. Pass dependencies explicitly
when needed rather than introducing a container.

Signal handling stays in `main` through `signal.NotifyContext`: SIGINT and SIGTERM
cancel the context passed to Cobra. Long-running commands should observe
`cmd.Context().Done()` and release their own resources. No separate signal loop,
shutdown callback, or panic recovery helper is needed for this sample.

## Tests

Run `make test` from this directory, or `make PROJECT=cmdPattern test` from the
repository root. For additional checks:

```sh
go test -race ./...
go build ./...
go vet ./...
```

Tests cover CLI registration and validation, version and sample consumer output,
and writer errors.
