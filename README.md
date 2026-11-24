# Go Samples

Small, standalone Go examples of useful application patterns. Choose a sample
from the contents table to read its detailed guide.

## Contents

| Pattern | What it demonstrates | Guide |
| --- | --- | --- |
| Repository Pattern | PostgreSQL data access, cache decorators, and storage interfaces | [repoPattern/README.md](repoPattern/README.md) |
| Command Pattern | A Cobra CLI with command groups, version reporting, and explicit wiring | [cmdPattern/README.md](cmdPattern/README.md) |
| Mutation Testing | Test invoice rules against small changes to their logic | [mutationTesting/README.md](mutationTesting/README.md) |
| Data Serialization for Cache | Cache invoices in Valkey as JSON, Avro + zstd, Protobuf, or MessagePack, and compare them | [dataSerializationForCache/README.md](dataSerializationForCache/README.md) |

## Running a sample

You need Go and Make. The Repository Pattern also requires Docker Compose to
start PostgreSQL and Redis, and Data Serialization for Cache requires it to start
Valkey. The Command Pattern has no external service dependencies.

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
`bench`, `generate`, and `help`. The Repository Pattern and Data Serialization
for Cache samples also provide `up` and `down` targets. You can also run Make from either sample directory without
setting `PROJECT`.

The Mutation Testing example provides `test` and `mutate` targets:

```sh
make PROJECT=mutationTesting test
make PROJECT=mutationTesting mutate
```

Use `test-weak` and `mutate-weak` to compare an intentionally incomplete test
suite; the latter reports surviving mutations.

The Data Serialization for Cache sample starts its own Valkey on port 6380 with
Docker Compose:

```sh
make PROJECT=dataSerializationForCache start
make PROJECT=dataSerializationForCache run ARGS='fetch --format avro'
make PROJECT=dataSerializationForCache bench
make PROJECT=dataSerializationForCache down
```
