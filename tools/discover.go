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

// Discovery tools — natural language query generation and bulk field exploration.

// --- Input schemas ---

type GenerateQueryParams struct {
	TeamID   int    `json:"team_id" jsonschema:"Team ID"`
	SourceID int    `json:"source_id" jsonschema:"Source ID"`
	Query    string `json:"query" jsonschema:"Natural language description of what you want to find (e.g. 'show me errors from api service in last hour')"`
}

type GetAllFieldDimensionsParams struct {
	TeamID    int    `json:"team_id" jsonschema:"Team ID"`
	SourceID  int    `json:"source_id" jsonschema:"Source ID"`
	StartTime string `json:"start_time" jsonschema:"Start time (RFC3339)"`
	EndTime   string `json:"end_time" jsonschema:"End time (RFC3339)"`
	Timezone  string `json:"timezone,omitempty" jsonschema:"Timezone (default UTC)"`
	Limit     int    `json:"limit,omitempty" jsonschema:"Max values per field (default 10 max 100)"`
}

// --- Output schemas ---

type GenerateQueryResult struct {
	SQL string `json:"sql" jsonschema:"Generated ClickHouse SQL query"`
}

// --- Handlers ---

func handleGenerateQuery(ctx context.Context, request mcp.CallToolRequest, params GenerateQueryParams) (GenerateQueryResult, error) {
	lc := mcplogchef.LogchefClientFromContext(ctx)
	if lc == nil {
		return GenerateQueryResult{}, fmt.Errorf("logchef client not configured")
	}

	resp, err := lc.GenerateAISQL(ctx, params.TeamID, params.SourceID, client.GenerateSQLRequest{
		NaturalLanguageQuery: params.Query,
	})
	if err != nil {
		return GenerateQueryResult{}, fmt.Errorf("generate query failed: %w", err)
	}

	return GenerateQueryResult{
		SQL: resp.Data.SQLQuery,
	}, nil
}

// get_all_field_dimensions returns dynamic data — typed handler
func handleGetAllFieldDimensions(ctx context.Context, request mcp.CallToolRequest, params GetAllFieldDimensionsParams) (*mcp.CallToolResult, error) {
	lc := mcplogchef.LogchefClientFromContext(ctx)
	if lc == nil {
		return mcp.NewToolResultError("logchef client not configured"), nil
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	timezone := params.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	resp, err := lc.GetAllFieldValues(ctx, params.TeamID, params.SourceID,
		params.StartTime, params.EndTime, timezone, limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("get all field dimensions failed: %v", err)), nil
	}

	out, _ := json.MarshalIndent(resp.Data, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

func AddDiscoverTools(s *server.MCPServer) {
	generateTool := mcp.NewTool("generate_query",
		mcp.WithDescription("Generate a ClickHouse SQL query from a natural language description. Uses AI to translate your intent into a query based on the source's schema. Requires AI to be enabled on the Logchef instance. Example: 'show me 500 errors from the api service in the last hour'."),
		mcp.WithInputSchema[GenerateQueryParams](),
		mcp.WithOutputSchema[GenerateQueryResult](),
		mcp.WithTitleAnnotation("Generate Query from Natural Language"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(generateTool, mcp.NewStructuredToolHandler(handleGenerateQuery))

	dimensionsTool := mcp.NewTool("get_all_field_dimensions",
		mcp.WithDescription("Get top values for all LowCardinality fields in one call. Returns a map of field names to their top values with counts. Much faster than calling get_field_values for each field individually. Use this for initial source exploration to understand what dimensions exist."),
		mcp.WithInputSchema[GetAllFieldDimensionsParams](),
		mcp.WithTitleAnnotation("Get All Field Dimensions"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(dimensionsTool, mcp.NewTypedToolHandler(handleGetAllFieldDimensions))
}
