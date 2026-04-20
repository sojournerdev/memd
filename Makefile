SHELL := /bin/sh
.DEFAULT_GOAL := run

MAIN_PKG := ./cmd/memd
BIN_DIR := ./.bin
TOOLS_DIR := ./.tools
GOVULNCHECK := $(TOOLS_DIR)/govulncheck
GOLANGCI_LINT := $(TOOLS_DIR)/golangci-lint
MEMD_BIN := $(BIN_DIR)/memd
GOVULNCHECK_VERSION ?= v1.1.4
GOLANGCI_LINT_VERSION ?= v2.11.2
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X main.version=$(VERSION)

.PHONY: build
build:
	@set -eu; \
	go mod tidy; \
	go fmt ./...; \
	go vet ./...; \
	go test -count=1 ./...; \
	mkdir -p "$(BIN_DIR)"; \
	go build -trimpath -ldflags "$(LDFLAGS)" -o "$(MEMD_BIN)" $(MAIN_PKG)

.PHONY: run
run: build
	@set -eu; \
	"$(MEMD_BIN)"

.PHONY: tools
tools:
	@set -eu; \
	mkdir -p "$(TOOLS_DIR)"; \
	echo "==> install golangci-lint"; \
	GOBIN="$(abspath $(TOOLS_DIR))" go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	echo "==> install govulncheck"; \
	GOBIN="$(abspath $(TOOLS_DIR))" go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION); \

.PHONY: ci
ci: tools
	@set -eu; \
	echo "==> tidy"; \
	go mod tidy; \
	git diff --exit-code -- go.mod go.sum; \
	echo "==> mod verify"; \
	go mod verify; \
	echo "==> fmt"; \
	test -z "$$(gofmt -l .)" || (echo "gofmt needed:"; gofmt -l .; exit 1); \
	echo "==> lint"; \
	"$(GOLANGCI_LINT)" run --config .golangci.yml ./...; \
	echo "==> vulncheck"; \
	"$(GOVULNCHECK)" ./...; \
	echo "==> test"; \
	go test -count=1 ./...; \
	echo "==> build"; \
	mkdir -p "$(BIN_DIR)"; \
	go build -trimpath -ldflags "$(LDFLAGS)" -o "$(MEMD_BIN)" $(MAIN_PKG); \
