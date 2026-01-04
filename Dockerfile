# ==============================================================================
# Dockerfile for NexusPointWG
# Copy pre-built binaries and frontend files from _output directory
# ==============================================================================

FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata wget

# Create app user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy backend binary from _output (built locally)
COPY _output/NexusPointWG /app/NexusPointWG

# Copy frontend build output from _output/dist (built locally)
COPY _output/dist /app/ui

# Copy config file (if exists)
COPY configs/NexusPointWG-Example.yaml /app/configs/NexusPointWG.yaml

# Change ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8001

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8001/livez || exit 1

# Run the application
ENTRYPOINT ["/app/NexusPointWG", "-c", "/app/configs/NexusPointWG.yaml"]

