# Data Serialization for Cache

Stores the same invoice in Valkey as **JSON**, **Avro** (zstd-compressed),
**Protobuf** and **MessagePack**, so you can compare formats for cached values.
The CLI is a small [Cobra](https://github.com/spf13/cobra) app in `cmd/cli`.

## Quick start

```sh
make up                                   # start Valkey on 127.0.0.1:6380
make run ARGS='compare'                   # store every format, print size and retrieval speed
make run ARGS='store --format avro --id INV-1001 --ttl 10m'
make run ARGS='fetch --format avro --id INV-1001'
make down
```

`make start` runs `make up` first, then `compare`. Set `VALKEY_PORT` to
publish a different port; the CLI uses that port by default. `VALKEY_ADDR` or
`--addr` can override the address.
The default port is 6380; `repoPattern` uses 6379. Each Compose project has an
explicit distinct name, and this sample prefixes its Valkey keys with
`dataSerializationForCache:` so the projects remain isolated even if configured
to share a Valkey instance.

Example output:

```text
$ cli compare --runs 2000
invoice INV-1001 with 40 lines, averages over 2000 runs
    FORMAT  BYTES  ENCODE      SET  SET TOTAL      GET  DECODE  GET TOTAL    TOTAL  ROUND-TRIP
      json   5121  14.4µs  390.9µs    405.3µs  314.2µs  24.5µs    338.7µs  744.1µs          ok
      avro    544  14.8µs    283µs    297.8µs  255.4µs  13.1µs    268.6µs  566.4µs          ok
  protobuf   2887   4.6µs  385.3µs    389.8µs  370.8µs  14.2µs    384.9µs  774.8µs          ok
   msgpack   4892  10.5µs    355µs    365.5µs  346.3µs  19.5µs    365.8µs  731.3µs          ok

$ cli fetch --format avro --id INV-1001
fetched invoice:avro:INV-1001 (544 bytes) in 477.5µs (get 320.4µs, decode 157.1µs)
invoice INV-1001 for CUST-42 issued 2026-01-15: 40 lines, total 4154.70 EUR
```

The sample invoice has 40 lines cycling through 8 products, so it is about 5 KB as JSON.
Its repeated descriptions are why zstd makes the Avro payload so small.

`compare` columns, each the average of `--runs` (default 500) iterations:

| Column | Meaning |
| --- | --- |
| `BYTES` | Size of the value stored in Valkey. |
| `ENCODE` | Time to serialize the invoice (plus zstd compression for Avro). |
| `SET` | Time for the Valkey `SET`, including the network round trip. |
| `SET TOTAL` | `ENCODE` + `SET`: how long it takes to write a cached invoice. |
| `GET` | Time for the Valkey `GET`, including the network round trip. Bigger payloads take longer. |
| `DECODE` | Time to deserialize the value (plus zstd decompression for Avro). |
| `GET TOTAL` | `GET` + `DECODE`: how long it takes to read a cached invoice. |
| `TOTAL` | `SET TOTAL` + `GET TOTAL`: one full write and read. |

`fetch` reports one measurement, so its numbers are noisier than `compare`'s.
Timings depend on the machine and on Valkey running locally.

## Commands

| Command | What it does |
| --- | --- |
| `store --format <f> --id <id> --ttl <d>` | Serializes the sample invoice and runs `SET invoice:<f>:<id>` with a TTL (`0` = no expiry). |
| `fetch --format <f> --id <id>` | Runs `GET`, deserializes with the same format, and prints a summary with the `GET` and decode times. |
| `compare --id <id> --runs <n>` | Stores the invoice in every format, reads each back `n` times, and prints the size, average encode/`SET`/`GET`/decode times with `SET TOTAL`, `GET TOTAL` and `TOTAL` sums, and whether it round-trips unchanged. |

All `cache` commands accept `--addr` (default `127.0.0.1:6380`); the root
command accepts `--verbose`. Valid formats are `json`, `avro`, `protobuf` and
`msgpack`.

## Formats

| Format | Library | Schema | Notes |
| --- | --- | --- | --- |
| JSON | `encoding/json` | None | Human readable; field names are repeated in every value, so it is the largest and slowest to decode. |
| MessagePack | `github.com/vmihailenco/msgpack/v5` | None | A binary JSON: same struct tags, smaller and faster, but field names are still included. |
| Protobuf | `google.golang.org/protobuf` | `invoice.proto` + generated code | Fields are numeric tags, so payloads are compact. The domain type is mapped to the generated message. |
| Avro + zstd | `github.com/confluentinc/confluent-avro-go/v2` + `github.com/klauspost/compress/zstd` | JSON schema in `avro.go` | No field names or tags, so the reader needs the writer's schema. The payload is zstd-compressed before caching. |

### Serialization formats for caching

| Feature | JSON | Avro | Protobuf | MessagePack |
| --- | --- | --- | --- | --- |
| Storage efficiency | Low | Very high | Very high | High |
| Serialization speed | Moderate | High | High | High |
| Deserialization speed | Moderate | High | High | High |
| Schema evolution | Manual | Excellent | Excellent | Manual |
| Schema required | No | Yes | Yes | No |
| Human-readable | Yes | No | No | No |
| Implementation complexity | Low | Medium–high | Medium | Low |
| External schema management | No | Optional | No | No |

Avro can resolve old data against a new schema, but only if the reader has the
schema the data was written with. Many teams use a Schema Registry for this,
which is why it is "Optional" above; this sample embeds the schema in code.

Benchmarks for this invoice (`make bench`) on one machine:

| Format | Bytes | Marshal | Unmarshal |
| --- | ---: | ---: | ---: |
| JSON | 5121 | ~8.9 µs | ~15.4 µs |
| MessagePack | 4892 | ~7.1 µs | ~9.3 µs |
| Protobuf | 2887 | ~4.0 µs | ~5.0 µs |
| Avro + zstd | 544 | ~9.5 µs | ~5.2 µs |

The benchmarks run in memory, without Valkey. Here the `GET` round trip costs more than
decoding, so smaller payloads are read faster. zstd makes Avro slower to write, but shrinks it
from 2652 bytes uncompressed (close to Protobuf) to 544, about 9x smaller than JSON.

### Choosing a format for a cache

- **JSON**: easy to debug with `valkey-cli GET`, and works for any client language.
- **MessagePack**: a drop-in way to shrink JSON values without maintaining a schema.
- **Protobuf**: compact and fast, and fields can be added safely by giving them new numbers.
- **Avro**: the most compact, and suited to Kafka and Schema Registry setups. Every reader needs the
  writer schema, so put a schema version in the key or value when the schema changes.

When a schema changes, old cache entries may not decode. Put a version in the
key (for example `invoice:v2:avro:<id>`) or use short TTLs.

## Structure

- `cmd/cli/main.go`: `main`, the root command with the `--addr` flag, and the Valkey connection.
- `cmd/cli/commands.go`: the `store`, `fetch` and `compare` commands. The connect function is passed
  in, so tests in `main_test.go` use `kvstore.MemStore` or the generated `kvstore.MockStore` instead of Valkey.
- `pkg/kvstore/`: the key/value store from [repoPattern](../repoPattern/pkg/kvstore), copied so the
  module stays standalone. The CLI uses `ValKeyStore`. `Set`/`Get` work with strings, so payloads are
  converted between `[]byte` and `string`; the bytes are not changed.
- `internal/serializer/`: the `Codec` interface, one file per format, round-trip tests and benchmarks.
- `internal/serializer/invoicepb/`: `invoice.proto` and its generated code. Run `make generate` to
  regenerate it with `buf`; no local `protoc` is needed.
- `internal/invoice/`: the domain model and the sample invoice.
- `compose.yaml`: the Valkey service.

## Make targets

| Target | Purpose |
| --- | --- |
| `make up` / `make down` | Start or stop Valkey. |
| `make start` | Start Valkey, then run `ARGS` (default `compare`). |
| `make run ARGS='...'` | Run the CLI against an already running Valkey. |
| `make test` | Unit tests; Valkey is not needed. |
| `make bench` | Serializer benchmarks (`BENCH`, `BENCHTIME`). |
| `make generate` | Regenerate the Protobuf code. |
