SHELL := /bin/sh
.DEFAULT_GOAL := dev

MAIN_PKG := ./cmd/memd
TOOLS_BIN := ./.bin
GOVULNCHECK := $(TOOLS_BIN)/govulncheck
MEMD_BIN := $(TOOLS_BIN)/memd
GOVULNCHECK_VERSION ?= v1.1.4

VERSION  ?= dev
GIT_SHA  := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_TS ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

GO_LDFLAGS := -s -w
GO_LDFLAGS += -X github.com/sojournerdev/memd/internal/commands.VersionString=$(VERSION)
GO_LDFLAGS += -X github.com/sojournerdev/memd/internal/commands.CommitHash=$(GIT_SHA)
GO_LDFLAGS += -X github.com/sojournerdev/memd/internal/commands.BuildDate=$(BUILD_TS)

.PHONY: dev
dev:
	@set -eu; \
	go mod tidy; \
	go fmt ./...; \
	go vet ./...; \
	go test -count=1 ./...; \
	go install -trimpath -ldflags "$(GO_LDFLAGS)" $(MAIN_PKG)

.PHONY: ci
ci:
	@set -eu; \
	echo "==> tidy"; \
	go mod tidy; \
	git diff --exit-code -- go.mod go.sum; \
	echo "==> mod verify"; \
	go mod verify; \
	echo "==> fmt"; \
	test -z "$$(gofmt -l .)" || (echo "gofmt needed:"; gofmt -l .; exit 1); \
	echo "==> vet"; \
	go vet ./...; \
	echo "==> vulncheck (bootstrap)"; \
	mkdir -p "$(TOOLS_BIN)"; \
	GOBIN="$(abspath $(TOOLS_BIN))" go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION); \
	"$(GOVULNCHECK)" ./...; \
	echo "==> test"; \
	go test -count=1 ./...; \
	echo "==> build"; \
	go build -trimpath -ldflags "$(GO_LDFLAGS)" -o "$(MEMD_BIN)" $(MAIN_PKG); \
