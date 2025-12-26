// Package providers defines interfaces for external book service providers.
package providers

import (
	"context"

	"github.com/user/booklife-mcp/internal/models"
)

// HardcoverProvider defines the interface for Hardcover reading tracker operations.
type HardcoverProvider interface {
	// SearchBooks searches for books by query string.
	SearchBooks(ctx context.Context, query string, limit int) ([]models.Book, error)

	// GetBook retrieves a single book by its Hardcover ID.
	GetBook(ctx context.Context, bookID string) (*models.Book, error)

	// GetUserBooks retrieves books from the user's library by status.
	// Valid statuses: "reading", "read", "want-to-read", "dnf"
	GetUserBooks(ctx context.Context, status string, limit int) ([]models.Book, error)

	// UpdateBookStatus updates the user's status for a book.
	UpdateBookStatus(ctx context.Context, bookID, status string, progress int, rating float64) error

	// AddBook adds a book to the user's library.
	// Returns the new user book ID.
	AddBook(ctx context.Context, isbn, title, author, status string) (string, error)

	// GetReadingStats retrieves reading statistics for a given year.
	GetReadingStats(ctx context.Context, year int) (*models.ReadingStats, error)
}

// LibbyProvider defines the interface for Libby/OverDrive library operations.
type LibbyProvider interface {
	// Search searches the library catalog.
	Search(ctx context.Context, query string, formats []string, available bool) ([]models.Book, error)

	// CheckAvailability checks if a book is available at the library.
	CheckAvailability(ctx context.Context, isbn, title, author string) (*models.LibraryAvailability, error)

	// GetLoans retrieves the user's current library loans.
	GetLoans(ctx context.Context) ([]models.LibbyLoan, error)

	// GetHolds retrieves the user's current library holds.
	GetHolds(ctx context.Context) ([]models.LibbyHold, error)

	// PlaceHold places a hold on a library item.
	// Returns the hold ID.
	PlaceHold(ctx context.Context, mediaID, format string, autoBorrow bool) (string, error)
}

// OpenLibraryProvider defines the interface for Open Library metadata operations.
type OpenLibraryProvider interface {
	// GetByISBN retrieves book metadata by ISBN.
	GetByISBN(ctx context.Context, isbn string) (*models.Book, error)

	// Search searches for books by query string.
	Search(ctx context.Context, query string, limit int) ([]models.Book, error)

	// GetCoverURL returns the URL for a book cover image.
	GetCoverURL(isbn string, size string) string

	// GetDescription retrieves a book's description by Open Library work ID.
	GetDescription(ctx context.Context, workID string) (string, error)
}
