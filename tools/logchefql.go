package tools

import (
	"context"
	"encoding/json"
	"fmt"

	mcplogchef "github.com/mr-karan/logchef-mcp"
	"github.com/mr-karan/logchef-mcp/client"

	"github.com/mark3labs/mcp-go/server"
)

// LogchefQL tools — query, translate, validate using Logchef's native search syntax.

type QueryLogchefQLParams struct {
	TeamID       int    `json:"team_id" jsonschema:"description=Team ID" jsonschema_extras:"required=true"`
	SourceID     int    `json:"source_id" jsonschema:"description=Source ID" jsonschema_extras:"required=true"`
	Query        string `json:"query" jsonschema:"description=LogchefQL filter expression (e.g. severity_text=ERROR and service=api). Empty string returns all logs." jsonschema_extras:"required=true"`
	StartTime    string `json:"start_time" jsonschema:"description=Start time in YYYY-MM-DD HH:MM:SS format" jsonschema_extras:"required=true"`
	EndTime      string `json:"end_time" jsonschema:"description=End time in YYYY-MM-DD HH:MM:SS format" jsonschema_extras:"required=true"`
	Limit        int    `json:"limit,omitempty" jsonschema:"description=Max rows to return (1-500 default 100)"`
	Timezone     string `json:"timezone,omitempty" jsonschema:"description=Timezone (default UTC)"`
	QueryTimeout *int   `json:"query_timeout,omitempty" jsonschema:"description=Query timeout in seconds (default 60)"`
}

type TranslateLogchefQLParams struct {
	TeamID    int    `json:"team_id" jsonschema:"description=Team ID" jsonschema_extras:"required=true"`
	SourceID  int    `json:"source_id" jsonschema:"description=Source ID" jsonschema_extras:"required=true"`
	Query     string `json:"query" jsonschema:"description=LogchefQL filter expression to translate to SQL" jsonschema_extras:"required=true"`
	StartTime string `json:"start_time" jsonschema:"description=Start time in YYYY-MM-DD HH:MM:SS format" jsonschema_extras:"required=true"`
	EndTime   string `json:"end_time" jsonschema:"description=End time in YYYY-MM-DD HH:MM:SS format" jsonschema_extras:"required=true"`
	Limit     int    `json:"limit,omitempty" jsonschema:"description=Row limit for generated SQL"`
	Timezone  string `json:"timezone,omitempty" jsonschema:"description=Timezone (default UTC)"`
}

func queryLogchefQL(ctx context.Context, params QueryLogchefQLParams) (string, error) {
	lc := mcplogchef.LogchefClientFromContext(ctx)
	if lc == nil {
		return "", fmt.Errorf("logchef client not configured")
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	resp, err := lc.QueryLogchefQL(ctx, params.TeamID, params.SourceID, client.LogchefQLQueryRequest{
		Query:        params.Query,
		Limit:        limit,
		StartTime:    params.StartTime,
		EndTime:      params.EndTime,
		Timezone:     params.Timezone,
		QueryTimeout: params.QueryTimeout,
	})
	if err != nil {
		return "", fmt.Errorf("logchefql query failed: %w", err)
	}

	result := map[string]any{
		"logs":          resp.Data.Logs,
		"columns":       resp.Data.Columns,
		"stats":         resp.Data.Stats,
		"query_id":      resp.Data.QueryID,
		"generated_sql": resp.Data.GeneratedSQL,
		"row_count":     len(resp.Data.Logs),
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}

func translateLogchefQL(ctx context.Context, params TranslateLogchefQLParams) (string, error) {
	lc := mcplogchef.LogchefClientFromContext(ctx)
	if lc == nil {
		return "", fmt.Errorf("logchef client not configured")
	}

	resp, err := lc.TranslateLogchefQL(ctx, params.TeamID, params.SourceID, client.LogchefQLTranslateRequest{
		Query:     params.Query,
		StartTime: params.StartTime,
		EndTime:   params.EndTime,
		Timezone:  params.Timezone,
		Limit:     params.Limit,
	})
	if err != nil {
		return "", fmt.Errorf("logchefql translate failed: %w", err)
	}

	result := map[string]any{
		"sql":   resp.Data.SQL,
		"valid": resp.Data.Valid,
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}

func AddLogchefQLTools(s *server.MCPServer) {
	tools := []mcplogchef.Tool{
		mcplogchef.MustTool[QueryLogchefQLParams, string](
			"query_logchefql",
			"Execute a LogchefQL query against a log source. LogchefQL is a simple filter syntax (e.g. 'severity_text=ERROR and service=api'). Time range is specified separately. Returns logs, columns, stats, and the generated SQL.",
			queryLogchefQL,
		),
		mcplogchef.MustTool[TranslateLogchefQLParams, string](
			"translate_logchefql",
			"Translate a LogchefQL expression to ClickHouse SQL without executing it. Useful for understanding what SQL will be generated.",
			translateLogchefQL,
		),
	}

	for _, t := range tools {
		s.AddTool(t.Tool, t.Handler)
	}
}
