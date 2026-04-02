# Changelog

All notable changes to logchef-mcp will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added
- **Typed tool registration** — All tools now use mcp-go v0.46.0 native `WithInputSchema[T]` / `WithOutputSchema[T]` with `NewTypedToolHandler` and `NewStructuredToolHandler`, replacing the custom reflection-based wrapper.
- **Structured output schemas** — Tools with fixed response shapes (profile, teams, sources, collections, admin CRUD) return typed structured content with JSON schema descriptions.
- **Tool annotations** — Every tool has `ReadOnlyHintAnnotation`, `DestructiveHintAnnotation`, and `TitleAnnotation` so AI assistants understand safety implications.
- **Resource templates** — 3 MCP resources for read-only data access:
  - `logchef://team/{team_id}/source/{source_id}/schema` — ClickHouse schema
  - `logchef://team/{team_id}/source/{source_id}/collections` — Saved query list
  - `logchef://team/{team_id}/source/{source_id}/collection/{collection_id}` — Single saved query
- **Investigation prompts** — 2 guided workflows:
  - `investigate_error_spike` — Schema discovery, error volume, pattern identification, timeline correlation, root cause analysis
  - `investigate_alert` — Alert config review, evaluation history, query reproduction, context exploration
- **Analysis tools**:
  - `compare_windows` — Run the same LogchefQL query across two time windows and compare row counts with delta
  - `top_values` — Get top distinct values for multiple fields in one call (parallelized)
- **Discovery tools**:
  - `generate_query` — Natural language to ClickHouse SQL via Logchef AI endpoint
  - `get_all_field_dimensions` — Bulk fetch top values for all LowCardinality fields in one call
- **LogchefQL validation** — `validate_logchefql` tool checks syntax without executing
- **Query telemetry** — `get_query_telemetry` tool reads ClickHouse `system.query_log` for query performance data (query text excluded for privacy)
- **Server capabilities** — `WithResourceCapabilities`, `WithPromptCapabilities`, `WithRecovery()` on MCP server
- **Conditional prompts/resources** — Only registered when their dependent tool categories are enabled
- **Documentation** — `docs/setup.md` with per-provider setup (Claude Code, Claude Desktop, Cursor, VS Code, Codex CLI, Windsurf, Docker) and `docs/tools.md` with full tool/resource/prompt reference

### Changed
- **Handler error pattern** — Structured handlers return Go errors (SDK converts to tool errors); typed handlers use `mcp.NewToolResultError()` for flexible output tools
- **get_sources parallelized** — Fetches team sources concurrently instead of sequentially (N+1 fix)
- **top_values parallelized** — Fetches field values concurrently across all requested fields

### Fixed
- **jsonschema tag format** — All struct tags updated from `jsonschema:"description=X,required"` (invopop format) to `jsonschema:"X"` (google/jsonschema-go format). The old format silently produced empty input schemas.
- **URL parameter injection** — `GetFieldValues` client method now uses `url.PathEscape` and `url.Values` instead of raw string interpolation
- **Unbounded log context** — `get_log_context` before/after limits capped at 100 (was unbounded)
- **Query text leakage** — `get_query_telemetry` no longer returns the `query` column from `system.query_log`, preventing other users' query text from being exposed
- **Alert prompt args** — `investigate_alert` prompt now passes `team_id` and `source_id` to `get_alert_history` (was missing, would fail argument validation)

### Removed
- **Custom tool wrapper** — Deleted `tools.go` (228 lines) with `MustTool`/`ConvertTool`/`Tool` types and the `invopop/jsonschema` reflection-based schema generator
- **5 transitive dependencies** — `invopop/jsonschema`, `go-ordered-map/v2`, `easyjson`, `generic-list-go`, `buger/jsonparser`

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
