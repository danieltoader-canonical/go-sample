.DEFAULT_GOAL := help

GO ?= go
GREMLINS_VERSION ?= v0.6.0
BASE_REF ?= origin/main

.PHONY: help test test-weak mutate mutate-weak mutate-diff

help:
	@printf '%s\n' \
		'make test         Run corrected invoice tests' \
		'make test-weak    Run intentionally weak invoice tests with coverage' \
		'make mutate       Run mutation testing with corrected tests' \
		'make mutate-weak  Run mutation testing with weak tests (mutants survive)' \
		'make mutate-diff  Mutate changed lines against BASE_REF (default: origin/main)'

test:
	$(GO) test ./...

test-weak:
	$(GO) test -tags=weak -cover ./...

mutate:
	$(GO) run github.com/go-gremlins/gremlins/cmd/gremlins@$(GREMLINS_VERSION) unleash --workers 1 --timeout-coefficient 100

mutate-weak:
	$(GO) run github.com/go-gremlins/gremlins/cmd/gremlins@$(GREMLINS_VERSION) unleash --workers 1 --timeout-coefficient 100 --tags weak

mutate-diff:
	@export GIT_CONFIG_COUNT=2 GIT_CONFIG_KEY_0=diff.relative GIT_CONFIG_VALUE_0=true \
		GIT_CONFIG_KEY_1=diff.context GIT_CONFIG_VALUE_1=0; \
	git diff --quiet --merge-base "$(BASE_REF)" -- .; status=$$?; \
	if [ "$$status" -eq 0 ]; then \
		echo "No tracked changes in this module relative to $(BASE_REF); skipping mutation testing."; \
	elif [ "$$status" -eq 1 ]; then \
		$(GO) run github.com/go-gremlins/gremlins/cmd/gremlins@$(GREMLINS_VERSION) unleash --workers 1 --timeout-coefficient 100 --diff "$(BASE_REF)"; \
	else \
		echo "Could not compare against $(BASE_REF)." >&2; exit "$$status"; \
	fi
