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

type GetCollectionsParams struct {
	TeamID   int `json:"team_id" jsonschema:"The ID of the team that has access to the source"`
	SourceID int `json:"source_id" jsonschema:"The ID of the source to get collections for"`
}

type CreateCollectionParams struct {
	TeamID      int    `json:"team_id" jsonschema:"The ID of the team that has access to the source"`
	SourceID    int    `json:"source_id" jsonschema:"The ID of the source to create the collection for"`
	Name        string `json:"name" jsonschema:"Name of the collection"`
	Description string `json:"description,omitempty" jsonschema:"Optional description of the collection"`
	Query       string `json:"query" jsonschema:"The ClickHouse SQL query to save in the collection"`
}

type GetCollectionParams struct {
	TeamID       int `json:"team_id" jsonschema:"The ID of the team that has access to the source"`
	SourceID     int `json:"source_id" jsonschema:"The ID of the source that contains the collection"`
	CollectionID int `json:"collection_id" jsonschema:"The ID of the collection to retrieve"`
}

type UpdateCollectionParams struct {
	TeamID       int    `json:"team_id" jsonschema:"The ID of the team that has access to the source"`
	SourceID     int    `json:"source_id" jsonschema:"The ID of the source that contains the collection"`
	CollectionID int    `json:"collection_id" jsonschema:"The ID of the collection to update"`
	Name         string `json:"name" jsonschema:"Name of the collection"`
	Description  string `json:"description,omitempty" jsonschema:"Optional description of the collection"`
	Query        string `json:"query" jsonschema:"The ClickHouse SQL query to save in the collection"`
}

type DeleteCollectionParams struct {
	TeamID       int `json:"team_id" jsonschema:"The ID of the team that has access to the source"`
	SourceID     int `json:"source_id" jsonschema:"The ID of the source that contains the collection"`
	CollectionID int `json:"collection_id" jsonschema:"The ID of the collection to delete"`
}

// --- Output schemas ---

type SchemaColumnResult struct {
	Name string `json:"name" jsonschema:"Column name"`
	Type string `json:"type" jsonschema:"ClickHouse column type"`
}

type CollectionResult struct {
	ID          int    `json:"id" jsonschema:"Collection ID"`
	Name        string `json:"name" jsonschema:"Collection name"`
	Description string `json:"description" jsonschema:"Collection description"`
	TeamID      int    `json:"team_id" jsonschema:"Team ID"`
	SourceID    int    `json:"source_id" jsonschema:"Source ID"`
	Query       string `json:"query" jsonschema:"Saved ClickHouse SQL query"`
	CreatedAt   string `json:"created_at" jsonschema:"Creation timestamp"`
	UpdatedAt   string `json:"updated_at" jsonschema:"Last update timestamp"`
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

func handleGetCollections(ctx context.Context, request mcp.CallToolRequest, args GetCollectionsParams) ([]CollectionResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("logchef client not configured")
	}

	collections, err := c.GetCollections(ctx, args.TeamID, args.SourceID)
	if err != nil {
		return nil, fmt.Errorf("get collections: %w", err)
	}

	result := make([]CollectionResult, len(collections.Data))
	for i, col := range collections.Data {
		result[i] = collectionToResult(col)
	}
	return result, nil
}

func handleCreateCollection(ctx context.Context, request mcp.CallToolRequest, args CreateCollectionParams) (CollectionResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return CollectionResult{}, fmt.Errorf("logchef client not configured")
	}

	collection, err := c.CreateCollection(ctx, args.TeamID, args.SourceID, client.CollectionRequest{
		Name: args.Name, Description: args.Description, Query: args.Query,
	})
	if err != nil {
		return CollectionResult{}, fmt.Errorf("create collection: %w", err)
	}

	return collectionToResult(collection.Data), nil
}

func handleGetCollection(ctx context.Context, request mcp.CallToolRequest, args GetCollectionParams) (CollectionResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return CollectionResult{}, fmt.Errorf("logchef client not configured")
	}

	collection, err := c.GetCollection(ctx, args.TeamID, args.SourceID, args.CollectionID)
	if err != nil {
		return CollectionResult{}, fmt.Errorf("get collection: %w", err)
	}

	return collectionToResult(collection.Data), nil
}

func handleUpdateCollection(ctx context.Context, request mcp.CallToolRequest, args UpdateCollectionParams) (CollectionResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return CollectionResult{}, fmt.Errorf("logchef client not configured")
	}

	collection, err := c.UpdateCollection(ctx, args.TeamID, args.SourceID, args.CollectionID, client.CollectionRequest{
		Name: args.Name, Description: args.Description, Query: args.Query,
	})
	if err != nil {
		return CollectionResult{}, fmt.Errorf("update collection: %w", err)
	}

	return collectionToResult(collection.Data), nil
}

func handleDeleteCollection(ctx context.Context, request mcp.CallToolRequest, args DeleteCollectionParams) (SuccessResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return SuccessResult{}, fmt.Errorf("logchef client not configured")
	}

	if err := c.DeleteCollection(ctx, args.TeamID, args.SourceID, args.CollectionID); err != nil {
		return SuccessResult{}, fmt.Errorf("delete collection: %w", err)
	}

	return SuccessResult{Success: true, Message: "Collection deleted successfully"}, nil
}

func collectionToResult(c client.Collection) CollectionResult {
	return CollectionResult{
		ID: c.ID, Name: c.Name, Description: c.Description,
		TeamID: c.TeamID, SourceID: c.SourceID, Query: c.Query,
		CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
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

	getCollectionsTool := mcp.NewTool("get_collections",
		mcp.WithDescription("Get all saved query collections for a specific team and source. Collections are saved queries that can be reused for common log analysis patterns."),
		mcp.WithInputSchema[GetCollectionsParams](),
		mcp.WithOutputSchema[[]CollectionResult](),
		mcp.WithTitleAnnotation("Get Collections"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(getCollectionsTool, mcp.NewStructuredToolHandler(handleGetCollections))

	createCollectionTool := mcp.NewTool("create_collection",
		mcp.WithDescription("Create a new saved query collection for a specific team and source. Provide a name, optional description, and the ClickHouse SQL query to save."),
		mcp.WithInputSchema[CreateCollectionParams](),
		mcp.WithOutputSchema[CollectionResult](),
		mcp.WithTitleAnnotation("Create Collection"),
		mcp.WithDestructiveHintAnnotation(false),
	)
	s.AddTool(createCollectionTool, mcp.NewStructuredToolHandler(handleCreateCollection))

	getCollectionTool := mcp.NewTool("get_collection",
		mcp.WithDescription("Get a specific saved query collection by ID. Returns the collection details including name, description, query, and metadata."),
		mcp.WithInputSchema[GetCollectionParams](),
		mcp.WithOutputSchema[CollectionResult](),
		mcp.WithTitleAnnotation("Get Collection"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(getCollectionTool, mcp.NewStructuredToolHandler(handleGetCollection))

	updateCollectionTool := mcp.NewTool("update_collection",
		mcp.WithDescription("Update an existing saved query collection. All fields are required - provide the current values for fields you don't want to change."),
		mcp.WithInputSchema[UpdateCollectionParams](),
		mcp.WithOutputSchema[CollectionResult](),
		mcp.WithTitleAnnotation("Update Collection"),
		mcp.WithDestructiveHintAnnotation(false),
	)
	s.AddTool(updateCollectionTool, mcp.NewStructuredToolHandler(handleUpdateCollection))

	deleteCollectionTool := mcp.NewTool("delete_collection",
		mcp.WithDescription("Delete a saved query collection by ID. This permanently removes the collection and cannot be undone."),
		mcp.WithInputSchema[DeleteCollectionParams](),
		mcp.WithOutputSchema[SuccessResult](),
		mcp.WithTitleAnnotation("Delete Collection"),
		mcp.WithDestructiveHintAnnotation(true),
	)
	s.AddTool(deleteCollectionTool, mcp.NewStructuredToolHandler(handleDeleteCollection))
}
