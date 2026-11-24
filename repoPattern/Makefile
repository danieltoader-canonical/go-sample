.DEFAULT_GOAL := help

GO ?= go
ARGS ?= --customer-id 1
BENCH ?= .
BENCHTIME ?= 1s

.PHONY: help start run up down test bench generate

help:
	@printf '%s\n' \
		'make start     Start PostgreSQL and Valkey, then run the sample' \
		'make run       Run against existing services (ARGS="--customer-id 1")' \
		'make up        Start PostgreSQL and Valkey and wait until healthy' \
		'make down      Stop services without deleting database volumes' \
		'make test      Run Go tests' \
		'make bench     Run benchmarks with allocation statistics (BENCH=. BENCHTIME=1s)' \
		'make generate  Regenerate mocks'

start: up
	$(GO) run ./cmd/app $(ARGS)

run:
	$(GO) run ./cmd/app $(ARGS)

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
