# Multi-stage build for smaller image size
# Stage 1: Build the Go binary
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy shared package
COPY shared/ ./shared/

# Copy service code
COPY user-service/ ./user-service/

# Build the binary
# CGO_ENABLED=0: Static binary (no C dependencies)
# -ldflags="-w -s": Strip debug info for smaller size
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/user-service \
    ./user-service

# Stage 2: Create minimal runtime image
FROM alpine:latest

# Install CA certificates (for HTTPS calls)
RUN apk --no-cache add ca-certificates

# Create non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/user-service .

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check (Kubernetes uses /health endpoint)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the binary
ENTRYPOINT ["/app/user-service"]