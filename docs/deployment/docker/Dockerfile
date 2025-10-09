# Multi-stage build for goimgserver
FROM golang:1.21-bookworm AS builder

# Install libvips development dependencies
RUN apt-get update && apt-get install -y \
    libvips-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /build

# Copy go module files
COPY src/go.mod src/go.sum ./
RUN go mod download

# Copy source code
COPY src/ ./

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w" \
    -o goimgserver \
    main.go

# Runtime stage
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    libvips \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN useradd --system --no-create-home --shell /bin/false goimgserver

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/goimgserver /app/goimgserver

# Create directories
RUN mkdir -p /app/images /app/cache && \
    chown -R goimgserver:goimgserver /app

# Switch to non-root user
USER goimgserver

# Expose port
EXPOSE 9000

# Volume mounts
VOLUME ["/app/images", "/app/cache"]

# Environment variables
ENV PORT=9000 \
    IMAGES_DIR=/app/images \
    CACHE_DIR=/app/cache \
    GOMAXPROCS=4 \
    GOGC=100

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/goimgserver", "--help"] || exit 1

# Run the application
ENTRYPOINT ["/app/goimgserver"]
CMD ["--port=9000", "--imagesdir=/app/images", "--cachedir=/app/cache", "--precache=true"]
