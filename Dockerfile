# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git (needed for go mod sometimes)
RUN apk add --no-cache git

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy full source
COPY . .

# Build HTTP server
RUN CGO_ENABLED=0 GOOS=linux go build -o http-server ./cmd/server

# Build gRPC server
RUN CGO_ENABLED=0 GOOS=linux go build -o grpc-server ./cmd/grpc

# Stage 2: Runtime
FROM alpine:3.18

WORKDIR /app

# Copy both binaries
COPY --from=builder /app/http-server .
COPY --from=builder /app/grpc-server .

# Expose ports (documentation only)
EXPOSE 8080
EXPOSE 50051

# Default command (can be overridden by docker-compose)
CMD ["./http-server"]