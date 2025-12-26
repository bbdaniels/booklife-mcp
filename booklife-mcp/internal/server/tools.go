package server

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/user/booklife-mcp/internal/models"
)

// Tool input types with JSON schema annotations

// SearchBooksInput for the search_books tool
type SearchBooksInput struct {
	Query   string   `json:"query" jsonschema:"required,description=Title, author, or ISBN"`
	Sources []string `json:"sources,omitempty" jsonschema:"description=Sources to search: hardcover, openlibrary, wikidata"`
	Limit   int      `json:"limit,omitempty" jsonschema:"default=10,maximum=50"`
}

// GetMyLibraryInput for the get_my_library tool
type GetMyLibraryInput struct {
	Status string `json:"status,omitempty" jsonschema:"enum=reading,read,want-to-read,dnf,all"`
	SortBy string `json:"sort_by,omitempty" jsonschema:"enum=date_added,title,author,rating"`
	Limit  int    `json:"limit,omitempty" jsonschema:"default=20"`
}

// UpdateReadingStatusInput for the update_reading_status tool
type UpdateReadingStatusInput struct {
	BookID   string  `json:"book_id" jsonschema:"required,description=Hardcover book ID"`
	Status   string  `json:"status" jsonschema:"required,enum=reading,read,want-to-read,dnf"`
	Progress int     `json:"progress,omitempty" jsonschema:"description=Percentage 0-100"`
	Rating   float64 `json:"rating,omitempty" jsonschema:"description=Rating 0-5 with 0.25 increments"`
}

// AddToLibraryInput for the add_to_library tool
type AddToLibraryInput struct {
	ISBN      string `json:"isbn,omitempty" jsonschema:"description=ISBN-10 or ISBN-13"`
	Title     string `json:"title,omitempty" jsonschema:"description=Book title if ISBN unknown"`
	Author    string `json:"author,omitempty" jsonschema:"description=Author name if ISBN unknown"`
	Status    string `json:"status,omitempty" jsonschema:"default=want-to-read"`
	PlaceHold bool   `json:"place_hold,omitempty" jsonschema:"description=Also place library hold if available"`
}

// SearchLibraryInput for the search_library tool
type SearchLibraryInput struct {
	Query     string   `json:"query" jsonschema:"required"`
	Format    []string `json:"format,omitempty" jsonschema:"description=ebook, audiobook, or all"`
	Available bool     `json:"available,omitempty" jsonschema:"description=Only show currently available"`
}

// CheckAvailabilityInput for the check_availability tool
type CheckAvailabilityInput struct {
	ISBN   string `json:"isbn,omitempty"`
	Title  string `json:"title,omitempty"`
	Author string `json:"author,omitempty"`
}

// PlaceHoldInput for the place_hold tool
type PlaceHoldInput struct {
	MediaID    string `json:"media_id" jsonschema:"required,description=OverDrive media ID"`
	Format     string `json:"format" jsonschema:"required,enum=ebook,audiobook"`
	AutoBorrow bool   `json:"auto_borrow,omitempty" jsonschema:"default=true"`
}

// FindBookEverywhereInput for the find_book_everywhere tool
type FindBookEverywhereInput struct {
	Query string `json:"query" jsonschema:"required,description=Title, author, or ISBN"`
}

// BestWayToReadInput for the best_way_to_read tool
type BestWayToReadInput struct {
	BookID      string   `json:"book_id,omitempty"`
	ISBN        string   `json:"isbn,omitempty"`
	Preferences []string `json:"preferences,omitempty" jsonschema:"description=Priority: library, local, ebook, any"`
}

// AddToTBRInput for the add_to_tbr tool
type AddToTBRInput struct {
	ISBN          string `json:"isbn,omitempty"`
	Title         string `json:"title,omitempty"`
	Author        string `json:"author,omitempty"`
	PlaceHold     bool   `json:"place_hold,omitempty" jsonschema:"description=Auto-place library hold if available"`
	NotifyInStock bool   `json:"notify_in_stock,omitempty" jsonschema:"description=Notify when available at library"`
}

// registerTools registers all BookLife tools with the MCP server
func (s *Server) registerTools() {
	// === Library Management (Hardcover) ===
	
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "search_books",
		Description: "Search for books across Hardcover, Open Library, and Wikidata",
	}, s.handleSearchBooks)
	
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_my_library",
		Description: "Get your reading list from Hardcover (currently reading, TBR, finished, etc.)",
	}, s.handleGetMyLibrary)
	
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "update_reading_status",
		Description: "Update a book's status, progress, or rating in Hardcover",
	}, s.handleUpdateReadingStatus)
	
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "add_to_library",
		Description: "Add a book to your Hardcover library",
	}, s.handleAddToLibrary)
	
	// === Library Access (Libby) ===
	
	if s.libby != nil {
		mcp.AddTool(s.mcpServer, &mcp.Tool{
			Name:        "search_library",
			Description: "Search Austin Public Library catalog via Libby for ebooks and audiobooks",
		}, s.handleSearchLibrary)
		
		mcp.AddTool(s.mcpServer, &mcp.Tool{
			Name:        "check_availability",
			Description: "Check if a specific book is available at the library",
		}, s.handleCheckAvailability)
		
		mcp.AddTool(s.mcpServer, &mcp.Tool{
			Name:        "get_loans",
			Description: "Get your current Libby loans with due dates",
		}, s.handleGetLoans)
		
		mcp.AddTool(s.mcpServer, &mcp.Tool{
			Name:        "get_holds",
			Description: "Get your current library holds and queue position",
		}, s.handleGetHolds)
		
		mcp.AddTool(s.mcpServer, &mcp.Tool{
			Name:        "place_hold",
			Description: "Place a hold on a library ebook or audiobook",
		}, s.handlePlaceHold)
	}
	
	// === Unified Actions ===
	
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "find_book_everywhere",
		Description: "Search all sources for a book and show availability across library, local stores, and online",
	}, s.handleFindBookEverywhere)
	
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "best_way_to_read",
		Description: "Determine the best way to access a book (library, local bookstore, ebook, etc.)",
	}, s.handleBestWayToRead)
	
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "add_to_tbr",
		Description: "Add a book to your TBR list, optionally placing a library hold",
	}, s.handleAddToTBR)
}

// === Tool Handlers ===

func (s *Server) handleSearchBooks(ctx context.Context, req *mcp.CallToolRequest, input SearchBooksInput) (*mcp.CallToolResult, any, error) {
	if input.Query == "" {
		return nil, nil, fmt.Errorf("query is required")
	}
	
	limit := input.Limit
	if limit == 0 {
		limit = 10
	}
	
	var results []models.Book
	
	// Search Hardcover first
	if s.hardcover != nil {
		books, err := s.hardcover.SearchBooks(ctx, input.Query, limit)
		if err == nil {
			results = append(results, books...)
		}
	}
	
	// Enrich with Open Library data
	if s.openlibrary != nil {
		for i := range results {
			if results[i].ISBN13 != "" {
				olData, err := s.openlibrary.GetByISBN(ctx, results[i].ISBN13)
				if err == nil {
					// Merge Open Library data
					if results[i].Description == "" {
						results[i].Description = olData.Description
					}
				}
			}
		}
	}
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Found %d books", len(results)),
			},
		},
	}, results, nil
}

func (s *Server) handleGetMyLibrary(ctx context.Context, req *mcp.CallToolRequest, input GetMyLibraryInput) (*mcp.CallToolResult, any, error) {
	if s.hardcover == nil {
		return nil, nil, fmt.Errorf("Hardcover is not configured")
	}
	
	status := input.Status
	if status == "" {
		status = "all"
	}
	
	limit := input.Limit
	if limit == 0 {
		limit = 20
	}
	
	books, err := s.hardcover.GetUserBooks(ctx, status, limit)
	if err != nil {
		return nil, nil, fmt.Errorf("fetching library: %w", err)
	}
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Retrieved %d books from your library", len(books)),
			},
		},
	}, books, nil
}

func (s *Server) handleUpdateReadingStatus(ctx context.Context, req *mcp.CallToolRequest, input UpdateReadingStatusInput) (*mcp.CallToolResult, any, error) {
	if s.hardcover == nil {
		return nil, nil, fmt.Errorf("Hardcover is not configured")
	}
	
	err := s.hardcover.UpdateBookStatus(ctx, input.BookID, input.Status, input.Progress, input.Rating)
	if err != nil {
		return nil, nil, fmt.Errorf("updating status: %w", err)
	}
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Updated book %s to status '%s'", input.BookID, input.Status),
			},
		},
	}, nil, nil
}

func (s *Server) handleAddToLibrary(ctx context.Context, req *mcp.CallToolRequest, input AddToLibraryInput) (*mcp.CallToolResult, any, error) {
	if s.hardcover == nil {
		return nil, nil, fmt.Errorf("Hardcover is not configured")
	}
	
	status := input.Status
	if status == "" {
		status = "want-to-read"
	}
	
	bookID, err := s.hardcover.AddBook(ctx, input.ISBN, input.Title, input.Author, status)
	if err != nil {
		return nil, nil, fmt.Errorf("adding book: %w", err)
	}
	
	result := fmt.Sprintf("Added book to library with ID %s", bookID)
	
	// Optionally place library hold
	if input.PlaceHold && s.libby != nil {
		avail, err := s.libby.CheckAvailability(ctx, input.ISBN, input.Title, input.Author)
		if err == nil && avail != nil {
			_, err := s.libby.PlaceHold(ctx, avail.MediaID, "ebook", true)
			if err == nil {
				result += "; also placed library hold"
			}
		}
	}
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result},
		},
	}, map[string]string{"book_id": bookID}, nil
}

func (s *Server) handleSearchLibrary(ctx context.Context, req *mcp.CallToolRequest, input SearchLibraryInput) (*mcp.CallToolResult, any, error) {
	if s.libby == nil {
		return nil, nil, fmt.Errorf("Libby is not configured")
	}
	
	results, err := s.libby.Search(ctx, input.Query, input.Format, input.Available)
	if err != nil {
		return nil, nil, fmt.Errorf("searching library: %w", err)
	}
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Found %d results in library catalog", len(results)),
			},
		},
	}, results, nil
}

func (s *Server) handleCheckAvailability(ctx context.Context, req *mcp.CallToolRequest, input CheckAvailabilityInput) (*mcp.CallToolResult, any, error) {
	if s.libby == nil {
		return nil, nil, fmt.Errorf("Libby is not configured")
	}
	
	avail, err := s.libby.CheckAvailability(ctx, input.ISBN, input.Title, input.Author)
	if err != nil {
		return nil, nil, fmt.Errorf("checking availability: %w", err)
	}
	
	if avail == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Book not found in library catalog"},
			},
		}, nil, nil
	}
	
	var statusText string
	if avail.EbookAvailable || avail.AudiobookAvailable {
		statusText = "Available now!"
	} else {
		statusText = fmt.Sprintf("Not currently available. Estimated wait: %d days", avail.EstimatedWaitDays)
	}
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: statusText},
		},
	}, avail, nil
}

func (s *Server) handleGetLoans(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
	if s.libby == nil {
		return nil, nil, fmt.Errorf("Libby is not configured")
	}
	
	loans, err := s.libby.GetLoans(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("getting loans: %w", err)
	}
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("You have %d active loans", len(loans)),
			},
		},
	}, loans, nil
}

func (s *Server) handleGetHolds(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
	if s.libby == nil {
		return nil, nil, fmt.Errorf("Libby is not configured")
	}
	
	holds, err := s.libby.GetHolds(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("getting holds: %w", err)
	}
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("You have %d active holds", len(holds)),
			},
		},
	}, holds, nil
}

func (s *Server) handlePlaceHold(ctx context.Context, req *mcp.CallToolRequest, input PlaceHoldInput) (*mcp.CallToolResult, any, error) {
	if s.libby == nil {
		return nil, nil, fmt.Errorf("Libby is not configured")
	}
	
	holdID, err := s.libby.PlaceHold(ctx, input.MediaID, input.Format, input.AutoBorrow)
	if err != nil {
		return nil, nil, fmt.Errorf("placing hold: %w", err)
	}
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Hold placed successfully (ID: %s)", holdID),
			},
		},
	}, map[string]string{"hold_id": holdID}, nil
}

func (s *Server) handleFindBookEverywhere(ctx context.Context, req *mcp.CallToolRequest, input FindBookEverywhereInput) (*mcp.CallToolResult, any, error) {
	result := models.UnifiedBookResult{
		Query: input.Query,
	}
	
	// Search Hardcover for metadata
	if s.hardcover != nil {
		books, err := s.hardcover.SearchBooks(ctx, input.Query, 1)
		if err == nil && len(books) > 0 {
			result.Book = &books[0]
		}
	}
	
	// Check library availability
	if s.libby != nil && result.Book != nil {
		avail, err := s.libby.CheckAvailability(ctx, result.Book.ISBN13, result.Book.Title, "")
		if err == nil {
			result.LibraryAvailability = avail
		}
	}
	
	// Generate recommendation
	result.Recommendation = generateRecommendation(&result)
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result.Recommendation},
		},
	}, result, nil
}

func (s *Server) handleBestWayToRead(ctx context.Context, req *mcp.CallToolRequest, input BestWayToReadInput) (*mcp.CallToolResult, any, error) {
	// Similar to find_book_everywhere but returns prioritized options
	var options []string
	
	// Check library first
	if s.libby != nil {
		avail, err := s.libby.CheckAvailability(ctx, input.ISBN, "", "")
		if err == nil && avail != nil {
			if avail.EbookAvailable || avail.AudiobookAvailable {
				options = append(options, "📚 Available NOW at Austin Public Library via Libby (FREE)")
			} else {
				options = append(options, fmt.Sprintf("📚 Can place hold at APL - estimated %d day wait (FREE)", avail.EstimatedWaitDays))
			}
		}
	}
	
	// TODO: Check local bookstores
	options = append(options, "🏪 Check BookPeople for local purchase")
	
	// Online fallback
	options = append(options, "💻 Available for purchase online")
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Options for reading this book:\n- %s", 
					fmt.Sprintf("\n- ")),
			},
		},
	}, options, nil
}

func (s *Server) handleAddToTBR(ctx context.Context, req *mcp.CallToolRequest, input AddToTBRInput) (*mcp.CallToolResult, any, error) {
	// Add to Hardcover
	addInput := AddToLibraryInput{
		ISBN:      input.ISBN,
		Title:     input.Title,
		Author:    input.Author,
		Status:    "want-to-read",
		PlaceHold: input.PlaceHold,
	}
	
	return s.handleAddToLibrary(ctx, req, addInput)
}

// Helper function to generate reading recommendations
func generateRecommendation(result *models.UnifiedBookResult) string {
	if result.LibraryAvailability != nil {
		if result.LibraryAvailability.EbookAvailable {
			return "🎉 Great news! This book is available RIGHT NOW as an ebook at Austin Public Library. Borrow it free with Libby!"
		}
		if result.LibraryAvailability.AudiobookAvailable {
			return "🎧 This audiobook is available RIGHT NOW at Austin Public Library. Listen free with Libby!"
		}
		if result.LibraryAvailability.EstimatedWaitDays < 14 {
			return fmt.Sprintf("📚 This book has a short wait at APL (~%d days). Place a hold for free access!", result.LibraryAvailability.EstimatedWaitDays)
		}
	}
	
	return "📖 Check local bookstores like BookPeople, or place a library hold for free access later."
}
