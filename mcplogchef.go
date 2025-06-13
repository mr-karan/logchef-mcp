package mcplogchef

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/server"

	"github.com/mr-karan/logchef-mcp/client"
)

const (
	defaultLogchefHost = "localhost:5173"
	defaultLogchefURL  = "http://" + defaultLogchefHost

	logchefURLEnvVar = "LOGCHEF_URL"
	logchefAPIEnvVar = "LOGCHEF_API_KEY"

	logchefURLHeader    = "X-Logchef-URL"
	logchefAPIKeyHeader = "X-Logchef-API-Key"
)

func urlAndAPIKeyFromEnv() (string, string) {
	u := strings.TrimRight(os.Getenv(logchefURLEnvVar), "/")
	apiKey := os.Getenv(logchefAPIEnvVar)
	return u, apiKey
}

func urlAndAPIKeyFromHeaders(req *http.Request) (string, string) {
	u := req.Header.Get(logchefURLHeader)
	apiKey := req.Header.Get(logchefAPIKeyHeader)
	return u, apiKey
}

type logchefURLKey struct{}
type logchefAPIKeyKey struct{}

// logchefDebugKey is the context key for the Logchef transport's debug flag.
type logchefDebugKey struct{}

// WithLogchefDebug adds the Logchef debug flag to the context.
func WithLogchefDebug(ctx context.Context, debug bool) context.Context {
	if debug {
		slog.Info("Logchef transport debug mode enabled")
	}
	return context.WithValue(ctx, logchefDebugKey{}, debug)
}

// LogchefDebugFromContext extracts the Logchef debug flag from the context.
// If the flag is not set, it returns false.
func LogchefDebugFromContext(ctx context.Context) bool {
	if debug, ok := ctx.Value(logchefDebugKey{}).(bool); ok {
		return debug
	}
	return false
}

// ExtractLogchefInfoFromEnv is a StdioContextFunc that extracts Logchef configuration
// from environment variables and injects the configuration into the context.
var ExtractLogchefInfoFromEnv server.StdioContextFunc = func(ctx context.Context) context.Context {
	u, apiKey := urlAndAPIKeyFromEnv()
	if u == "" {
		u = defaultLogchefURL
	}
	parsedURL, err := url.Parse(u)
	if err != nil {
		panic(fmt.Errorf("invalid Logchef URL %s: %w", u, err))
	}
	slog.Info("Using Logchef configuration", "url", parsedURL.Redacted(), "api_key_set", apiKey != "")
	return WithLogchefURL(WithLogchefAPIKey(ctx, apiKey), u)
}

// httpContextFunc is a function that can be used as a `server.HTTPContextFunc` or a
// `server.SSEContextFunc`. It is necessary because, while the two types are functionally
// identical, they have distinct types and cannot be passed around interchangeably.
type httpContextFunc func(ctx context.Context, req *http.Request) context.Context

// ExtractLogchefInfoFromHeaders is a HTTPContextFunc that extracts Logchef configuration
// from request headers and injects the configuration into the context.
var ExtractLogchefInfoFromHeaders httpContextFunc = func(ctx context.Context, req *http.Request) context.Context {
	u, apiKey := urlAndAPIKeyFromHeaders(req)
	uEnv, apiKeyEnv := urlAndAPIKeyFromEnv()
	if u == "" {
		u = uEnv
	}
	if u == "" {
		u = defaultLogchefURL
	}
	if apiKey == "" {
		apiKey = apiKeyEnv
	}
	return WithLogchefURL(WithLogchefAPIKey(ctx, apiKey), u)
}

// WithLogchefURL adds the Logchef URL to the context.
func WithLogchefURL(ctx context.Context, url string) context.Context {
	return context.WithValue(ctx, logchefURLKey{}, url)
}

// WithLogchefAPIKey adds the Logchef API key to the context.
func WithLogchefAPIKey(ctx context.Context, apiKey string) context.Context {
	return context.WithValue(ctx, logchefAPIKeyKey{}, apiKey)
}

// LogchefURLFromContext extracts the Logchef URL from the context.
func LogchefURLFromContext(ctx context.Context) string {
	if u, ok := ctx.Value(logchefURLKey{}).(string); ok {
		return u
	}
	return defaultLogchefURL
}

// LogchefAPIKeyFromContext extracts the Logchef API key from the context.
func LogchefAPIKeyFromContext(ctx context.Context) string {
	if k, ok := ctx.Value(logchefAPIKeyKey{}).(string); ok {
		return k
	}
	return ""
}

type logchefClientKey struct{}

// NewLogchefClient creates a Logchef client with the provided URL and API key.
func NewLogchefClient(ctx context.Context, logchefURL, apiKey string) *client.Client {
	if logchefURL == "" {
		logchefURL = defaultLogchefURL
	}

	parsedURL, err := url.Parse(logchefURL)
	if err != nil {
		panic(fmt.Errorf("invalid Logchef URL: %w", err))
	}

	slog.Debug("Creating Logchef client", "url", parsedURL.Redacted(), "api_key_set", apiKey != "")
	return client.New(client.Config{
		BaseURL: logchefURL,
		APIKey:  apiKey,
	})
}

// ExtractLogchefClientFromEnv is a StdioContextFunc that extracts Logchef configuration
// from environment variables and injects a configured client into the context.
var ExtractLogchefClientFromEnv server.StdioContextFunc = func(ctx context.Context) context.Context {
	// Extract transport config from env vars
	logchefURL, ok := os.LookupEnv(logchefURLEnvVar)
	if !ok {
		logchefURL = defaultLogchefURL
	}
	apiKey := os.Getenv(logchefAPIEnvVar)

	logchefClient := NewLogchefClient(ctx, logchefURL, apiKey)
	return context.WithValue(ctx, logchefClientKey{}, logchefClient)
}

// ExtractLogchefClientFromHeaders is a HTTPContextFunc that extracts Logchef configuration
// from request headers and injects a configured client into the context.
var ExtractLogchefClientFromHeaders httpContextFunc = func(ctx context.Context, req *http.Request) context.Context {
	// Extract transport config from request headers, and set it on the context.
	u, apiKey := urlAndAPIKeyFromHeaders(req)
	uEnv, apiKeyEnv := urlAndAPIKeyFromEnv()
	if u == "" {
		u = uEnv
	}
	if u == "" {
		u = defaultLogchefURL
	}
	if apiKey == "" {
		apiKey = apiKeyEnv
	}

	logchefClient := NewLogchefClient(ctx, u, apiKey)
	return WithLogchefClient(ctx, logchefClient)
}

// WithLogchefClient sets the Logchef client in the context.
//
// It can be retrieved using LogchefClientFromContext.
func WithLogchefClient(ctx context.Context, client *client.Client) context.Context {
	return context.WithValue(ctx, logchefClientKey{}, client)
}

// LogchefClientFromContext retrieves the Logchef client from the context.
func LogchefClientFromContext(ctx context.Context) *client.Client {
	c, ok := ctx.Value(logchefClientKey{}).(*client.Client)
	if !ok {
		return nil
	}
	return c
}

// ComposeStdioContextFuncs composes multiple StdioContextFuncs into a single one.
func ComposeStdioContextFuncs(funcs ...server.StdioContextFunc) server.StdioContextFunc {
	return func(ctx context.Context) context.Context {
		for _, f := range funcs {
			ctx = f(ctx)
		}
		return ctx
	}
}

// ComposeSSEContextFuncs composes multiple SSEContextFuncs into a single one.
func ComposeSSEContextFuncs(funcs ...httpContextFunc) server.SSEContextFunc {
	return func(ctx context.Context, req *http.Request) context.Context {
		for _, f := range funcs {
			ctx = f(ctx, req)
		}
		return ctx
	}
}

// ComposeHTTPContextFuncs composes multiple HTTPContextFuncs into a single one.
func ComposeHTTPContextFuncs(funcs ...httpContextFunc) server.HTTPContextFunc {
	return func(ctx context.Context, req *http.Request) context.Context {
		for _, f := range funcs {
			ctx = f(ctx, req)
		}
		return ctx
	}
}

// ComposedStdioContextFunc returns a StdioContextFunc that comprises all predefined StdioContextFuncs,
// as well as the Logchef debug flag.
func ComposedStdioContextFunc(debug bool) server.StdioContextFunc {
	return ComposeStdioContextFuncs(
		func(ctx context.Context) context.Context {
			return WithLogchefDebug(ctx, debug)
		},
		ExtractLogchefInfoFromEnv,
		ExtractLogchefClientFromEnv,
	)
}

// ComposedSSEContextFunc is a SSEContextFunc that comprises all predefined SSEContextFuncs.
func ComposedSSEContextFunc(debug bool) server.SSEContextFunc {
	return ComposeSSEContextFuncs(
		func(ctx context.Context, req *http.Request) context.Context {
			return WithLogchefDebug(ctx, debug)
		},
		ExtractLogchefInfoFromHeaders,
		ExtractLogchefClientFromHeaders,
	)
}

// ComposedHTTPContextFunc is a HTTPContextFunc that comprises all predefined HTTPContextFuncs.
func ComposedHTTPContextFunc(debug bool) server.HTTPContextFunc {
	return ComposeHTTPContextFuncs(
		func(ctx context.Context, req *http.Request) context.Context {
			return WithLogchefDebug(ctx, debug)
		},
		ExtractLogchefInfoFromHeaders,
		ExtractLogchefClientFromHeaders,
	)
}