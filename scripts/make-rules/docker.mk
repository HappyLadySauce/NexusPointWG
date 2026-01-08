# ==============================================================================
# Makefile helper functions for Docker
# ==============================================================================

# Version extraction from pkg/environment/version.go
# Extract the dev constant value as the default version
# This ensures single source of truth for version management
# Use awk for better compatibility (works on most Unix systems)
DOCKER_VERSION_FROM_SRC := $(shell awk '/^[[:space:]]*dev[[:space:]]*=[[:space:]]*"/ {match($$0, /"[^"]+"/); print substr($$0, RSTART+1, RLENGTH-2); exit}' $(ROOT_DIR)/pkg/environment/version.go)
ifeq ($(DOCKER_VERSION_FROM_SRC),)
  # Fallback: try with grep (if available)
  DOCKER_VERSION_FROM_SRC := $(shell grep -E '^\s*dev\s*=\s*"' $(ROOT_DIR)/pkg/environment/version.go | sed -E 's/.*dev\s*=\s*"([^"]+)".*/\1/' | head -1)
endif
ifeq ($(DOCKER_VERSION_FROM_SRC),)
  # Final fallback
  DOCKER_VERSION_FROM_SRC := 1.0.0-dev
endif

# Binary name with version (use from main Makefile if available, otherwise compute)
DOCKER_BINARY_NAME := $(if $(BINARY_NAME),$(BINARY_NAME),NexusPointWG-$(DOCKER_VERSION_FROM_SRC))

# Docker Compose file paths
DOCKER_COMPOSE_DEV := $(ROOT_DIR)/docker/docker-compose.dev.yml
DOCKER_COMPOSE_RELEASE := $(ROOT_DIR)/docker/docker-compose.release.yml
DOCKERFILE := $(ROOT_DIR)/docker/Dockerfile

# ==============================================================================
# Docker build targets
# ==============================================================================

## docker.build.dev: Build dev image (default).
.PHONY: docker.build.dev
docker.build.dev:
	@echo "===========> Building dev binary and Docker image (version: $(DOCKER_VERSION_FROM_SRC))"
	@echo "===========> Cleaning _output directory"
	@rm -rf $(OUTPUT_DIR)/*
	@$(MAKE) BUILD_VERSION= BINARY_NAME=$(DOCKER_BINARY_NAME) go.build ui.build
	@echo "===========> Building Docker image with tag: $(DOCKER_VERSION_FROM_SRC)"
	@docker build --build-arg BUILD_VERSION=$(DOCKER_VERSION_FROM_SRC) --build-arg BINARY_NAME=$(DOCKER_BINARY_NAME) -t nexuspointwg:$(DOCKER_VERSION_FROM_SRC) -f $(DOCKERFILE) $(ROOT_DIR)

## docker.build.release: Build release image.
.PHONY: docker.build.release
docker.build.release:
	@echo "===========> Building release binary and Docker image"
	@echo "===========> Cleaning _output directory"
	@rm -rf $(OUTPUT_DIR)/*
	@set -e; \
	RELEASE_VERSION=$$(awk '/^[[:space:]]*release[[:space:]]*=[[:space:]]*"/ {match($$0, /"[^"]+"/); print substr($$0, RSTART+1, RLENGTH-2); exit}' $(ROOT_DIR)/pkg/environment/version.go); \
	if [ -z "$$RELEASE_VERSION" ]; then \
		RELEASE_VERSION=$$(grep -E '^\s*release\s*=\s*"' $(ROOT_DIR)/pkg/environment/version.go | sed -E 's/.*release\s*=\s*"([^"]+)".*/\1/' | head -1); \
	fi; \
	if [ -z "$$RELEASE_VERSION" ]; then \
		RELEASE_VERSION="1.0.0"; \
	fi; \
	echo "Release version: $$RELEASE_VERSION"; \
	RELEASE_BINARY_NAME="NexusPointWG-$$RELEASE_VERSION"; \
	$(MAKE) BUILD_VERSION=$$RELEASE_VERSION BINARY_NAME=$$RELEASE_BINARY_NAME go.build ui.build; \
	echo "===========> Building Docker image with tag: $$RELEASE_VERSION"; \
	docker build --build-arg BUILD_VERSION=$$RELEASE_VERSION --build-arg BINARY_NAME=$$RELEASE_BINARY_NAME -t nexuspointwg:$$RELEASE_VERSION -f $(DOCKERFILE) $(ROOT_DIR)

## docker.build: Default to dev build.
.PHONY: docker.build
docker.build: docker.build.dev

# ==============================================================================
# Docker run targets
# ==============================================================================

## docker.run.dev: Run dev environment with docker-compose.dev.yml.
.PHONY: docker.run.dev
docker.run.dev:
	@echo "===========> Starting dev services with docker compose ($(DOCKER_COMPOSE_DEV), version: $(DOCKER_VERSION_FROM_SRC))"
	@if docker ps -a --format '{{.Names}}' | grep -q '^nexuspointwg-dev$$'; then \
		echo "Removing existing dev container..."; \
		docker rm -f nexuspointwg-dev || true; \
	fi
	@IMAGE_TAG=$(DOCKER_VERSION_FROM_SRC) docker compose -f $(DOCKER_COMPOSE_DEV) up

## docker.run.release: Run release environment with docker-compose.release.yml.
.PHONY: docker.run.release
docker.run.release:
	@echo "===========> Starting release services with docker compose ($(DOCKER_COMPOSE_RELEASE))"
	@set -e; \
	RELEASE_VERSION=$$(awk '/^[[:space:]]*release[[:space:]]*=[[:space:]]*"/ {match($$0, /"[^"]+"/); print substr($$0, RSTART+1, RLENGTH-2); exit}' $(ROOT_DIR)/pkg/environment/version.go); \
	if [ -z "$$RELEASE_VERSION" ]; then \
		RELEASE_VERSION=$$(grep -E '^\s*release\s*=\s*"' $(ROOT_DIR)/pkg/environment/version.go | sed -E 's/.*release\s*=\s*"([^"]+)".*/\1/' | head -1); \
	fi; \
	if [ -z "$$RELEASE_VERSION" ]; then \
		RELEASE_VERSION="1.0.0"; \
	fi; \
	echo "Release version: $$RELEASE_VERSION"; \
	if docker ps -a --format '{{.Names}}' | grep -q '^nexuspointwg$$'; then \
		echo "Removing existing release container..."; \
		docker rm -f nexuspointwg || true; \
	fi; \
	IMAGE_TAG=$$RELEASE_VERSION docker compose -f $(DOCKER_COMPOSE_RELEASE) up

## docker.run: Default to dev run.
.PHONY: docker.run
docker.run: docker.run.dev

## docker.up: Backward-compatible alias for docker.run.
.PHONY: docker.up
docker.up: docker.run

# ==============================================================================
# Docker management targets
# ==============================================================================

## docker.down: Stop and remove services.
.PHONY: docker.down
docker.down:
	@echo "===========> Stopping services"
	@if docker ps -a --format '{{.Names}}' | grep -q '^nexuspointwg-dev$$'; then \
		echo "Stopping dev services..."; \
		IMAGE_TAG=$(DOCKER_VERSION_FROM_SRC) docker compose -f $(DOCKER_COMPOSE_DEV) down || true; \
	fi
	@if docker ps -a --format '{{.Names}}' | grep -q '^nexuspointwg$$'; then \
		echo "Stopping release services..."; \
		RELEASE_VERSION=$$(awk '/^[[:space:]]*release[[:space:]]*=[[:space:]]*"/ {match($$0, /"[^"]+"/); print substr($$0, RSTART+1, RLENGTH-2); exit}' $(ROOT_DIR)/pkg/environment/version.go 2>/dev/null || echo "1.0.0"); \
		IMAGE_TAG=$$RELEASE_VERSION docker compose -f $(DOCKER_COMPOSE_RELEASE) down || true; \
	fi

## docker.restart: Restart services.
.PHONY: docker.restart
docker.restart:
	@echo "===========> Restarting services"
	@if docker ps -a --format '{{.Names}}' | grep -q '^nexuspointwg-dev$$'; then \
		echo "Restarting dev services..."; \
		IMAGE_TAG=$(DOCKER_VERSION_FROM_SRC) docker compose -f $(DOCKER_COMPOSE_DEV) restart || true; \
	fi
	@if docker ps -a --format '{{.Names}}' | grep -q '^nexuspointwg$$'; then \
		echo "Restarting release services..."; \
		RELEASE_VERSION=$$(awk '/^[[:space:]]*release[[:space:]]*=[[:space:]]*"/ {match($$0, /"[^"]+"/); print substr($$0, RSTART+1, RLENGTH-2); exit}' $(ROOT_DIR)/pkg/environment/version.go 2>/dev/null || echo "1.0.0"); \
		IMAGE_TAG=$$RELEASE_VERSION docker compose -f $(DOCKER_COMPOSE_RELEASE) restart || true; \
	fi

## docker.push: Push release Docker image to Docker Hub (happlelaoganma/nexuspointwg).
.PHONY: docker.push
docker.push:
	@echo "===========> Pushing release Docker image to Docker Hub"
	@set -e; \
	RELEASE_VERSION=$$(awk '/^[[:space:]]*release[[:space:]]*=[[:space:]]*"/ {match($$0, /"[^"]+"/); print substr($$0, RSTART+1, RLENGTH-2); exit}' $(ROOT_DIR)/pkg/environment/version.go); \
	if [ -z "$$RELEASE_VERSION" ]; then \
		RELEASE_VERSION=$$(grep -E '^\s*release\s*=\s*"' $(ROOT_DIR)/pkg/environment/version.go | sed -E 's/.*release\s*=\s*"([^"]+)".*/\1/' | head -1); \
	fi; \
	if [ -z "$$RELEASE_VERSION" ]; then \
		RELEASE_VERSION="1.0.0"; \
	fi; \
	echo "Release version: $$RELEASE_VERSION"; \
	SOURCE_IMAGE="nexuspointwg:$$RELEASE_VERSION"; \
	TARGET_IMAGE_VERSION="happlelaoganma/nexuspointwg:$$RELEASE_VERSION"; \
	TARGET_IMAGE_LATEST="happlelaoganma/nexuspointwg:latest"; \
	echo "Checking if source image exists: $$SOURCE_IMAGE"; \
	if ! docker image inspect $$SOURCE_IMAGE >/dev/null 2>&1; then \
		echo "Error: Source image $$SOURCE_IMAGE does not exist."; \
		echo "Please run 'make docker.build.release' first to build the release image."; \
		exit 1; \
	fi; \
	echo "Tagging image as $$TARGET_IMAGE_VERSION"; \
	docker tag $$SOURCE_IMAGE $$TARGET_IMAGE_VERSION; \
	echo "Tagging image as $$TARGET_IMAGE_LATEST"; \
	docker tag $$SOURCE_IMAGE $$TARGET_IMAGE_LATEST; \
	echo "Pushing $$TARGET_IMAGE_VERSION"; \
	docker push $$TARGET_IMAGE_VERSION; \
	echo "Pushing $$TARGET_IMAGE_LATEST"; \
	docker push $$TARGET_IMAGE_LATEST; \
	echo "===========> Successfully pushed release image to Docker Hub"
