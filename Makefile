.DEFAULT_GOAL := help

PROJECT ?= repoPattern

.PHONY: help start run up down test bench generate

help start run up down test bench generate:
	$(MAKE) -C $(PROJECT) $@
