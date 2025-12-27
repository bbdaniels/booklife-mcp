package server

import (
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/user/booklife-mcp/internal/models"
)

// ===== Shared Types =====

// PaginationParams provides common pagination for list requests
// Page is 1-indexed (default: 1)
// PageSize is items per page (default: 20, max: 100)
type PaginationParams struct {
	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`
}

// getPagination returns page (default 1) and pageSize (default 20, max 100)
func getPagination(params PaginationParams) (page, pageSize int) {
	page = params.Page
	if page < 1 {
		page = 1
	}

	pageSize = params.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return page, pageSize
}

// offset returns the calculated offset for pagination
func (p PaginationParams) offset() int {
	page, pageSize := getPagination(p)
	return (page - 1) * pageSize
}

// calculatePagedResult creates a PagedResult from page, pageSize, and totalCount
func calculatePagedResult(page, pageSize, totalCount int) PagedResult {
	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	return PagedResult{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// PagedResult contains pagination metadata for list responses
type PagedResult struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// formatPagingFooter adds pagination info to the output
// itemCount is the number of items on the current page
func formatPagingFooter(paged PagedResult, itemCount int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n--- Page %d of %d (%d items shown, %d total) ---",
		paged.Page, paged.TotalPages, itemCount, paged.TotalCount))
	if paged.HasPrev {
		sb.WriteString(" (prev page available)")
	}
	if paged.HasNext {
		sb.WriteString(" (next page available)")
	}
	return sb.String()
}

// ===== Format Helpers for Cross-Tool Data Roundtripping =====

// formatBookForDisplay creates a detailed, human-readable representation with actionable IDs
// Returns text that shows all identifiers needed for cross-tool usage:
// - book_id → for update_reading_status, get_my_library
// - isbn → for add_to_library, check_availability
func formatBookForDisplay(book models.Book, index int) string {
	var sb strings.Builder

	// Header with index
	if index > 0 {
		sb.WriteString(fmt.Sprintf("[%d] ", index))
	}

	// Title line with subtitle
	sb.WriteString(book.Title)
	if book.Subtitle != "" {
		sb.WriteString(fmt.Sprintf(": %s", book.Subtitle))
	}
	sb.WriteString("\n")

	// Authors
	if len(book.Authors) > 0 {
		authorNames := make([]string, 0, len(book.Authors))
		for _, a := range book.Authors {
			authorNames = append(authorNames, a.Name)
		}
		sb.WriteString(fmt.Sprintf("    by %s\n", strings.Join(authorNames, ", ")))
	}

	// === CRITICAL: Identifiers for cross-tool usage ===
	var ids []string
	if book.HardcoverID != "" {
		ids = append(ids, fmt.Sprintf("book_id=%s", book.HardcoverID))
	}
	if book.ISBN13 != "" {
		ids = append(ids, fmt.Sprintf("isbn=%s", book.ISBN13))
	} else if book.ISBN10 != "" {
		ids = append(ids, fmt.Sprintf("isbn=%s", book.ISBN10))
	}
	if len(ids) > 0 {
		sb.WriteString(fmt.Sprintf("    IDs: %s\n", strings.Join(ids, ", ")))
	}

	// User status if available
	if book.UserStatus != nil {
		statusInfo := book.UserStatus.Status
		if book.UserStatus.Progress > 0 {
			statusInfo += fmt.Sprintf(" (%d%%)", book.UserStatus.Progress)
		}
		if book.UserStatus.Rating > 0 {
			statusInfo += fmt.Sprintf(" ⭐ %.1f", book.UserStatus.Rating)
		}
		sb.WriteString(fmt.Sprintf("    Status: %s\n", statusInfo))
	}

	// Series info
	if book.Series != nil {
		series := book.Series.Name
		if book.Series.Position > 0 {
			series += fmt.Sprintf(" #%g", book.Series.Position)
		}
		if book.Series.Total > 0 {
			series += fmt.Sprintf(" of %d", book.Series.Total)
		}
		sb.WriteString(fmt.Sprintf("    Series: %s\n", series))
	}

	// Publication info
	if book.Publisher != "" || book.PublishedDate != "" {
		pubInfo := ""
		if book.Publisher != "" {
			pubInfo = book.Publisher
		}
		if book.PublishedDate != "" {
			if pubInfo != "" {
				pubInfo += ", "
			}
			pubInfo += book.PublishedDate
		}
		sb.WriteString(fmt.Sprintf("    Publisher: %s\n", pubInfo))
	}

	// Page count
	if book.PageCount > 0 {
		sb.WriteString(fmt.Sprintf("    Pages: %d\n", book.PageCount))
	}

	// Genres
	if len(book.Genres) > 0 {
		sb.WriteString(fmt.Sprintf("    Genres: %s\n", strings.Join(book.Genres, ", ")))
	}

	// Community rating
	if book.HardcoverRating > 0 {
		rating := fmt.Sprintf("⭐ %.1f", book.HardcoverRating)
		if book.HardcoverCount > 0 {
			rating += fmt.Sprintf(" (%d ratings)", book.HardcoverCount)
		}
		sb.WriteString(fmt.Sprintf("    Community Rating: %s\n", rating))
	}

	// Library availability
	if book.LibraryAvailability != nil {
		sb.WriteString(formatLibraryAvailabilityForDisplay(book.LibraryAvailability, "    "))
	}

	return sb.String()
}

// formatLibraryAvailabilityForDisplay shows detailed availability with media_id
func formatLibraryAvailabilityForDisplay(avail *models.LibraryAvailability, indent string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%sLibrary Availability (%s):\n", indent, avail.LibraryName))

	// === CRITICAL: media_id for place_hold ===
	sb.WriteString(fmt.Sprintf("%s  media_id: %s\n", indent, avail.MediaID))

	if avail.EbookAvailable {
		sb.WriteString(fmt.Sprintf("%s  ✅ Ebook: Available now (%d copies)\n", indent, avail.EbookCopies))
	} else if avail.EbookWaitlistSize > 0 {
		sb.WriteString(fmt.Sprintf("%s  📚 Ebook: %d in wait list\n", indent, avail.EbookWaitlistSize))
	}

	if avail.AudiobookAvailable {
		sb.WriteString(fmt.Sprintf("%s  ✅ Audiobook: Available now (%d copies)\n", indent, avail.AudiobookCopies))
	} else if avail.AudiobookWaitlistSize > 0 {
		sb.WriteString(fmt.Sprintf("%s  🎧 Audiobook: %d in wait list\n", indent, avail.AudiobookWaitlistSize))
	}

	if avail.EstimatedWaitDays > 0 {
		sb.WriteString(fmt.Sprintf("%s  Estimated wait: ~%d days\n", indent, avail.EstimatedWaitDays))
	}

	return sb.String()
}

// formatAuthorsList formats authors for display
func formatAuthorsList(authors []models.Contributor) string {
	if len(authors) == 0 {
		return "Unknown"
	}
	names := make([]string, 0, len(authors))
	for _, a := range authors {
		names = append(names, a.Name)
	}
	return strings.Join(names, ", ")
}

// registerTools registers all BookLife tools with the MCP server
func (s *Server) registerTools() {
	// === Hardcover (Reading Tracker) ===

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name: "hardcover_search_books",
		Description: `Search for books in Hardcover catalog.
Returns detailed book information with identifiers for cross-tool usage.
Example: {"query": "Project Hail Mary"}
Example: {"query": "Andy Weir", "page_size": 5, "sort_by": "rating"}`,
	}, s.handleSearchBooks)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name: "hardcover_get_my_library",
		Description: `Get your reading list from Hardcover.
Returns detailed book information with your reading status.
Example: {"status": "reading"} - currently reading
Example: {"status": "want-to-read"} - TBR list
Example: {"status": "read", "page_size": 10} - recently finished`,
	}, s.handleGetMyLibrary)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name: "hardcover_update_reading_status",
		Description: `Update a book's status, progress, or rating in Hardcover.
Requires book_id from search_books or get_my_library.
Example: {"book_id": "123", "status": "reading", "progress": 50}
Example: {"book_id": "123", "status": "read", "rating": 4.5}`,
	}, s.handleUpdateReadingStatus)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name: "hardcover_add_to_library",
		Description: `Add a book to your Hardcover library.
Can use ISBN from search_books or title/author.
Example: {"isbn": "9780593135204", "status": "want-to-read"}
Example: {"title": "Project Hail Mary", "author": "Andy Weir"}`,
	}, s.handleAddToLibrary)

	// === Libby (Library Access) ===

	if s.libby != nil {
		mcp.AddTool(s.mcpServer, &mcp.Tool{
			Name: "libby_search",
			Description: `Search library catalog via Libby for ebooks and audiobooks.
Returns detailed results with media_id for placing holds.
Example: {"query": "Project Hail Mary"}
Example: {"query": "Brandon Sanderson", "available": true}`,
		}, s.handleSearchLibrary)

		mcp.AddTool(s.mcpServer, &mcp.Tool{
			Name: "libby_check_availability",
			Description: `Check if a book is available at the library.
Returns detailed availability with media_id for placing holds.
Example: {"isbn": "9780593135204"}
Example: {"title": "Mistborn", "author": "Brandon Sanderson"}`,
		}, s.handleCheckAvailability)

		mcp.AddTool(s.mcpServer, &mcp.Tool{
			Name: "libby_get_loans",
			Description: `Get your current Libby loans with due dates.
Returns detailed loan information.
Example: {} - returns all current checkouts`,
		}, s.handleGetLoans)

		mcp.AddTool(s.mcpServer, &mcp.Tool{
			Name: "libby_get_holds",
			Description: `Get your current library holds and queue position.
Returns detailed hold information with media_id.
Example: {} - returns all active holds`,
		}, s.handleGetHolds)

		mcp.AddTool(s.mcpServer, &mcp.Tool{
			Name: "libby_place_hold",
			Description: `Place a hold on a library ebook or audiobook.
Requires media_id from libby_search or libby_check_availability.
Example: {"media_id": "12345", "format": "ebook"}
Example: {"media_id": "12345", "format": "audiobook", "auto_borrow": true}`,
		}, s.handlePlaceHold)

		mcp.AddTool(s.mcpServer, &mcp.Tool{
			Name: "libby_get_tags",
			Description: `Get your Libby book tags for organization.
Returns all tags and the media IDs tagged with each.
Example: {} - returns all tags
Example: {"tag": "favorites"} - returns books tagged as "favorites"`,
		}, s.handleGetTags)

		mcp.AddTool(s.mcpServer, &mcp.Tool{
			Name: "libby_add_tag",
			Description: `Add a tag to a library book for organization.
Requires media_id from libby_search or libby_get_loans.
Example: {"media_id": "12345", "tag": "favorites"}
Example: {"media_id": "12345", "tag": "sci-fi"}`,
		}, s.handleAddTag)

		mcp.AddTool(s.mcpServer, &mcp.Tool{
			Name: "libby_remove_tag",
			Description: `Remove a tag from a library book.
Requires media_id and the tag to remove.
Example: {"media_id": "12345", "tag": "favorites"}`,
		}, s.handleRemoveTag)
	}

	// === Unified Actions ===

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name: "booklife_find_book_everywhere",
		Description: `Search all sources for a book and show availability.
Returns comprehensive availability with all actionable IDs.
Example: {"query": "The Name of the Wind"}`,
	}, s.handleFindBookEverywhere)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name: "booklife_best_way_to_read",
		Description: `Determine the best way to access a book (library, bookstore, etc.).
Returns prioritized options with identifiers.
Example: {"isbn": "9780756404741"}`,
	}, s.handleBestWayToRead)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name: "booklife_add_to_tbr",
		Description: `Add a book to your TBR list, optionally placing a library hold.
Can use ISBN from search results.
Example: {"isbn": "9780756404741", "place_hold": true}
Example: {"title": "The Way of Kings", "author": "Brandon Sanderson"}`,
	}, s.handleAddToTBR)

	// === Local History Store ===

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "history_import_timeline",
		Description: `Import Libby reading history from a timeline export URL.
This imports your complete Libby reading history into the local store.
Example: {"url": "https://share.libbyapp.com/data/{uuid}/libbytimeline-all-loans.json"}`,
	}, s.handleImportTimeline)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "history_sync_current_loans",
		Description: `Sync current Libby loans to local history store.
This captures your current checkouts for historical tracking.
Example: {}`,
	}, s.handleSyncCurrentLoans)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "history_get",
		Description: `Get reading history from local store with pagination.
Returns all imported timeline entries and synced loans.
Example: {"page": 1, "page_size": 20}`,
	}, s.handleGetLocalHistory)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "history_search",
		Description: `Search reading history by title or author.
Searches all imported timeline entries for matches.
Example: {"query": "Sanderson", "page": 1, "page_size": 20}
Example: {"query": "Mistborn"}`,
	}, s.handleSearchLocalHistory)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "history_stats",
		Description: `Get reading statistics from local history store.
Returns breakdowns by format, library, year, and totals.
Example: {}`,
	}, s.handleGetHistoryStats)
}
