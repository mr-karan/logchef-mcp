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

// Investigation tools — field values, log context, and alerts for incident analysis.

type GetFieldValuesParams struct {
	TeamID    int    `json:"team_id" jsonschema:"Team ID"`
	SourceID  int    `json:"source_id" jsonschema:"Source ID"`
	FieldName string `json:"field_name" jsonschema:"Column name to get distinct values for"`
	FieldType string `json:"field_type" jsonschema:"ClickHouse column type (e.g. String or LowCardinality(String))"`
	StartTime string `json:"start_time" jsonschema:"Start time (RFC3339)"`
	EndTime   string `json:"end_time" jsonschema:"End time (RFC3339)"`
	Limit     int    `json:"limit,omitempty" jsonschema:"Max values to return (default 20 max 100)"`
}

type GetLogContextParams struct {
	TeamID      int   `json:"team_id" jsonschema:"Team ID"`
	SourceID    int   `json:"source_id" jsonschema:"Source ID"`
	Timestamp   int64 `json:"timestamp" jsonschema:"Target timestamp in milliseconds (from a log entry)"`
	BeforeLimit int   `json:"before_limit,omitempty" jsonschema:"Number of logs before the target (default 10)"`
	AfterLimit  int   `json:"after_limit,omitempty" jsonschema:"Number of logs after the target (default 10)"`
}

type ListAlertsParams struct {
	TeamID   int `json:"team_id" jsonschema:"Team ID"`
	SourceID int `json:"source_id" jsonschema:"Source ID"`
}

type GetAlertHistoryParams struct {
	TeamID   int `json:"team_id" jsonschema:"Team ID"`
	SourceID int `json:"source_id" jsonschema:"Source ID"`
	AlertID  int `json:"alert_id" jsonschema:"Alert ID"`
}

// --- Handlers ---

// get_field_values returns dynamic field data — typed handler
func handleGetFieldValues(ctx context.Context, request mcp.CallToolRequest, params GetFieldValuesParams) (*mcp.CallToolResult, error) {
	lc := mcplogchef.LogchefClientFromContext(ctx)
	if lc == nil {
		return mcp.NewToolResultError("logchef client not configured"), nil
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	resp, err := lc.GetFieldValues(ctx, params.TeamID, params.SourceID,
		params.FieldName, params.FieldType, params.StartTime, params.EndTime, limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("get field values failed: %v", err)), nil
	}

	out, _ := json.MarshalIndent(resp.Data, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

// get_log_context returns flexible log data — typed handler
func handleGetLogContext(ctx context.Context, request mcp.CallToolRequest, params GetLogContextParams) (*mcp.CallToolResult, error) {
	lc := mcplogchef.LogchefClientFromContext(ctx)
	if lc == nil {
		return mcp.NewToolResultError("logchef client not configured"), nil
	}

	beforeLimit := params.BeforeLimit
	if beforeLimit <= 0 {
		beforeLimit = 10
	}
	if beforeLimit > 100 {
		beforeLimit = 100
	}
	afterLimit := params.AfterLimit
	if afterLimit <= 0 {
		afterLimit = 10
	}
	if afterLimit > 100 {
		afterLimit = 100
	}

	resp, err := lc.GetLogContext(ctx, params.TeamID, params.SourceID, client.LogContextRequest{
		Timestamp:   params.Timestamp,
		BeforeLimit: beforeLimit,
		AfterLimit:  afterLimit,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("get log context failed: %v", err)), nil
	}

	result := map[string]any{
		"target_timestamp": resp.Data.TargetTimestamp,
		"before_logs":      resp.Data.BeforeLogs,
		"target_logs":      resp.Data.TargetLogs,
		"after_logs":       resp.Data.AfterLogs,
		"stats":            resp.Data.Stats,
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

// list_alerts and get_alert_history return API JSON directly — typed handlers
func handleListAlerts(ctx context.Context, request mcp.CallToolRequest, params ListAlertsParams) (*mcp.CallToolResult, error) {
	lc := mcplogchef.LogchefClientFromContext(ctx)
	if lc == nil {
		return mcp.NewToolResultError("logchef client not configured"), nil
	}

	resp, err := lc.ListAlerts(ctx, params.TeamID, params.SourceID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("list alerts failed: %v", err)), nil
	}

	out, _ := json.MarshalIndent(resp.Data, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

func handleGetAlertHistory(ctx context.Context, request mcp.CallToolRequest, params GetAlertHistoryParams) (*mcp.CallToolResult, error) {
	lc := mcplogchef.LogchefClientFromContext(ctx)
	if lc == nil {
		return mcp.NewToolResultError("logchef client not configured"), nil
	}

	resp, err := lc.GetAlertHistory(ctx, params.TeamID, params.SourceID, params.AlertID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("get alert history failed: %v", err)), nil
	}

	out, _ := json.MarshalIndent(resp.Data, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

func AddInvestigateTools(s *server.MCPServer) {
	fieldValuesTool := mcp.NewTool("get_field_values",
		mcp.WithDescription("Get the top distinct values for a specific field in a time range. Useful for exploring dimensions (e.g. top severity levels, service names, status codes) before writing queries."),
		mcp.WithInputSchema[GetFieldValuesParams](),
		mcp.WithTitleAnnotation("Get Field Values"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(fieldValuesTool, mcp.NewTypedToolHandler(handleGetFieldValues))

	logContextTool := mcp.NewTool("get_log_context",
		mcp.WithDescription("Get surrounding log entries (before and after) a specific timestamp. Useful for investigating what happened around a particular event."),
		mcp.WithInputSchema[GetLogContextParams](),
		mcp.WithTitleAnnotation("Get Log Context"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(logContextTool, mcp.NewTypedToolHandler(handleGetLogContext))

	listAlertsTool := mcp.NewTool("list_alerts",
		mcp.WithDescription("List all alert rules configured for a source. Shows name, severity, active status, query mode, and last state (firing/resolved)."),
		mcp.WithInputSchema[ListAlertsParams](),
		mcp.WithTitleAnnotation("List Alerts"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(listAlertsTool, mcp.NewTypedToolHandler(handleListAlerts))

	alertHistoryTool := mcp.NewTool("get_alert_history",
		mcp.WithDescription("Get the evaluation history for a specific alert. Shows when it triggered, resolved, or errored, with the actual metric values."),
		mcp.WithInputSchema[GetAlertHistoryParams](),
		mcp.WithTitleAnnotation("Get Alert History"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(alertHistoryTool, mcp.NewTypedToolHandler(handleGetAlertHistory))
}
