# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the migration binary
RUN go build -o migrate ./migrate/migrate.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy the migration binary from builder
COPY --from=builder /app/migrate .

# Create a directory for SQLite file mounting
RUN mkdir -p /data

# The migration tool expects a SQLite database path as argument
ENTRYPOINT ["./migrate"]