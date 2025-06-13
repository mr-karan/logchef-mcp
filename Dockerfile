# Build stage
FROM golang:1.24-bullseye AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown
RUN go build -ldflags="-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" -o logchef-mcp.bin ./cmd/logchef-mcp

# Final stage
FROM debian:bullseye-slim

# Install ca-certificates for HTTPS requests
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Create a non-root user
RUN useradd -r -u 1000 -m logchef-mcp

# Set the working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder --chown=1000:1000 /app/logchef-mcp.bin /app/

# Use the non-root user
USER logchef-mcp

# Expose the port the app runs on
EXPOSE 8000

# Run the application
ENTRYPOINT ["/app/logchef-mcp.bin", "--transport", "sse", "--address", "0.0.0.0:8000"]
