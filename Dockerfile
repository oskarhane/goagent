# Multi-stage Dockerfile for GoAgent
# Stage 1: Build
FROM golang:1.26-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with version injection
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.version=${VERSION}" \
    -o goagent \
    ./cmd/goagent

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 goagent && \
    adduser -D -u 1000 -G goagent goagent

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/goagent /app/goagent

# Change ownership to non-root user
RUN chown -R goagent:goagent /app

# Switch to non-root user
USER goagent

# Expose port for potential health check endpoint (if added in future)
EXPOSE 8080

# Health check (using binary version check as basic health indicator)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/goagent", "--version"]

# Set environment variables with defaults
ENV LOG_LEVEL=info \
    MAX_ITERATIONS=10 \
    TIMEOUT=300

# Default entrypoint
ENTRYPOINT ["/app/goagent"]

# Default command (can be overridden)
CMD ["--help"]
