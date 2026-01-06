# ==============================================================================
# Dockerfile for NexusPointWG
# Copy pre-built binaries and frontend files from _output directory
# ==============================================================================

FROM alpine:latest

# Install runtime dependencies
# util-linux provides nsenter to access host systemd
# curl is needed for public IP detection
RUN apk add --no-cache ca-certificates tzdata wget curl dbus util-linux

# Create systemctl wrapper script to communicate with host systemd
# If running with --pid=host, systemctl is available directly
# Otherwise, uses nsenter to access host's PID namespace
RUN echo '#!/bin/sh' > /usr/local/bin/systemctl && \
    echo '# Wrapper to execute systemctl on the host system' >> /usr/local/bin/systemctl && \
    echo 'if [ -f /proc/1/ns/pid ] && [ "$(readlink /proc/self/ns/pid)" != "$(readlink /proc/1/ns/pid)" ]; then' >> /usr/local/bin/systemctl && \
    echo '  # Not in host PID namespace, use nsenter' >> /usr/local/bin/systemctl && \
    echo '  exec nsenter -t 1 -m -u -i -n -p systemctl "$@"' >> /usr/local/bin/systemctl && \
    echo 'else' >> /usr/local/bin/systemctl && \
    echo '  # In host PID namespace, use systemctl directly' >> /usr/local/bin/systemctl && \
    echo '  # Note: systemctl binary needs to be available in PATH' >> /usr/local/bin/systemctl && \
    echo '  # For Alpine, we use nsenter as fallback since systemctl is not available' >> /usr/local/bin/systemctl && \
    echo '  exec nsenter -t 1 -m -u -i -n -p systemctl "$@"' >> /usr/local/bin/systemctl && \
    echo 'fi' >> /usr/local/bin/systemctl && \
    chmod +x /usr/local/bin/systemctl

# Create app user
RUN addgroup -g 51830 nexuspointwg && \
    adduser -D -u 51830 -G nexuspointwg nexuspointwg

# Set working directory
WORKDIR /app

# Copy backend binary from _output (built locally)
COPY _output/NexusPointWG /app/NexusPointWG

# Copy frontend build output from _output/dist (built locally)
COPY _output/dist /app/ui

# Copy config file
COPY configs/NexusPointWG.yaml /app/configs/NexusPointWG.yaml

# Change ownership
RUN chown -R nexuspointwg:nexuspointwg /app

# Switch to non-root user
USER nexuspointwg

# Expose port
EXPOSE 51830

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:51830/livez || exit 1

# Run the application
ENTRYPOINT ["/app/NexusPointWG", "-c", "/app/configs/NexusPointWG.yaml"]

