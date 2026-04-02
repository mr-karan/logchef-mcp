# Logchef MCP Server Setup

Connect your AI assistant to [Logchef](https://logchef.app) for natural language log exploration and analysis powered by ClickHouse.

## Prerequisites

- A running Logchef instance
- A Logchef API token (generate one from Profile > API Tokens in the Logchef UI)

## Quick Start

Choose your AI tool below to get started.

---

### Claude Code

```bash
claude mcp add logchef -- logchef-mcp
```

Set your credentials:

```bash
export LOGCHEF_URL=https://your-logchef-instance.com
export LOGCHEF_API_KEY=your_api_token
```

Verify it works:

```
/mcp
```

You should see `logchef` listed with its tools.

### Claude Desktop

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "logchef": {
      "command": "logchef-mcp",
      "env": {
        "LOGCHEF_URL": "https://your-logchef-instance.com",
        "LOGCHEF_API_KEY": "your_api_token"
      }
    }
  }
}
```

> If you see `Error: spawn logchef-mcp ENOENT`, use the full path to the binary.

### Cursor

Open Cursor Settings > MCP Servers, then add:

```json
{
  "mcpServers": {
    "logchef": {
      "command": "logchef-mcp",
      "env": {
        "LOGCHEF_URL": "https://your-logchef-instance.com",
        "LOGCHEF_API_KEY": "your_api_token"
      }
    }
  }
}
```

### VS Code (Copilot)

Add to `.vscode/settings.json`:

```json
{
  "mcp": {
    "servers": {
      "logchef": {
        "command": "logchef-mcp",
        "env": {
          "LOGCHEF_URL": "https://your-logchef-instance.com",
          "LOGCHEF_API_KEY": "your_api_token"
        }
      }
    }
  }
}
```

For a remote MCP server (streamable HTTP mode):

```json
{
  "mcp": {
    "servers": {
      "logchef": {
        "type": "http",
        "url": "http://localhost:8000/mcp",
        "headers": {
          "X-Logchef-URL": "https://your-logchef-instance.com",
          "X-Logchef-API-Key": "your_api_token"
        }
      }
    }
  }
}
```

### Codex CLI

```bash
# Add MCP server
codex mcp add logchef -- logchef-mcp

# Or with explicit env vars
LOGCHEF_URL=https://your-logchef-instance.com \
LOGCHEF_API_KEY=your_api_token \
codex mcp add logchef -- logchef-mcp
```

### Windsurf

Open Windsurf Settings > MCP, then add:

```json
{
  "mcpServers": {
    "logchef": {
      "command": "logchef-mcp",
      "env": {
        "LOGCHEF_URL": "https://your-logchef-instance.com",
        "LOGCHEF_API_KEY": "your_api_token"
      }
    }
  }
}
```

### Docker

For any tool that supports MCP via stdio:

```json
{
  "mcpServers": {
    "logchef": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i",
        "-e", "LOGCHEF_URL",
        "-e", "LOGCHEF_API_KEY",
        "ghcr.io/mr-karan/logchef-mcp:latest",
        "-t", "stdio"
      ],
      "env": {
        "LOGCHEF_URL": "https://your-logchef-instance.com",
        "LOGCHEF_API_KEY": "your_api_token"
      }
    }
  }
}
```

---

## Installation

### Binary (recommended)

Download the latest release from the [releases page](https://github.com/mr-karan/logchef-mcp/releases) and place it in your `$PATH`.

### Go install

```bash
go install github.com/mr-karan/logchef-mcp/cmd/logchef-mcp@latest
```

### Build from source

```bash
git clone https://github.com/mr-karan/logchef-mcp.git
cd logchef-mcp
go build -o logchef-mcp ./cmd/logchef-mcp
```

### Docker

```bash
docker pull ghcr.io/mr-karan/logchef-mcp:latest
```

---

## Transport Modes

The server supports three transport modes:

| Mode | Flag | Use case |
|------|------|----------|
| **stdio** (default) | `-t stdio` | Direct integration with AI assistants |
| **SSE** | `-t sse` | Legacy web-based clients |
| **Streamable HTTP** | `-t streamable-http` | Multi-client HTTP deployments |

For HTTP modes, the server listens on `localhost:8000` by default. Override with `--address`.

---

## Authentication

### Environment Variables (stdio mode)

```bash
export LOGCHEF_URL=https://your-logchef-instance.com
export LOGCHEF_API_KEY=your_api_token
```

### HTTP Headers (SSE / Streamable HTTP mode)

When running in HTTP mode, clients can pass credentials via headers:

- `X-Logchef-URL` — Logchef instance URL
- `X-Logchef-API-Key` — API token

Headers take precedence over environment variables. If headers are absent, the server falls back to env vars.

---

## Tool Configuration

Selectively enable or disable tool categories:

```bash
# Only enable log querying and profile tools
logchef-mcp --enabled-tools profile,sources,logs,logchefql

# Disable admin tools
logchef-mcp --disable-admin

# Disable telemetry tools
logchef-mcp --disable-telemetry
```

Available categories: `profile`, `sources`, `logs`, `logchefql`, `investigate`, `admin`, `analysis`, `telemetry`, `discover`

---

## Debug Mode

Enable verbose HTTP logging between the MCP server and Logchef API:

```bash
logchef-mcp -debug
```

In Claude Desktop config:

```json
{
  "mcpServers": {
    "logchef": {
      "command": "logchef-mcp",
      "args": ["-debug"],
      "env": {
        "LOGCHEF_URL": "https://your-logchef-instance.com",
        "LOGCHEF_API_KEY": "your_api_token"
      }
    }
  }
}
```
