# Stage 1: Build
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum from root
COPY go.mod go.sum ./
RUN go mod download

# Copy entire project
COPY . .

# Build the Go binary from cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

# Stage 2: Minimal runtime
FROM alpine:3.18

WORKDIR /app

# Copy built binary
COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]