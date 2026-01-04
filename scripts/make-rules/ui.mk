# ==============================================================================
# Makefile helper functions for frontend (UI)
# ==============================================================================

# Try to find pnpm binary
PNPM := $(shell \
	if command -v pnpm >/dev/null 2>&1; then \
		echo pnpm; \
	else \
		echo ""; \
	fi)

# Check if pnpm is available
.PHONY: ui.check
ui.check:
	@if [ -z "$(PNPM)" ]; then \
		echo "Error: pnpm is not installed. Please install it first:"; \
		echo "  npm install -g pnpm"; \
		exit 1; \
	fi

# Install frontend dependencies
.PHONY: ui.install
ui.install: ui.check
	@echo "===========> Installing frontend dependencies"
	@cd $(ROOT_DIR)/ui && $(PNPM) install

# Build frontend
.PHONY: ui.build
ui.build: ui.check
	@echo "===========> Building frontend"
	@mkdir -p $(OUTPUT_DIR)/dist
	@cd $(ROOT_DIR)/ui && $(PNPM) build

# Clean frontend build output
.PHONY: ui.clean
ui.clean:
	@echo "===========> Cleaning frontend build output"
	@rm -rf $(OUTPUT_DIR)/dist
	@rm -rf $(ROOT_DIR)/ui/dist
	@rm -rf $(ROOT_DIR)/ui/node_modules/.vite

