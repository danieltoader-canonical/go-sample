.DEFAULT_GOAL := help

GO ?= go
ARGS ?= compare
BENCH ?= .
BENCHTIME ?= 1s

.PHONY: help start run up down test bench generate

help:
	@printf '%s\n' \
		'make start     Start Valkey, then run the sample (ARGS="compare")' \
		'make run       Run against an existing Valkey (ARGS="store --format avro")' \
		'make up        Start Valkey and wait until healthy' \
		'make down      Stop Valkey' \
		'make test      Run Go tests (no Valkey needed)' \
		'make bench     Run serializer benchmarks (BENCH=. BENCHTIME=1s)' \
		'make generate  Regenerate Protobuf code from invoice.proto'

start: up
	$(GO) run ./cmd/cli $(ARGS)

run:
	$(GO) run ./cmd/cli $(ARGS)

up:
	docker compose up -d --wait

down:
	docker compose down

test:
	$(GO) test ./...

bench:
	$(GO) test -run='^$$' -bench='$(BENCH)' -benchmem -benchtime='$(BENCHTIME)' ./...

generate:
	$(GO) generate ./...
