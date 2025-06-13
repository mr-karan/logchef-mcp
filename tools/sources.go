package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/server"

	mcplogchef "github.com/mr-karan/logchef-mcp"
	"github.com/mr-karan/logchef-mcp/client"
)

// GetTeamSourcesParams represents the parameters for getting sources for a specific team.
type GetTeamSourcesParams struct {
	TeamID int `json:"team_id" jsonschema:"description=The ID of the team to get sources for,required"`
}

// GetSourcesParams represents the parameters for getting all sources accessible to the user.
// Since this aggregates across all teams, no parameters are needed.
type GetSourcesParams struct{}


// SourcesAggregateResponse represents the aggregated response with sources and their team associations
type SourcesAggregateResponse struct {
	Sources []SourceWithTeams `json:"sources"`
}

// SourceWithTeams represents a source with information about which teams it belongs to
type SourceWithTeams struct {
	*client.SourceResponse
	Teams []TeamInfo `json:"teams"`
}

// TeamInfo represents basic team information for source association
type TeamInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

func getTeamSources(ctx context.Context, args GetTeamSourcesParams) (*client.TeamSourcesResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	sources, err := c.GetTeamSources(ctx, args.TeamID)
	if err != nil {
		return nil, fmt.Errorf("get team sources: %w", err)
	}

	return sources, nil
}

func getSources(ctx context.Context, args GetSourcesParams) (*SourcesAggregateResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	// First, get all teams the user belongs to
	teamsResp, err := c.GetTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("get user teams: %w", err)
	}

	// Map to aggregate sources and their team associations
	sourceMap := make(map[int]*SourceWithTeams)

	// For each team, get its sources
	for _, team := range teamsResp.Data {
		teamSources, err := c.GetTeamSources(ctx, team.ID)
		if err != nil {
			return nil, fmt.Errorf("get sources for team %d: %w", team.ID, err)
		}

		// Add each source to the map, tracking team associations
		for _, source := range teamSources.Data {
			if existing, exists := sourceMap[source.ID]; exists {
				// Source already exists, add this team to its team list
				existing.Teams = append(existing.Teams, TeamInfo{
					ID:   team.ID,
					Name: team.Name,
					Role: team.Role,
				})
			} else {
				// New source, create entry with first team
				sourceMap[source.ID] = &SourceWithTeams{
					SourceResponse: source,
					Teams: []TeamInfo{{
						ID:   team.ID,
						Name: team.Name,
						Role: team.Role,
					}},
				}
			}
		}
	}

	// Convert map to slice
	sources := make([]SourceWithTeams, 0, len(sourceMap))
	for _, source := range sourceMap {
		sources = append(sources, *source)
	}

	return &SourcesAggregateResponse{Sources: sources}, nil
}


var GetTeamSources = mcplogchef.MustTool(
	"get_team_sources",
	"Get the sources that belong to a specific team. Requires the team ID as a parameter.",
	getTeamSources,
)

var GetSources = mcplogchef.MustTool(
	"get_sources",
	"Get all sources that the current user has access to across all their team memberships. Returns sources with their team associations and the user's role in each team.",
	getSources,
)


func AddSourcesTools(mcp *server.MCPServer) {
	GetTeamSources.Register(mcp)
	GetSources.Register(mcp)
}