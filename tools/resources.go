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

	// Collections list resource template
	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"logchef://team/{team_id}/source/{source_id}/collections",
			"Saved Query Collections",
			mcp.WithTemplateDescription("List of saved query collections for a log source. Collections are reusable queries for common log analysis patterns."),
			mcp.WithTemplateMIMEType("application/json"),
		),
		handleCollectionsListResource,
	)

	// Single collection resource template
	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"logchef://team/{team_id}/source/{source_id}/collection/{collection_id}",
			"Saved Query Collection",
			mcp.WithTemplateDescription("A single saved query collection with its name, description, and ClickHouse SQL query."),
			mcp.WithTemplateMIMEType("application/json"),
		),
		handleCollectionResource,
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

func handleCollectionsListResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	teamID, sourceID, err := parseTeamSourceURI(request.Params.URI)
	if err != nil {
		return nil, err
	}

	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("logchef client not configured")
	}

	collections, err := c.GetCollections(ctx, teamID, sourceID)
	if err != nil {
		return nil, fmt.Errorf("get collections: %w", err)
	}

	out, _ := json.MarshalIndent(collections.Data, "", "  ")
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(out),
		},
	}, nil
}

func handleCollectionResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	teamID, sourceID, collectionID, err := parseCollectionURI(request.Params.URI)
	if err != nil {
		return nil, err
	}

	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("logchef client not configured")
	}

	collection, err := c.GetCollection(ctx, teamID, sourceID, collectionID)
	if err != nil {
		return nil, fmt.Errorf("get collection: %w", err)
	}

	out, _ := json.MarshalIndent(collection.Data, "", "  ")
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

// parseCollectionURI extracts team_id, source_id, and collection_id from URIs like
// logchef://team/{team_id}/source/{source_id}/collection/{collection_id}
func parseCollectionURI(uri string) (int, int, int, error) {
	parts := strings.Split(strings.TrimPrefix(uri, "logchef://"), "/")
	// Expected: team/{id}/source/{id}/collection/{id}
	if len(parts) < 6 || parts[0] != "team" || parts[2] != "source" || parts[4] != "collection" {
		return 0, 0, 0, fmt.Errorf("invalid collection URI format: %s", uri)
	}

	teamID, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid team_id in URI: %s", parts[1])
	}

	sourceID, err := strconv.Atoi(parts[3])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid source_id in URI: %s", parts[3])
	}

	collectionID, err := strconv.Atoi(parts[5])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid collection_id in URI: %s", parts[5])
	}

	return teamID, sourceID, collectionID, nil
}
