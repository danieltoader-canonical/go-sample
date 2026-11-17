# go-sample

A collection of Go sample patterns that might be useful.


- [Repository Pattern](repoPattern/README.md)

From the repository root, use `make start` to start PostgreSQL and Redis and run
the Repository Pattern sample. Use `make run ARGS='--customer-id 1'` against
existing services, and `make down` to stop them without deleting database data.
Docker with Compose, Go (see `repoPattern/go.mod`), and Make are required.

The root Makefile delegates to `repoPattern/Makefile`. Both support `make test`,
`make bench`, `make generate`, and `make help`.
