# Go Samples

Small, standalone Go examples of useful application patterns. Choose a sample
from the contents table to read its detailed guide.

## Contents

| Pattern | What it demonstrates | Guide |
| --- | --- | --- |
| Repository Pattern | PostgreSQL data access, cache decorators, and storage interfaces | [repoPattern/README.md](repoPattern/README.md) |
| Command Pattern | A Cobra CLI with command groups, version reporting, and explicit wiring | [cmdPattern/README.md](cmdPattern/README.md) |
| Mutation Testing | Test invoice rules against small changes to their logic | [mutationTesting/README.md](mutationTesting/README.md) |

## Running a sample

You need Go and Make. The Repository Pattern also requires Docker Compose to
start PostgreSQL and Redis; the Command Pattern has no external service
dependencies.

From the repository root, Make uses the Repository Pattern by default. For
example:

```sh
make start
make run ARGS='--customer-id 1'
make test
make down
```

Select the Command Pattern with `PROJECT=cmdPattern`:

```sh
make PROJECT=cmdPattern start
make PROJECT=cmdPattern run ARGS='command get-version'
make PROJECT=cmdPattern test
```

The Repository and Command Pattern samples support `start`, `run`, `test`,
`bench`, `generate`, and `help`. Only the Repository Pattern provides `up` and
`down` targets. You can also run Make from either sample directory without
setting `PROJECT`.

The Mutation Testing example provides `test` and `mutate` targets:

```sh
make PROJECT=mutationTesting test
make PROJECT=mutationTesting mutate
```

Use `test-weak` and `mutate-weak` to compare an intentionally incomplete test
suite; the latter reports surviving mutations.
