# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install required build tools and dependencies
RUN apk add --no-cache git protoc protobuf-dev make

# Set up Go environment
ENV PATH="/app/bin:${PATH}"
ENV GO111MODULE=on
ENV CGO_ENABLED=0

# Install protoc plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

# Copy source code
COPY . .

# Generate proto files
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pkg/proto/auth/v1/auth.proto \
    pkg/proto/document/v1/document.proto \
    pkg/proto/collaboration/v1/collaboration.proto

# Update go.mod to include all dependencies
RUN go mod tidy

# Build the application
RUN go build -o server cmd/server/main.go

# Final stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    curl \
    tzdata

WORKDIR /app

# Copy the binary and configurations
COPY --from=builder /app/server .
COPY --from=builder /app/init/init.sql ./init/
COPY --from=builder /app/mosquitto/config/mosquitto.conf ./mosquitto/config/

# Create necessary directories
RUN mkdir -p /app/mosquitto/data /app/mosquitto/log

# Expose port
EXPOSE 50051

# Run the application
CMD ["./server"]