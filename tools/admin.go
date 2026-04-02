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

type ListAllTeamsParams struct{}

type GetTeamParams struct {
	TeamID int `json:"team_id" jsonschema:"The ID of the team to retrieve"`
}

type CreateTeamParams struct {
	Name        string `json:"name" jsonschema:"Name of the team"`
	Description string `json:"description,omitempty" jsonschema:"Optional description of the team"`
}

type UpdateTeamParams struct {
	TeamID      int     `json:"team_id" jsonschema:"The ID of the team to update"`
	Name        *string `json:"name,omitempty" jsonschema:"New name for the team"`
	Description *string `json:"description,omitempty" jsonschema:"New description for the team"`
}

type DeleteTeamParams struct {
	TeamID int `json:"team_id" jsonschema:"The ID of the team to delete"`
}

type ListTeamMembersParams struct {
	TeamID int `json:"team_id" jsonschema:"The ID of the team to list members for"`
}

type AddTeamMemberParams struct {
	TeamID int    `json:"team_id" jsonschema:"The ID of the team to add the member to"`
	UserID int    `json:"user_id" jsonschema:"The ID of the user to add to the team"`
	Role   string `json:"role" jsonschema:"Role to assign: owner admin editor or member"`
}

type RemoveTeamMemberParams struct {
	TeamID int `json:"team_id" jsonschema:"The ID of the team to remove the member from"`
	UserID int `json:"user_id" jsonschema:"The ID of the user to remove from the team"`
}

type LinkSourceToTeamParams struct {
	TeamID   int `json:"team_id" jsonschema:"The ID of the team to link the source to"`
	SourceID int `json:"source_id" jsonschema:"The ID of the source to link to the team"`
}

type UnlinkSourceFromTeamParams struct {
	TeamID   int `json:"team_id" jsonschema:"The ID of the team to unlink the source from"`
	SourceID int `json:"source_id" jsonschema:"The ID of the source to unlink from the team"`
}

type ListAllUsersParams struct{}

type GetUserParams struct {
	UserID int `json:"user_id" jsonschema:"The ID of the user to retrieve"`
}

type CreateUserParams struct {
	Email    string `json:"email" jsonschema:"Email address of the user"`
	FullName string `json:"full_name" jsonschema:"Full name of the user"`
	Role     string `json:"role" jsonschema:"Role of the user: admin or member"`
	Status   string `json:"status" jsonschema:"Status of the user: active or inactive"`
}

type UpdateUserParams struct {
	UserID   int     `json:"user_id" jsonschema:"The ID of the user to update"`
	Email    *string `json:"email,omitempty" jsonschema:"New email address"`
	FullName *string `json:"full_name,omitempty" jsonschema:"New full name"`
	Role     *string `json:"role,omitempty" jsonschema:"New role: admin or member"`
	Status   *string `json:"status,omitempty" jsonschema:"New status: active or inactive"`
}

type DeleteUserParams struct {
	UserID int `json:"user_id" jsonschema:"The ID of the user to delete"`
}

type ListAPITokensParams struct{}

type CreateAPITokenParams struct {
	Name      string  `json:"name" jsonschema:"Name for the API token"`
	ExpiresAt *string `json:"expires_at,omitempty" jsonschema:"Optional expiration date in ISO 8601 format"`
}

type DeleteAPITokenParams struct {
	TokenID int `json:"token_id" jsonschema:"The ID of the API token to delete"`
}

type ListAllSourcesParams struct{}

type CreateSourceParams struct {
	Name              string                   `json:"name" jsonschema:"Name of the source"`
	Description       string                   `json:"description,omitempty" jsonschema:"Optional description of the source"`
	Host              string                   `json:"host" jsonschema:"ClickHouse host"`
	Database          string                   `json:"database" jsonschema:"ClickHouse database name"`
	TableName         string                   `json:"table_name" jsonschema:"ClickHouse table name"`
	MetaIsAutoCreated bool                     `json:"_meta_is_auto_created" jsonschema:"Whether the table should be auto-created"`
	MetaTsField       string                   `json:"_meta_ts_field" jsonschema:"Timestamp field name (defaults to timestamp)"`
	MetaSeverityField string                   `json:"_meta_severity_field,omitempty" jsonschema:"Optional severity field name"`
	TTLDays           int                      `json:"ttl_days" jsonschema:"Time-to-live in days for log data"`
	Schema            []map[string]interface{} `json:"schema,omitempty" jsonschema:"Optional table schema for auto-creation"`
}

type ValidateSourceConnectionParams struct {
	Host           string `json:"host" jsonschema:"ClickHouse host"`
	Database       string `json:"database" jsonschema:"ClickHouse database name"`
	TableName      string `json:"table_name" jsonschema:"ClickHouse table name"`
	TimestampField string `json:"timestamp_field,omitempty" jsonschema:"Optional timestamp field to validate"`
	SeverityField  string `json:"severity_field,omitempty" jsonschema:"Optional severity field to validate"`
}

type DeleteSourceParams struct {
	SourceID int `json:"source_id" jsonschema:"The ID of the source to delete"`
}

type GetAdminSourceStatsParams struct {
	SourceID int `json:"source_id" jsonschema:"The ID of the source to get statistics for"`
}

// --- Output schemas ---

type AdminTeamResult struct {
	ID          int    `json:"id" jsonschema:"Team ID"`
	Name        string `json:"name" jsonschema:"Team name"`
	Description string `json:"description" jsonschema:"Team description"`
	MemberCount int    `json:"member_count" jsonschema:"Number of members"`
	CreatedAt   string `json:"created_at" jsonschema:"Creation timestamp"`
	UpdatedAt   string `json:"updated_at" jsonschema:"Last update timestamp"`
}

type TeamMemberResult struct {
	TeamID   int    `json:"team_id" jsonschema:"Team ID"`
	UserID   int    `json:"user_id" jsonschema:"User ID"`
	Role     string `json:"role" jsonschema:"Member role in the team"`
	Email    string `json:"email" jsonschema:"Member email"`
	FullName string `json:"full_name" jsonschema:"Member full name"`
}

type AdminUserResult struct {
	ID          int     `json:"id" jsonschema:"User ID"`
	Email       string  `json:"email" jsonschema:"User email"`
	FullName    string  `json:"full_name" jsonschema:"User full name"`
	Role        string  `json:"role" jsonschema:"User role"`
	Status      string  `json:"status" jsonschema:"User status"`
	LastLoginAt *string `json:"last_login_at,omitempty" jsonschema:"Last login timestamp"`
	CreatedAt   string  `json:"created_at" jsonschema:"Creation timestamp"`
}

type APITokenResult struct {
	ID         int     `json:"id" jsonschema:"Token ID"`
	Name       string  `json:"name" jsonschema:"Token name"`
	Prefix     string  `json:"prefix" jsonschema:"Token prefix"`
	LastUsedAt *string `json:"last_used_at,omitempty" jsonschema:"Last used timestamp"`
	ExpiresAt  *string `json:"expires_at,omitempty" jsonschema:"Expiration timestamp"`
	CreatedAt  string  `json:"created_at" jsonschema:"Creation timestamp"`
}

type APITokenCreateResult struct {
	Token string         `json:"token" jsonschema:"The full API token value (only shown once)"`
	Info  APITokenResult `json:"info" jsonschema:"Token metadata"`
}

type ValidationResult struct {
	IsValid      bool            `json:"is_valid" jsonschema:"Whether the connection is valid"`
	Message      string          `json:"message" jsonschema:"Validation message"`
	ErrorDetails []string        `json:"error_details,omitempty" jsonschema:"Detailed error messages if validation failed"`
	TableExists  bool            `json:"table_exists" jsonschema:"Whether the table exists"`
	ColumnChecks map[string]bool `json:"column_checks,omitempty" jsonschema:"Per-column validation results"`
}

// Helper function to check if user has admin role
func checkAdminRole(ctx context.Context, c *client.Client) error {
	profile, err := c.GetProfile(ctx)
	if err != nil {
		return fmt.Errorf("failed to get user profile: %w", err)
	}
	if profile.Data.User.Role != "admin" {
		return fmt.Errorf("access denied: admin role required")
	}
	return nil
}

// --- Conversion helpers ---

func teamToAdminResult(t client.Team) AdminTeamResult {
	return AdminTeamResult{
		ID: t.ID, Name: t.Name, Description: t.Description,
		MemberCount: t.MemberCount, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}
}

func memberToResult(m client.TeamMember) TeamMemberResult {
	return TeamMemberResult{
		TeamID: m.TeamID, UserID: m.UserID, Role: m.Role,
		Email: m.Email, FullName: m.FullName,
	}
}

func userToResult(u client.User) AdminUserResult {
	return AdminUserResult{
		ID: u.ID, Email: u.Email, FullName: u.FullName,
		Role: u.Role, Status: u.Status, LastLoginAt: u.LastLoginAt,
		CreatedAt: u.CreatedAt,
	}
}

func tokenToResult(t client.APIToken) APITokenResult {
	return APITokenResult{
		ID: t.ID, Name: t.Name, Prefix: t.Prefix,
		LastUsedAt: t.LastUsedAt, ExpiresAt: t.ExpiresAt, CreatedAt: t.CreatedAt,
	}
}

// --- Team management handlers ---

func handleListAllTeams(ctx context.Context, request mcp.CallToolRequest, args ListAllTeamsParams) ([]AdminTeamResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("logchef client not configured")
	}
	if err := checkAdminRole(ctx, c); err != nil {
		return nil, err
	}
	teams, err := c.ListAllTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all teams: %w", err)
	}
	result := make([]AdminTeamResult, len(teams.Data))
	for i, t := range teams.Data {
		result[i] = teamToAdminResult(t)
	}
	return result, nil
}

func handleGetTeam(ctx context.Context, request mcp.CallToolRequest, args GetTeamParams) (AdminTeamResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return AdminTeamResult{}, fmt.Errorf("logchef client not configured")
	}
	team, err := c.GetTeamByID(ctx, args.TeamID)
	if err != nil {
		return AdminTeamResult{}, fmt.Errorf("get team: %w", err)
	}
	return teamToAdminResult(team.Data), nil
}

func handleCreateTeam(ctx context.Context, request mcp.CallToolRequest, args CreateTeamParams) (AdminTeamResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return AdminTeamResult{}, fmt.Errorf("logchef client not configured")
	}
	if err := checkAdminRole(ctx, c); err != nil {
		return AdminTeamResult{}, err
	}
	team, err := c.CreateTeam(ctx, client.TeamRequest{Name: args.Name, Description: args.Description})
	if err != nil {
		return AdminTeamResult{}, fmt.Errorf("create team: %w", err)
	}
	return teamToAdminResult(team.Data), nil
}

func handleUpdateTeam(ctx context.Context, request mcp.CallToolRequest, args UpdateTeamParams) (AdminTeamResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return AdminTeamResult{}, fmt.Errorf("logchef client not configured")
	}
	team, err := c.UpdateTeam(ctx, args.TeamID, client.TeamUpdateRequest{Name: args.Name, Description: args.Description})
	if err != nil {
		return AdminTeamResult{}, fmt.Errorf("update team: %w", err)
	}
	return teamToAdminResult(team.Data), nil
}

func handleDeleteTeam(ctx context.Context, request mcp.CallToolRequest, args DeleteTeamParams) (SuccessResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return SuccessResult{}, fmt.Errorf("logchef client not configured")
	}
	if err := checkAdminRole(ctx, c); err != nil {
		return SuccessResult{}, err
	}
	if err := c.DeleteTeam(ctx, args.TeamID); err != nil {
		return SuccessResult{}, fmt.Errorf("delete team: %w", err)
	}
	return SuccessResult{Success: true, Message: "Team deleted successfully"}, nil
}

func handleListTeamMembers(ctx context.Context, request mcp.CallToolRequest, args ListTeamMembersParams) ([]TeamMemberResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("logchef client not configured")
	}
	members, err := c.ListTeamMembers(ctx, args.TeamID)
	if err != nil {
		return nil, fmt.Errorf("list team members: %w", err)
	}
	result := make([]TeamMemberResult, len(members.Data))
	for i, m := range members.Data {
		result[i] = memberToResult(m)
	}
	return result, nil
}

func handleAddTeamMember(ctx context.Context, request mcp.CallToolRequest, args AddTeamMemberParams) (SuccessResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return SuccessResult{}, fmt.Errorf("logchef client not configured")
	}
	if err := c.AddTeamMember(ctx, args.TeamID, client.TeamMemberRequest{UserID: args.UserID, Role: args.Role}); err != nil {
		return SuccessResult{}, fmt.Errorf("add team member: %w", err)
	}
	return SuccessResult{Success: true, Message: "Team member added successfully"}, nil
}

func handleRemoveTeamMember(ctx context.Context, request mcp.CallToolRequest, args RemoveTeamMemberParams) (SuccessResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return SuccessResult{}, fmt.Errorf("logchef client not configured")
	}
	if err := c.RemoveTeamMember(ctx, args.TeamID, args.UserID); err != nil {
		return SuccessResult{}, fmt.Errorf("remove team member: %w", err)
	}
	return SuccessResult{Success: true, Message: "Team member removed successfully"}, nil
}

func handleLinkSourceToTeam(ctx context.Context, request mcp.CallToolRequest, args LinkSourceToTeamParams) (SuccessResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return SuccessResult{}, fmt.Errorf("logchef client not configured")
	}
	if err := c.LinkSourceToTeam(ctx, args.TeamID, client.TeamSourceRequest{SourceID: args.SourceID}); err != nil {
		return SuccessResult{}, fmt.Errorf("link source to team: %w", err)
	}
	return SuccessResult{Success: true, Message: "Source linked to team successfully"}, nil
}

func handleUnlinkSourceFromTeam(ctx context.Context, request mcp.CallToolRequest, args UnlinkSourceFromTeamParams) (SuccessResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return SuccessResult{}, fmt.Errorf("logchef client not configured")
	}
	if err := c.UnlinkSourceFromTeam(ctx, args.TeamID, args.SourceID); err != nil {
		return SuccessResult{}, fmt.Errorf("unlink source from team: %w", err)
	}
	return SuccessResult{Success: true, Message: "Source unlinked from team successfully"}, nil
}

// --- User management handlers ---

func handleListAllUsers(ctx context.Context, request mcp.CallToolRequest, args ListAllUsersParams) ([]AdminUserResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("logchef client not configured")
	}
	if err := checkAdminRole(ctx, c); err != nil {
		return nil, err
	}
	users, err := c.ListAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all users: %w", err)
	}
	result := make([]AdminUserResult, len(users.Data))
	for i, u := range users.Data {
		result[i] = userToResult(u)
	}
	return result, nil
}

func handleGetUser(ctx context.Context, request mcp.CallToolRequest, args GetUserParams) (AdminUserResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return AdminUserResult{}, fmt.Errorf("logchef client not configured")
	}
	if err := checkAdminRole(ctx, c); err != nil {
		return AdminUserResult{}, err
	}
	user, err := c.GetUserByID(ctx, args.UserID)
	if err != nil {
		return AdminUserResult{}, fmt.Errorf("get user: %w", err)
	}
	return userToResult(user.Data), nil
}

func handleCreateUser(ctx context.Context, request mcp.CallToolRequest, args CreateUserParams) (AdminUserResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return AdminUserResult{}, fmt.Errorf("logchef client not configured")
	}
	if err := checkAdminRole(ctx, c); err != nil {
		return AdminUserResult{}, err
	}
	user, err := c.CreateUser(ctx, client.UserRequest{
		Email: args.Email, FullName: args.FullName, Role: args.Role, Status: args.Status,
	})
	if err != nil {
		return AdminUserResult{}, fmt.Errorf("create user: %w", err)
	}
	return userToResult(user.Data), nil
}

func handleUpdateUser(ctx context.Context, request mcp.CallToolRequest, args UpdateUserParams) (AdminUserResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return AdminUserResult{}, fmt.Errorf("logchef client not configured")
	}
	if err := checkAdminRole(ctx, c); err != nil {
		return AdminUserResult{}, err
	}
	user, err := c.UpdateUser(ctx, args.UserID, client.UserUpdateRequest{
		Email: args.Email, FullName: args.FullName, Role: args.Role, Status: args.Status,
	})
	if err != nil {
		return AdminUserResult{}, fmt.Errorf("update user: %w", err)
	}
	return userToResult(user.Data), nil
}

func handleDeleteUser(ctx context.Context, request mcp.CallToolRequest, args DeleteUserParams) (SuccessResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return SuccessResult{}, fmt.Errorf("logchef client not configured")
	}
	if err := checkAdminRole(ctx, c); err != nil {
		return SuccessResult{}, err
	}
	if err := c.DeleteUser(ctx, args.UserID); err != nil {
		return SuccessResult{}, fmt.Errorf("delete user: %w", err)
	}
	return SuccessResult{Success: true, Message: "User deleted successfully"}, nil
}

// --- API Token handlers ---

func handleListAPITokens(ctx context.Context, request mcp.CallToolRequest, args ListAPITokensParams) ([]APITokenResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("logchef client not configured")
	}
	tokens, err := c.ListAPITokens(ctx)
	if err != nil {
		return nil, fmt.Errorf("list API tokens: %w", err)
	}
	result := make([]APITokenResult, len(tokens.Data))
	for i, t := range tokens.Data {
		result[i] = tokenToResult(t)
	}
	return result, nil
}

func handleCreateAPIToken(ctx context.Context, request mcp.CallToolRequest, args CreateAPITokenParams) (APITokenCreateResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return APITokenCreateResult{}, fmt.Errorf("logchef client not configured")
	}
	token, err := c.CreateAPIToken(ctx, client.APITokenRequest{Name: args.Name, ExpiresAt: args.ExpiresAt})
	if err != nil {
		return APITokenCreateResult{}, fmt.Errorf("create API token: %w", err)
	}
	return APITokenCreateResult{
		Token: token.Data.Token,
		Info:  tokenToResult(token.Data.APIToken),
	}, nil
}

func handleDeleteAPIToken(ctx context.Context, request mcp.CallToolRequest, args DeleteAPITokenParams) (SuccessResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return SuccessResult{}, fmt.Errorf("logchef client not configured")
	}
	if err := c.DeleteAPIToken(ctx, args.TokenID); err != nil {
		return SuccessResult{}, fmt.Errorf("delete API token: %w", err)
	}
	return SuccessResult{Success: true, Message: "API token deleted successfully"}, nil
}

// --- Admin source handlers ---

func handleListAllSources(ctx context.Context, request mcp.CallToolRequest, args ListAllSourcesParams) ([]SourceResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("logchef client not configured")
	}
	if err := checkAdminRole(ctx, c); err != nil {
		return nil, err
	}
	sources, err := c.ListAllSources(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all sources: %w", err)
	}
	result := make([]SourceResult, len(sources.Data))
	for i, s := range sources.Data {
		result[i] = SourceResult{
			ID: s.ID, Name: s.Name, Description: s.Description,
			Connection: ConnectionResult{Host: s.Connection.Host, Database: s.Connection.Database, TableName: s.Connection.TableName},
			TsField: s.MetaTsField, IsConnected: s.IsConnected, TTLDays: s.TTLDays, CreatedAt: s.CreatedAt,
		}
	}
	return result, nil
}

func handleCreateSource(ctx context.Context, request mcp.CallToolRequest, args CreateSourceParams) (SourceResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return SourceResult{}, fmt.Errorf("logchef client not configured")
	}
	if err := checkAdminRole(ctx, c); err != nil {
		return SourceResult{}, err
	}

	var schema []client.LogColumn
	if args.Schema != nil {
		schema = make([]client.LogColumn, 0, len(args.Schema))
		for _, col := range args.Schema {
			name, _ := col["name"].(string)
			colType, _ := col["type"].(string)
			if name != "" && colType != "" {
				schema = append(schema, client.LogColumn{Name: name, Type: colType})
			}
		}
	}

	source, err := c.CreateSource(ctx, client.SourceRequest{
		Name: args.Name, Description: args.Description,
		Connection:        client.SourceConnection{Host: args.Host, Database: args.Database, TableName: args.TableName},
		MetaIsAutoCreated: args.MetaIsAutoCreated, MetaTsField: args.MetaTsField, MetaSeverityField: args.MetaSeverityField,
		TTLDays: args.TTLDays, Schema: schema,
	})
	if err != nil {
		return SourceResult{}, fmt.Errorf("create source: %w", err)
	}
	s := source.Data
	return SourceResult{
		ID: s.ID, Name: s.Name, Description: s.Description,
		Connection: ConnectionResult{Host: s.Connection.Host, Database: s.Connection.Database, TableName: s.Connection.TableName},
		TsField: s.MetaTsField, IsConnected: s.IsConnected, TTLDays: s.TTLDays, CreatedAt: s.CreatedAt,
	}, nil
}

func handleValidateSourceConnection(ctx context.Context, request mcp.CallToolRequest, args ValidateSourceConnectionParams) (ValidationResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return ValidationResult{}, fmt.Errorf("logchef client not configured")
	}
	if err := checkAdminRole(ctx, c); err != nil {
		return ValidationResult{}, err
	}
	v, err := c.ValidateSourceConnection(ctx, client.SourceValidationRequest{
		Host: args.Host, Database: args.Database, TableName: args.TableName,
		TimestampField: args.TimestampField, SeverityField: args.SeverityField,
	})
	if err != nil {
		return ValidationResult{}, fmt.Errorf("validate source connection: %w", err)
	}
	return ValidationResult{
		IsValid: v.Data.IsValid, Message: v.Data.Message,
		ErrorDetails: v.Data.ErrorDetails, TableExists: v.Data.TableExists,
		ColumnChecks: v.Data.ColumnChecks,
	}, nil
}

func handleDeleteSource(ctx context.Context, request mcp.CallToolRequest, args DeleteSourceParams) (SuccessResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return SuccessResult{}, fmt.Errorf("logchef client not configured")
	}
	if err := checkAdminRole(ctx, c); err != nil {
		return SuccessResult{}, err
	}
	if err := c.DeleteSource(ctx, args.SourceID); err != nil {
		return SuccessResult{}, fmt.Errorf("delete source: %w", err)
	}
	return SuccessResult{Success: true, Message: "Source deleted successfully"}, nil
}

func handleGetAdminSourceStats(ctx context.Context, request mcp.CallToolRequest, args GetAdminSourceStatsParams) (*mcp.CallToolResult, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return mcp.NewToolResultError("logchef client not configured"), nil
	}
	if err := checkAdminRole(ctx, c); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	stats, err := c.GetAdminSourceStats(ctx, args.SourceID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("get admin source stats: %v", err)), nil
	}
	// Stats have nested/dynamic structure — use typed handler with JSON output
	out, _ := json.MarshalIndent(stats.Data, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

func AddAdminTools(s *server.MCPServer) {
	// Team management tools
	s.AddTool(mcp.NewTool("list_all_teams",
		mcp.WithDescription("List all teams in the system (admin only). Returns all teams with details including member counts."),
		mcp.WithInputSchema[ListAllTeamsParams](),
		mcp.WithOutputSchema[[]AdminTeamResult](),
		mcp.WithTitleAnnotation("List All Teams"),
		mcp.WithReadOnlyHintAnnotation(true),
	), mcp.NewStructuredToolHandler(handleListAllTeams))

	s.AddTool(mcp.NewTool("get_team",
		mcp.WithDescription("Get detailed information about a specific team by ID. Available to team members and admins."),
		mcp.WithInputSchema[GetTeamParams](),
		mcp.WithOutputSchema[AdminTeamResult](),
		mcp.WithTitleAnnotation("Get Team"),
		mcp.WithReadOnlyHintAnnotation(true),
	), mcp.NewStructuredToolHandler(handleGetTeam))

	s.AddTool(mcp.NewTool("create_team",
		mcp.WithDescription("Create a new team (admin only). Provide a team name and optional description."),
		mcp.WithInputSchema[CreateTeamParams](),
		mcp.WithOutputSchema[AdminTeamResult](),
		mcp.WithTitleAnnotation("Create Team"),
		mcp.WithDestructiveHintAnnotation(false),
	), mcp.NewStructuredToolHandler(handleCreateTeam))

	s.AddTool(mcp.NewTool("update_team",
		mcp.WithDescription("Update an existing team's name and/or description. Requires team admin or global admin role."),
		mcp.WithInputSchema[UpdateTeamParams](),
		mcp.WithOutputSchema[AdminTeamResult](),
		mcp.WithTitleAnnotation("Update Team"),
		mcp.WithDestructiveHintAnnotation(false),
	), mcp.NewStructuredToolHandler(handleUpdateTeam))

	s.AddTool(mcp.NewTool("delete_team",
		mcp.WithDescription("Delete a team permanently (admin only). Cannot be undone — removes all team associations."),
		mcp.WithInputSchema[DeleteTeamParams](),
		mcp.WithOutputSchema[SuccessResult](),
		mcp.WithTitleAnnotation("Delete Team"),
		mcp.WithDestructiveHintAnnotation(true),
	), mcp.NewStructuredToolHandler(handleDeleteTeam))

	s.AddTool(mcp.NewTool("list_team_members",
		mcp.WithDescription("List all members of a specific team. Returns member details including roles and join dates."),
		mcp.WithInputSchema[ListTeamMembersParams](),
		mcp.WithOutputSchema[[]TeamMemberResult](),
		mcp.WithTitleAnnotation("List Team Members"),
		mcp.WithReadOnlyHintAnnotation(true),
	), mcp.NewStructuredToolHandler(handleListTeamMembers))

	s.AddTool(mcp.NewTool("add_team_member",
		mcp.WithDescription("Add a user to a team with a specific role. Requires team admin or global admin. Valid roles: owner, admin, editor, member."),
		mcp.WithInputSchema[AddTeamMemberParams](),
		mcp.WithOutputSchema[SuccessResult](),
		mcp.WithTitleAnnotation("Add Team Member"),
		mcp.WithDestructiveHintAnnotation(false),
	), mcp.NewStructuredToolHandler(handleAddTeamMember))

	s.AddTool(mcp.NewTool("remove_team_member",
		mcp.WithDescription("Remove a user from a team. Requires team admin or global admin. Revokes access to team resources."),
		mcp.WithInputSchema[RemoveTeamMemberParams](),
		mcp.WithOutputSchema[SuccessResult](),
		mcp.WithTitleAnnotation("Remove Team Member"),
		mcp.WithDestructiveHintAnnotation(true),
	), mcp.NewStructuredToolHandler(handleRemoveTeamMember))

	s.AddTool(mcp.NewTool("link_source_to_team",
		mcp.WithDescription("Link a log source to a team, granting team members access. Requires team admin or global admin."),
		mcp.WithInputSchema[LinkSourceToTeamParams](),
		mcp.WithOutputSchema[SuccessResult](),
		mcp.WithTitleAnnotation("Link Source to Team"),
		mcp.WithDestructiveHintAnnotation(false),
	), mcp.NewStructuredToolHandler(handleLinkSourceToTeam))

	s.AddTool(mcp.NewTool("unlink_source_from_team",
		mcp.WithDescription("Remove a log source from a team, revoking access. Requires team admin or global admin."),
		mcp.WithInputSchema[UnlinkSourceFromTeamParams](),
		mcp.WithOutputSchema[SuccessResult](),
		mcp.WithTitleAnnotation("Unlink Source from Team"),
		mcp.WithDestructiveHintAnnotation(true),
	), mcp.NewStructuredToolHandler(handleUnlinkSourceFromTeam))

	// User management tools (admin only)
	s.AddTool(mcp.NewTool("list_all_users",
		mcp.WithDescription("List all users in the system (admin only). Returns users with roles, status, and activity info."),
		mcp.WithInputSchema[ListAllUsersParams](),
		mcp.WithOutputSchema[[]AdminUserResult](),
		mcp.WithTitleAnnotation("List All Users"),
		mcp.WithReadOnlyHintAnnotation(true),
	), mcp.NewStructuredToolHandler(handleListAllUsers))

	s.AddTool(mcp.NewTool("get_user",
		mcp.WithDescription("Get detailed user information by ID (admin only). Returns email, role, status, and timestamps."),
		mcp.WithInputSchema[GetUserParams](),
		mcp.WithOutputSchema[AdminUserResult](),
		mcp.WithTitleAnnotation("Get User"),
		mcp.WithReadOnlyHintAnnotation(true),
	), mcp.NewStructuredToolHandler(handleGetUser))

	s.AddTool(mcp.NewTool("create_user",
		mcp.WithDescription("Create a new user (admin only). Provide email, full name, role (admin/member), and status (active/inactive)."),
		mcp.WithInputSchema[CreateUserParams](),
		mcp.WithOutputSchema[AdminUserResult](),
		mcp.WithTitleAnnotation("Create User"),
		mcp.WithDestructiveHintAnnotation(false),
	), mcp.NewStructuredToolHandler(handleCreateUser))

	s.AddTool(mcp.NewTool("update_user",
		mcp.WithDescription("Update a user's information (admin only). All fields are optional — provide only fields to change."),
		mcp.WithInputSchema[UpdateUserParams](),
		mcp.WithOutputSchema[AdminUserResult](),
		mcp.WithTitleAnnotation("Update User"),
		mcp.WithDestructiveHintAnnotation(false),
	), mcp.NewStructuredToolHandler(handleUpdateUser))

	s.AddTool(mcp.NewTool("delete_user",
		mcp.WithDescription("Delete a user permanently (admin only). Cannot be undone. Cannot delete the last admin user."),
		mcp.WithInputSchema[DeleteUserParams](),
		mcp.WithOutputSchema[SuccessResult](),
		mcp.WithTitleAnnotation("Delete User"),
		mcp.WithDestructiveHintAnnotation(true),
	), mcp.NewStructuredToolHandler(handleDeleteUser))

	// Source management tools (admin only)
	s.AddTool(mcp.NewTool("list_all_sources",
		mcp.WithDescription("List all log sources in the system (admin only). Returns sources with connection details and metadata."),
		mcp.WithInputSchema[ListAllSourcesParams](),
		mcp.WithOutputSchema[[]SourceResult](),
		mcp.WithTitleAnnotation("List All Sources"),
		mcp.WithReadOnlyHintAnnotation(true),
	), mcp.NewStructuredToolHandler(handleListAllSources))

	s.AddTool(mcp.NewTool("create_source",
		mcp.WithDescription("Create a new log source (admin only). Provide ClickHouse connection details, metadata config, and optional schema for auto-creation."),
		mcp.WithInputSchema[CreateSourceParams](),
		mcp.WithOutputSchema[SourceResult](),
		mcp.WithTitleAnnotation("Create Source"),
		mcp.WithDestructiveHintAnnotation(false),
	), mcp.NewStructuredToolHandler(handleCreateSource))

	s.AddTool(mcp.NewTool("validate_source_connection",
		mcp.WithDescription("Validate ClickHouse connection details before creating a source (admin only). Tests connectivity and table existence."),
		mcp.WithInputSchema[ValidateSourceConnectionParams](),
		mcp.WithOutputSchema[ValidationResult](),
		mcp.WithTitleAnnotation("Validate Source Connection"),
		mcp.WithReadOnlyHintAnnotation(true),
	), mcp.NewStructuredToolHandler(handleValidateSourceConnection))

	s.AddTool(mcp.NewTool("delete_source",
		mcp.WithDescription("Delete a log source permanently (admin only). Cannot be undone — removes all team associations."),
		mcp.WithInputSchema[DeleteSourceParams](),
		mcp.WithOutputSchema[SuccessResult](),
		mcp.WithTitleAnnotation("Delete Source"),
		mcp.WithDestructiveHintAnnotation(true),
	), mcp.NewStructuredToolHandler(handleDeleteSource))

	// Source stats uses typed handler (dynamic nested data)
	s.AddTool(mcp.NewTool("get_admin_source_stats",
		mcp.WithDescription("Get detailed statistics for a log source (admin only). Returns ClickHouse table stats including row count, sizes, compression ratio, and column statistics."),
		mcp.WithInputSchema[GetAdminSourceStatsParams](),
		mcp.WithTitleAnnotation("Get Source Statistics"),
		mcp.WithReadOnlyHintAnnotation(true),
	), mcp.NewTypedToolHandler(handleGetAdminSourceStats))

	// API token management tools
	s.AddTool(mcp.NewTool("list_api_tokens",
		mcp.WithDescription("List all API tokens for the current authenticated user. Returns token names, prefixes, last used, and expiration dates."),
		mcp.WithInputSchema[ListAPITokensParams](),
		mcp.WithOutputSchema[[]APITokenResult](),
		mcp.WithTitleAnnotation("List API Tokens"),
		mcp.WithReadOnlyHintAnnotation(true),
	), mcp.NewStructuredToolHandler(handleListAPITokens))

	s.AddTool(mcp.NewTool("create_api_token",
		mcp.WithDescription("Create a new API token for the current user. Returns the full token value (only shown once) and metadata."),
		mcp.WithInputSchema[CreateAPITokenParams](),
		mcp.WithOutputSchema[APITokenCreateResult](),
		mcp.WithTitleAnnotation("Create API Token"),
		mcp.WithDestructiveHintAnnotation(false),
	), mcp.NewStructuredToolHandler(handleCreateAPIToken))

	s.AddTool(mcp.NewTool("delete_api_token",
		mcp.WithDescription("Delete an API token. Immediately revokes access for any applications using it. Cannot be undone."),
		mcp.WithInputSchema[DeleteAPITokenParams](),
		mcp.WithOutputSchema[SuccessResult](),
		mcp.WithTitleAnnotation("Delete API Token"),
		mcp.WithDestructiveHintAnnotation(true),
	), mcp.NewStructuredToolHandler(handleDeleteAPIToken))
}
