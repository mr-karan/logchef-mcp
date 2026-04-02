package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	mcplogchef "github.com/mr-karan/logchef-mcp"
	"github.com/mr-karan/logchef-mcp/client"
)

// Query telemetry tool — inspects recent query performance via system.query_log.
// Excludes the raw query text to avoid leaking other users' queries.

type GetQueryTelemetryParams struct {
	TeamID        int `json:"team_id" jsonschema:"Team ID"`
	SourceID      int `json:"source_id" jsonschema:"Source ID"`
	Limit         int `json:"limit,omitempty" jsonschema:"Max queries to return (default 20 max 50)"`
	MinDurationMs int `json:"min_duration_ms,omitempty" jsonschema:"Only show queries slower than this (milliseconds)"`
}

func handleGetQueryTelemetry(ctx context.Context, request mcp.CallToolRequest, params GetQueryTelemetryParams) (*mcp.CallToolResult, error) {
	lc := mcplogchef.LogchefClientFromContext(ctx)
	if lc == nil {
		return mcp.NewToolResultError("logchef client not configured"), nil
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	// Build a query against system.query_log.
	// Note: the query column is intentionally excluded to avoid leaking
	// other users' query text. Only performance metrics are returned.
	durationFilter := ""
	if params.MinDurationMs > 0 {
		// Safe: MinDurationMs is an int, fmt.Sprintf with %d prevents injection.
		durationFilter = fmt.Sprintf("AND query_duration_ms >= %d", params.MinDurationMs)
	}

	sql := fmt.Sprintf(`SELECT
    query_id,
    type,
    query_duration_ms,
    read_rows,
    read_bytes,
    result_rows,
    result_bytes,
    memory_usage,
    event_time
FROM system.query_log
WHERE type = 'QueryFinish'
    AND is_initial_query = 1
    AND query NOT LIKE '%%system.query_log%%'
    %s
ORDER BY event_time DESC
LIMIT %d`, durationFilter, limit)

	resp, err := lc.QueryLogs(ctx, params.TeamID, params.SourceID, client.LogQueryRequest{
		RawSQL: sql,
		Limit:  limit,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("query telemetry failed: %v", err)), nil
	}

	out, _ := json.MarshalIndent(resp.Data, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

func AddTelemetryTools(s *server.MCPServer) {
	telemetryTool := mcp.NewTool("get_query_telemetry",
		mcp.WithDescription("Get recent query performance data from ClickHouse system.query_log. Shows query duration, rows/bytes read, memory usage, and timing. Useful for identifying slow queries. Note: query text is excluded for privacy; requires system.query_log access which may not be available on all deployments."),
		mcp.WithInputSchema[GetQueryTelemetryParams](),
		mcp.WithTitleAnnotation("Get Query Telemetry"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(telemetryTool, mcp.NewTypedToolHandler(handleGetQueryTelemetry))
}
