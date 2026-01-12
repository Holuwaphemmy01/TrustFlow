# Build Stage
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install git (required for fetching dependencies)
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 is used since we are using modernc.org/sqlite (pure Go)
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server src/cmd/server/main.go

# Run Stage
FROM alpine:latest

# Install CA certificates for HTTPS (RPC calls)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/server .

# Expose the API port
EXPOSE 8081

# Run the server
CMD ["./server"]
