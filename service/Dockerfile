# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tmux \
    git

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/server .

# Expose port
EXPOSE 8080

# Set environment variables
ENV HOST=0.0.0.0
ENV PORT=8080

# Run the application
CMD ["./server"]
