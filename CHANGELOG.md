# Changelog

All notable changes to logchef-mcp will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

## [0.3.0] - 2026-05-13

Tracks logchef v1.6.0. Saved queries and alerts are no longer team-scoped on
the server — their endpoints in this client are rewired and the team-scoped
URLs are gone.

### Changed (breaking)
- **Saved queries rewired to `/api/v1/saved-queries`.** The old
  `/api/v1/teams/:teamID/sources/:sourceID/collections` endpoints are gone.
  Tools renamed: `get_collections` → `list_saved_queries`, `get_collection`
  → `get_saved_query`, `create_collection` → `create_saved_query`,
  `update_collection` → `update_saved_query`, `delete_collection` →
  `delete_saved_query`. The client functions follow the same naming.
- **Saved-query create body shape changed** to match logchef
  `CreateSavedQueryRequest`: `source_id`, `name`, `description`,
  `query_type` (`"logchefql"` or `"sql"`), `query_content` (JSON envelope).
  The previous `query` field is replaced by `query_content` carrying the
  full envelope.
- **Update no longer accepts `source_id`** — saved queries can't be
  retargeted to a different source. Create a new query instead.
- **Alerts rewired to `/api/v1/alerts`.** `list_alerts` drops the
  `team_id` argument and now accepts an optional `source_id` filter.
  `get_alert_history` drops `team_id` + `source_id` — the new endpoint is
  `/api/v1/alerts/:alertID/history`.
- **`investigate_alert` prompt** drops the required `team_id` and
  `source_id` arguments; only `alert_id` is required now (team/source are
  optional hints for schema/log lookups).
- **Resource URI templates updated**:
  - `logchef://team/{team_id}/source/{source_id}/collections` →
    `logchef://source/{source_id}/saved-queries`
  - `logchef://team/{team_id}/source/{source_id}/collection/{collection_id}`
    → `logchef://saved-query/{query_id}`

### Changed (non-breaking)
- **Dependencies refreshed.** `mark3labs/mcp-go` v0.46.0 → v0.53.0,
  `spf13/cast` v1.9.2 → v1.10.0, `google/jsonschema-go` v0.4.2 → v0.4.3.
  Go toolchain bumped to 1.25.5.

### Added (rolled in from the previously-unreleased buffer)
- **Typed tool registration** — All tools use mcp-go's native
  `WithInputSchema[T]` / `WithOutputSchema[T]` with `NewTypedToolHandler`
  and `NewStructuredToolHandler`, replacing the custom reflection-based
  wrapper.
- **Structured output schemas** — Tools with fixed response shapes
  (profile, teams, sources, saved queries, admin CRUD) return typed
  structured content with JSON schema descriptions.
- **Tool annotations** — Every tool has `ReadOnlyHintAnnotation`,
  `DestructiveHintAnnotation`, and `TitleAnnotation` so AI assistants
  understand safety implications.
- **Resource templates** — 3 MCP resources for read-only data access:
  - `logchef://team/{team_id}/source/{source_id}/schema` — ClickHouse schema
  - `logchef://source/{source_id}/saved-queries` — Saved query list
  - `logchef://saved-query/{query_id}` — Single saved query
- **Investigation prompts** — 2 guided workflows:
  - `investigate_error_spike` — Schema discovery, error volume, pattern
    identification, timeline correlation, root cause analysis
  - `investigate_alert` — Alert config review, evaluation history, query
    reproduction, context exploration
- **Analysis tools**: `compare_windows` (two-window diff with row-count
  delta), `top_values` (parallel multi-field top values).
- **Discovery tools**: `generate_query` (natural language → ClickHouse
  SQL), `get_all_field_dimensions` (bulk LowCardinality top values).
- **`validate_logchefql`** — syntax check without executing.
- **`get_query_telemetry`** — recent query performance from
  `system.query_log` (query text excluded for privacy).
- **Server capabilities** — `WithResourceCapabilities`,
  `WithPromptCapabilities`, `WithRecovery()` on the MCP server.
- **Conditional prompts/resources** — only registered when their
  dependent tool categories are enabled.
- **Documentation** — `docs/setup.md` (per-provider setup) and
  `docs/tools.md` (full tool/resource/prompt reference).

### Changed (rolled in)
- **Handler error pattern** — Structured handlers return Go errors (SDK
  converts to tool errors); typed handlers use `mcp.NewToolResultError()`
  for flexible output tools.
- **`get_sources` parallelized** — Fetches team sources concurrently
  instead of sequentially (N+1 fix).
- **`top_values` parallelized** — Fetches field values concurrently across
  all requested fields.

### Fixed (rolled in)
- **jsonschema tag format** — Struct tags updated from
  `jsonschema:"description=X,required"` (invopop format) to
  `jsonschema:"X"` (google/jsonschema-go format). The old format silently
  produced empty input schemas.
- **URL parameter injection** — `GetFieldValues` uses `url.PathEscape` /
  `url.Values` instead of raw string interpolation.
- **Unbounded log context** — `get_log_context` before/after limits
  capped at 100 (was unbounded).
- **Query text leakage** — `get_query_telemetry` no longer returns the
  `query` column from `system.query_log`.
- **Alert prompt args** — `investigate_alert` prompt argument plumbing
  no longer drops required IDs.

### Removed
- **Custom tool wrapper** — Deleted `tools.go` (228 lines) with the
  `MustTool` / `ConvertTool` / `Tool` types and the `invopop/jsonschema`
  reflection-based schema generator.
- **5 transitive dependencies** — `invopop/jsonschema`,
  `go-ordered-map/v2`, `easyjson`, `generic-list-go`, `buger/jsonparser`.

## [0.2.0] - 2026-04-01

### Added
- **LogchefQL tools** — `query_logchefql` and `translate_logchefql` for Logchef's native query syntax
- **Investigation tools** — `get_field_values`, `get_log_context`, `list_alerts`, `get_alert_history`
- **Admin tools** — Full team/user/source/token CRUD (16 tools)
- **mcp-go v0.46.0** — Upgraded from v0.32.0

## [0.1.0] - 2025-06-13

### Added
- Initial release with profile, sources, logs, collections, and schema tools
- stdio, SSE, and streamable-http transport support
- Docker image at `ghcr.io/mr-karan/logchef-mcp`
- Environment variable and HTTP header authentication
