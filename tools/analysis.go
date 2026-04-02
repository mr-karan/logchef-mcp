package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"golang.org/x/sync/errgroup"

	mcplogchef "github.com/mr-karan/logchef-mcp"
	"github.com/mr-karan/logchef-mcp/client"
)

// Composite analysis tools that orchestrate multiple API calls.

// --- Input schemas ---

type CompareWindowsParams struct {
	TeamID    int    `json:"team_id" jsonschema:"Team ID"`
	SourceID  int    `json:"source_id" jsonschema:"Source ID"`
	Query     string `json:"query" jsonschema:"LogchefQL filter expression to run in both windows"`
	Window1Start string `json:"window1_start" jsonschema:"Start time for window 1 (YYYY-MM-DD HH:MM:SS)"`
	Window1End   string `json:"window1_end" jsonschema:"End time for window 1 (YYYY-MM-DD HH:MM:SS)"`
	Window2Start string `json:"window2_start" jsonschema:"Start time for window 2 (YYYY-MM-DD HH:MM:SS)"`
	Window2End   string `json:"window2_end" jsonschema:"End time for window 2 (YYYY-MM-DD HH:MM:SS)"`
	Limit        int    `json:"limit,omitempty" jsonschema:"Max rows per window (default 100)"`
	Timezone     string `json:"timezone,omitempty" jsonschema:"Timezone (default UTC)"`
}

type TopValuesParams struct {
	TeamID    int      `json:"team_id" jsonschema:"Team ID"`
	SourceID  int      `json:"source_id" jsonschema:"Source ID"`
	Fields    []string `json:"fields" jsonschema:"List of field names to get top values for"`
	StartTime string   `json:"start_time" jsonschema:"Start time (RFC3339)"`
	EndTime   string   `json:"end_time" jsonschema:"End time (RFC3339)"`
	Limit     int      `json:"limit,omitempty" jsonschema:"Max values per field (default 10 max 50)"`
}

// --- Output schemas ---

type CompareWindowsResult struct {
	Window1 WindowResult `json:"window1" jsonschema:"Results from time window 1"`
	Window2 WindowResult `json:"window2" jsonschema:"Results from time window 2"`
	Delta   DeltaResult  `json:"delta" jsonschema:"Difference between the two windows"`
}

type WindowResult struct {
	Start    string `json:"start" jsonschema:"Window start time"`
	End      string `json:"end" jsonschema:"Window end time"`
	RowCount int    `json:"row_count" jsonschema:"Number of matching rows"`
	QueryID  string `json:"query_id" jsonschema:"ClickHouse query ID"`
}

type DeltaResult struct {
	RowCountDiff    int     `json:"row_count_diff" jsonschema:"Row count difference (window2 - window1)"`
	RowCountPercent float64 `json:"row_count_percent" jsonschema:"Percentage change in row count"`
}

type TopValuesResult struct {
	Fields []FieldTopValues `json:"fields" jsonschema:"Top values for each requested field"`
}

type FieldTopValues struct {
	FieldName string       `json:"field_name" jsonschema:"The field name"`
	Values    []FieldValue `json:"values" jsonschema:"Top values with counts"`
}

type FieldValue struct {
	Value string `json:"value" jsonschema:"The field value"`
	Count int64  `json:"count" jsonschema:"Number of occurrences"`
}

// --- Handlers ---

func handleCompareWindows(ctx context.Context, request mcp.CallToolRequest, params CompareWindowsParams) (CompareWindowsResult, error) {
	lc := mcplogchef.LogchefClientFromContext(ctx)
	if lc == nil {
		return CompareWindowsResult{}, fmt.Errorf("logchef client not configured")
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	// Query window 1
	resp1, err := lc.QueryLogchefQL(ctx, params.TeamID, params.SourceID, client.LogchefQLQueryRequest{
		Query:     params.Query,
		Limit:     limit,
		StartTime: params.Window1Start,
		EndTime:   params.Window1End,
		Timezone:  params.Timezone,
	})
	if err != nil {
		return CompareWindowsResult{}, fmt.Errorf("window 1 query failed: %w", err)
	}

	// Query window 2
	resp2, err := lc.QueryLogchefQL(ctx, params.TeamID, params.SourceID, client.LogchefQLQueryRequest{
		Query:     params.Query,
		Limit:     limit,
		StartTime: params.Window2Start,
		EndTime:   params.Window2End,
		Timezone:  params.Timezone,
	})
	if err != nil {
		return CompareWindowsResult{}, fmt.Errorf("window 2 query failed: %w", err)
	}

	count1 := len(resp1.Data.Logs)
	count2 := len(resp2.Data.Logs)
	diff := count2 - count1
	var pct float64
	if count1 > 0 {
		pct = float64(diff) / float64(count1) * 100
	}

	return CompareWindowsResult{
		Window1: WindowResult{
			Start: params.Window1Start, End: params.Window1End,
			RowCount: count1, QueryID: resp1.Data.QueryID,
		},
		Window2: WindowResult{
			Start: params.Window2Start, End: params.Window2End,
			RowCount: count2, QueryID: resp2.Data.QueryID,
		},
		Delta: DeltaResult{
			RowCountDiff:    diff,
			RowCountPercent: pct,
		},
	}, nil
}

func handleTopValues(ctx context.Context, request mcp.CallToolRequest, params TopValuesParams) (TopValuesResult, error) {
	lc := mcplogchef.LogchefClientFromContext(ctx)
	if lc == nil {
		return TopValuesResult{}, fmt.Errorf("logchef client not configured")
	}

	if len(params.Fields) == 0 {
		return TopValuesResult{}, fmt.Errorf("at least one field is required")
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	// First get the schema to determine field types
	schema, err := lc.GetSourceSchema(ctx, params.TeamID, params.SourceID)
	if err != nil {
		return TopValuesResult{}, fmt.Errorf("get schema: %w", err)
	}

	fieldTypes := make(map[string]string, len(schema.Data))
	for _, col := range schema.Data {
		fieldTypes[col.Name] = col.Type
	}

	// Fetch field values in parallel.
	fieldResults := make([]FieldTopValues, len(params.Fields))
	g, gctx := errgroup.WithContext(ctx)

	for i, fieldName := range params.Fields {
		fieldType, ok := fieldTypes[fieldName]
		if !ok {
			fieldResults[i] = FieldTopValues{FieldName: fieldName, Values: nil}
			continue
		}

		g.Go(func() error {
			resp, err := lc.GetFieldValues(gctx, params.TeamID, params.SourceID,
				fieldName, fieldType, params.StartTime, params.EndTime, limit)
			if err != nil {
				// Non-fatal: skip this field rather than fail the whole request.
				fieldResults[i] = FieldTopValues{FieldName: fieldName, Values: nil}
				return nil
			}
			fieldResults[i] = FieldTopValues{FieldName: fieldName, Values: parseFieldValues(resp.Data)}
			return nil
		})
	}
	_ = g.Wait()

	return TopValuesResult{Fields: fieldResults}, nil
}

// parseFieldValues extracts value/count pairs from the field values API response.
func parseFieldValues(data any) []FieldValue {
	// The API returns a JSON array of objects; try to parse as generic JSON
	raw, err := json.Marshal(data)
	if err != nil {
		return nil
	}

	// Try array of objects with "value" and "count" keys
	var items []map[string]any
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil
	}

	values := make([]FieldValue, 0, len(items))
	for _, item := range items {
		v := fmt.Sprintf("%v", item["value"])
		var count int64
		switch c := item["count"].(type) {
		case float64:
			count = int64(c)
		case json.Number:
			count, _ = c.Int64()
		}
		values = append(values, FieldValue{Value: v, Count: count})
	}
	return values
}

func AddAnalysisTools(s *server.MCPServer) {
	compareWindowsTool := mcp.NewTool("compare_windows",
		mcp.WithDescription("Compare log query results across two time windows. Runs the same LogchefQL query in both windows and returns row counts with the delta. Useful for before/after analysis of deployments, incidents, or config changes."),
		mcp.WithInputSchema[CompareWindowsParams](),
		mcp.WithOutputSchema[CompareWindowsResult](),
		mcp.WithTitleAnnotation("Compare Time Windows"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(compareWindowsTool, mcp.NewStructuredToolHandler(handleCompareWindows))

	topValuesTool := mcp.NewTool("top_values",
		mcp.WithDescription("Get the top distinct values for multiple fields in one call. Fetches the schema to determine field types, then queries each field's top values. Useful for quickly exploring the dimensions of a log source."),
		mcp.WithInputSchema[TopValuesParams](),
		mcp.WithOutputSchema[TopValuesResult](),
		mcp.WithTitleAnnotation("Top Field Values"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(topValuesTool, mcp.NewStructuredToolHandler(handleTopValues))
}
