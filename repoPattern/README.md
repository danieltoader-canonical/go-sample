# Repository Pattern

A Go sample project demonstrating a PostgreSQL customer repository, a cache layer,
and testable storage interfaces. Use it as a reference when these patterns are
useful in your own projects.

The example includes:

- A customer repository using `jackc/pgx/v5` and a connection pool, with Redis caching.
- A `kvstore.Store` interface with Redis, Valkey, and in-memory implementations.
- Generated GoMock repository and storage mocks, with Testify assertions.
- A Cobra CLI for reading a customer or updating their name.

`internal/cmd/root` owns the Cobra command, flags, and JSON output.
`cmd/app` wires PostgreSQL,
Redis, and the cached repository, including cleanup. Connections are opened only
after command arguments have been validated, not for `--help`.

## Run

With Go (see `go.mod`), Make, and Docker Compose installed, start the services
and run the sample with:

```sh
make start
make run ARGS='--customer-id 1 --name "Alice"'
make down
```

`make start` waits for PostgreSQL and Redis to become healthy before running the
CLI. `make run` uses existing services; `make down` preserves database volumes.
These targets also work from the repository root via its delegating Makefile.

Start PostgreSQL and Redis with Docker Compose:

```sh
docker compose up -d --wait
```

From this directory, run the sample:

```sh
go run ./cmd/app --customer-id 1
go run ./cmd/app --customer-id 1 --name "Alice"
```

Run the tests with:

```sh
go test ./...
```

Alternatively, use `make test`.

The defaults connect to PostgreSQL using
`postgres://postgres@localhost:5432/sample?sslmode=disable` and Redis at
`localhost:6379`. Override them with `POSTGRES_URI` and `REDIS_ADDR`.
PostgreSQL initialization creates the `customer` table and seeds customer `1`
when the database volume is first created. The Compose database uses trust
authentication and a loopback-only port for local development; use credentials
and TLS outside this local example.

Reads return a zero customer for missing rows and log database failures.
Both updates use parameterized SQL with `ON CONFLICT` upserts: changing a name
preserves the anonymous status, and changing the status preserves the name.
The pgx client explicitly uses `QueryExecModeCacheStatement`: each connection
prepares queries on first use and reuses the server-side statements for subsequent
calls. This is pgx's default mode, now enforced even if the URI specifies another
execution mode. `statement_cache_capacity` must remain positive (default: 512).
New rows default to an empty name and a false anonymous status. The public
repository interface and JSON/cache encoding are unchanged.

Run the PostgreSQL integration test against the initialized database with:

```sh
POSTGRES_TEST_URI='postgres://postgres@localhost:5432/sample?sslmode=disable' make test
```

It uses temporary customer IDs and deletes its own rows afterward. Without
`POSTGRES_TEST_URI`, only this integration test is skipped; unit tests use GoMock.
Existing MongoDB data is not migrated. Old MongoDB containers and volumes are
not removed automatically; the new setup uses a separate PostgreSQL volume.
Existing Redis entries also remain until their TTL expires; use a fresh Redis
instance or explicitly invalidate customer keys when switching an existing dataset.

## Benchmarks

Run the benchmarks without PostgreSQL, Redis, or Docker:

```sh
make bench
make bench BENCH=BenchmarkRepoCacheHit BENCHTIME=2s
```

Benchmarks report time and allocations for in-memory store hits, misses,
parallel reads, writes with and without TTL, and serial/parallel customer cache
hits including JSON decoding. Setup and cleanup are excluded from timing.
The customer benchmarks use a real in-memory store and reject any repository
access; they do not measure database or network latency. Write benchmarks
overwrite a fixed key rather than measuring cache growth or eviction.

## Cache concurrency

By default, updates write only to the repository and leave cached entries
unchanged. Construct the decorator with `WithRefreshCache()` to invalidate entries
after successful updates. This option also bypasses cached reads and repopulates
the cache from the repository when a customer is found.

The cache decorator delegates explicitly to a named repository field and does not
serialize reads or writes. Repository and storage implementations must support
concurrent use. Concurrent cache misses may issue duplicate database reads, and
an in-flight read can cache stale data after an update invalidates the entry.
Stale entries can remain until the 24-hour TTL expires, an update through a
decorator configured with `WithRefreshCache()` invalidates them, or a forced
refresh replaces them; this sample does not provide strict cache consistency.

## Generate the mocks

pgx imports stay in `pkg/postgres`. Other packages use `postgres.Client`,
`postgres.Row`, and `postgres.ErrNoRows`. The client owns a concurrency-safe
connection pool; repository operations apply read/write deadlines.

After changing the `Repository`, `kvstore.Store`, or `postgres.Client` interface,
regenerate its mock from this directory:

```sh
go generate ./...
```

Alternatively, use `make generate`.

The directives in `internal/dbstore/customer`, `pkg/kvstore/interface.go`, and
`pkg/postgres/client.go` run MockGen pinned to `v0.6.0` via `go run`, so no separate
installation is needed. The customer directive reads `Repository` from
`customer.go`; its mock can stand in for either `Repo` or `RepoCache`.
`repo_mock.go`, `store_mock.go`, and
`client_mock.go` stay in their respective packages so tests in other packages
can use them. Tests create them with `NewMockRepository`, `NewMockStore`, or
`NewMockClient`, passing `gomock.NewController(t)`, and declare
calls with `EXPECT()`. The controller automatically checks expectations
when each test finishes.
