package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	mcplogchef "github.com/mr-karan/logchef-mcp"
)

// GetProfileParams represents the parameters for getting user profile.
type GetProfileParams struct{}

// GetTeamsParams represents the parameters for getting user teams.
type GetTeamsParams struct{}

// GetMetaParams represents the parameters for getting server metadata.
type GetMetaParams struct{}

// --- Output schemas ---

type ProfileResult struct {
	User       ProfileUser     `json:"user" jsonschema:"User details"`
	APIToken   ProfileAPIToken `json:"api_token" jsonschema:"Current API token details"`
	AuthMethod string          `json:"auth_method" jsonschema:"Authentication method used (api_token or oidc)"`
}

type ProfileUser struct {
	ID          int    `json:"id" jsonschema:"User ID"`
	Email       string `json:"email" jsonschema:"User email address"`
	FullName    string `json:"full_name" jsonschema:"User full name"`
	Role        string `json:"role" jsonschema:"User role (admin or member)"`
	Status      string `json:"status" jsonschema:"User status (active or inactive)"`
	LastLoginAt string `json:"last_login_at" jsonschema:"Last login timestamp"`
	CreatedAt   string `json:"created_at" jsonschema:"Account creation timestamp"`
}

type ProfileAPIToken struct {
	ID         int    `json:"id" jsonschema:"Token ID"`
	Name       string `json:"name" jsonschema:"Token name"`
	Prefix     string `json:"prefix" jsonschema:"Token prefix for identification"`
	LastUsedAt string `json:"last_used_at" jsonschema:"Last usage timestamp"`
	CreatedAt  string `json:"created_at" jsonschema:"Token creation timestamp"`
}

type TeamResult struct {
	ID          int    `json:"id" jsonschema:"Team ID"`
	Name        string `json:"name" jsonschema:"Team name"`
	Role        string `json:"role" jsonschema:"Current user role in this team"`
	MemberCount int    `json:"member_count" jsonschema:"Number of members in the team"`
	CreatedAt   string `json:"created_at" jsonschema:"Team creation timestamp"`
	UpdatedAt   string `json:"updated_at" jsonschema:"Team last update timestamp"`
}

type MetaResult struct {
	Version           string `json:"version" jsonschema:"Server version"`
	HTTPServerTimeout string `json:"http_server_timeout" jsonschema:"HTTP server timeout setting"`
}

// --- Handlers ---

func handleGetProfile(ctx context.Context, request mcp.CallToolRequest, args GetProfileParams) (ProfileResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return ProfileResult{}, fmt.Errorf("logchef client not configured")
	}

	profile, err := c.GetProfile(ctx)
	if err != nil {
		return ProfileResult{}, fmt.Errorf("get profile: %w", err)
	}

	d := profile.Data
	return ProfileResult{
		User: ProfileUser{
			ID:          d.User.ID,
			Email:       d.User.Email,
			FullName:    d.User.FullName,
			Role:        d.User.Role,
			Status:      d.User.Status,
			LastLoginAt: d.User.LastLoginAt,
			CreatedAt:   d.User.CreatedAt,
		},
		APIToken: ProfileAPIToken{
			ID:         d.APIToken.ID,
			Name:       d.APIToken.Name,
			Prefix:     d.APIToken.Prefix,
			LastUsedAt: d.APIToken.LastUsedAt,
			CreatedAt:  d.APIToken.CreatedAt,
		},
		AuthMethod: d.AuthMethod,
	}, nil
}

func handleGetTeams(ctx context.Context, request mcp.CallToolRequest, args GetTeamsParams) ([]TeamResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("logchef client not configured")
	}

	teams, err := c.GetTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("get teams: %w", err)
	}

	result := make([]TeamResult, len(teams.Data))
	for i, t := range teams.Data {
		result[i] = TeamResult{
			ID:          t.ID,
			Name:        t.Name,
			Role:        t.Role,
			MemberCount: t.MemberCount,
			CreatedAt:   t.CreatedAt,
			UpdatedAt:   t.UpdatedAt,
		}
	}
	return result, nil
}

func handleGetMeta(ctx context.Context, request mcp.CallToolRequest, args GetMetaParams) (MetaResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return MetaResult{}, fmt.Errorf("logchef client not configured")
	}

	meta, err := c.GetMeta(ctx)
	if err != nil {
		return MetaResult{}, fmt.Errorf("get meta: %w", err)
	}

	return MetaResult{
		Version:           meta.Data.Version,
		HTTPServerTimeout: meta.Data.HTTPServerTimeout,
	}, nil
}

func AddProfileTools(s *server.MCPServer) {
	profileTool := mcp.NewTool("get_profile",
		mcp.WithDescription("Get the current user profile information from Logchef, including user details and API token information."),
		mcp.WithInputSchema[GetProfileParams](),
		mcp.WithOutputSchema[ProfileResult](),
		mcp.WithTitleAnnotation("Get Profile"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(profileTool, mcp.NewStructuredToolHandler(handleGetProfile))

	teamsTool := mcp.NewTool("get_teams",
		mcp.WithDescription("Get the teams that the current user belongs to, including their role in each team and member count."),
		mcp.WithInputSchema[GetTeamsParams](),
		mcp.WithOutputSchema[[]TeamResult](),
		mcp.WithTitleAnnotation("Get My Teams"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(teamsTool, mcp.NewStructuredToolHandler(handleGetTeams))

	metaTool := mcp.NewTool("get_meta",
		mcp.WithDescription("Get server metadata including version information and configuration details."),
		mcp.WithInputSchema[GetMetaParams](),
		mcp.WithOutputSchema[MetaResult](),
		mcp.WithTitleAnnotation("Get Server Metadata"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(metaTool, mcp.NewStructuredToolHandler(handleGetMeta))
}
