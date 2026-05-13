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

// --- Input schemas ---

type QueryLogsParams struct {
	TeamID       int    `json:"team_id" jsonschema:"The ID of the team that has access to the source"`
	SourceID     int    `json:"source_id" jsonschema:"The ID of the source to query logs from"`
	RawSQL       string `json:"raw_sql" jsonschema:"The ClickHouse SQL query to execute. Use get_source_schema first to understand available columns. Include WHERE clauses with timestamp filters and ORDER BY and LIMIT clauses."`
	Limit        int    `json:"limit,omitempty" jsonschema:"Maximum number of log entries to return (1-100 default 100)"`
	QueryTimeout *int   `json:"query_timeout,omitempty" jsonschema:"Query timeout in seconds (default 30)"`
}

type GetSourceSchemaParams struct {
	TeamID   int `json:"team_id" jsonschema:"The ID of the team that has access to the source"`
	SourceID int `json:"source_id" jsonschema:"The ID of the source to get the schema for"`
}

type GetLogHistogramParams struct {
	TeamID       int    `json:"team_id" jsonschema:"The ID of the team that has access to the source"`
	SourceID     int    `json:"source_id" jsonschema:"The ID of the source to generate histogram for"`
	RawSQL       string `json:"raw_sql" jsonschema:"The ClickHouse SQL query to analyze with proper WHERE clauses and timestamp filters"`
	Window       string `json:"window,omitempty" jsonschema:"Time window for histogram buckets (e.g. 1m 5m 1h 1d). Defaults to 1m."`
	GroupBy      string `json:"group_by,omitempty" jsonschema:"Optional field to group histogram data by (e.g. severity_text or service_name)"`
	Timezone     string `json:"timezone,omitempty" jsonschema:"Timezone for histogram timestamps (default UTC)"`
	QueryTimeout *int   `json:"query_timeout,omitempty" jsonschema:"Query timeout in seconds (default 30)"`
}

type ListSavedQueriesParams struct {
	SourceID int `json:"source_id,omitempty" jsonschema:"Optional source ID to filter by. Omit to list every saved query visible to the caller."`
}

type CreateSavedQueryParams struct {
	SourceID     int    `json:"source_id" jsonschema:"The ID of the source the saved query runs against"`
	Name         string `json:"name" jsonschema:"Name of the saved query"`
	Description  string `json:"description,omitempty" jsonschema:"Optional description of the saved query"`
	QueryType    string `json:"query_type" jsonschema:"Either 'logchefql' or 'sql'"`
	QueryContent string `json:"query_content" jsonschema:"The query payload. JSON envelope containing version, sourceId, timeRange, limit, and content."`
}

type GetSavedQueryParams struct {
	QueryID int `json:"query_id" jsonschema:"The ID of the saved query to retrieve"`
}

type UpdateSavedQueryParams struct {
	QueryID      int    `json:"query_id" jsonschema:"The ID of the saved query to update"`
	Name         string `json:"name" jsonschema:"Name of the saved query"`
	Description  string `json:"description,omitempty" jsonschema:"Optional description of the saved query"`
	QueryType    string `json:"query_type" jsonschema:"Either 'logchefql' or 'sql'"`
	QueryContent string `json:"query_content" jsonschema:"The query payload. JSON envelope containing version, sourceId, timeRange, limit, and content."`
}

type DeleteSavedQueryParams struct {
	QueryID int `json:"query_id" jsonschema:"The ID of the saved query to delete"`
}

// --- Output schemas ---

type SchemaColumnResult struct {
	Name string `json:"name" jsonschema:"Column name"`
	Type string `json:"type" jsonschema:"ClickHouse column type"`
}

type SavedQueryResult struct {
	ID                int    `json:"id" jsonschema:"Saved query ID"`
	Name              string `json:"name" jsonschema:"Saved query name"`
	Description       string `json:"description" jsonschema:"Saved query description"`
	SourceID          int    `json:"source_id" jsonschema:"ID of the source the query runs against"`
	CreatedFromTeamID *int   `json:"created_from_team_id,omitempty" jsonschema:"Team the query was originally saved from (resolver preference hint, not an ACL)"`
	QueryType         string `json:"query_type" jsonschema:"Either 'logchefql' or 'sql'"`
	QueryContent      string `json:"query_content" jsonschema:"Query payload (JSON envelope)"`
	CreatedBy         *int   `json:"created_by,omitempty" jsonschema:"Creator user ID; null on legacy queries"`
	SourceName        string `json:"source_name,omitempty" jsonschema:"Human-readable source name (when included)"`
	CreatedAt         string `json:"created_at" jsonschema:"Creation timestamp"`
	UpdatedAt         string `json:"updated_at" jsonschema:"Last update timestamp"`
}

type SuccessResult struct {
	Success bool   `json:"success" jsonschema:"Whether the operation succeeded"`
	Message string `json:"message" jsonschema:"Human-readable result message"`
}

// --- Handlers ---

// query_logs returns flexible log data — uses typed handler
func handleQueryLogs(ctx context.Context, request mcp.CallToolRequest, args QueryLogsParams) (*mcp.CallToolResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return mcp.NewToolResultError("logchef client not configured"), nil
	}

	if args.Limit < 0 {
		args.Limit = 0
	}
	if args.Limit > 100 {
		args.Limit = 100
	}

	logs, err := c.QueryLogs(ctx, args.TeamID, args.SourceID, client.LogQueryRequest{
		RawSQL:       args.RawSQL,
		Limit:        args.Limit,
		QueryTimeout: args.QueryTimeout,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("query logs: %v", err)), nil
	}

	out, _ := json.MarshalIndent(logs.Data, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

func handleGetSourceSchema(ctx context.Context, request mcp.CallToolRequest, args GetSourceSchemaParams) ([]SchemaColumnResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("logchef client not configured")
	}

	schema, err := c.GetSourceSchema(ctx, args.TeamID, args.SourceID)
	if err != nil {
		return nil, fmt.Errorf("get source schema: %w", err)
	}

	result := make([]SchemaColumnResult, len(schema.Data))
	for i, col := range schema.Data {
		result[i] = SchemaColumnResult{Name: col.Name, Type: col.Type}
	}
	return result, nil
}

// get_log_histogram returns flexible histogram data — uses typed handler
func handleGetLogHistogram(ctx context.Context, request mcp.CallToolRequest, args GetLogHistogramParams) (*mcp.CallToolResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return mcp.NewToolResultError("logchef client not configured"), nil
	}

	histogram, err := c.GetLogHistogram(ctx, args.TeamID, args.SourceID, client.HistogramRequest{
		RawSQL:       args.RawSQL,
		Window:       args.Window,
		GroupBy:      args.GroupBy,
		Timezone:     args.Timezone,
		QueryTimeout: args.QueryTimeout,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("get log histogram: %v", err)), nil
	}

	out, _ := json.MarshalIndent(histogram.Data, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

func handleListSavedQueries(ctx context.Context, request mcp.CallToolRequest, args ListSavedQueriesParams) ([]SavedQueryResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("logchef client not configured")
	}

	queries, err := c.ListSavedQueries(ctx, args.SourceID)
	if err != nil {
		return nil, fmt.Errorf("list saved queries: %w", err)
	}

	result := make([]SavedQueryResult, len(queries.Data))
	for i, q := range queries.Data {
		result[i] = savedQueryToResult(q)
	}
	return result, nil
}

func handleCreateSavedQuery(ctx context.Context, request mcp.CallToolRequest, args CreateSavedQueryParams) (SavedQueryResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return SavedQueryResult{}, fmt.Errorf("logchef client not configured")
	}

	created, err := c.CreateSavedQuery(ctx, client.CreateSavedQueryRequest{
		Name:         args.Name,
		Description:  args.Description,
		SourceID:     args.SourceID,
		QueryType:    args.QueryType,
		QueryContent: args.QueryContent,
	})
	if err != nil {
		return SavedQueryResult{}, fmt.Errorf("create saved query: %w", err)
	}

	return savedQueryToResult(created.Data), nil
}

func handleGetSavedQuery(ctx context.Context, request mcp.CallToolRequest, args GetSavedQueryParams) (SavedQueryResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return SavedQueryResult{}, fmt.Errorf("logchef client not configured")
	}

	got, err := c.GetSavedQuery(ctx, args.QueryID)
	if err != nil {
		return SavedQueryResult{}, fmt.Errorf("get saved query: %w", err)
	}

	return savedQueryToResult(got.Data), nil
}

func handleUpdateSavedQuery(ctx context.Context, request mcp.CallToolRequest, args UpdateSavedQueryParams) (SavedQueryResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return SavedQueryResult{}, fmt.Errorf("logchef client not configured")
	}

	updated, err := c.UpdateSavedQuery(ctx, args.QueryID, client.UpdateSavedQueryRequest{
		Name:         args.Name,
		Description:  args.Description,
		QueryType:    args.QueryType,
		QueryContent: args.QueryContent,
	})
	if err != nil {
		return SavedQueryResult{}, fmt.Errorf("update saved query: %w", err)
	}

	return savedQueryToResult(updated.Data), nil
}

func handleDeleteSavedQuery(ctx context.Context, request mcp.CallToolRequest, args DeleteSavedQueryParams) (SuccessResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return SuccessResult{}, fmt.Errorf("logchef client not configured")
	}

	if err := c.DeleteSavedQuery(ctx, args.QueryID); err != nil {
		return SuccessResult{}, fmt.Errorf("delete saved query: %w", err)
	}

	return SuccessResult{Success: true, Message: "Saved query deleted successfully"}, nil
}

func savedQueryToResult(q client.SavedQuery) SavedQueryResult {
	return SavedQueryResult{
		ID:                q.ID,
		Name:              q.Name,
		Description:       q.Description,
		SourceID:          q.SourceID,
		CreatedFromTeamID: q.CreatedFromTeamID,
		QueryType:         q.QueryType,
		QueryContent:      q.QueryContent,
		CreatedBy:         q.CreatedBy,
		SourceName:        q.SourceName,
		CreatedAt:         q.CreatedAt,
		UpdatedAt:         q.UpdatedAt,
	}
}

func AddLogsTools(s *server.MCPServer) {
	// query_logs returns flexible log data — typed handler
	queryLogsTool := mcp.NewTool("query_logs",
		mcp.WithDescription("Execute a ClickHouse SQL query against a specific log source within a team. Use get_source_schema first to understand available columns. The query should include proper WHERE clauses with timestamp filters, ORDER BY, and LIMIT. Maximum 100 results per query."),
		mcp.WithInputSchema[QueryLogsParams](),
		mcp.WithTitleAnnotation("Query Logs (SQL)"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(queryLogsTool, mcp.NewTypedToolHandler(handleQueryLogs))

	schemaTool := mcp.NewTool("get_source_schema",
		mcp.WithDescription("Get the ClickHouse table schema (column names and types) for a specific log source within a team. Use this before querying logs to understand what fields are available."),
		mcp.WithInputSchema[GetSourceSchemaParams](),
		mcp.WithOutputSchema[[]SchemaColumnResult](),
		mcp.WithTitleAnnotation("Get Source Schema"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(schemaTool, mcp.NewStructuredToolHandler(handleGetSourceSchema))

	// get_log_histogram returns flexible histogram data — typed handler
	histogramTool := mcp.NewTool("get_log_histogram",
		mcp.WithDescription("Generate time-based histogram data for log analysis. Creates a time series showing log volume over specified time windows, with optional grouping by fields like severity or service. Useful for identifying traffic patterns, spikes, and trends."),
		mcp.WithInputSchema[GetLogHistogramParams](),
		mcp.WithTitleAnnotation("Get Log Histogram"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(histogramTool, mcp.NewTypedToolHandler(handleGetLogHistogram))

	listSavedQueriesTool := mcp.NewTool("list_saved_queries",
		mcp.WithDescription("List saved queries visible to the caller. Optionally filter by source_id. Saved queries are reusable LogchefQL or SQL queries pinned to a specific source; any user with source access via any team can see them."),
		mcp.WithInputSchema[ListSavedQueriesParams](),
		mcp.WithOutputSchema[[]SavedQueryResult](),
		mcp.WithTitleAnnotation("List Saved Queries"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(listSavedQueriesTool, mcp.NewStructuredToolHandler(handleListSavedQueries))

	createSavedQueryTool := mcp.NewTool("create_saved_query",
		mcp.WithDescription("Create a new saved query bound to a source. Provide source_id, name, query_type ('logchefql' or 'sql'), and query_content (a JSON envelope with version, sourceId, timeRange, limit, and the query text)."),
		mcp.WithInputSchema[CreateSavedQueryParams](),
		mcp.WithOutputSchema[SavedQueryResult](),
		mcp.WithTitleAnnotation("Create Saved Query"),
		mcp.WithDestructiveHintAnnotation(false),
	)
	s.AddTool(createSavedQueryTool, mcp.NewStructuredToolHandler(handleCreateSavedQuery))

	getSavedQueryTool := mcp.NewTool("get_saved_query",
		mcp.WithDescription("Get a single saved query by ID. Returns name, description, query_type, query_content, source, and metadata."),
		mcp.WithInputSchema[GetSavedQueryParams](),
		mcp.WithOutputSchema[SavedQueryResult](),
		mcp.WithTitleAnnotation("Get Saved Query"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(getSavedQueryTool, mcp.NewStructuredToolHandler(handleGetSavedQuery))

	updateSavedQueryTool := mcp.NewTool("update_saved_query",
		mcp.WithDescription("Update an existing saved query. Only the creator (or a global admin) can update. The source cannot be changed — create a new query if you need to retarget."),
		mcp.WithInputSchema[UpdateSavedQueryParams](),
		mcp.WithOutputSchema[SavedQueryResult](),
		mcp.WithTitleAnnotation("Update Saved Query"),
		mcp.WithDestructiveHintAnnotation(false),
	)
	s.AddTool(updateSavedQueryTool, mcp.NewStructuredToolHandler(handleUpdateSavedQuery))

	deleteSavedQueryTool := mcp.NewTool("delete_saved_query",
		mcp.WithDescription("Delete a saved query by ID. Only the creator (or a global admin) can delete. Permanent."),
		mcp.WithInputSchema[DeleteSavedQueryParams](),
		mcp.WithOutputSchema[SuccessResult](),
		mcp.WithTitleAnnotation("Delete Saved Query"),
		mcp.WithDestructiveHintAnnotation(true),
	)
	s.AddTool(deleteSavedQueryTool, mcp.NewStructuredToolHandler(handleDeleteSavedQuery))
}
