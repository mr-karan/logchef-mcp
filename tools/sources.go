package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"golang.org/x/sync/errgroup"

	mcplogchef "github.com/mr-karan/logchef-mcp"
	"github.com/mr-karan/logchef-mcp/client"
)

// GetTeamSourcesParams represents the parameters for getting sources for a specific team.
type GetTeamSourcesParams struct {
	TeamID int `json:"team_id" jsonschema:"The ID of the team to get sources for"`
}

// GetSourcesParams represents the parameters for getting all sources accessible to the user.
type GetSourcesParams struct{}

// --- Output schemas ---

type SourceResult struct {
	ID          int              `json:"id" jsonschema:"Source ID"`
	Name        string           `json:"name" jsonschema:"Source name"`
	Description string           `json:"description" jsonschema:"Source description"`
	Connection  ConnectionResult `json:"connection" jsonschema:"ClickHouse connection details"`
	TsField     string           `json:"ts_field" jsonschema:"Timestamp field name"`
	IsConnected bool             `json:"is_connected" jsonschema:"Whether the source is currently connected"`
	TTLDays     int              `json:"ttl_days" jsonschema:"Data retention in days"`
	CreatedAt   string           `json:"created_at" jsonschema:"Creation timestamp"`
}

type ConnectionResult struct {
	Host      string `json:"host" jsonschema:"ClickHouse host"`
	Database  string `json:"database" jsonschema:"ClickHouse database name"`
	TableName string `json:"table_name" jsonschema:"ClickHouse table name"`
}

type SourceWithTeamsResult struct {
	SourceResult
	Teams []TeamInfo `json:"teams" jsonschema:"Teams this source belongs to"`
}

type TeamInfo struct {
	ID   int    `json:"id" jsonschema:"Team ID"`
	Name string `json:"name" jsonschema:"Team name"`
	Role string `json:"role" jsonschema:"User role in this team"`
}

type SourcesAggregateResult struct {
	Sources []SourceWithTeamsResult `json:"sources" jsonschema:"All accessible sources with team associations"`
}

// --- Handlers ---

func handleGetTeamSources(ctx context.Context, request mcp.CallToolRequest, args GetTeamSourcesParams) ([]SourceResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("logchef client not configured")
	}

	sources, err := c.GetTeamSources(ctx, args.TeamID)
	if err != nil {
		return nil, fmt.Errorf("get team sources: %w", err)
	}

	result := make([]SourceResult, len(sources.Data))
	for i, s := range sources.Data {
		result[i] = SourceResult{
			ID:          s.ID,
			Name:        s.Name,
			Description: s.Description,
			Connection: ConnectionResult{
				Host:      s.Connection.Host,
				Database:  s.Connection.Database,
				TableName: s.Connection.TableName,
			},
			TsField:     s.MetaTsField,
			IsConnected: s.IsConnected,
			TTLDays:     s.TTLDays,
			CreatedAt:   s.CreatedAt,
		}
	}
	return result, nil
}

func handleGetSources(ctx context.Context, request mcp.CallToolRequest, args GetSourcesParams) (SourcesAggregateResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return SourcesAggregateResult{}, fmt.Errorf("logchef client not configured")
	}

	teamsResp, err := c.GetTeams(ctx)
	if err != nil {
		return SourcesAggregateResult{}, fmt.Errorf("get user teams: %w", err)
	}

	// Fetch sources for all teams in parallel to avoid N+1.
	type teamSourceResult struct {
		team    TeamInfo
		sources []*client.SourceResponse
	}

	results := make([]teamSourceResult, len(teamsResp.Data))
	g, gctx := errgroup.WithContext(ctx)

	for i, team := range teamsResp.Data {
		g.Go(func() error {
			teamSources, err := c.GetTeamSources(gctx, team.ID)
			if err != nil {
				return fmt.Errorf("get sources for team %d: %w", team.ID, err)
			}
			results[i] = teamSourceResult{
				team:    TeamInfo{ID: team.ID, Name: team.Name, Role: team.Role},
				sources: teamSources.Data,
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return SourcesAggregateResult{}, err
	}

	// Merge results into deduplicated map.
	sourceMap := make(map[int]*SourceWithTeamsResult)
	for _, r := range results {
		for _, s := range r.sources {
			if existing, exists := sourceMap[s.ID]; exists {
				existing.Teams = append(existing.Teams, r.team)
			} else {
				sourceMap[s.ID] = &SourceWithTeamsResult{
					SourceResult: SourceResult{
						ID: s.ID, Name: s.Name, Description: s.Description,
						Connection: ConnectionResult{
							Host: s.Connection.Host, Database: s.Connection.Database, TableName: s.Connection.TableName,
						},
						TsField: s.MetaTsField, IsConnected: s.IsConnected, TTLDays: s.TTLDays, CreatedAt: s.CreatedAt,
					},
					Teams: []TeamInfo{r.team},
				}
			}
		}
	}

	sources := make([]SourceWithTeamsResult, 0, len(sourceMap))
	for _, entry := range sourceMap {
		sources = append(sources, *entry)
	}

	return SourcesAggregateResult{Sources: sources}, nil
}

func AddSourcesTools(s *server.MCPServer) {
	teamSourcesTool := mcp.NewTool("get_team_sources",
		mcp.WithDescription("Get the sources that belong to a specific team. Requires the team ID. If the user gives a team name instead of an ID, call get_teams first to find the numeric ID."),
		mcp.WithInputSchema[GetTeamSourcesParams](),
		mcp.WithOutputSchema[[]SourceResult](),
		mcp.WithTitleAnnotation("Get Team Sources"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(teamSourcesTool, mcp.NewStructuredToolHandler(handleGetTeamSources))

	sourcesTool := mcp.NewTool("get_sources",
		mcp.WithDescription("Get all sources the current user can access across all teams. Returns sources with team associations. Use this when the user mentions a source by name — find the matching source and use its team_id and source_id for subsequent queries."),
		mcp.WithInputSchema[GetSourcesParams](),
		mcp.WithOutputSchema[SourcesAggregateResult](),
		mcp.WithTitleAnnotation("Get All Accessible Sources"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(sourcesTool, mcp.NewStructuredToolHandler(handleGetSources))
}
