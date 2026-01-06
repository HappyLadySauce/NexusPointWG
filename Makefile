
# Makefile for NexusPointWG

# Build all by default, even if it's not first
.DEFAULT_GOAL := all

# Silence "Entering/Leaving directory ..." output from recursive make invocations.
MAKEFLAGS += --no-print-directory

.PHONY: all
all: build

# ==============================================================================
# Build options

ROOT_PACKAGE=github.com/HappyLadySauce/NexusPointWG

# ==============================================================================
# Version extraction from pkg/environment/version.go
# Extract the dev constant value as the default version
# This ensures single source of truth for version management
# Use awk for better compatibility (works on most Unix systems)
VERSION_FROM_SRC := $(shell awk '/^[[:space:]]*dev[[:space:]]*=[[:space:]]*"/ {match($$0, /"[^"]+"/); print substr($$0, RSTART+1, RLENGTH-2); exit}' pkg/environment/version.go)
ifeq ($(VERSION_FROM_SRC),)
  # Fallback: try with grep (if available)
  VERSION_FROM_SRC := $(shell grep -E '^\s*dev\s*=\s*"' pkg/environment/version.go | sed -E 's/.*dev\s*=\s*"([^"]+)".*/\1/' | head -1)
endif
ifeq ($(VERSION_FROM_SRC),)
  # Final fallback
  VERSION_FROM_SRC := 1.0.0-dev
endif

# Binary name with version
BINARY_NAME := NexusPointWG-$(VERSION_FROM_SRC)

# ==============================================================================
# Includes

include scripts/make-rules/common.mk
include scripts/make-rules/config.mk
include scripts/make-rules/golang.mk
include scripts/make-rules/tools.mk
include scripts/make-rules/swagger.mk
include scripts/make-rules/ui.mk

# ==============================================================================
# Usage

define USAGE_OPTIONS

Options:
  DEBUG            Whether to generate debug symbols. Default is 0.
  BINS             The binaries to build. Default is all of cmd.
                   This option is available when using: make build/build.multiarch
                   Example: make build BINS="iam-apiserver iam-authz-server"
  IMAGES           Backend images to make. Default is all of cmd starting with iam-.
                   This option is available when using: make image/image.multiarch/push/push.multiarch
                   Example: make image.multiarch IMAGES="iam-apiserver iam-authz-server"
  REGISTRY_PREFIX  Docker registry prefix. Default is marmotedu. 
                   Example: make push REGISTRY_PREFIX=ccr.ccs.tencentyun.com/marmotedu VERSION=v1.6.2
  PLATFORMS        The multiple platforms to build. Default is linux_amd64 and linux_arm64.
                   This option is available when using: make build.multiarch/image.multiarch/push.multiarch
                   Example: make image.multiarch IMAGES="iam-apiserver iam-pump" PLATFORMS="linux_amd64 linux_arm64"
  VERSION          The version information compiled into binaries.
                   The default is obtained from gsemver or git.
  V                Set to 1 enable verbose build. Default is 0.
endef
export USAGE_OPTIONS

# ==============================================================================
# Targets

## build: Build source code for host platform.
.PHONY: build
build:
	@$(MAKE) go.build

## swagger: Generate swagger document.
.PHONY: swagger
swagger:
	@$(MAKE) swagger.run

## run: Run NexusPointWG with config from configs/NexusPointWG.yaml (values can be provided via scripts/make-rules/config.mk).
.PHONY: run
run:
	@$(MAKE) go.run

## ui.build: Build frontend application to _output/dist.

## docker.build.dev: Build dev image (default).
.PHONY: docker.build.dev
docker.build.dev:
	@echo "===========> Building dev binary and Docker image (version: $(VERSION_FROM_SRC))"
	@$(MAKE) BUILD_VERSION= BINARY_NAME=$(BINARY_NAME) go.build ui.build
	@echo "===========> Building Docker image with tag: $(VERSION_FROM_SRC)"
	@docker build --build-arg BUILD_VERSION=$(VERSION_FROM_SRC) --build-arg BINARY_NAME=$(BINARY_NAME) -t nexuspointwg:$(VERSION_FROM_SRC) -f Dockerfile .

## docker.build.release: Build release image.
.PHONY: docker.build.release
docker.build.release:
	@echo "===========> Building release binary and Docker image"
	@RELEASE_VERSION=$$(awk '/^[[:space:]]*release[[:space:]]*=[[:space:]]*"/ {match($$0, /"[^"]+"/); print substr($$0, RSTART+1, RLENGTH-2); exit}' pkg/environment/version.go); \
	if [ -z "$$RELEASE_VERSION" ]; then \
		RELEASE_VERSION=$$(grep -E '^\s*release\s*=\s*"' pkg/environment/version.go | sed -E 's/.*release\s*=\s*"([^"]+)".*/\1/' | head -1); \
	fi; \
	if [ -z "$$RELEASE_VERSION" ]; then \
		RELEASE_VERSION="1.0.0"; \
	fi; \
	echo "Release version: $$RELEASE_VERSION"; \
	RELEASE_BINARY_NAME="NexusPointWG-$$RELEASE_VERSION"; \
	$(MAKE) BUILD_VERSION=$$RELEASE_VERSION BINARY_NAME=$$RELEASE_BINARY_NAME go.build ui.build; \
	echo "===========> Building Docker image with tag: $$RELEASE_VERSION"; \
	docker build --build-arg BUILD_VERSION=$$RELEASE_VERSION --build-arg BINARY_NAME=$$RELEASE_BINARY_NAME -t nexuspointwg:$$RELEASE_VERSION -f Dockerfile .

## docker.build: Default to dev build.
.PHONY: docker.build
docker.build: docker.build.dev

## docker.run.dev: Run dev environment with docker-compose.dev.yml.
.PHONY: docker.run.dev
docker.run.dev:
	@echo "===========> Starting dev services with docker compose (docker-compose.dev.yml, version: $(VERSION_FROM_SRC))"
	@if docker ps -a --format '{{.Names}}' | grep -q '^nexuspointwg-dev$$'; then \
		echo "Removing existing dev container..."; \
		docker rm -f nexuspointwg-dev || true; \
	fi
	@IMAGE_TAG=$(VERSION_FROM_SRC) docker compose -f docker-compose.dev.yml up

## docker.run.release: Run release environment with docker-compose.release.yml.
.PHONY: docker.run.release
docker.run.release:
	@echo "===========> Starting release services with docker compose (docker-compose.release.yml)"
	@RELEASE_VERSION=$$(awk '/^[[:space:]]*release[[:space:]]*=[[:space:]]*"/ {match($$0, /"[^"]+"/); print substr($$0, RSTART+1, RLENGTH-2); exit}' pkg/environment/version.go); \
	if [ -z "$$RELEASE_VERSION" ]; then \
		RELEASE_VERSION=$$(grep -E '^\s*release\s*=\s*"' pkg/environment/version.go | sed -E 's/.*release\s*=\s*"([^"]+)".*/\1/' | head -1); \
	fi; \
	if [ -z "$$RELEASE_VERSION" ]; then \
		RELEASE_VERSION="1.0.0"; \
	fi; \
	echo "Release version: $$RELEASE_VERSION"; \
	if docker ps -a --format '{{.Names}}' | grep -q '^nexuspointwg$$'; then \
		echo "Removing existing release container..."; \
		docker rm -f nexuspointwg || true; \
	fi; \
	IMAGE_TAG=$$RELEASE_VERSION docker compose -f docker-compose.release.yml up

## docker.run: Default to dev run.
.PHONY: docker.run
docker.run: docker.run.dev

## docker.up: Backward-compatible alias for docker.run.
.PHONY: docker.up
docker.up: docker.run

## docker.down: Stop and remove services.
.PHONY: docker.down
docker.down:
	@echo "===========> Stopping services"
	@docker compose down

## docker.restart: Restart services.
.PHONY: docker.restart
docker.restart:
	@echo "===========> Restarting services"
	@docker compose restart

## docker.push: Push Docker image to registry.
.PHONY: docker.push
docker.push: REGISTRY_PREFIX ?= 
docker.push: VERSION ?= latest
docker.push:
	@if [ -z "$(REGISTRY_PREFIX)" ]; then \
		echo "Error: REGISTRY_PREFIX is required. Example: make docker.push REGISTRY_PREFIX=your-registry.com/namespace VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "===========> Pushing Docker image"
	@docker tag nexuspointwg:latest $(REGISTRY_PREFIX)/nexuspointwg:$(VERSION)
	@docker push $(REGISTRY_PREFIX)/nexuspointwg:$(VERSION)

.PHONY: tidy
tidy:
	@$(GO) mod tidy

## help: Show this help info.
.PHONY: help
help: Makefile
	@printf "\nUsage: make <TARGETS> <OPTIONS> ...\n\nTargets:\n"
	@sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/ /'
	@echo "$$USAGE_OPTIONS"