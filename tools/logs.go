package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/server"

	mcplogchef "github.com/mr-karan/logchef-mcp"
	"github.com/mr-karan/logchef-mcp/client"
)

// QueryLogsParams represents the parameters for querying logs from a specific source.
type QueryLogsParams struct {
	TeamID       int    `json:"team_id" jsonschema:"description=The ID of the team that has access to the source,required"`
	SourceID     int    `json:"source_id" jsonschema:"description=The ID of the source to query logs from,required"`
	RawSQL       string `json:"raw_sql" jsonschema:"description=The ClickHouse SQL query to execute. Use the get_source_schema tool first to understand available columns and their types. The query should include proper WHERE clauses with timestamp filters\\, ORDER BY clauses\\, and LIMIT clauses. Example: 'SELECT * FROM default.logs WHERE timestamp BETWEEN toDateTime(\\'2025-06-12 05:00:00\\') AND toDateTime(\\'2025-06-12 17:00:00\\') ORDER BY timestamp DESC LIMIT 50',required"`
	Limit        int    `json:"limit,omitempty" jsonschema:"description=Maximum number of log entries to return. Must be between 1 and 100. Defaults to 100 if not specified."`
	QueryTimeout *int   `json:"query_timeout,omitempty" jsonschema:"description=Query timeout in seconds. Defaults to 30 seconds if not specified."`
}

// GetSourceSchemaParams represents the parameters for getting the schema of a source.
type GetSourceSchemaParams struct {
	TeamID   int `json:"team_id" jsonschema:"description=The ID of the team that has access to the source,required"`
	SourceID int `json:"source_id" jsonschema:"description=The ID of the source to get the schema for,required"`
}

// GetLogHistogramParams represents the parameters for getting log histogram data.
type GetLogHistogramParams struct {
	TeamID       int    `json:"team_id" jsonschema:"description=The ID of the team that has access to the source,required"`
	SourceID     int    `json:"source_id" jsonschema:"description=The ID of the source to generate histogram for,required"`
	RawSQL       string `json:"raw_sql" jsonschema:"description=The ClickHouse SQL query to analyze. Should include proper WHERE clauses with timestamp filters. Example: 'SELECT * FROM default.logs WHERE timestamp BETWEEN toDateTime(\\'2025-06-12 05:00:00\\') AND toDateTime(\\'2025-06-12 17:00:00\\')  ORDER BY timestamp DESC',required"`
	Window       string `json:"window,omitempty" jsonschema:"description=Time window for histogram buckets. Examples: '1m' (1 minute)\\, '5m' (5 minutes)\\, '1h' (1 hour)\\, '1d' (1 day). Defaults to '1m' if not specified."`
	GroupBy      string `json:"group_by,omitempty" jsonschema:"description=Optional field to group histogram data by. Example: 'severity_text' to group by log level\\, 'service_name' to group by service."`
	Timezone     string `json:"timezone,omitempty" jsonschema:"description=Timezone for histogram timestamps. Defaults to 'UTC' if not specified. Examples: 'UTC'\\, 'America/New_York'\\, 'Europe/London'."`
	QueryTimeout *int   `json:"query_timeout,omitempty" jsonschema:"description=Query timeout in seconds. Defaults to 30 seconds if not specified."`
}

// GetCollectionsParams represents the parameters for getting collections.
type GetCollectionsParams struct {
	TeamID   int `json:"team_id" jsonschema:"description=The ID of the team that has access to the source,required"`
	SourceID int `json:"source_id" jsonschema:"description=The ID of the source to get collections for,required"`
}

// CreateCollectionParams represents the parameters for creating a collection.
type CreateCollectionParams struct {
	TeamID      int    `json:"team_id" jsonschema:"description=The ID of the team that has access to the source,required"`
	SourceID    int    `json:"source_id" jsonschema:"description=The ID of the source to create the collection for,required"`
	Name        string `json:"name" jsonschema:"description=Name of the collection,required"`
	Description string `json:"description,omitempty" jsonschema:"description=Optional description of the collection"`
	Query       string `json:"query" jsonschema:"description=The ClickHouse SQL query to save in the collection,required"`
}

// GetCollectionParams represents the parameters for getting a specific collection.
type GetCollectionParams struct {
	TeamID       int `json:"team_id" jsonschema:"description=The ID of the team that has access to the source,required"`
	SourceID     int `json:"source_id" jsonschema:"description=The ID of the source that contains the collection,required"`
	CollectionID int `json:"collection_id" jsonschema:"description=The ID of the collection to retrieve,required"`
}

// UpdateCollectionParams represents the parameters for updating a collection.
type UpdateCollectionParams struct {
	TeamID       int    `json:"team_id" jsonschema:"description=The ID of the team that has access to the source,required"`
	SourceID     int    `json:"source_id" jsonschema:"description=The ID of the source that contains the collection,required"`
	CollectionID int    `json:"collection_id" jsonschema:"description=The ID of the collection to update,required"`
	Name         string `json:"name" jsonschema:"description=Name of the collection,required"`
	Description  string `json:"description,omitempty" jsonschema:"description=Optional description of the collection"`
	Query        string `json:"query" jsonschema:"description=The ClickHouse SQL query to save in the collection,required"`
}

// DeleteCollectionParams represents the parameters for deleting a collection.
type DeleteCollectionParams struct {
	TeamID       int `json:"team_id" jsonschema:"description=The ID of the team that has access to the source,required"`
	SourceID     int `json:"source_id" jsonschema:"description=The ID of the source that contains the collection,required"`
	CollectionID int `json:"collection_id" jsonschema:"description=The ID of the collection to delete,required"`
}


func queryLogs(ctx context.Context, args QueryLogsParams) (*client.LogQueryResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	// Validate limit
	if args.Limit < 0 {
		args.Limit = 0 // Will be set to default by client
	}
	if args.Limit > 100 {
		args.Limit = 100
	}

	request := client.LogQueryRequest{
		RawSQL:       args.RawSQL,
		Limit:        args.Limit,
		QueryTimeout: args.QueryTimeout,
	}

	logs, err := c.QueryLogs(ctx, args.TeamID, args.SourceID, request)
	if err != nil {
		return nil, fmt.Errorf("query logs: %w", err)
	}

	return logs, nil
}

func getSourceSchema(ctx context.Context, args GetSourceSchemaParams) (*client.SchemaResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	schema, err := c.GetSourceSchema(ctx, args.TeamID, args.SourceID)
	if err != nil {
		return nil, fmt.Errorf("get source schema: %w", err)
	}

	return schema, nil
}

func getLogHistogram(ctx context.Context, args GetLogHistogramParams) (*client.HistogramResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	request := client.HistogramRequest{
		RawSQL:       args.RawSQL,
		Window:       args.Window,
		GroupBy:      args.GroupBy,
		Timezone:     args.Timezone,
		QueryTimeout: args.QueryTimeout,
	}

	histogram, err := c.GetLogHistogram(ctx, args.TeamID, args.SourceID, request)
	if err != nil {
		return nil, fmt.Errorf("get log histogram: %w", err)
	}

	return histogram, nil
}

func getCollections(ctx context.Context, args GetCollectionsParams) (*client.CollectionsResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	collections, err := c.GetCollections(ctx, args.TeamID, args.SourceID)
	if err != nil {
		return nil, fmt.Errorf("get collections: %w", err)
	}

	return collections, nil
}

func createCollection(ctx context.Context, args CreateCollectionParams) (*client.CollectionResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	request := client.CollectionRequest{
		Name:        args.Name,
		Description: args.Description,
		Query:       args.Query,
	}

	collection, err := c.CreateCollection(ctx, args.TeamID, args.SourceID, request)
	if err != nil {
		return nil, fmt.Errorf("create collection: %w", err)
	}

	return collection, nil
}

func getCollection(ctx context.Context, args GetCollectionParams) (*client.CollectionResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	collection, err := c.GetCollection(ctx, args.TeamID, args.SourceID, args.CollectionID)
	if err != nil {
		return nil, fmt.Errorf("get collection: %w", err)
	}

	return collection, nil
}

func updateCollection(ctx context.Context, args UpdateCollectionParams) (*client.CollectionResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	request := client.CollectionRequest{
		Name:        args.Name,
		Description: args.Description,
		Query:       args.Query,
	}

	collection, err := c.UpdateCollection(ctx, args.TeamID, args.SourceID, args.CollectionID, request)
	if err != nil {
		return nil, fmt.Errorf("update collection: %w", err)
	}

	return collection, nil
}

type DeleteCollectionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func deleteCollection(ctx context.Context, args DeleteCollectionParams) (*DeleteCollectionResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	err := c.DeleteCollection(ctx, args.TeamID, args.SourceID, args.CollectionID)
	if err != nil {
		return nil, fmt.Errorf("delete collection: %w", err)
	}

	return &DeleteCollectionResponse{
		Success: true,
		Message: "Collection deleted successfully",
	}, nil
}

var QueryLogs = mcplogchef.MustTool(
	"query_logs",
	"Execute a ClickHouse SQL query against a specific log source within a team. This tool allows you to fetch log entries using raw SQL queries. Before using this tool, use get_source_schema to understand the available columns and their types. The query should include proper WHERE clauses with timestamp filters, ORDER BY clauses, and LIMIT clauses. Returns log entries with execution statistics and column metadata. Maximum 100 results per query.",
	queryLogs,
)

var GetSourceSchema = mcplogchef.MustTool(
	"get_source_schema",
	"Get the ClickHouse table schema (column names and types) for a specific log source within a team. This tool returns the structure of the underlying ClickHouse table, including column names and their data types. Use this tool before querying logs to understand what fields are available and how to construct proper SQL queries.",
	getSourceSchema,
)

var GetLogHistogram = mcplogchef.MustTool(
	"get_log_histogram",
	"Generate time-based histogram data for log analysis within a team source. Creates a time series showing log volume over specified time windows, with optional grouping by fields like severity or service. Useful for identifying traffic patterns, spikes, and trends in log data. Requires a SQL query with proper time filters.",
	getLogHistogram,
)

var GetCollections = mcplogchef.MustTool(
	"get_collections",
	"Get all saved query collections for a specific team and source. Collections are saved queries that can be reused for common log analysis patterns. Returns a list of collections with their names, descriptions, queries, and metadata.",
	getCollections,
)

var CreateCollection = mcplogchef.MustTool(
	"create_collection",
	"Create a new saved query collection for a specific team and source. Collections allow you to save frequently used queries for easy reuse. Provide a name, optional description, and the ClickHouse SQL query to save.",
	createCollection,
)

var GetCollection = mcplogchef.MustTool(
	"get_collection",
	"Get a specific saved query collection by ID. Returns the collection details including name, description, query, and metadata. Use this to retrieve a previously saved query for execution or modification.",
	getCollection,
)

var UpdateCollection = mcplogchef.MustTool(
	"update_collection",
	"Update an existing saved query collection. You can modify the name, description, and query of an existing collection. All fields are required - provide the current values for fields you don't want to change.",
	updateCollection,
)

var DeleteCollection = mcplogchef.MustTool(
	"delete_collection",
	"Delete a saved query collection by ID. This permanently removes the collection and cannot be undone. Returns a success confirmation when the collection is deleted.",
	deleteCollection,
)

func AddLogsTools(mcp *server.MCPServer) {
	QueryLogs.Register(mcp)
	GetSourceSchema.Register(mcp)
	GetLogHistogram.Register(mcp)
	GetCollections.Register(mcp)
	CreateCollection.Register(mcp)
	GetCollection.Register(mcp)
	UpdateCollection.Register(mcp)
	DeleteCollection.Register(mcp)
}