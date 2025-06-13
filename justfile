# Logchef MCP Server - Justfile

# Build variables
VERSION := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`
COMMIT := `git rev-parse --short HEAD 2>/dev/null || echo "unknown"`
DATE := `date -u +"%Y-%m-%dT%H:%M:%SZ"`
LDFLAGS := "-X main.version=" + VERSION + " -X main.commit=" + COMMIT + " -X main.date=" + DATE

# Default recipe to display help
default:
    @just --list

# Build the binary with version info
build:
    mkdir -p dist
    go build -ldflags="{{LDFLAGS}}" -o dist/logchef-mcp ./cmd/logchef-mcp

# Build for Linux (primary target)
build-linux:
    mkdir -p dist
    GOOS=linux GOARCH=amd64 go build -ldflags="{{LDFLAGS}}" -o dist/logchef-mcp-linux-amd64 ./cmd/logchef-mcp

# Build for macOS (development convenience)
build-darwin:
    mkdir -p dist
    GOOS=darwin GOARCH=amd64 go build -ldflags="{{LDFLAGS}}" -o dist/logchef-mcp-darwin-amd64 ./cmd/logchef-mcp
    GOOS=darwin GOARCH=arm64 go build -ldflags="{{LDFLAGS}}" -o dist/logchef-mcp-darwin-arm64 ./cmd/logchef-mcp

# Build for Windows (development convenience)
build-windows:
    mkdir -p dist
    GOOS=windows GOARCH=amd64 go build -ldflags="{{LDFLAGS}}" -o dist/logchef-mcp-windows-amd64.exe ./cmd/logchef-mcp

# Build for multiple platforms (development only - releases use GoReleaser)
build-all: build-linux build-darwin build-windows

# Build the Docker image with version info
build-image:
    docker build \
        --build-arg VERSION={{VERSION}} \
        --build-arg COMMIT={{COMMIT}} \
        --build-arg DATE={{DATE}} \
        -t logchef-mcp:{{VERSION}} \
        -t logchef-mcp:latest \
        .

# Build the GoReleaser Docker image locally for testing
build-goreleaser-image: build-linux
    cp dist/logchef-mcp-linux-amd64 logchef-mcp
    docker build \
        -f Dockerfile.goreleaser \
        -t logchef-mcp:goreleaser \
        .
    rm logchef-mcp

# Run the MCP server in stdio mode
run:
    go run -ldflags="{{LDFLAGS}}" ./cmd/logchef-mcp

# Run the MCP server in SSE mode with debug logging
run-sse:
    go run -ldflags="{{LDFLAGS}}" ./cmd/logchef-mcp --transport sse --log-level debug --debug

# Run the MCP server in StreamableHTTP mode with debug logging  
run-streamable-http:
    go run -ldflags="{{LDFLAGS}}" ./cmd/logchef-mcp --transport streamable-http --log-level debug --debug

# Run unit tests
test:
    go test -v ./...

# Run tests with coverage
test-cover:
    go test -v -cover ./...

# Run Go linter
lint:
    go vet ./...
    go fmt ./...

# Clean build artifacts
clean:
    rm -rf dist/
    go clean

# Show version information
version:
    @echo "Version: {{VERSION}}"
    @echo "Commit: {{COMMIT}}"
    @echo "Date: {{DATE}}"

# Show help for the binary
help:
    ./dist/logchef-mcp --help

# Show version of built binary
show-version: build
    ./dist/logchef-mcp --version

# Install dependencies
deps:
    go mod download
    go mod tidy

# Run Docker container in stdio mode
docker-run-stdio: build-image
    docker run --rm -i \
        -e LOGCHEF_URL=http://localhost:5173 \
        -e LOGCHEF_API_KEY=your_token_here \
        logchef-mcp:latest -t stdio

# Run Docker container in SSE mode
docker-run-sse: build-image
    docker run --rm -p 8000:8000 \
        -e LOGCHEF_URL=http://localhost:5173 \
        -e LOGCHEF_API_KEY=your_token_here \
        logchef-mcp:latest

# Show project info
info:
    @echo "Logchef MCP Server"
    @echo "=================="
    @echo "Version: {{VERSION}}"
    @echo "Commit: {{COMMIT}}"
    @echo "Date: {{DATE}}"
    @echo "Go version: $(go version)"
    @echo "Module: $(head -1 go.mod)"
    @echo ""
    @echo "Available commands:"
    @just --list --unsorted