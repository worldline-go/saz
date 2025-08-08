BINARY    := saz
MAIN_FILE := cmd/$(BINARY)/main.go

LOCAL_BIN_DIR := $(PWD)/bin

BUILD_DATE := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
BUILD_COMMIT := $(shell git rev-parse --short HEAD)
VERSION := $(or $(IMAGE_TAG),$(shell git describe --tags --first-parent --match "v*" 2> /dev/null || echo v0.0.0))

.DEFAULT_GOAL := help

.PHONY: build
build: CGO_ENABLED ?= 0
build: GOOS ?= linux
build: GOARCH ?= amd64
build: ## Build the binary
	go build -trimpath -ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(BUILD_COMMIT) -X main.date=$(BUILD_DATE)" -o bin/$(BINARY_NAME) $(BINARY_PATH)

.PHONY: run
run: export LOG_LEVEL ?= debug
run: ## Run the application
	go run $(MAIN_FILE)

.PHONY: lint
lint: ## Lint Go files
	@GOPATH="$(shell dirname $(PWD))" golangci-lint run ./...

.PHONY: test
test: ## Run unit tests
	@go test -v -race ./...

.PHONY: env
env: ## Create environment
	@echo "> Creating environment $(BINARY)"
	docker compose --project-name=$(BINARY) --file=env/docker-compose.yaml up -d

.PHONY: env-down
env-down: ## Destroy environment
	@echo "> Destroying environment $(BINARY)"
	docker compose --project-name=$(BINARY) down --volumes

.PHONY: help
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
