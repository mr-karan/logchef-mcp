package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	mcplogchef "github.com/mr-karan/logchef-mcp"
)

// AddResourceTemplates registers MCP resource templates for Logchef entities.
func AddResourceTemplates(s *server.MCPServer) {
	// Source schema resource template
	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"logchef://team/{team_id}/source/{source_id}/schema",
			"Source Schema",
			mcp.WithTemplateDescription("ClickHouse table schema (column names and types) for a log source. Use this to understand available fields before writing queries."),
			mcp.WithTemplateMIMEType("application/json"),
		),
		handleSourceSchemaResource,
	)

	// Saved queries for a source (logchef v1.6.0+ — source-scoped, not
	// team-scoped). The team_id segment is kept in the URI so callers can
	// continue to use the same team-pivoted browsing UX, but is ignored
	// when looking up the saved queries themselves.
	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"logchef://source/{source_id}/saved-queries",
			"Saved Queries",
			mcp.WithTemplateDescription("Saved queries pinned to a log source. Each entry has a query_type (logchefql or sql), query_content (JSON envelope), and source metadata."),
			mcp.WithTemplateMIMEType("application/json"),
		),
		handleSavedQueriesListResource,
	)

	// Single saved query by ID. Saved queries are no longer scoped under a
	// team/source path — the canonical lookup is by global query ID.
	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"logchef://saved-query/{query_id}",
			"Saved Query",
			mcp.WithTemplateDescription("A single saved query with its name, description, query_type, query_content, and source metadata."),
			mcp.WithTemplateMIMEType("application/json"),
		),
		handleSavedQueryResource,
	)
}

func handleSourceSchemaResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	teamID, sourceID, err := parseTeamSourceURI(request.Params.URI)
	if err != nil {
		return nil, err
	}

	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("logchef client not configured")
	}

	schema, err := c.GetSourceSchema(ctx, teamID, sourceID)
	if err != nil {
		return nil, fmt.Errorf("get source schema: %w", err)
	}

	out, _ := json.MarshalIndent(schema.Data, "", "  ")
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(out),
		},
	}, nil
}

func handleSavedQueriesListResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	sourceID, err := parseSourceURI(request.Params.URI)
	if err != nil {
		return nil, err
	}

	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("logchef client not configured")
	}

	queries, err := c.ListSavedQueries(ctx, sourceID)
	if err != nil {
		return nil, fmt.Errorf("list saved queries: %w", err)
	}

	out, _ := json.MarshalIndent(queries.Data, "", "  ")
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(out),
		},
	}, nil
}

func handleSavedQueryResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	queryID, err := parseSavedQueryURI(request.Params.URI)
	if err != nil {
		return nil, err
	}

	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("logchef client not configured")
	}

	got, err := c.GetSavedQuery(ctx, queryID)
	if err != nil {
		return nil, fmt.Errorf("get saved query: %w", err)
	}

	out, _ := json.MarshalIndent(got.Data, "", "  ")
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(out),
		},
	}, nil
}

// parseTeamSourceURI extracts team_id and source_id from URIs like
// logchef://team/{team_id}/source/{source_id}/...
func parseTeamSourceURI(uri string) (int, int, error) {
	parts := strings.Split(strings.TrimPrefix(uri, "logchef://"), "/")
	if len(parts) < 4 || parts[0] != "team" || parts[2] != "source" {
		return 0, 0, fmt.Errorf("invalid URI format: %s", uri)
	}

	teamID, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid team_id in URI: %s", parts[1])
	}

	sourceID, err := strconv.Atoi(parts[3])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid source_id in URI: %s", parts[3])
	}

	return teamID, sourceID, nil
}

// parseSourceURI extracts source_id from URIs like
// logchef://source/{source_id}/saved-queries
func parseSourceURI(uri string) (int, error) {
	parts := strings.Split(strings.TrimPrefix(uri, "logchef://"), "/")
	if len(parts) < 2 || parts[0] != "source" {
		return 0, fmt.Errorf("invalid source URI format: %s", uri)
	}

	sourceID, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid source_id in URI: %s", parts[1])
	}

	return sourceID, nil
}

// parseSavedQueryURI extracts query_id from URIs like
// logchef://saved-query/{query_id}
func parseSavedQueryURI(uri string) (int, error) {
	parts := strings.Split(strings.TrimPrefix(uri, "logchef://"), "/")
	if len(parts) < 2 || parts[0] != "saved-query" {
		return 0, fmt.Errorf("invalid saved-query URI format: %s", uri)
	}

	queryID, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid query_id in URI: %s", parts[1])
	}

	return queryID, nil
}
