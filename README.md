# Logchef MCP Server

A [Model Context Protocol][mcp] (MCP) server for Logchef.

This provides access to your Logchef instance for querying logs, managing teams, and exploring log data through natural language interactions.

## Features

_The following features are currently available in the MCP server. This list is for informational purposes only and does not represent a roadmap or commitment to future features._

### Profile Management
- **Get user profile:** Retrieve current user profile information including user details and API token information
- **Get teams:** List all teams that the authenticated user belongs to, including their role in each team and member count
- **Get server metadata:** Retrieve server version and configuration information

### Source Management
- **Get user sources:** List all log sources accessible to the user across all their team memberships
- **Get team sources:** List log sources for a specific team
- **Get source schema:** Retrieve ClickHouse table schema (column names and types) for a specific source

### Log Querying
- **Query logs:** Execute ClickHouse SQL queries against log sources to retrieve structured log data
- **Natural language to SQL:** Ask questions in natural language and get help constructing efficient ClickHouse queries
- **Schema exploration:** Understand available fields and data types before querying

The list of tools is configurable, so you can choose which tools you want to make available to the MCP client.
This is useful if you don't use certain functionality or if you don't want to take up too much of the context window.
To disable a category of tools, use the `--disable-<category>` flag when starting the server. For example, to disable
the logs tools, use `--disable-logs`.

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

## Usage

1. Generate an API token in Logchef and copy it for use in the configuration.
   Follow your Logchef instance documentation for details on creating API tokens.

2. You have several options to install `logchef-mcp`:

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

If you're using VSCode and running the MCP server in SSE mode (which is the default when using the Docker image without overriding the transport), make sure your `.vscode/settings.json` includes the following:

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

## Log Querying Workflow

The MCP server enables natural language log querying through this workflow:

1. **Explore Sources**: Use `get_sources` or `get_team_sources` to see available log sources
2. **Get Schema**: Use `get_source_schema` to understand the ClickHouse table structure
3. **Query Logs**: Use `query_logs` with raw SQL to retrieve log entries

Example interaction:
- "Show me all available sources" → `get_sources`
- "What fields are available in source 2?" → `get_source_schema` 
- "Get the last 50 error logs from today" → `query_logs` with appropriate ClickHouse SQL

The AI assistant can help construct efficient ClickHouse queries based on the schema and your natural language requests.

## Development

Contributions are welcome! Please open an issue or submit a pull request if you have any suggestions or improvements.

This project is written in Go. Install Go following the instructions for your platform.

To run the server locally in STDIO mode (which is the default for local development), use:

```bash
just run
```

Or manually:

```bash
go run ./cmd/logchef-mcp
```

To run the server locally in SSE mode, use:

```bash
just run-sse
```

Or manually:

```bash
go run ./cmd/logchef-mcp --transport sse
```

You can also run the server using the SSE transport inside a custom built Docker image. Just like the published Docker image, this custom image's entrypoint defaults to SSE mode. To build the image, use:

```bash
just build-image
```

And to run the image in SSE mode (the default), use:

```bash
just docker-run-sse
```

If you need to run it in STDIO mode instead, use:

```bash
just docker-run-stdio
```

### Testing

To test the server, you can build and run it with a local Logchef instance:

```bash
just build
LOGCHEF_URL=http://localhost:5173 LOGCHEF_API_KEY=your_token ./dist/logchef-mcp.bin
```

You can also check the version:

```bash
just show-version
```

## License

This project is licensed under the [Apache License, Version 2.0](LICENSE).

*The design of this repository was inspired by [mcp-grafana](https://github.com/grafana/mcp-grafana).*

[mcp]: https://modelcontextprotocol.io/