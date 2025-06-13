package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/server"

	mcplogchef "github.com/mr-karan/logchef-mcp"
	"github.com/mr-karan/logchef-mcp/client"
)

// GetProfileParams represents the parameters for getting user profile.
// Since the /api/v1/me endpoint doesn't require any parameters, this struct is empty.
type GetProfileParams struct{}

// GetTeamsParams represents the parameters for getting user teams.
// Since the /api/v1/me/teams endpoint doesn't require any parameters, this struct is empty.
type GetTeamsParams struct{}

// GetMetaParams represents the parameters for getting server metadata.
// Since the /api/v1/meta endpoint doesn't require any parameters, this struct is empty.
type GetMetaParams struct{}

func getProfile(ctx context.Context, args GetProfileParams) (*client.ProfileResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	profile, err := c.GetProfile(ctx)
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}

	return profile, nil
}

func getTeams(ctx context.Context, args GetTeamsParams) (*client.TeamsResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	teams, err := c.GetTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("get teams: %w", err)
	}

	return teams, nil
}

func getMeta(ctx context.Context, args GetMetaParams) (*client.MetaResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	meta, err := c.GetMeta(ctx)
	if err != nil {
		return nil, fmt.Errorf("get meta: %w", err)
	}

	return meta, nil
}

var GetProfile = mcplogchef.MustTool(
	"get_profile",
	"Get the current user profile information from Logchef, including user details and API token information.",
	getProfile,
)

var GetTeams = mcplogchef.MustTool(
	"get_teams",
	"Get the teams that the current user belongs to, including their role in each team and member count.",
	getTeams,
)

var GetMeta = mcplogchef.MustTool(
	"get_meta",
	"Get server metadata including version information and configuration details.",
	getMeta,
)

func AddProfileTools(mcp *server.MCPServer) {
	GetProfile.Register(mcp)
	GetTeams.Register(mcp)
	GetMeta.Register(mcp)
}