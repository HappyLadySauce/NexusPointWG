# ==============================================================================
# Makefile helper functions for golang
#

# Try to find go binary: prefer GOROOT (if set), otherwise fall back to PATH.
# This avoids toolchain/stdlib mismatches when the environment sets GOROOT.
GO := $(shell \
	if [ -n "$$GOROOT" ] && [ -f "$$GOROOT/bin/go" ]; then \
		echo $$GOROOT/bin/go; \
	elif command -v go >/dev/null 2>&1; then \
		echo go; \
	else \
		echo go; \
	fi)

# Get GOPATH from go env, fallback to environment variable or default
GOPATH := $(shell $(GO) env GOPATH 2>/dev/null || echo $$GOPATH || echo $(HOME)/go)

# Get GOROOT from go env, fallback to environment variable
GOROOT := $(shell $(GO) env GOROOT 2>/dev/null || echo $$GOROOT)

.PHONY: go.build
go.build:
	@echo "===========> Building binary"
	@$(GO) build -o $(OUTPUT_DIR)/NexusPointWG cmd/main.go

.PHONY: go.run
go.run:
	@echo "===========> Running NexusPointWG"
	@mkdir -p "$(OUTPUT_DIR)" "$(dir $(NEXUS_POINT_WG_LOGS_LOG_FILE))"
	@$(GO) run cmd/main.go -c $(CONFIG_FILE)

