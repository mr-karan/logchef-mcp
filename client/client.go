package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Config holds the configuration for the Logchef client
type Config struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

// Client represents a Logchef API client
type Client struct {
	config     Config
	httpClient *http.Client
}

// New creates a new Logchef client
func New(config Config) *Client {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// ProfileResponse represents the response from the /api/v1/me endpoint
type ProfileResponse struct {
	Status string `json:"status"`
	Data   struct {
		APIToken struct {
			CreatedAt  string `json:"created_at"`
			ID         int    `json:"id"`
			LastUsedAt string `json:"last_used_at"`
			Name       string `json:"name"`
			Prefix     string `json:"prefix"`
		} `json:"api_token"`
		AuthMethod string `json:"auth_method"`
		User       struct {
			ID          int    `json:"id"`
			Email       string `json:"email"`
			FullName    string `json:"full_name"`
			Role        string `json:"role"`
			Status      string `json:"status"`
			LastLoginAt string `json:"last_login_at"`
			CreatedAt   string `json:"created_at"`
			UpdatedAt   string `json:"updated_at"`
		} `json:"user"`
	} `json:"data"`
}

// UserTeamDetails represents the details of a team that a user belongs to
type UserTeamDetails struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	MemberCount int    `json:"member_count"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// TeamsResponse represents the response from the /api/v1/me/teams endpoint
type TeamsResponse struct {
	Status string             `json:"status"`
	Data   []*UserTeamDetails `json:"data"`
}

// SourceConnection represents the connection details for a source
type SourceConnection struct {
	Host      string `json:"host"`
	Database  string `json:"database"`
	TableName string `json:"table_name"`
}

// SourceResponse represents a source with its details
type SourceResponse struct {
	ID                   int               `json:"id"`
	Name                 string            `json:"name"`
	Description          string            `json:"description"`
	Connection           SourceConnection  `json:"connection"`
	MetaIsAutoCreated    bool              `json:"_meta_is_auto_created"`
	MetaTsField          string            `json:"_meta_ts_field"`
	MetaSeverityField    string            `json:"_meta_severity_field"`
	TTLDays              int               `json:"ttl_days"`
	IsConnected          bool              `json:"is_connected"`
	CreatedAt            string            `json:"created_at"`
	UpdatedAt            string            `json:"updated_at"`
}

// TeamSourcesResponse represents the response from the /api/v1/teams/:teamID/sources endpoint
type TeamSourcesResponse struct {
	Status string            `json:"status"`
	Data   []*SourceResponse `json:"data"`
}

// MetaResponse represents the response from the /api/v1/meta endpoint
type MetaResponse struct {
	Status string `json:"status"`
	Data   struct {
		Version           string `json:"version"`
		HTTPServerTimeout string `json:"http_server_timeout"`
	} `json:"data"`
}

// LogQueryRequest represents the request body for querying logs
type LogQueryRequest struct {
	RawSQL       string `json:"raw_sql"`
	Limit        int    `json:"limit,omitempty"`
	QueryTimeout *int   `json:"query_timeout,omitempty"`
}

// LogEntry represents a single log entry with flexible schema
type LogEntry map[string]interface{}

// LogQueryStats represents the execution statistics for a log query
type LogQueryStats struct {
	ExecutionTimeMs int `json:"execution_time_ms"`
	RowsRead        int `json:"rows_read"`
}

// LogColumn represents a column in the log schema
type LogColumn struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// LogQueryResponse represents the response from the log query endpoint
type LogQueryResponse struct {
	Status string `json:"status"`
	Data   struct {
		Logs    []LogEntry      `json:"logs"`
		Stats   LogQueryStats   `json:"stats"`
		Columns []LogColumn     `json:"columns"`
	} `json:"data"`
}

// SchemaResponse represents the response from the schema endpoint
type SchemaResponse struct {
	Status string      `json:"status"`
	Data   []LogColumn `json:"data"`
}

// HistogramRequest represents the request body for histogram endpoint
type HistogramRequest struct {
	RawSQL       string `json:"raw_sql"`
	Window       string `json:"window,omitempty"`
	GroupBy      string `json:"group_by,omitempty"`
	Timezone     string `json:"timezone,omitempty"`
	QueryTimeout *int   `json:"query_timeout,omitempty"`
}

// HistogramDataPoint represents a single point in the histogram
type HistogramDataPoint struct {
	Bucket     string `json:"bucket"`
	LogCount   int64  `json:"log_count"`
	GroupValue string `json:"group_value"`
}

// HistogramResponse represents the response from the histogram endpoint
type HistogramResponse struct {
	Status string `json:"status"`
	Data   struct {
		Granularity string                `json:"granularity"`
		Data        []HistogramDataPoint  `json:"data"`
	} `json:"data"`
}

// TableStats represents table statistics from ClickHouse
type TableStats struct {
	Database     string  `json:"database"`
	Table        string  `json:"table"`
	Compressed   string  `json:"compressed"`
	Uncompressed string  `json:"uncompressed"`
	ComprRate    float64 `json:"compr_rate"`
	Rows         int64   `json:"rows"`
	PartCount    int     `json:"part_count"`
}

// ColumnStats represents column statistics from ClickHouse
type ColumnStats struct {
	Name string `json:"name"`
	Type string `json:"type"`
	// Add more fields as needed based on actual API response
}

// SourceStatsResponse represents the response from source stats endpoint
type SourceStatsResponse struct {
	Status string `json:"status"`
	Data   struct {
		TableStats  TableStats    `json:"table_stats"`
		ColumnStats []ColumnStats `json:"column_stats"`
	} `json:"data"`
}

// Collection represents a saved query collection
type Collection struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TeamID      int    `json:"team_id"`
	SourceID    int    `json:"source_id"`
	Query       string `json:"query"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// CollectionRequest represents the request body for creating/updating collections
type CollectionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Query       string `json:"query"`
}

// CollectionsResponse represents the response from the collections list endpoint
type CollectionsResponse struct {
	Status string       `json:"status"`
	Data   []Collection `json:"data"`
}

// CollectionResponse represents the response from single collection endpoints
type CollectionResponse struct {
	Status string     `json:"status"`
	Data   Collection `json:"data"`
}

// Team represents a team in the admin API
type Team struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MemberCount int    `json:"member_count"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// TeamRequest represents the request body for creating/updating teams
type TeamRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// TeamUpdateRequest represents the request body for updating teams (with optional fields)
type TeamUpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// TeamsListResponse represents the response from the admin teams list endpoint
type TeamsListResponse struct {
	Status string `json:"status"`
	Data   []Team `json:"data"`
}

// TeamResponse represents the response from single team endpoints
type TeamResponse struct {
	Status string `json:"status"`
	Data   Team   `json:"data"`
}

// TeamMember represents a team member
type TeamMember struct {
	TeamID    int    `json:"team_id"`
	UserID    int    `json:"user_id"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
	Email     string `json:"email,omitempty"`
	FullName  string `json:"full_name,omitempty"`
}

// TeamMemberRequest represents the request body for adding team members
type TeamMemberRequest struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
}

// TeamMembersResponse represents the response from team members endpoint
type TeamMembersResponse struct {
	Status string       `json:"status"`
	Data   []TeamMember `json:"data"`
}

// TeamSourceRequest represents the request body for linking sources to teams
type TeamSourceRequest struct {
	SourceID int `json:"source_id"`
}

// User represents a user in the admin API
type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	FullName     string    `json:"full_name"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	LastLoginAt  *string   `json:"last_login_at,omitempty"`
	LastActiveAt *string   `json:"last_active_at,omitempty"`
	CreatedAt    string    `json:"created_at"`
	UpdatedAt    string    `json:"updated_at"`
}

// UserRequest represents the request body for creating users
type UserRequest struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

// UserUpdateRequest represents the request body for updating users (with optional fields)
type UserUpdateRequest struct {
	Email    *string `json:"email,omitempty"`
	FullName *string `json:"full_name,omitempty"`
	Role     *string `json:"role,omitempty"`
	Status   *string `json:"status,omitempty"`
}

// UsersListResponse represents the response from the admin users list endpoint
type UsersListResponse struct {
	Status string `json:"status"`
	Data   []User `json:"data"`
}

// UserResponse represents the response from single user endpoints
type UserResponse struct {
	Status string `json:"status"`
	Data   User   `json:"data"`
}

// APIToken represents an API token
type APIToken struct {
	ID         int     `json:"id"`
	UserID     int     `json:"user_id"`
	Name       string  `json:"name"`
	Prefix     string  `json:"prefix"`
	LastUsedAt *string `json:"last_used_at,omitempty"`
	ExpiresAt  *string `json:"expires_at,omitempty"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

// APITokenRequest represents the request body for creating API tokens
type APITokenRequest struct {
	Name      string  `json:"name"`
	ExpiresAt *string `json:"expires_at,omitempty"`
}

// APITokenCreateResponse represents the response when creating an API token
type APITokenCreateResponse struct {
	Token    string   `json:"token"`
	APIToken APIToken `json:"api_token"`
}

// APITokensResponse represents the response from the API tokens list endpoint
type APITokensResponse struct {
	Status string     `json:"status"`
	Data   []APIToken `json:"data"`
}

// APITokenResponse represents the response from single API token endpoints
type APITokenResponse struct {
	Status string                  `json:"status"`
	Data   APITokenCreateResponse  `json:"data"`
}

// Admin Source Management structs

// Source represents a source in the admin API
type Source struct {
	ID                   int               `json:"id"`
	Name                 string            `json:"name"`
	Description          string            `json:"description"`
	Connection           SourceConnection  `json:"connection"`
	MetaIsAutoCreated    bool              `json:"_meta_is_auto_created"`
	MetaTsField          string            `json:"_meta_ts_field"`
	MetaSeverityField    string            `json:"_meta_severity_field"`
	TTLDays              int               `json:"ttl_days"`
	IsConnected          bool              `json:"is_connected"`
	CreatedAt            string            `json:"created_at"`
	UpdatedAt            string            `json:"updated_at"`
}

// SourceRequest represents the request body for creating sources
type SourceRequest struct {
	Name                 string            `json:"name"`
	Description          string            `json:"description"`
	Connection           SourceConnection  `json:"connection"`
	MetaIsAutoCreated    bool              `json:"_meta_is_auto_created"`
	MetaTsField          string            `json:"_meta_ts_field"`
	MetaSeverityField    string            `json:"_meta_severity_field"`
	TTLDays              int               `json:"ttl_days"`
	Schema               []LogColumn       `json:"schema,omitempty"`
}

// SourceValidationRequest represents the request body for validating source connections
type SourceValidationRequest struct {
	Host           string `json:"host"`
	Database       string `json:"database"`
	TableName      string `json:"table_name"`
	TimestampField string `json:"timestamp_field,omitempty"`
	SeverityField  string `json:"severity_field,omitempty"`
}

// SourceValidationResult represents the result of connection validation
type SourceValidationResult struct {
	IsValid       bool     `json:"is_valid"`
	Message       string   `json:"message"`
	ErrorDetails  []string `json:"error_details,omitempty"`
	TableExists   bool     `json:"table_exists,omitempty"`
	ColumnChecks  map[string]bool `json:"column_checks,omitempty"`
}

// SourcesListResponse represents the response from the admin sources list endpoint
type SourcesListResponse struct {
	Status string   `json:"status"`
	Data   []Source `json:"data"`
}

// AdminSourceResponse represents the response from single admin source endpoints
type AdminSourceResponse struct {
	Status string `json:"status"`
	Data   Source `json:"data"`
}

// SourceValidationResponse represents the response from source validation endpoint
type SourceValidationResponse struct {
	Status string                  `json:"status"`
	Data   SourceValidationResult  `json:"data"`
}


// GetProfile retrieves the current user profile
func (c *Client) GetProfile(ctx context.Context) (*ProfileResponse, error) {
	url := fmt.Sprintf("%s/api/v1/me", c.config.BaseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var profile ProfileResponse
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &profile, nil
}

// GetTeams retrieves the teams that the current user belongs to
func (c *Client) GetTeams(ctx context.Context) (*TeamsResponse, error) {
	url := fmt.Sprintf("%s/api/v1/me/teams", c.config.BaseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var teams TeamsResponse
	if err := json.Unmarshal(body, &teams); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &teams, nil
}

// GetTeamSources retrieves the sources that belong to a specific team
func (c *Client) GetTeamSources(ctx context.Context, teamID int) (*TeamSourcesResponse, error) {
	url := fmt.Sprintf("%s/api/v1/teams/%d/sources", c.config.BaseURL, teamID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var sources TeamSourcesResponse
	if err := json.Unmarshal(body, &sources); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &sources, nil
}

// GetMeta retrieves server metadata including version information
func (c *Client) GetMeta(ctx context.Context) (*MetaResponse, error) {
	url := fmt.Sprintf("%s/api/v1/meta", c.config.BaseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Meta endpoint doesn't require authentication
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var meta MetaResponse
	if err := json.Unmarshal(body, &meta); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &meta, nil
}

// QueryLogs executes a log query against a specific source within a team
func (c *Client) QueryLogs(ctx context.Context, teamID, sourceID int, request LogQueryRequest) (*LogQueryResponse, error) {
	url := fmt.Sprintf("%s/api/v1/teams/%d/sources/%d/logs/query", c.config.BaseURL, teamID, sourceID)

	// Set default limit if not specified
	if request.Limit <= 0 {
		request.Limit = 100
	}
	// Enforce max limit
	if request.Limit > 100 {
		request.Limit = 100
	}

	// Set default timeout if not specified
	if request.QueryTimeout == nil {
		defaultTimeout := 30
		request.QueryTimeout = &defaultTimeout
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var logResponse LogQueryResponse
	if err := json.Unmarshal(body, &logResponse); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &logResponse, nil
}

// GetSourceSchema retrieves the schema (column names and types) for a specific source within a team
func (c *Client) GetSourceSchema(ctx context.Context, teamID, sourceID int) (*SchemaResponse, error) {
	url := fmt.Sprintf("%s/api/v1/teams/%d/sources/%d/schema", c.config.BaseURL, teamID, sourceID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var schema SchemaResponse
	if err := json.Unmarshal(body, &schema); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &schema, nil
}

// GetSourceStats retrieves statistics for a specific source within a team
func (c *Client) GetSourceStats(ctx context.Context, teamID, sourceID int) (*SourceStatsResponse, error) {
	url := fmt.Sprintf("%s/api/v1/teams/%d/sources/%d/stats", c.config.BaseURL, teamID, sourceID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var stats SourceStatsResponse
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &stats, nil
}

// GetLogHistogram generates histogram data for logs within a team source
func (c *Client) GetLogHistogram(ctx context.Context, teamID, sourceID int, request HistogramRequest) (*HistogramResponse, error) {
	url := fmt.Sprintf("%s/api/v1/teams/%d/sources/%d/logs/histogram", c.config.BaseURL, teamID, sourceID)

	// Set default timeout if not specified
	if request.QueryTimeout == nil {
		defaultTimeout := 30
		request.QueryTimeout = &defaultTimeout
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var histogram HistogramResponse
	if err := json.Unmarshal(body, &histogram); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &histogram, nil
}

// GetCollections retrieves all collections for a specific team and source
func (c *Client) GetCollections(ctx context.Context, teamID, sourceID int) (*CollectionsResponse, error) {
	url := fmt.Sprintf("%s/api/v1/teams/%d/sources/%d/collections", c.config.BaseURL, teamID, sourceID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var collections CollectionsResponse
	if err := json.Unmarshal(body, &collections); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &collections, nil
}

// CreateCollection creates a new collection for a specific team and source
func (c *Client) CreateCollection(ctx context.Context, teamID, sourceID int, request CollectionRequest) (*CollectionResponse, error) {
	url := fmt.Sprintf("%s/api/v1/teams/%d/sources/%d/collections", c.config.BaseURL, teamID, sourceID)

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var collection CollectionResponse
	if err := json.Unmarshal(body, &collection); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &collection, nil
}

// GetCollection retrieves a specific collection by ID
func (c *Client) GetCollection(ctx context.Context, teamID, sourceID, collectionID int) (*CollectionResponse, error) {
	url := fmt.Sprintf("%s/api/v1/teams/%d/sources/%d/collections/%d", c.config.BaseURL, teamID, sourceID, collectionID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var collection CollectionResponse
	if err := json.Unmarshal(body, &collection); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &collection, nil
}

// UpdateCollection updates an existing collection
func (c *Client) UpdateCollection(ctx context.Context, teamID, sourceID, collectionID int, request CollectionRequest) (*CollectionResponse, error) {
	url := fmt.Sprintf("%s/api/v1/teams/%d/sources/%d/collections/%d", c.config.BaseURL, teamID, sourceID, collectionID)

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var collection CollectionResponse
	if err := json.Unmarshal(body, &collection); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &collection, nil
}

// DeleteCollection deletes a collection by ID
func (c *Client) DeleteCollection(ctx context.Context, teamID, sourceID, collectionID int) error {
	url := fmt.Sprintf("%s/api/v1/teams/%d/sources/%d/collections/%d", c.config.BaseURL, teamID, sourceID, collectionID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Admin Team Management Methods

// ListAllTeams retrieves all teams (admin only)
func (c *Client) ListAllTeams(ctx context.Context) (*TeamsListResponse, error) {
	url := fmt.Sprintf("%s/api/v1/admin/teams", c.config.BaseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var teams TeamsListResponse
	if err := json.Unmarshal(body, &teams); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &teams, nil
}

// GetTeamByID retrieves a specific team by ID
func (c *Client) GetTeamByID(ctx context.Context, teamID int) (*TeamResponse, error) {
	url := fmt.Sprintf("%s/api/v1/teams/%d", c.config.BaseURL, teamID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var team TeamResponse
	if err := json.Unmarshal(body, &team); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &team, nil
}

// CreateTeam creates a new team (admin only)
func (c *Client) CreateTeam(ctx context.Context, request TeamRequest) (*TeamResponse, error) {
	url := fmt.Sprintf("%s/api/v1/admin/teams", c.config.BaseURL)

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var team TeamResponse
	if err := json.Unmarshal(body, &team); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &team, nil
}

// UpdateTeam updates an existing team
func (c *Client) UpdateTeam(ctx context.Context, teamID int, request TeamUpdateRequest) (*TeamResponse, error) {
	url := fmt.Sprintf("%s/api/v1/teams/%d", c.config.BaseURL, teamID)

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var team TeamResponse
	if err := json.Unmarshal(body, &team); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &team, nil
}

// DeleteTeam deletes a team (admin only)
func (c *Client) DeleteTeam(ctx context.Context, teamID int) error {
	url := fmt.Sprintf("%s/api/v1/admin/teams/%d", c.config.BaseURL, teamID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListTeamMembers retrieves all members of a team
func (c *Client) ListTeamMembers(ctx context.Context, teamID int) (*TeamMembersResponse, error) {
	url := fmt.Sprintf("%s/api/v1/teams/%d/members", c.config.BaseURL, teamID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var members TeamMembersResponse
	if err := json.Unmarshal(body, &members); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &members, nil
}

// AddTeamMember adds a user to a team
func (c *Client) AddTeamMember(ctx context.Context, teamID int, request TeamMemberRequest) error {
	url := fmt.Sprintf("%s/api/v1/teams/%d/members", c.config.BaseURL, teamID)

	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// RemoveTeamMember removes a user from a team
func (c *Client) RemoveTeamMember(ctx context.Context, teamID, userID int) error {
	url := fmt.Sprintf("%s/api/v1/teams/%d/members/%d", c.config.BaseURL, teamID, userID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// LinkSourceToTeam links a source to a team
func (c *Client) LinkSourceToTeam(ctx context.Context, teamID int, request TeamSourceRequest) error {
	url := fmt.Sprintf("%s/api/v1/teams/%d/sources", c.config.BaseURL, teamID)

	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// UnlinkSourceFromTeam removes a source from a team
func (c *Client) UnlinkSourceFromTeam(ctx context.Context, teamID, sourceID int) error {
	url := fmt.Sprintf("%s/api/v1/teams/%d/sources/%d", c.config.BaseURL, teamID, sourceID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Admin User Management Methods

// ListAllUsers retrieves all users (admin only)
func (c *Client) ListAllUsers(ctx context.Context) (*UsersListResponse, error) {
	url := fmt.Sprintf("%s/api/v1/admin/users", c.config.BaseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var users UsersListResponse
	if err := json.Unmarshal(body, &users); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &users, nil
}

// GetUserByID retrieves a specific user by ID (admin only)
func (c *Client) GetUserByID(ctx context.Context, userID int) (*UserResponse, error) {
	url := fmt.Sprintf("%s/api/v1/admin/users/%d", c.config.BaseURL, userID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var user UserResponse
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &user, nil
}

// CreateUser creates a new user (admin only)
func (c *Client) CreateUser(ctx context.Context, request UserRequest) (*UserResponse, error) {
	url := fmt.Sprintf("%s/api/v1/admin/users", c.config.BaseURL)

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var user UserResponse
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &user, nil
}

// UpdateUser updates an existing user (admin only)
func (c *Client) UpdateUser(ctx context.Context, userID int, request UserUpdateRequest) (*UserResponse, error) {
	url := fmt.Sprintf("%s/api/v1/admin/users/%d", c.config.BaseURL, userID)

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var user UserResponse
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &user, nil
}

// DeleteUser deletes a user (admin only)
func (c *Client) DeleteUser(ctx context.Context, userID int) error {
	url := fmt.Sprintf("%s/api/v1/admin/users/%d", c.config.BaseURL, userID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// API Token Management Methods

// ListAPITokens retrieves all API tokens for the current user
func (c *Client) ListAPITokens(ctx context.Context) (*APITokensResponse, error) {
	url := fmt.Sprintf("%s/api/v1/me/tokens", c.config.BaseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var tokens APITokensResponse
	if err := json.Unmarshal(body, &tokens); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &tokens, nil
}

// CreateAPIToken creates a new API token for the current user
func (c *Client) CreateAPIToken(ctx context.Context, request APITokenRequest) (*APITokenResponse, error) {
	url := fmt.Sprintf("%s/api/v1/me/tokens", c.config.BaseURL)

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var token APITokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &token, nil
}

// DeleteAPIToken deletes an API token
func (c *Client) DeleteAPIToken(ctx context.Context, tokenID int) error {
	url := fmt.Sprintf("%s/api/v1/me/tokens/%d", c.config.BaseURL, tokenID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Admin Source Management Methods

// ListAllSources retrieves all sources (admin only)
func (c *Client) ListAllSources(ctx context.Context) (*SourcesListResponse, error) {
	url := fmt.Sprintf("%s/api/v1/admin/sources", c.config.BaseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var sources SourcesListResponse
	if err := json.Unmarshal(body, &sources); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &sources, nil
}

// CreateSource creates a new source (admin only)
func (c *Client) CreateSource(ctx context.Context, request SourceRequest) (*AdminSourceResponse, error) {
	url := fmt.Sprintf("%s/api/v1/admin/sources", c.config.BaseURL)

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var source AdminSourceResponse
	if err := json.Unmarshal(body, &source); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &source, nil
}

// ValidateSourceConnection validates a source connection (admin only)
func (c *Client) ValidateSourceConnection(ctx context.Context, request SourceValidationRequest) (*SourceValidationResponse, error) {
	url := fmt.Sprintf("%s/api/v1/admin/sources/validate", c.config.BaseURL)

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var validation SourceValidationResponse
	if err := json.Unmarshal(body, &validation); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &validation, nil
}

// DeleteSource deletes a source (admin only)
func (c *Client) DeleteSource(ctx context.Context, sourceID int) error {
	url := fmt.Sprintf("%s/api/v1/admin/sources/%d", c.config.BaseURL, sourceID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetAdminSourceStats retrieves source statistics (admin only)
func (c *Client) GetAdminSourceStats(ctx context.Context, sourceID int) (*SourceStatsResponse, error) {
	url := fmt.Sprintf("%s/api/v1/admin/sources/%d/stats", c.config.BaseURL, sourceID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var stats SourceStatsResponse
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &stats, nil
}

