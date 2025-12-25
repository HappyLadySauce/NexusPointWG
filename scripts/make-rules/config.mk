
# Centralized runtime configuration for `make run`.
# Values are passed to the application through environment variables and expanded
# in `configs/NexusPointWG.yaml` via ${ENV_VAR} syntax.

# Config file path used by `go.run`.
CONFIG_FILE ?= $(ROOT_DIR)/configs/NexusPointWG.yaml

# --- insecure ---
export ORGANIZE_TOYS_SERVER_INSECURE_BIND_ADDRESS ?= 127.0.0.1
export ORGANIZE_TOYS_SERVER_INSECURE_BIND_PORT ?= 8001

# --- logs ---
# Write logs into _output/logs by default (keeps repo clean).
export NEXUS_POINT_WG_LOGS_LOG_FILE ?= $(ROOT_DIR)/logs/NexusPointWG.log
export NEXUS_POINT_WG_LOGS_LOG_MAX_SIZE ?= 100
export NEXUS_POINT_WG_LOGS_LOG_MAX_BACKUPS ?= 3
export NEXUS_POINT_WG_LOGS_LOG_MAX_AGE ?= 28
export NEXUS_POINT_WG_LOGS_LOG_COMPRESS ?= true

# --- sqlite ---
export NEXUS_POINT_WG_SQLITE_DATA_SOURCE_NAME ?= $(ROOT_DIR)/nexuspointwg.db

