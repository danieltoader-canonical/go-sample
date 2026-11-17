.DEFAULT_GOAL := help

GO ?= go
ARGS ?= command get-version
BENCH ?= .
BENCHTIME ?= 1s

.PHONY: help start run test bench generate

help:
	@printf '%s\n' \
		'make start     Run the sample (no external services needed)' \
		'make run       Run the CLI (ARGS="command get-version")' \
		'make test      Run Go tests' \
		'make bench     Run benchmarks with allocation statistics (BENCH=. BENCHTIME=1s)' \
		'make generate  Run Go generation directives'

start: run

run:
	$(GO) run ./cmd/cli $(ARGS)

test:
	$(GO) test ./...

bench:
	$(GO) test -run='^$$' -bench='$(BENCH)' -benchmem -benchtime='$(BENCHTIME)' ./...

generate:
	$(GO) generate ./...
