
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

## docker.build: Build backend and frontend, then build Docker image.
.PHONY: docker.build
docker.build: go.build ui.build
	@echo "===========> Building Docker image"
	@docker build -t nexuspointwg:latest -f Dockerfile .

## docker.run: Run NexusPointWG in Docker container.
.PHONY: docker.run
docker.run:
	@echo "===========> Running Docker container"
	@if docker ps -a --format '{{.Names}}' | grep -q '^nexuspointwg$$'; then \
		echo "Removing existing container..."; \
		docker rm -f nexuspointwg || true; \
	fi
	@docker run -d --name nexuspointwg \
		-p 8001:8001 \
		-v $(ROOT_DIR)/configs:/app/configs:ro \
		-v $(ROOT_DIR)/nexuspointwg.db:/app/nexuspointwg.db \
		nexuspointwg:latest
	@echo "Container started. Use 'docker logs nexuspointwg' to view logs."

## docker.stop: Stop Docker container.
.PHONY: docker.stop
docker.stop:
	@echo "===========> Stopping Docker container"
	@docker stop nexuspointwg || true

## docker.rm: Remove Docker container.
.PHONY: docker.rm
docker.rm:
	@echo "===========> Removing Docker container"
	@docker rm -f nexuspointwg || true

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