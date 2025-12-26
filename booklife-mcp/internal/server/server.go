package server

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/user/booklife-mcp/internal/config"
	"github.com/user/booklife-mcp/internal/providers"
	"github.com/user/booklife-mcp/internal/providers/hardcover"
	"github.com/user/booklife-mcp/internal/providers/libby"
	"github.com/user/booklife-mcp/internal/providers/openlibrary"
)

// Server wraps the MCP server with BookLife providers
type Server struct {
	cfg       *config.Config
	mcpServer *mcp.Server

	// Providers (using interfaces for testability)
	hardcover   providers.HardcoverProvider
	libby       providers.LibbyProvider
	openlibrary providers.OpenLibraryProvider
}

// New creates a new BookLife MCP server
func New(cfg *config.Config) (*Server, error) {
	s := &Server{
		cfg: cfg,
	}
	
	// Initialize MCP server
	s.mcpServer = mcp.NewServer(&mcp.Implementation{
		Name:    cfg.Server.Name,
		Version: cfg.Server.Version,
	}, nil)
	
	// Initialize providers
	if err := s.initProviders(); err != nil {
		return nil, fmt.Errorf("initializing providers: %w", err)
	}
	
	// Register tools
	s.registerTools()
	
	// Register resources
	s.registerResources()
	
	// Register prompts
	s.registerPrompts()
	
	return s, nil
}

func (s *Server) initProviders() error {
	// Hardcover
	if s.cfg.Providers.Hardcover.Enabled {
		client, err := hardcover.NewClient(
			s.cfg.Providers.Hardcover.Endpoint,
			s.cfg.Providers.Hardcover.APIKey,
		)
		if err != nil {
			return fmt.Errorf("initializing Hardcover client: %w", err)
		}
		s.hardcover = client
	}
	
	// Libby
	if s.cfg.Providers.Libby.Enabled {
		client, err := libby.NewClient(s.cfg.Providers.Libby.CloneCode)
		if err != nil {
			return fmt.Errorf("initializing Libby client: %w", err)
		}
		s.libby = client
	}
	
	// Open Library
	if s.cfg.Providers.OpenLibrary.Enabled {
		s.openlibrary = openlibrary.NewClient(
			s.cfg.Providers.OpenLibrary.Endpoint,
			s.cfg.Providers.OpenLibrary.CoversEndpoint,
		)
	}
	
	return nil
}

// Run starts the MCP server
func (s *Server) Run(ctx context.Context) error {
	// Use stdio transport
	transport := &mcp.StdioTransport{}
	return s.mcpServer.Run(ctx, transport)
}
