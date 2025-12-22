# ==============================================================================
# Makefile helper functions for golang
#

GO := go

.PHONY: go.build
go.build:
	@echo "===========> Building binary"
	@$(GO) build -o $(OUTPUT_DIR)/NexusPointWG cmd/main.go

