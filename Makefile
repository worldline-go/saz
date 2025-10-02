PROJECT    := saz
MAIN_FILE := cmd/$(PROJECT)/main.go

LOCAL_BIN_DIR := $(PWD)/bin

BUILD_DATE := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
BUILD_COMMIT := $(shell git rev-parse --short HEAD)
VERSION := $(or $(IMAGE_TAG),$(shell git describe --tags --first-parent --match "v*" 2> /dev/null || echo v0.0.0))

.DEFAULT_GOAL := help

.PHONY: build-front
build-front: ## Build the front-end application (manually)
	@echo "> Building front-end application"
	cd _ui && pnpm install && pnpm run build
	@echo "> Replacing build output"
	rm -rf internal/server/dist 2> /dev/null
	mv _ui/dist internal/server/dist

.PHONY: build-releaser
build-releaser: ## Build the binary with goreleaser
	goreleaser build --snapshot --clean --single-target

.PHONY: build-container
build-container: ## Build the container image with test tag
	docker build -t $(PROJECT):test -f ci/Dockerfile .

.PHONY: run
run: export LOG_LEVEL ?= debug
run: ## Run the application
	go run $(MAIN_FILE)

.PHONY: build-in
build-in: ## Build binary inside of the container
	docker run -it --rm \
		-v $(PWD):/workspace \
		-v $(HOME)/.cache:/.cache \
		-v $(HOME)/go/pkg/mod:/go/pkg/mod \
		-w /workspace \
		-u $(shell id -u):$(shell id -g) \
		ghcr.io/rytsh/dock/build/go:1.25.1 \
		make build-releaser

.PHONY: lint
lint: ## Lint Go files
	golangci-lint run ./...

.PHONY: test
test: ## Run unit tests
	@go test -v -race ./...

.PHONY: env
env: ## Create environment
	@echo "> Creating environment $(PROJECT)"
	docker compose --project-name=$(PROJECT) --file=env/docker-compose.yaml up -d

.PHONY: env-down
env-down: ## Destroy environment
	@echo "> Destroying environment $(PROJECT)"
	docker compose --project-name=$(PROJECT) down --volumes

.PHONY: help
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
