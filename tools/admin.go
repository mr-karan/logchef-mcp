package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/server"

	mcplogchef "github.com/mr-karan/logchef-mcp"
	"github.com/mr-karan/logchef-mcp/client"
)

// Admin tool parameters and functions with role validation

// ListAllTeamsParams represents the parameters for listing all teams (admin only).
type ListAllTeamsParams struct{}

// GetTeamParams represents the parameters for getting a specific team.
type GetTeamParams struct {
	TeamID int `json:"team_id" jsonschema:"description=The ID of the team to retrieve,required"`
}

// CreateTeamParams represents the parameters for creating a team (admin only).
type CreateTeamParams struct {
	Name        string `json:"name" jsonschema:"description=Name of the team,required"`
	Description string `json:"description,omitempty" jsonschema:"description=Optional description of the team"`
}

// UpdateTeamParams represents the parameters for updating a team.
type UpdateTeamParams struct {
	TeamID      int     `json:"team_id" jsonschema:"description=The ID of the team to update,required"`
	Name        *string `json:"name,omitempty" jsonschema:"description=New name for the team (optional)"`
	Description *string `json:"description,omitempty" jsonschema:"description=New description for the team (optional)"`
}

// DeleteTeamParams represents the parameters for deleting a team (admin only).
type DeleteTeamParams struct {
	TeamID int `json:"team_id" jsonschema:"description=The ID of the team to delete,required"`
}

// ListTeamMembersParams represents the parameters for listing team members.
type ListTeamMembersParams struct {
	TeamID int `json:"team_id" jsonschema:"description=The ID of the team to list members for,required"`
}

// AddTeamMemberParams represents the parameters for adding a team member.
type AddTeamMemberParams struct {
	TeamID int    `json:"team_id" jsonschema:"description=The ID of the team to add the member to,required"`
	UserID int    `json:"user_id" jsonschema:"description=The ID of the user to add to the team,required"`
	Role   string `json:"role" jsonschema:"description=The role to assign to the user in the team. Valid values: owner\\\\, admin\\\\, editor\\\\, member,required"`
}

// RemoveTeamMemberParams represents the parameters for removing a team member.
type RemoveTeamMemberParams struct {
	TeamID int `json:"team_id" jsonschema:"description=The ID of the team to remove the member from,required"`
	UserID int `json:"user_id" jsonschema:"description=The ID of the user to remove from the team,required"`
}

// LinkSourceToTeamParams represents the parameters for linking a source to a team.
type LinkSourceToTeamParams struct {
	TeamID   int `json:"team_id" jsonschema:"description=The ID of the team to link the source to,required"`
	SourceID int `json:"source_id" jsonschema:"description=The ID of the source to link to the team,required"`
}

// UnlinkSourceFromTeamParams represents the parameters for unlinking a source from a team.
type UnlinkSourceFromTeamParams struct {
	TeamID   int `json:"team_id" jsonschema:"description=The ID of the team to unlink the source from,required"`
	SourceID int `json:"source_id" jsonschema:"description=The ID of the source to unlink from the team,required"`
}

// User management parameters

// ListAllUsersParams represents the parameters for listing all users (admin only).
type ListAllUsersParams struct{}

// GetUserParams represents the parameters for getting a specific user (admin only).
type GetUserParams struct {
	UserID int `json:"user_id" jsonschema:"description=The ID of the user to retrieve,required"`
}

// CreateUserParams represents the parameters for creating a user (admin only).
type CreateUserParams struct {
	Email    string `json:"email" jsonschema:"description=Email address of the user,required"`
	FullName string `json:"full_name" jsonschema:"description=Full name of the user,required"`
	Role     string `json:"role" jsonschema:"description=Role of the user. Valid values: admin\\\\, member,required"`
	Status   string `json:"status" jsonschema:"description=Status of the user. Valid values: active\\\\, inactive,required"`
}

// UpdateUserParams represents the parameters for updating a user (admin only).
type UpdateUserParams struct {
	UserID   int     `json:"user_id" jsonschema:"description=The ID of the user to update,required"`
	Email    *string `json:"email,omitempty" jsonschema:"description=New email address for the user (optional)"`
	FullName *string `json:"full_name,omitempty" jsonschema:"description=New full name for the user (optional)"`
	Role     *string `json:"role,omitempty" jsonschema:"description=New role for the user. Valid values: admin\\\\, member (optional)"`
	Status   *string `json:"status,omitempty" jsonschema:"description=New status for the user. Valid values: active\\\\, inactive (optional)"`
}

// DeleteUserParams represents the parameters for deleting a user (admin only).
type DeleteUserParams struct {
	UserID int `json:"user_id" jsonschema:"description=The ID of the user to delete,required"`
}

// API Token management parameters

// ListAPITokensParams represents the parameters for listing API tokens.
type ListAPITokensParams struct{}

// CreateAPITokenParams represents the parameters for creating an API token.
type CreateAPITokenParams struct {
	Name      string  `json:"name" jsonschema:"description=Name for the API token,required"`
	ExpiresAt *string `json:"expires_at,omitempty" jsonschema:"description=Optional expiration date for the token in ISO 8601 format (e.g. '2025-12-31T23:59:59Z')"`
}

// DeleteAPITokenParams represents the parameters for deleting an API token.
type DeleteAPITokenParams struct {
	TokenID int `json:"token_id" jsonschema:"description=The ID of the API token to delete,required"`
}

// Admin Source management parameters

// ListAllSourcesParams represents the parameters for listing all sources (admin only).
type ListAllSourcesParams struct{}

// CreateSourceParams represents the parameters for creating a source (admin only).
type CreateSourceParams struct {
	Name                 string                     `json:"name" jsonschema:"description=Name of the source,required"`
	Description          string                     `json:"description,omitempty" jsonschema:"description=Optional description of the source"`
	Host                 string                     `json:"host" jsonschema:"description=ClickHouse host,required"`
	Database             string                     `json:"database" jsonschema:"description=ClickHouse database name,required"`
	TableName            string                     `json:"table_name" jsonschema:"description=ClickHouse table name,required"`
	MetaIsAutoCreated    bool                       `json:"_meta_is_auto_created" jsonschema:"description=Whether the table should be auto-created if it doesn't exist"`
	MetaTsField          string                     `json:"_meta_ts_field" jsonschema:"description=Timestamp field name (defaults to 'timestamp'),required"`
	MetaSeverityField    string                     `json:"_meta_severity_field,omitempty" jsonschema:"description=Optional severity field name"`
	TTLDays              int                        `json:"ttl_days" jsonschema:"description=Time-to-live in days for log data,required"`
	Schema               []map[string]interface{}   `json:"schema,omitempty" jsonschema:"description=Optional table schema for auto-creation"`
}

// ValidateSourceConnectionParams represents the parameters for validating a source connection (admin only).
type ValidateSourceConnectionParams struct {
	Host           string `json:"host" jsonschema:"description=ClickHouse host,required"`
	Database       string `json:"database" jsonschema:"description=ClickHouse database name,required"`
	TableName      string `json:"table_name" jsonschema:"description=ClickHouse table name,required"`
	TimestampField string `json:"timestamp_field,omitempty" jsonschema:"description=Optional timestamp field to validate"`
	SeverityField  string `json:"severity_field,omitempty" jsonschema:"description=Optional severity field to validate"`
}

// DeleteSourceParams represents the parameters for deleting a source (admin only).
type DeleteSourceParams struct {
	SourceID int `json:"source_id" jsonschema:"description=The ID of the source to delete,required"`
}

// GetAdminSourceStatsParams represents the parameters for getting source statistics (admin only).
type GetAdminSourceStatsParams struct {
	SourceID int `json:"source_id" jsonschema:"description=The ID of the source to get statistics for,required"`
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

// Admin tool functions

func listAllTeams(ctx context.Context, args ListAllTeamsParams) (*client.TeamsListResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	// Check admin role
	if err := checkAdminRole(ctx, c); err != nil {
		return nil, err
	}

	teams, err := c.ListAllTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all teams: %w", err)
	}

	return teams, nil
}

func getTeam(ctx context.Context, args GetTeamParams) (*client.TeamResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	team, err := c.GetTeamByID(ctx, args.TeamID)
	if err != nil {
		return nil, fmt.Errorf("get team: %w", err)
	}

	return team, nil
}

func createTeam(ctx context.Context, args CreateTeamParams) (*client.TeamResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	// Check admin role
	if err := checkAdminRole(ctx, c); err != nil {
		return nil, err
	}

	request := client.TeamRequest{
		Name:        args.Name,
		Description: args.Description,
	}

	team, err := c.CreateTeam(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("create team: %w", err)
	}

	return team, nil
}

func updateTeam(ctx context.Context, args UpdateTeamParams) (*client.TeamResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	request := client.TeamUpdateRequest{
		Name:        args.Name,
		Description: args.Description,
	}

	team, err := c.UpdateTeam(ctx, args.TeamID, request)
	if err != nil {
		return nil, fmt.Errorf("update team: %w", err)
	}

	return team, nil
}

type DeleteTeamResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func deleteTeam(ctx context.Context, args DeleteTeamParams) (*DeleteTeamResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	// Check admin role
	if err := checkAdminRole(ctx, c); err != nil {
		return nil, err
	}

	err := c.DeleteTeam(ctx, args.TeamID)
	if err != nil {
		return nil, fmt.Errorf("delete team: %w", err)
	}

	return &DeleteTeamResponse{
		Success: true,
		Message: "Team deleted successfully",
	}, nil
}

func listTeamMembers(ctx context.Context, args ListTeamMembersParams) (*client.TeamMembersResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	members, err := c.ListTeamMembers(ctx, args.TeamID)
	if err != nil {
		return nil, fmt.Errorf("list team members: %w", err)
	}

	return members, nil
}

type AddTeamMemberResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func addTeamMember(ctx context.Context, args AddTeamMemberParams) (*AddTeamMemberResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	request := client.TeamMemberRequest{
		UserID: args.UserID,
		Role:   args.Role,
	}

	err := c.AddTeamMember(ctx, args.TeamID, request)
	if err != nil {
		return nil, fmt.Errorf("add team member: %w", err)
	}

	return &AddTeamMemberResponse{
		Success: true,
		Message: "Team member added successfully",
	}, nil
}

type RemoveTeamMemberResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func removeTeamMember(ctx context.Context, args RemoveTeamMemberParams) (*RemoveTeamMemberResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	err := c.RemoveTeamMember(ctx, args.TeamID, args.UserID)
	if err != nil {
		return nil, fmt.Errorf("remove team member: %w", err)
	}

	return &RemoveTeamMemberResponse{
		Success: true,
		Message: "Team member removed successfully",
	}, nil
}

type LinkSourceResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func linkSourceToTeam(ctx context.Context, args LinkSourceToTeamParams) (*LinkSourceResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	request := client.TeamSourceRequest{
		SourceID: args.SourceID,
	}

	err := c.LinkSourceToTeam(ctx, args.TeamID, request)
	if err != nil {
		return nil, fmt.Errorf("link source to team: %w", err)
	}

	return &LinkSourceResponse{
		Success: true,
		Message: "Source linked to team successfully",
	}, nil
}

type UnlinkSourceResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func unlinkSourceFromTeam(ctx context.Context, args UnlinkSourceFromTeamParams) (*UnlinkSourceResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	err := c.UnlinkSourceFromTeam(ctx, args.TeamID, args.SourceID)
	if err != nil {
		return nil, fmt.Errorf("unlink source from team: %w", err)
	}

	return &UnlinkSourceResponse{
		Success: true,
		Message: "Source unlinked from team successfully",
	}, nil
}

// User management tool functions

func listAllUsers(ctx context.Context, args ListAllUsersParams) (*client.UsersListResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	// Check admin role
	if err := checkAdminRole(ctx, c); err != nil {
		return nil, err
	}

	users, err := c.ListAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all users: %w", err)
	}

	return users, nil
}

func getUser(ctx context.Context, args GetUserParams) (*client.UserResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	// Check admin role
	if err := checkAdminRole(ctx, c); err != nil {
		return nil, err
	}

	user, err := c.GetUserByID(ctx, args.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	return user, nil
}

func createUser(ctx context.Context, args CreateUserParams) (*client.UserResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	// Check admin role
	if err := checkAdminRole(ctx, c); err != nil {
		return nil, err
	}

	request := client.UserRequest{
		Email:    args.Email,
		FullName: args.FullName,
		Role:     args.Role,
		Status:   args.Status,
	}

	user, err := c.CreateUser(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func updateUser(ctx context.Context, args UpdateUserParams) (*client.UserResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	// Check admin role
	if err := checkAdminRole(ctx, c); err != nil {
		return nil, err
	}

	request := client.UserUpdateRequest{
		Email:    args.Email,
		FullName: args.FullName,
		Role:     args.Role,
		Status:   args.Status,
	}

	user, err := c.UpdateUser(ctx, args.UserID, request)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	return user, nil
}

type DeleteUserResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func deleteUser(ctx context.Context, args DeleteUserParams) (*DeleteUserResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	// Check admin role
	if err := checkAdminRole(ctx, c); err != nil {
		return nil, err
	}

	err := c.DeleteUser(ctx, args.UserID)
	if err != nil {
		return nil, fmt.Errorf("delete user: %w", err)
	}

	return &DeleteUserResponse{
		Success: true,
		Message: "User deleted successfully",
	}, nil
}

// API Token management tool functions

func listAPITokens(ctx context.Context, args ListAPITokensParams) (*client.APITokensResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	tokens, err := c.ListAPITokens(ctx)
	if err != nil {
		return nil, fmt.Errorf("list API tokens: %w", err)
	}

	return tokens, nil
}

func createAPIToken(ctx context.Context, args CreateAPITokenParams) (*client.APITokenResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	request := client.APITokenRequest{
		Name:      args.Name,
		ExpiresAt: args.ExpiresAt,
	}

	token, err := c.CreateAPIToken(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("create API token: %w", err)
	}

	return token, nil
}

type DeleteAPITokenResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func deleteAPIToken(ctx context.Context, args DeleteAPITokenParams) (*DeleteAPITokenResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	err := c.DeleteAPIToken(ctx, args.TokenID)
	if err != nil {
		return nil, fmt.Errorf("delete API token: %w", err)
	}

	return &DeleteAPITokenResponse{
		Success: true,
		Message: "API token deleted successfully",
	}, nil
}

// Admin source management tool functions

func listAllSources(ctx context.Context, args ListAllSourcesParams) (*client.SourcesListResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	// Check admin role
	if err := checkAdminRole(ctx, c); err != nil {
		return nil, err
	}

	sources, err := c.ListAllSources(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all sources: %w", err)
	}

	return sources, nil
}

func createSource(ctx context.Context, args CreateSourceParams) (*client.AdminSourceResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	// Check admin role
	if err := checkAdminRole(ctx, c); err != nil {
		return nil, err
	}

	// Convert schema from map to LogColumn if provided
	var schema []client.LogColumn
	if args.Schema != nil {
		schema = make([]client.LogColumn, len(args.Schema))
		for i, col := range args.Schema {
			if name, ok := col["name"].(string); ok {
				if colType, ok := col["type"].(string); ok {
					schema[i] = client.LogColumn{
						Name: name,
						Type: colType,
					}
				}
			}
		}
	}

	request := client.SourceRequest{
		Name:        args.Name,
		Description: args.Description,
		Connection: client.SourceConnection{
			Host:      args.Host,
			Database:  args.Database,
			TableName: args.TableName,
		},
		MetaIsAutoCreated: args.MetaIsAutoCreated,
		MetaTsField:       args.MetaTsField,
		MetaSeverityField: args.MetaSeverityField,
		TTLDays:           args.TTLDays,
		Schema:            schema,
	}

	source, err := c.CreateSource(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("create source: %w", err)
	}

	return source, nil
}

func validateSourceConnection(ctx context.Context, args ValidateSourceConnectionParams) (*client.SourceValidationResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	// Check admin role
	if err := checkAdminRole(ctx, c); err != nil {
		return nil, err
	}

	request := client.SourceValidationRequest{
		Host:           args.Host,
		Database:       args.Database,
		TableName:      args.TableName,
		TimestampField: args.TimestampField,
		SeverityField:  args.SeverityField,
	}

	validation, err := c.ValidateSourceConnection(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("validate source connection: %w", err)
	}

	return validation, nil
}

type DeleteSourceResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func deleteSource(ctx context.Context, args DeleteSourceParams) (*DeleteSourceResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	// Check admin role
	if err := checkAdminRole(ctx, c); err != nil {
		return nil, err
	}

	err := c.DeleteSource(ctx, args.SourceID)
	if err != nil {
		return nil, fmt.Errorf("delete source: %w", err)
	}

	return &DeleteSourceResponse{
		Success: true,
		Message: "Source deleted successfully",
	}, nil
}

func getAdminSourceStats(ctx context.Context, args GetAdminSourceStatsParams) (*client.SourceStatsResponse, error) {
	c := mcplogchef.LogchefClientFromContext(ctx)
	if c == nil {
		return nil, fmt.Errorf("Logchef client not found in context")
	}

	// Check admin role
	if err := checkAdminRole(ctx, c); err != nil {
		return nil, err
	}

	stats, err := c.GetAdminSourceStats(ctx, args.SourceID)
	if err != nil {
		return nil, fmt.Errorf("get admin source stats: %w", err)
	}

	return stats, nil
}

// Tool definitions

var ListAllTeams = mcplogchef.MustTool(
	"list_all_teams",
	"List all teams in the system. This is an admin-only operation that requires admin role privileges. Returns a list of all teams with their details including member counts.",
	listAllTeams,
)

var GetTeam = mcplogchef.MustTool(
	"get_team",
	"Get detailed information about a specific team by ID. Returns team details including name, description, member count, and timestamps. Available to team members and admins.",
	getTeam,
)

var CreateTeam = mcplogchef.MustTool(
	"create_team",
	"Create a new team. This is an admin-only operation that requires admin role privileges. Provide a team name and optional description to create a new team.",
	createTeam,
)

var UpdateTeam = mcplogchef.MustTool(
	"update_team",
	"Update an existing team's name and/or description. Requires team admin privileges or global admin role. Provide the team ID and the fields you want to update.",
	updateTeam,
)

var DeleteTeam = mcplogchef.MustTool(
	"delete_team",
	"Delete a team permanently. This is an admin-only operation that requires admin role privileges. This action cannot be undone and will remove all team associations.",
	deleteTeam,
)

var ListTeamMembers = mcplogchef.MustTool(
	"list_team_members",
	"List all members of a specific team. Returns member details including user information, roles, and join dates. Available to team members and admins.",
	listTeamMembers,
)

var AddTeamMember = mcplogchef.MustTool(
	"add_team_member",
	"Add a user to a team with a specific role. Requires team admin privileges or global admin role. Valid roles are: owner, admin, editor, member.",
	addTeamMember,
)

var RemoveTeamMember = mcplogchef.MustTool(
	"remove_team_member",
	"Remove a user from a team. Requires team admin privileges or global admin role. This will revoke the user's access to team resources.",
	removeTeamMember,
)

var LinkSourceToTeam = mcplogchef.MustTool(
	"link_source_to_team",
	"Link a log source to a team, granting team members access to query logs from that source. Requires team admin privileges or global admin role.",
	linkSourceToTeam,
)

var UnlinkSourceFromTeam = mcplogchef.MustTool(
	"unlink_source_from_team",
	"Remove a log source from a team, revoking team members' access to that source. Requires team admin privileges or global admin role.",
	unlinkSourceFromTeam,
)

var ListAllUsers = mcplogchef.MustTool(
	"list_all_users",
	"List all users in the system. This is an admin-only operation that requires admin role privileges. Returns a list of all users with their details including roles, status, and activity information.",
	listAllUsers,
)

var GetUser = mcplogchef.MustTool(
	"get_user",
	"Get detailed information about a specific user by ID. This is an admin-only operation that requires admin role privileges. Returns user details including email, role, status, and timestamps.",
	getUser,
)

var CreateUser = mcplogchef.MustTool(
	"create_user",
	"Create a new user in the system. This is an admin-only operation that requires admin role privileges. Provide email, full name, role (admin/member), and status (active/inactive).",
	createUser,
)

var UpdateUser = mcplogchef.MustTool(
	"update_user",
	"Update an existing user's information. This is an admin-only operation that requires admin role privileges. You can update email, full name, role, and status. All fields are optional - provide only the fields you want to change.",
	updateUser,
)

var DeleteUser = mcplogchef.MustTool(
	"delete_user",
	"Delete a user from the system. This is an admin-only operation that requires admin role privileges. This action cannot be undone and will remove all user associations. Cannot delete the last admin user.",
	deleteUser,
)

var ListAPITokens = mcplogchef.MustTool(
	"list_api_tokens",
	"List all API tokens for the current authenticated user. Returns token details including names, prefixes, last used timestamps, and expiration dates. Does not require admin privileges.",
	listAPITokens,
)

var CreateAPIToken = mcplogchef.MustTool(
	"create_api_token",
	"Create a new API token for the current authenticated user. Provide a name for the token and optionally an expiration date. Returns the full token value (only shown once) and token metadata.",
	createAPIToken,
)

var DeleteAPIToken = mcplogchef.MustTool(
	"delete_api_token",
	"Delete an API token belonging to the current authenticated user. This will immediately revoke access for any applications using this token. This action cannot be undone.",
	deleteAPIToken,
)

var ListAllSources = mcplogchef.MustTool(
	"list_all_sources",
	"List all log sources in the system. This is an admin-only operation that requires admin role privileges. Returns a list of all sources with their connection details, status, and metadata.",
	listAllSources,
)

var CreateSource = mcplogchef.MustTool(
	"create_source",
	"Create a new log source in the system. This is an admin-only operation that requires admin role privileges. Provide ClickHouse connection details, metadata configuration, and optional table schema for auto-creation.",
	createSource,
)

var ValidateSourceConnection = mcplogchef.MustTool(
	"validate_source_connection",
	"Validate ClickHouse connection details before creating a source. This is an admin-only operation that requires admin role privileges. Tests connectivity, table existence, and optionally validates timestamp and severity fields.",
	validateSourceConnection,
)

var DeleteSource = mcplogchef.MustTool(
	"delete_source",
	"Delete a log source from the system. This is an admin-only operation that requires admin role privileges. This action cannot be undone and will remove all source associations with teams.",
	deleteSource,
)

var GetAdminSourceStats = mcplogchef.MustTool(
	"get_admin_source_stats",
	"Get detailed statistics for a specific log source. This is an admin-only operation that requires admin role privileges. Returns ClickHouse table statistics including row count, compressed/uncompressed sizes, compression ratio, part count, and column statistics.",
	getAdminSourceStats,
)

func AddAdminTools(mcp *server.MCPServer) {
	// Team management tools
	ListAllTeams.Register(mcp)
	GetTeam.Register(mcp)
	CreateTeam.Register(mcp)
	UpdateTeam.Register(mcp)
	DeleteTeam.Register(mcp)
	ListTeamMembers.Register(mcp)
	AddTeamMember.Register(mcp)
	RemoveTeamMember.Register(mcp)
	LinkSourceToTeam.Register(mcp)
	UnlinkSourceFromTeam.Register(mcp)
	
	// User management tools (admin only)
	ListAllUsers.Register(mcp)
	GetUser.Register(mcp)
	CreateUser.Register(mcp)
	UpdateUser.Register(mcp)
	DeleteUser.Register(mcp)
	
	// Source management tools (admin only)
	ListAllSources.Register(mcp)
	CreateSource.Register(mcp)
	ValidateSourceConnection.Register(mcp)
	DeleteSource.Register(mcp)
	GetAdminSourceStats.Register(mcp)
	
	// API token management tools
	ListAPITokens.Register(mcp)
	CreateAPIToken.Register(mcp)
	DeleteAPIToken.Register(mcp)
}