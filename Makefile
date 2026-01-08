
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
include scripts/make-rules/docker.mk

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

## 1panel: Package 1panel app directory into tar.gz archive.
.PHONY: 1panel
1panel:
	@echo "Packaging 1panel app..."
	@cd docker/1panel && tar czvf nexuspointwg.tar.gz nexuspointwg/
	@echo "Package created: docker/1panel/nexuspointwg.tar.gz"

.PHONY: tidy
tidy:
	@$(GO) mod tidy

## help: Show this help info.
.PHONY: help
help: Makefile
	@printf "\nUsage: make <TARGETS> <OPTIONS> ...\n\nTargets:\n"
	@sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/ /'
	@echo "$$USAGE_OPTIONS"