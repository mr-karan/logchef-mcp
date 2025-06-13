package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/mark3labs/mcp-go/server"

	mcplogchef "github.com/mr-karan/logchef-mcp"
	"github.com/mr-karan/logchef-mcp/tools"
)

// Build-time variables
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func maybeAddTools(s *server.MCPServer, tf func(*server.MCPServer), enabledTools []string, disable bool, category string) {
	if !slices.Contains(enabledTools, category) {
		slog.Debug("Not enabling tools", "category", category)
		return
	}
	if disable {
		slog.Info("Disabling tools", "category", category)
		return
	}
	slog.Debug("Enabling tools", "category", category)
	tf(s)
}

// disabledTools indicates whether each category of tools should be disabled.
type disabledTools struct {
	enabledTools string
	profile      bool
	sources      bool
	logs         bool
	admin        bool
}

// Configuration for the Logchef client.
type logchefConfig struct {
	// Whether to enable debug mode for the Logchef transport.
	debug bool
}

func (dt *disabledTools) addFlags() {
	flag.StringVar(&dt.enabledTools, "enabled-tools", "profile,sources,logs,admin", "A comma separated list of tools enabled for this server. Can be overwritten entirely or by disabling specific components, e.g. --disable-profile.")
	flag.BoolVar(&dt.profile, "disable-profile", false, "Disable profile tools")
	flag.BoolVar(&dt.sources, "disable-sources", false, "Disable sources tools")
	flag.BoolVar(&dt.logs, "disable-logs", false, "Disable logs tools")
	flag.BoolVar(&dt.admin, "disable-admin", false, "Disable admin tools")
}

func (lc *logchefConfig) addFlags() {
	flag.BoolVar(&lc.debug, "debug", false, "Enable debug mode for the Logchef transport")
}

func (dt *disabledTools) addTools(s *server.MCPServer) {
	enabledTools := strings.Split(dt.enabledTools, ",")
	maybeAddTools(s, tools.AddProfileTools, enabledTools, dt.profile, "profile")
	maybeAddTools(s, tools.AddSourcesTools, enabledTools, dt.sources, "sources")
	maybeAddTools(s, tools.AddLogsTools, enabledTools, dt.logs, "logs")
	maybeAddTools(s, tools.AddAdminTools, enabledTools, dt.admin, "admin")
}

func newServer(dt disabledTools) *server.MCPServer {
	s := server.NewMCPServer(
		"logchef-mcp",
		version,
	)
	dt.addTools(s)
	return s
}

func run(transport, addr, basePath string, endpointPath string, logLevel slog.Level, dt disabledTools, lc logchefConfig) error {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))
	s := newServer(dt)

	switch transport {
	case "stdio":
		slog.Info("Starting Logchef MCP server using stdio transport")
		return server.ServeStdio(s, server.WithStdioContextFunc(mcplogchef.ComposedStdioContextFunc(lc.debug)))
	case "sse":
		srv := server.NewSSEServer(s,
			server.WithSSEContextFunc(mcplogchef.ComposedSSEContextFunc(lc.debug)),
			server.WithStaticBasePath(basePath),
		)
		slog.Info("Starting Logchef MCP server using SSE transport", "address", addr, "basePath", basePath)
		if err := srv.Start(addr); err != nil {
			return fmt.Errorf("Server error: %v", err)
		}
	case "streamable-http":
		srv := server.NewStreamableHTTPServer(s, server.WithHTTPContextFunc(mcplogchef.ComposedHTTPContextFunc(lc.debug)),
			server.WithStateLess(true),
			server.WithEndpointPath(endpointPath),
		)
		slog.Info("Starting Logchef MCP server using StreamableHTTP transport", "address", addr, "endpointPath", endpointPath)
		if err := srv.Start(addr); err != nil {
			return fmt.Errorf("Server error: %v", err)
		}
	default:
		return fmt.Errorf(
			"Invalid transport type: %s. Must be 'stdio', 'sse', or 'streamable-http'",
			transport,
		)
	}
	return nil
}

func main() {
	var transport string
	flag.StringVar(&transport, "t", "stdio", "Transport type (stdio, sse, or streamable-http)")
	flag.StringVar(
		&transport,
		"transport",
		"stdio",
		"Transport type (stdio, sse, or streamable-http)",
	)
	addr := flag.String("address", "localhost:8000", "The host and port to start the server on")
	basePath := flag.String("base-path", "", "Base path for the sse server")
	endpointPath := flag.String("endpoint-path", "/mcp", "Endpoint path for the streamable-http server")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	showVersion := flag.Bool("version", false, "Show version information")
	var dt disabledTools
	dt.addFlags()
	var lc logchefConfig
	lc.addFlags()
	flag.Parse()

	if *showVersion {
		fmt.Printf("logchef-mcp %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("date: %s\n", date)
		return
	}

	if err := run(transport, *addr, *basePath, *endpointPath, parseLevel(*logLevel), dt, lc); err != nil {
		panic(err)
	}
}

func parseLevel(level string) slog.Level {
	var l slog.Level
	if err := l.UnmarshalText([]byte(level)); err != nil {
		return slog.LevelInfo
	}
	return l
}