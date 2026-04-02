# Tools Reference

Complete reference for all tools, resources, and prompts available in the Logchef MCP server.

## Tools

### Profile & Metadata

| Tool | Description |
|------|-------------|
| `get_profile` | Get current user profile, including email, role, and API token info |
| `get_teams` | List teams you belong to with your role and member counts |
| `get_meta` | Server version and configuration details |

### Source Management

| Tool | Description |
|------|-------------|
| `get_sources` | All sources accessible to you, aggregated across all team memberships |
| `get_team_sources` | Sources belonging to a specific team |

### Log Querying

| Tool | Description |
|------|-------------|
| `query_logs` | Execute raw ClickHouse SQL against a log source (max 100 rows) |
| `query_logchefql` | Execute a LogchefQL query — simpler filter syntax (max 500 rows) |
| `translate_logchefql` | Translate LogchefQL to ClickHouse SQL without executing |
| `validate_logchefql` | Check LogchefQL syntax for errors without executing |
| `get_source_schema` | Get column names and ClickHouse types for a source |
| `get_log_histogram` | Time-series histogram of log volume with optional grouping |

### Saved Queries (Collections)

| Tool | Description |
|------|-------------|
| `get_collections` | List saved query collections for a team/source |
| `get_collection` | Get a specific saved query by ID |
| `create_collection` | Save a new query collection |
| `update_collection` | Update an existing saved query |
| `delete_collection` | Delete a saved query permanently |

### Investigation

| Tool | Description |
|------|-------------|
| `get_field_values` | Top distinct values for a field in a time range |
| `get_log_context` | Surrounding log entries before/after a timestamp |
| `list_alerts` | All alert rules configured for a source |
| `get_alert_history` | Evaluation history for a specific alert |

### Analysis

| Tool | Description |
|------|-------------|
| `compare_windows` | Run the same query across two time windows and compare row counts |
| `top_values` | Get top values for multiple fields in one call |

### Discovery

| Tool | Description |
|------|-------------|
| `generate_query` | Generate a ClickHouse SQL query from natural language (requires AI enabled on Logchef) |
| `get_all_field_dimensions` | Get top values for all LowCardinality fields in one call |

### Telemetry

| Tool | Description |
|------|-------------|
| `get_query_telemetry` | Recent query performance from ClickHouse system.query_log |

### Administration

| Tool | Description |
|------|-------------|
| `list_all_teams` | List all teams (admin only) |
| `get_team` | Get team details |
| `create_team` | Create a new team (admin only) |
| `update_team` | Update team name/description |
| `delete_team` | Delete a team permanently (admin only) |
| `list_team_members` | List members of a team |
| `add_team_member` | Add a user to a team with a role |
| `remove_team_member` | Remove a user from a team |
| `link_source_to_team` | Grant a team access to a source |
| `unlink_source_from_team` | Revoke a team's access to a source |
| `list_all_users` | List all users (admin only) |
| `get_user` | Get user details (admin only) |
| `create_user` | Create a user (admin only) |
| `update_user` | Update user info (admin only) |
| `delete_user` | Delete a user (admin only) |
| `list_all_sources` | List all sources (admin only) |
| `create_source` | Create a ClickHouse log source (admin only) |
| `validate_source_connection` | Test ClickHouse connectivity (admin only) |
| `delete_source` | Delete a source (admin only) |
| `get_admin_source_stats` | ClickHouse table stats — row count, sizes, compression (admin only) |
| `list_api_tokens` | List your API tokens |
| `create_api_token` | Create a new API token |
| `delete_api_token` | Delete an API token |

---

## Resources

Resources provide read-only data that AI assistants can access without explicit tool calls.

| URI Template | Description |
|-------------|-------------|
| `logchef://team/{team_id}/source/{source_id}/schema` | ClickHouse schema for a source |
| `logchef://team/{team_id}/source/{source_id}/collections` | List of saved queries |
| `logchef://team/{team_id}/source/{source_id}/collection/{collection_id}` | A single saved query |

---

## Prompts

Prompts provide guided investigation workflows that instruct the AI assistant through multi-step analysis.

### `investigate_error_spike`

Walk through error spike diagnosis: schema discovery, error volume assessment, pattern identification, timeline correlation, and root cause analysis.

**Arguments:**
- `team_id` (required) — Team ID
- `source_id` (required) — Source ID
- `time_range` (optional) — e.g. "last 1h", defaults to last 1 hour

### `investigate_alert`

Investigate a specific alert: review configuration, check evaluation history, reproduce the alert query, explore context, and summarize findings.

**Arguments:**
- `team_id` (required) — Team ID
- `source_id` (required) — Source ID
- `alert_id` (required) — Alert ID

---

## Tool Annotations

Every tool includes MCP annotations to help AI assistants understand safety:

- **Read-only tools** (`readOnlyHint: true`): All query, list, and get operations
- **Non-destructive writes** (`destructiveHint: false`): Create and update operations
- **Destructive operations** (`destructiveHint: true`): Delete operations

AI assistants use these hints to decide when to ask for user confirmation before proceeding.
