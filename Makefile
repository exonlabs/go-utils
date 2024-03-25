
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
