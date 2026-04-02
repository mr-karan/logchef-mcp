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

// LogchefQL tools — query, translate, validate using Logchef's native search syntax.

type QueryLogchefQLParams struct {
	TeamID       int    `json:"team_id" jsonschema:"Team ID"`
	SourceID     int    `json:"source_id" jsonschema:"Source ID"`
	Query        string `json:"query" jsonschema:"LogchefQL filter expression (e.g. severity_text=ERROR and service=api). Empty string returns all logs."`
	StartTime    string `json:"start_time" jsonschema:"Start time in YYYY-MM-DD HH:MM:SS format"`
	EndTime      string `json:"end_time" jsonschema:"End time in YYYY-MM-DD HH:MM:SS format"`
	Limit        int    `json:"limit,omitempty" jsonschema:"Max rows to return (1-500 default 100)"`
	Timezone     string `json:"timezone,omitempty" jsonschema:"Timezone (default UTC)"`
	QueryTimeout *int   `json:"query_timeout,omitempty" jsonschema:"Query timeout in seconds (default 60)"`
}

type TranslateLogchefQLParams struct {
	TeamID    int    `json:"team_id" jsonschema:"Team ID"`
	SourceID  int    `json:"source_id" jsonschema:"Source ID"`
	Query     string `json:"query" jsonschema:"LogchefQL filter expression to translate to SQL"`
	StartTime string `json:"start_time" jsonschema:"Start time in YYYY-MM-DD HH:MM:SS format"`
	EndTime   string `json:"end_time" jsonschema:"End time in YYYY-MM-DD HH:MM:SS format"`
	Limit     int    `json:"limit,omitempty" jsonschema:"Row limit for generated SQL"`
	Timezone  string `json:"timezone,omitempty" jsonschema:"Timezone (default UTC)"`
}

// --- Output schemas ---

type TranslateResult struct {
	SQL   string `json:"sql" jsonschema:"Generated ClickHouse SQL query"`
	Valid bool   `json:"valid" jsonschema:"Whether the LogchefQL expression is valid"`
}

// query_logchefql returns flexible log data, so it uses NewTypedToolHandler
// with manual JSON serialization instead of structured output.
func handleQueryLogchefQL(ctx context.Context, request mcp.CallToolRequest, params QueryLogchefQLParams) (*mcp.CallToolResult, error) {
	lc := mcplogchef.LogchefClientFromContext(ctx)
	if lc == nil {
		return mcp.NewToolResultError("logchef client not configured"), nil
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
		return mcp.NewToolResultError(fmt.Sprintf("logchefql query failed: %v", err)), nil
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
	return mcp.NewToolResultText(string(out)), nil
}

func handleTranslateLogchefQL(ctx context.Context, request mcp.CallToolRequest, params TranslateLogchefQLParams) (TranslateResult, error) {
	lc := mcplogchef.LogchefClientFromContext(ctx)
	if lc == nil {
		return TranslateResult{}, fmt.Errorf("logchef client not configured")
	}

	resp, err := lc.TranslateLogchefQL(ctx, params.TeamID, params.SourceID, client.LogchefQLTranslateRequest{
		Query:     params.Query,
		StartTime: params.StartTime,
		EndTime:   params.EndTime,
		Timezone:  params.Timezone,
		Limit:     params.Limit,
	})
	if err != nil {
		return TranslateResult{}, fmt.Errorf("logchefql translate failed: %w", err)
	}

	return TranslateResult{
		SQL:   resp.Data.SQL,
		Valid: resp.Data.Valid,
	}, nil
}

type ValidateLogchefQLParams struct {
	TeamID   int    `json:"team_id" jsonschema:"Team ID"`
	SourceID int    `json:"source_id" jsonschema:"Source ID"`
	Query    string `json:"query" jsonschema:"LogchefQL expression to validate"`
}

type ValidateResult struct {
	Valid bool   `json:"valid" jsonschema:"Whether the LogchefQL expression is syntactically valid"`
	Error string `json:"error,omitempty" jsonschema:"Syntax error description if invalid"`
}

func handleValidateLogchefQL(ctx context.Context, request mcp.CallToolRequest, params ValidateLogchefQLParams) (ValidateResult, error) {
	lc := mcplogchef.LogchefClientFromContext(ctx)
	if lc == nil {
		return ValidateResult{}, fmt.Errorf("logchef client not configured")
	}

	resp, err := lc.ValidateLogchefQL(ctx, params.TeamID, params.SourceID, client.LogchefQLValidateRequest{
		Query: params.Query,
	})
	if err != nil {
		return ValidateResult{}, fmt.Errorf("logchefql validate failed: %w", err)
	}

	return ValidateResult{
		Valid: resp.Data.Valid,
		Error: resp.Data.Error,
	}, nil
}

func AddLogchefQLTools(s *server.MCPServer) {
	// query_logchefql returns flexible log data — uses typed handler
	queryTool := mcp.NewTool("query_logchefql",
		mcp.WithDescription("Execute a LogchefQL query against a log source. LogchefQL is a simple filter syntax (e.g. 'severity_text=ERROR and service=api'). Time range is specified separately. Returns logs, columns, stats, and the generated SQL."),
		mcp.WithInputSchema[QueryLogchefQLParams](),
		mcp.WithTitleAnnotation("Query LogchefQL"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(queryTool, mcp.NewTypedToolHandler(handleQueryLogchefQL))

	// translate_logchefql has a fixed output shape — uses structured handler
	translateTool := mcp.NewTool("translate_logchefql",
		mcp.WithDescription("Translate a LogchefQL expression to ClickHouse SQL without executing it. Useful for understanding what SQL will be generated."),
		mcp.WithInputSchema[TranslateLogchefQLParams](),
		mcp.WithOutputSchema[TranslateResult](),
		mcp.WithTitleAnnotation("Translate LogchefQL to SQL"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(translateTool, mcp.NewStructuredToolHandler(handleTranslateLogchefQL))

	// validate_logchefql checks syntax without executing
	validateTool := mcp.NewTool("validate_logchefql",
		mcp.WithDescription("Validate a LogchefQL expression for syntax errors without executing it. Returns whether the expression is valid and any error details. Use this before executing queries to catch syntax issues early."),
		mcp.WithInputSchema[ValidateLogchefQLParams](),
		mcp.WithOutputSchema[ValidateResult](),
		mcp.WithTitleAnnotation("Validate LogchefQL"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(validateTool, mcp.NewStructuredToolHandler(handleValidateLogchefQL))
}
