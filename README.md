# Logchef MCP Server

A [Model Context Protocol][mcp] (MCP) server that connects AI assistants to your [Logchef](https://logchef.app) instance.

Logchef is a powerful log management platform that stores logs in ClickHouse, providing fast querying and analysis capabilities. This MCP server enables AI assistants to interact with your Logchef deployment, making log analysis and troubleshooting more accessible through natural conversation.

## What You Can Do

With this MCP server, you can ask AI assistants to help you:

- **Explore your log infrastructure:** See what teams you belong to and what log sources are available
- **Query logs effectively:** Execute ClickHouse SQL queries to find specific log entries, errors, or patterns
- **Understand your data:** Get schema information to know what fields are available in your logs
- **Analyze log patterns:** Generate histograms and time-series data for trend analysis
- **Manage saved queries:** Create and organize collections of frequently-used queries
- **Administer teams and users:** Handle team membership, user management, and source configuration (admin users)

## Tool Categories

The server provides tools organized into logical categories that can be enabled or disabled as needed:

### Profile & Metadata
Core user and server information including team memberships and server version details.

### Source Management  
Access to log sources across your teams, including schema exploration and source listing.

### Log Analysis
The core querying capabilities including SQL execution, histogram generation, and saved query management.

### Administration
Team management, user administration, source configuration, and API token management (requires admin privileges).

Use the `--disable-<category>` flag to turn off tool categories you don't need. For example, `--disable-admin` removes all administrative tools.

### Tools

| Tool                           | Category | Description                                                                                          |
| ------------------------------ | -------- | ---------------------------------------------------------------------------------------------------- |
| `get_profile`                  | Profile  | Get the current user profile information                                                             |
| `get_teams`                    | Profile  | Get teams that the current user belongs to                                                           |
| `get_meta`                     | Profile  | Get server metadata including version information                                                    |
| `get_sources`                  | Sources  | Get all sources accessible to the user across all team memberships                                  |
| `get_team_sources`             | Sources  | Get sources that belong to a specific team                                                           |
| `query_logs`                   | Logs     | Execute ClickHouse SQL queries against log sources to retrieve log entries                          |
| `get_source_schema`            | Logs     | Get the ClickHouse table schema (column names and types) for a specific source                      |
| `get_log_histogram`            | Logs     | Generate histogram data for logs with customizable time windows and grouping                        |
| `get_collections`              | Logs     | List saved query collections for a specific team and source                                         |
| `create_collection`            | Logs     | Create a new saved query collection                                                                  |
| `get_collection`               | Logs     | Get details of a specific saved query collection                                                     |
| `update_collection`            | Logs     | Update an existing saved query collection                                                            |
| `delete_collection`            | Logs     | Delete a saved query collection                                                                      |
| `list_all_teams`               | Admin    | List all teams in the system (admin only)                                                           |
| `get_team`                     | Admin    | Get detailed information about a specific team                                                       |
| `create_team`                  | Admin    | Create a new team (admin only)                                                                       |
| `update_team`                  | Admin    | Update an existing team's information                                                                |
| `delete_team`                  | Admin    | Delete a team permanently (admin only)                                                              |
| `list_team_members`            | Admin    | List all members of a specific team                                                                  |
| `add_team_member`              | Admin    | Add a user to a team with a specific role                                                           |
| `remove_team_member`           | Admin    | Remove a user from a team                                                                            |
| `link_source_to_team`          | Admin    | Link a log source to a team                                                                          |
| `unlink_source_from_team`      | Admin    | Remove a log source from a team                                                                      |
| `list_all_users`               | Admin    | List all users in the system (admin only)                                                           |
| `get_user`                     | Admin    | Get detailed information about a specific user (admin only)                                         |
| `create_user`                  | Admin    | Create a new user in the system (admin only)                                                        |
| `update_user`                  | Admin    | Update an existing user's information (admin only)                                                  |
| `delete_user`                  | Admin    | Delete a user from the system (admin only)                                                          |
| `list_api_tokens`              | Admin    | List all API tokens for the current user                                                            |
| `create_api_token`             | Admin    | Create a new API token for the current user                                                         |
| `delete_api_token`             | Admin    | Delete an API token                                                                                  |
| `list_all_sources`             | Admin    | List all log sources in the system (admin only)                                                     |
| `create_source`                | Admin    | Create a new log source (admin only)                                                                |
| `validate_source_connection`   | Admin    | Validate ClickHouse connection details (admin only)                                                 |
| `delete_source`                | Admin    | Delete a log source (admin only)                                                                    |
| `get_admin_source_stats`       | Admin    | Get detailed statistics for a log source (admin only)                                               |

## Getting Started

### Prerequisites

You'll need:
- A running Logchef instance ([learn more at logchef.app](https://logchef.app))
- A valid Logchef API token with appropriate permissions

### Generating an API Token

1. Log into your Logchef instance
2. Navigate to your profile settings
3. Create a new API token with the permissions you need
4. Copy the token for use in the MCP server configuration

### Installation

Choose one of the following installation methods:

   - **Docker image**: Use the pre-built Docker image.

     **Important**: The Docker image's entrypoint is configured to run the MCP server in SSE mode by default, but most users will want to use STDIO mode for direct integration with AI assistants like Claude Desktop:

     1. **STDIO Mode**: For stdio mode you must explicitly override the default with `-t stdio` and include the `-i` flag to keep stdin open:

     ```bash
     docker pull logchef-mcp
     docker run --rm -i -e LOGCHEF_URL=http://localhost:5173 -e LOGCHEF_API_KEY=<your_api_token> logchef-mcp -t stdio
     ```

     2. **SSE Mode**: In this mode, the server runs as an HTTP server that clients connect to. You must expose port 8000 using the `-p` flag:

     ```bash
     docker pull logchef-mcp
     docker run --rm -p 8000:8000 -e LOGCHEF_URL=http://localhost:5173 -e LOGCHEF_API_KEY=<your_api_token> logchef-mcp
     ```
     
     3. **Streamable HTTP Mode**: In this mode, the server operates as an independent process that can handle multiple client connections. You must expose port 8000 using the `-p` flag:

     ```bash
     docker pull logchef-mcp
     docker run --rm -p 8000:8000 -e LOGCHEF_URL=http://localhost:5173 -e LOGCHEF_API_KEY=<your_api_token> logchef-mcp -t streamable-http
     ```

   - **Download binary**: Download the latest release of `logchef-mcp` from the releases page and place it in your `$PATH`.

   - **Build from source**: If you have a Go toolchain installed you can also build and install it from source:

     ```bash
     just build
     ```

     Or manually:

     ```bash
     go build -o logchef-mcp.bin ./cmd/logchef-mcp
     ```

3. Add the server configuration to your client configuration file. For example, for Claude Desktop:

   **If using the binary:**

   ```json
   {
     "mcpServers": {
       "logchef": {
         "command": "logchef-mcp.bin",
         "args": [],
         "env": {
           "LOGCHEF_URL": "http://localhost:5173",
           "LOGCHEF_API_KEY": "<your_api_token>"
         }
       }
     }
   }
   ```

> Note: if you see `Error: spawn logchef-mcp.bin ENOENT` in Claude Desktop, you need to specify the full path to `logchef-mcp.bin`.

   **If using Docker:**

   ```json
   {
     "mcpServers": {
       "logchef": {
         "command": "docker",
         "args": [
           "run",
           "--rm",
           "-i",
           "-e",
           "LOGCHEF_URL",
           "-e",
           "LOGCHEF_API_KEY",
           "logchef-mcp",
           "-t",
           "stdio"
         ],
         "env": {
           "LOGCHEF_URL": "http://localhost:5173",
           "LOGCHEF_API_KEY": "<your_api_token>"
         }
       }
     }
   }
   ```

   > Note: The `-t stdio` argument is essential here because it overrides the default SSE mode in the Docker image.

**Using VSCode with remote MCP server**

If you're using VSCode and running the MCP server remotely, you can configure it to connect via HTTP or SSE transport. Make sure your `.vscode/settings.json` includes the following:

For HTTP transport (recommended):
```json
"mcp": {
  "servers": {
    "logchef": {
      "type": "http",
      "url": "http://localhost:8000/mcp",
      "headers": {
        "X-Logchef-URL": "https://your-logchef-instance.com",
        "X-Logchef-API-Key": "your_api_token_here"
      }
    }
  }
}
```

For SSE transport:
```json
"mcp": {
  "servers": {
    "logchef": {
      "type": "sse",
      "url": "http://localhost:8000/sse"
    }
  }
}
```

**Using HTTP Headers for Authentication**

When running the MCP server in SSE or streamable-http mode, you can pass Logchef credentials via HTTP headers instead of environment variables. This is useful when the server needs to handle multiple clients with different Logchef instances or API keys.

The server recognizes these headers:
- `X-Logchef-URL`: The Logchef instance URL  
- `X-Logchef-API-Key`: The API token for authentication

Example configuration for clients that support custom headers:

```json
{
  "mcpServers": {
    "logchef": {
      "type": "http",
      "url": "http://localhost:8000/mcp",
      "headers": {
        "X-Logchef-URL": "https://your-logchef-instance.com",
        "X-Logchef-API-Key": "your_api_token_here"
      }
    }
  }
}
```

If headers are not provided, the server will fall back to environment variables (`LOGCHEF_URL` and `LOGCHEF_API_KEY`).

### Debug Mode

You can enable debug mode for the Logchef transport by adding the `-debug` flag to the command. This will provide detailed logging of HTTP requests and responses between the MCP server and the Logchef API, which can be helpful for troubleshooting.

To use debug mode with the Claude Desktop configuration, update your config as follows:

**If using the binary:**

```json
{
  "mcpServers": {
    "logchef": {
      "command": "logchef-mcp.bin",
      "args": ["-debug"],
      "env": {
        "LOGCHEF_URL": "http://localhost:5173",
        "LOGCHEF_API_KEY": "<your_api_token>"
      }
    }
  }
}
```

**If using Docker:**

```json
{
  "mcpServers": {
    "logchef": {
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-i",
        "-e",
        "LOGCHEF_URL",
        "-e",
        "LOGCHEF_API_KEY",
        "logchef-mcp",
        "-t",
        "stdio",
        "-debug"
      ],
      "env": {
        "LOGCHEF_URL": "http://localhost:5173",
        "LOGCHEF_API_KEY": "<your_api_token>"
      }
    }
  }
}
```

> Note: As with the standard configuration, the `-t stdio` argument is required to override the default SSE mode in the Docker image.

### Tool Configuration

You can selectively enable or disable tool categories using command-line flags:

- `--enabled-tools`: Comma-separated list of tool categories to enable (default: "profile,sources,logs,admin")
- `--disable-profile`: Disable profile management tools
- `--disable-sources`: Disable source management tools  
- `--disable-logs`: Disable log querying tools
- `--disable-admin`: Disable admin tools

Example with selective tool enabling:

```json
{
  "mcpServers": {
    "logchef": {
      "command": "logchef-mcp.bin",
      "args": ["--enabled-tools", "profile,logs"],
      "env": {
        "LOGCHEF_URL": "http://localhost:5173",
        "LOGCHEF_API_KEY": "<your_api_token>"
      }
    }
  }
}
```

## Working with Logchef Through AI

This MCP server enables you to interact with your Logchef logs through natural conversation with AI assistants. Here's how it typically works:

### Discovery Workflow
1. **"What log sources do I have access to?"** → The AI uses `get_sources` to show your available sources
2. **"What data is in the nginx source?"** → The AI calls `get_source_schema` to explain the log structure
3. **"Show me recent errors"** → The AI constructs and executes a ClickHouse query using `query_logs`

### Practical Examples
- **Troubleshooting**: "Find all 500 errors in the last hour from the web service logs"
- **Analysis**: "Show me a histogram of log volume over the past day"
- **Investigation**: "What are the most common error messages in the database logs?"
- **Monitoring**: "Create a saved query for tracking API response times"

### AI-Assisted Query Building
The AI assistant understands ClickHouse SQL and can help you:
- Build complex queries with proper syntax
- Optimize queries for better performance  
- Explain what fields are available in your log data
- Suggest useful queries based on common log analysis patterns

Since Logchef uses ClickHouse as the storage backend, you get the full power of ClickHouse's analytical capabilities through natural language interaction.

## Development

This project is written in Go and uses [Just](https://github.com/casey/just) as a command runner for common development tasks.

### Prerequisites
- Go 1.21 or later
- A running Logchef instance for testing
- Valid Logchef API credentials

### Quick Start

```bash
# Install dependencies
just deps

# Run in STDIO mode (default for development)
just run

# Run in SSE mode for web-based clients
just run-sse

# Build the binary
just build

# Run tests
just test
```

### Environment Variables

For development, set these environment variables:

```bash
export LOGCHEF_URL=http://localhost:5173  # Your Logchef instance URL
export LOGCHEF_API_KEY=your_api_token_here
```

### Docker Development

Build and test the Docker image locally:

```bash
# Build development image
just build-image

# Test GoReleaser Docker setup
just build-goreleaser-image

# Run in different modes
just docker-run-stdio
just docker-run-sse
```

### Testing & Verification

Test the server with your Logchef instance:

```bash
# Build and run with your credentials
just build
LOGCHEF_URL=http://localhost:5173 LOGCHEF_API_KEY=your_token ./dist/logchef-mcp

# Check version info
just show-version

# Test basic functionality
just help
```

### Contributing

Contributions are welcome! When adding new tools or features:

1. Follow the existing patterns for tool definition and client methods
2. Add appropriate error handling and validation
3. Update the tools table in this README
4. Test with a real Logchef instance

Please open an issue to discuss larger changes before implementing them.

## License

This project is licensed under the [Apache License, Version 2.0](LICENSE).

*The design of this repository was inspired by [mcp-grafana](https://github.com/grafana/mcp-grafana).*

[mcp]: https://modelcontextprotocol.io/