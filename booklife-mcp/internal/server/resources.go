package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerResources registers all BookLife resources with the MCP server
func (s *Server) registerResources() {
	// Current reading list
	s.mcpServer.AddResource(&mcp.Resource{
		URI:         "booklife://library/current",
		Name:        "Currently Reading",
		Description: "Your currently reading books with progress",
		MIMEType:    "application/json",
	}, s.handleCurrentlyReading)

	// TBR list
	s.mcpServer.AddResource(&mcp.Resource{
		URI:         "booklife://library/tbr",
		Name:        "To Be Read",
		Description: "Your want-to-read list from Hardcover",
		MIMEType:    "application/json",
	}, s.handleTBR)

	// Recently finished
	s.mcpServer.AddResource(&mcp.Resource{
		URI:         "booklife://library/recent",
		Name:        "Recently Finished",
		Description: "Books finished in the last 30 days",
		MIMEType:    "application/json",
	}, s.handleRecentlyFinished)

	// Library loans
	if s.libby != nil {
		s.mcpServer.AddResource(&mcp.Resource{
			URI:         "booklife://loans",
			Name:        "Library Loans",
			Description: "Current Libby loans with due dates",
			MIMEType:    "application/json",
		}, s.handleLoansResource)

		// Library holds
		s.mcpServer.AddResource(&mcp.Resource{
			URI:         "booklife://holds",
			Name:        "Library Holds",
			Description: "Library hold queue with estimated wait times",
			MIMEType:    "application/json",
		}, s.handleHoldsResource)
	}

	// Reading stats
	s.mcpServer.AddResource(&mcp.Resource{
		URI:         "booklife://stats",
		Name:        "Reading Statistics",
		Description: "Your reading stats for the current year",
		MIMEType:    "application/json",
	}, s.handleStats)
}

// Resource template for dynamic book URIs
func (s *Server) registerResourceTemplates() {
	// Book detail template: booklife://book/{id}
	s.mcpServer.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "booklife://book/{id}",
		Name:        "Book Details",
		Description: "Complete book information from all sources",
		MIMEType:    "application/json",
	}, s.handleBookDetail)
}

// === Resource Handlers ===

func (s *Server) handleCurrentlyReading(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	if s.hardcover == nil {
		return nil, fmt.Errorf("Hardcover not configured")
	}

	books, err := s.hardcover.GetUserBooks(ctx, "reading", 50)
	if err != nil {
		return nil, fmt.Errorf("fetching currently reading: %w", err)
	}

	data, err := json.MarshalIndent(books, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling books: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:  req.Params.URI,
				Text: string(data),
			},
		},
	}, nil
}

func (s *Server) handleTBR(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	if s.hardcover == nil {
		return nil, fmt.Errorf("Hardcover not configured")
	}

	books, err := s.hardcover.GetUserBooks(ctx, "want-to-read", 100)
	if err != nil {
		return nil, fmt.Errorf("fetching TBR: %w", err)
	}

	data, err := json.MarshalIndent(books, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling books: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:  req.Params.URI,
				Text: string(data),
			},
		},
	}, nil
}

func (s *Server) handleRecentlyFinished(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	if s.hardcover == nil {
		return nil, fmt.Errorf("Hardcover not configured")
	}

	// TODO: Add date filter for "last 30 days"
	books, err := s.hardcover.GetUserBooks(ctx, "read", 20)
	if err != nil {
		return nil, fmt.Errorf("fetching recently finished: %w", err)
	}

	data, err := json.MarshalIndent(books, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling books: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:  req.Params.URI,
				Text: string(data),
			},
		},
	}, nil
}

func (s *Server) handleLoansResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	if s.libby == nil {
		return nil, fmt.Errorf("Libby not configured")
	}

	loans, err := s.libby.GetLoans(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching loans: %w", err)
	}

	data, err := json.MarshalIndent(loans, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling loans: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:  req.Params.URI,
				Text: string(data),
			},
		},
	}, nil
}

func (s *Server) handleHoldsResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	if s.libby == nil {
		return nil, fmt.Errorf("Libby not configured")
	}

	holds, err := s.libby.GetHolds(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching holds: %w", err)
	}

	data, err := json.MarshalIndent(holds, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling holds: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:  req.Params.URI,
				Text: string(data),
			},
		},
	}, nil
}

func (s *Server) handleStats(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	if s.hardcover == nil {
		return nil, fmt.Errorf("Hardcover not configured")
	}

	stats, err := s.hardcover.GetReadingStats(ctx, 2025)
	if err != nil {
		return nil, fmt.Errorf("fetching stats: %w", err)
	}

	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling stats: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:  req.Params.URI,
				Text: string(data),
			},
		},
	}, nil
}

func (s *Server) handleBookDetail(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	// Extract book ID from URI: booklife://book/{id}
	uri := req.Params.URI
	parts := strings.Split(uri, "/")
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid book URI: %s", uri)
	}
	bookID := parts[3]

	if s.hardcover == nil {
		return nil, fmt.Errorf("Hardcover not configured")
	}

	book, err := s.hardcover.GetBook(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("fetching book: %w", err)
	}

	// Enrich with library availability
	if s.libby != nil && book.ISBN13 != "" {
		avail, err := s.libby.CheckAvailability(ctx, book.ISBN13, "", "")
		if err == nil {
			book.LibraryAvailability = avail
		}
	}

	data, err := json.MarshalIndent(book, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling book: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:  req.Params.URI,
				Text: string(data),
			},
		},
	}, nil
}
