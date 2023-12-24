SHELL := /bin/bash

default: help
.PHONY: default

help: ## Display this help screen (default)
	@grep -h "##" $(MAKEFILE_LIST) | grep -vE '^#|grep' | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' | sort
.PHONY: help

lint: ## Run linter against codebase
	@golangci-lint -v run
.PHONY: lint

# CADDY_VERSION should match the equivalent go.mod entry
build: export CADDY_VERSION                ?= v2.7.5
build: export WITH_CADDY_ROUTE53           ?= github.com/caddy-dns/route53
build: export WITH_CADDY_VAULT_STORAGE     ?= github.com/mywordpress-io/caddy-vault-storage=.
build: export WITH_CERTMAGIC_VAULT_STORAGE ?= github.com/mywordpress-io/certmagic-vault-storage=../certmagic-vault-storage
build: mod-tidy lint build-setup ## Run 'xcaddy' to build vault storage plugin in to caddy binary
	@xcaddy build ${CADDY_VERSION} --output bin/caddy --with ${WITH_CADDY_ROUTE53} --with ${WITH_CADDY_VAULT_STORAGE} --with ${WITH_CERTMAGIC_VAULT_STORAGE}
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

mod-update: export MODULE ?=
mod-update: ## Update go proxy with latest module version: MODULE=github.com/mywordpress-io/caddy-vault-storage@v0.1.1 make mod-update
	@if [[ -n "${MODULE}" ]]; then                       \
		GOPROXY=proxy.golang.org go list -m ${MODULE};   \
	else                                                 \
		echo "ERROR: Missing 'MODULE', cannot continue"; \
		exit 1;                                          \
	fi
.PHONY: mod-update

localdev: ## Start a local 'vault' instance running in dev mode
	@vault server -dev -dev-root-token-id=dead-beef
.PHONY: localdev

test-setup: export VAULT_ADDR  := http://localhost:8200
test-setup: export VAULT_TOKEN := dead-beef
test-setup: ## Bootstrap local vault 'dev' server for tests
	@if ! vault secrets list | grep -ci 'secrets' >/dev/null 2>&1 ; then                              \
		vault secrets enable -version=2 -path=secrets kv 2>/dev/null;                                 \
		vault policy write vault-certmagic-storage localdev/vault-certmagic-storage-policy.hcl;       \
		vault auth enable approle 2>/dev/null;                                                        \
		vault write auth/approle/role/vault-certmagic-storage token_ttl=30s token_max_ttl=30s token_policies=default,vault-certmagic-storage 2>/dev/null;  \
	fi
.PHONY: localdev-setup

test: export VAULT_ADDR  ?= http://localhost:8200
test: export VAULT_TOKEN ?= dead-beef
test: export GINKGO_PATH ?= ./...
test: build test-setup ## Perform ginkgo tests against codebase: GINKGO_PATH=./... make test
	$(eval export VAULT_APPROLE_ROLE_ID=$(shell VAULT_ADDR="${VAULT_ADDR}" VAULT_TOKEN="${VAULT_TOKEN}" vault read -format=json auth/approle/role/vault-certmagic-storage/role-id | jq -r '.data.role_id'))
	$(eval export VAULT_APPROLE_SECRET_ID=$(shell VAULT_ADDR="${VAULT_ADDR}" VAULT_TOKEN="${VAULT_TOKEN}" vault write -format=json -f auth/approle/role/vault-certmagic-storage/secret-id | jq -r '.data.secret_id'))
	@ginkgo -r -v --race --cover --coverprofile code-coverage.out --trace --timeout 5m ${GINKGO_PATH}
.PHONY: test

test-coverage:
	@go tool cover -func code-coverage.out
	@gocover-cobertura < code-coverage.out > code-coverage.xml
.PHONY: test-coverage

clean: ## Clean up repo
	@rm -f bin/caddy 2>/dev/null || true
.PHONY: clean
