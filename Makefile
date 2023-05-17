SHELL := /bin/bash

default: help
.PHONY: default

help: ## Display this help screen (default)
	@grep -h "##" $(MAKEFILE_LIST) | grep -vE '^#|grep' | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' | sort
.PHONY: help

lint: ## Run linter against codebase
	@golangci-lint -v run
.PHONY: lint

build: export WITH_CADDY_ROUTE53           ?= github.com/caddy-dns/route53
build: export WITH_CADDY_VAULT_STORAGE     ?= github.com/mywordpress-io/caddy-vault-storage=.
build: export WITH_CERTMAGIC_VAULT_STORAGE ?= github.com/mywordpress-io/certmagic-vault-storage=../certmagic-vault-storage
build: lint build-setup ## Run 'xcaddy' to build vault storage plugin in to caddy binary
	@xcaddy build --output bin/caddy --with ${WITH_CADDY_ROUTE53} --with ${WITH_CADDY_VAULT_STORAGE} --with ${WITH_CERTMAGIC_VAULT_STORAGE}
.PHONY: build

build-setup:
	@if ! command -v xcaddy >/dev/null 2>&1; then                          \
		echo "ERROR: Missing 'xcaddy' binary on \$PATH, cannot continue";  \
	fi
.PHONY: build-setup

fmt: ## Run go-fmt against codebase
	@go fmt ./...
.PHONY: fmt

mod-download: ## Download go modules
	@go mod download
.PHONY: mod-download

mod-tidy: ## Make sure go modules are tidy
	@go mod tidy
.PHONY: mod-tidy

clean: ## Clean up repo
	@rm -f bin/caddy 2>/dev/null || true
.PHONY: clean
