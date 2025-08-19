# Multi-stage build for crypto trading bot
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build all binaries
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/dca-bot ./cmd/dca-bot
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/grid-bot ./cmd/grid-bot
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/combo-bot ./cmd/combo-bot
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/backtester ./cmd/backtester

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binaries from builder stage
COPY --from=builder /app/bin/ ./bin/
COPY --from=builder /app/configs/ ./configs/

# Create logs directory
RUN mkdir -p logs && chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose default port
EXPOSE 8080

# Set default command
CMD ["./bin/dca-bot", "-config", "configs/dca-config.json"]
