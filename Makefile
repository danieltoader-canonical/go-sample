.DEFAULT_GOAL := help

PROJECT ?= repoPattern

.PHONY: help start run up down test test-weak bench generate mutate mutate-weak mutate-diff

help start run up down test test-weak bench generate mutate mutate-weak mutate-diff:
	$(MAKE) -C $(PROJECT) $@
