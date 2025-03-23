
SHELL := /bin/bash

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*' $(MAKEFILE_LIST) \
		|sed 's|\(.*\):.*|\1|' |column
.PHONY: help

release:
	@bash scripts/release.sh
.PHONY: release

update-mod:
	go get -u ./...
	go mod tidy
.PHONY: update-mod

build-tests:
	@bash scripts/build_tests.sh
	@export GO_BIN=go1.20.14 ; bash scripts/build_tests.sh
.PHONY: build-tests

build-examples:
	@bash scripts/build_examples.sh
	@export GO_BIN=go1.20.14 ; bash scripts/build_examples.sh
.PHONY: build-examples
