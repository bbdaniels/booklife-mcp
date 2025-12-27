package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/user/booklife-mcp/internal/config"
	"github.com/user/booklife-mcp/internal/history"
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

	// Local history store
	historyStore *history.Store
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
	// Initialize local history store
	dataDir := os.Getenv("BOOKLIFE_DATA_DIR")
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = filepath.Join(home, ".local", "share", "booklife")
	}
	store, err := history.NewStore(dataDir)
	if err != nil {
		return fmt.Errorf("initializing history store: %w", err)
	}
	s.historyStore = store

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

	// Libby - uses saved identity from 'booklife libby-connect'
	if s.cfg.Providers.Libby.Enabled {
		if !libby.HasSavedIdentity() {
			return fmt.Errorf("Libby enabled but no saved identity found - run 'booklife libby-connect <code>' first")
		}
		client, err := libby.NewClientFromSavedIdentityWithOptions(s.cfg.Providers.Libby.SkipTLSVerify)
		if err != nil {
			return fmt.Errorf("initializing Libby client: %w", err)
		}
		s.libby = client

		// Auto-import timeline if URL is configured
		if s.cfg.Providers.Libby.TimelineURL != "" {
			importer := history.NewImporter(store)
			count, err := importer.ImportTimeline(s.cfg.Providers.Libby.TimelineURL)
			if err != nil {
				fmt.Printf("Warning: Failed to import timeline: %v\n", err)
			} else {
				fmt.Printf("Imported %d timeline entries\n", count)
			}
		}
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
