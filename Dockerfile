# Build stage - compile Go binary
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git for fetching dependencies
RUN apk add --no-cache git

# Copy dependency files first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /ckm ./cmd/kernel_main.go

# Production stage - minimal image
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS and docker CLI for container management
RUN apk --no-cache add ca-certificates docker-cli

# Create non-root user for security
RUN addgroup -S ckm && adduser -S ckm -G ckm
USER ckm

# Copy binary from builder
COPY --from=builder /ckm /app/ckm

# Copy config files
COPY --chown=ckm:ckm configs/ /app/configs/

# Expose API and metrics ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

# Run the binary
ENTRYPOINT ["/app/ckm"]
