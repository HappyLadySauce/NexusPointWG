# ==============================================================================
# Makefile helper functions for golang
#

# Try to find go binary: first check PATH, then try GOROOT env var
GO := $(shell \
	if command -v go >/dev/null 2>&1; then \
		echo go; \
	elif [ -n "$$GOROOT" ] && [ -f "$$GOROOT/bin/go" ]; then \
		echo $$GOROOT/bin/go; \
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

