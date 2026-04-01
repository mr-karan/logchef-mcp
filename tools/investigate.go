package tools

import (
	"context"
	"encoding/json"
	"fmt"

	mcplogchef "github.com/mr-karan/logchef-mcp"
	"github.com/mr-karan/logchef-mcp/client"

	"github.com/mark3labs/mcp-go/server"
)

// Investigation tools — field values, log context, and alerts for incident analysis.

type GetFieldValuesParams struct {
	TeamID    int    `json:"team_id" jsonschema:"description=Team ID" jsonschema_extras:"required=true"`
	SourceID  int    `json:"source_id" jsonschema:"description=Source ID" jsonschema_extras:"required=true"`
	FieldName string `json:"field_name" jsonschema:"description=Column name to get distinct values for" jsonschema_extras:"required=true"`
	FieldType string `json:"field_type" jsonschema:"description=ClickHouse column type (e.g. String LowCardinality(String))" jsonschema_extras:"required=true"`
	StartTime string `json:"start_time" jsonschema:"description=Start time (RFC3339)" jsonschema_extras:"required=true"`
	EndTime   string `json:"end_time" jsonschema:"description=End time (RFC3339)" jsonschema_extras:"required=true"`
	Limit     int    `json:"limit,omitempty" jsonschema:"description=Max values to return (default 20 max 100)"`
}

type GetLogContextParams struct {
	TeamID      int   `json:"team_id" jsonschema:"description=Team ID" jsonschema_extras:"required=true"`
	SourceID    int   `json:"source_id" jsonschema:"description=Source ID" jsonschema_extras:"required=true"`
	Timestamp   int64 `json:"timestamp" jsonschema:"description=Target timestamp in milliseconds (from a log entry)" jsonschema_extras:"required=true"`
	BeforeLimit int   `json:"before_limit,omitempty" jsonschema:"description=Number of logs before the target (default 10)"`
	AfterLimit  int   `json:"after_limit,omitempty" jsonschema:"description=Number of logs after the target (default 10)"`
}

type ListAlertsParams struct {
	TeamID   int `json:"team_id" jsonschema:"description=Team ID" jsonschema_extras:"required=true"`
	SourceID int `json:"source_id" jsonschema:"description=Source ID" jsonschema_extras:"required=true"`
}

type GetAlertHistoryParams struct {
	TeamID   int `json:"team_id" jsonschema:"description=Team ID" jsonschema_extras:"required=true"`
	SourceID int `json:"source_id" jsonschema:"description=Source ID" jsonschema_extras:"required=true"`
	AlertID  int `json:"alert_id" jsonschema:"description=Alert ID" jsonschema_extras:"required=true"`
}

func getFieldValues(ctx context.Context, params GetFieldValuesParams) (string, error) {
	lc := mcplogchef.LogchefClientFromContext(ctx)
	if lc == nil {
		return "", fmt.Errorf("logchef client not configured")
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
		return "", fmt.Errorf("get field values failed: %w", err)
	}

	out, _ := json.MarshalIndent(resp.Data, "", "  ")
	return string(out), nil
}

func getLogContext(ctx context.Context, params GetLogContextParams) (string, error) {
	lc := mcplogchef.LogchefClientFromContext(ctx)
	if lc == nil {
		return "", fmt.Errorf("logchef client not configured")
	}

	beforeLimit := params.BeforeLimit
	if beforeLimit <= 0 {
		beforeLimit = 10
	}
	afterLimit := params.AfterLimit
	if afterLimit <= 0 {
		afterLimit = 10
	}

	resp, err := lc.GetLogContext(ctx, params.TeamID, params.SourceID, client.LogContextRequest{
		Timestamp:   params.Timestamp,
		BeforeLimit: beforeLimit,
		AfterLimit:  afterLimit,
	})
	if err != nil {
		return "", fmt.Errorf("get log context failed: %w", err)
	}

	result := map[string]any{
		"target_timestamp": resp.Data.TargetTimestamp,
		"before_logs":      resp.Data.BeforeLogs,
		"target_logs":      resp.Data.TargetLogs,
		"after_logs":       resp.Data.AfterLogs,
		"stats":            resp.Data.Stats,
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}

func listAlerts(ctx context.Context, params ListAlertsParams) (string, error) {
	lc := mcplogchef.LogchefClientFromContext(ctx)
	if lc == nil {
		return "", fmt.Errorf("logchef client not configured")
	}

	resp, err := lc.ListAlerts(ctx, params.TeamID, params.SourceID)
	if err != nil {
		return "", fmt.Errorf("list alerts failed: %w", err)
	}

	out, _ := json.MarshalIndent(resp.Data, "", "  ")
	return string(out), nil
}

func getAlertHistory(ctx context.Context, params GetAlertHistoryParams) (string, error) {
	lc := mcplogchef.LogchefClientFromContext(ctx)
	if lc == nil {
		return "", fmt.Errorf("logchef client not configured")
	}

	resp, err := lc.GetAlertHistory(ctx, params.TeamID, params.SourceID, params.AlertID)
	if err != nil {
		return "", fmt.Errorf("get alert history failed: %w", err)
	}

	out, _ := json.MarshalIndent(resp.Data, "", "  ")
	return string(out), nil
}

func AddInvestigateTools(s *server.MCPServer) {
	tools := []mcplogchef.Tool{
		mcplogchef.MustTool[GetFieldValuesParams, string](
			"get_field_values",
			"Get the top distinct values for a specific field in a time range. Useful for exploring dimensions (e.g. top severity levels, service names, status codes) before writing queries.",
			getFieldValues,
		),
		mcplogchef.MustTool[GetLogContextParams, string](
			"get_log_context",
			"Get surrounding log entries (before and after) a specific timestamp. Useful for investigating what happened around a particular event.",
			getLogContext,
		),
		mcplogchef.MustTool[ListAlertsParams, string](
			"list_alerts",
			"List all alert rules configured for a source. Shows name, severity, active status, query mode, and last state (firing/resolved).",
			listAlerts,
		),
		mcplogchef.MustTool[GetAlertHistoryParams, string](
			"get_alert_history",
			"Get the evaluation history for a specific alert. Shows when it triggered, resolved, or errored, with the actual metric values.",
			getAlertHistory,
		),
	}

	for _, t := range tools {
		s.AddTool(t.Tool, t.Handler)
	}
}
