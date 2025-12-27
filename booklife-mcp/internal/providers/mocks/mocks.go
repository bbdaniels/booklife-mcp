// Package mocks provides test mock implementations for provider interfaces.
package mocks

import (
	"context"

	"github.com/user/booklife-mcp/internal/models"
)

// MockHardcoverProvider is a configurable mock for testing.
type MockHardcoverProvider struct {
	SearchBooksFunc      func(ctx context.Context, query string, offset, limit int) ([]models.Book, int, error)
	GetBookFunc          func(ctx context.Context, bookID string) (*models.Book, error)
	GetUserBooksFunc     func(ctx context.Context, status string, offset, limit int) ([]models.Book, int, error)
	UpdateBookStatusFunc func(ctx context.Context, bookID, status string, progress int, rating float64) error
	AddBookFunc          func(ctx context.Context, isbn, title, author, status string) (string, error)
	GetReadingStatsFunc  func(ctx context.Context, year int) (*models.ReadingStats, error)
}

func (m *MockHardcoverProvider) SearchBooks(ctx context.Context, query string, offset, limit int) ([]models.Book, int, error) {
	if m.SearchBooksFunc != nil {
		return m.SearchBooksFunc(ctx, query, offset, limit)
	}
	return nil, 0, nil
}

func (m *MockHardcoverProvider) GetBook(ctx context.Context, bookID string) (*models.Book, error) {
	if m.GetBookFunc != nil {
		return m.GetBookFunc(ctx, bookID)
	}
	return nil, nil
}

func (m *MockHardcoverProvider) GetUserBooks(ctx context.Context, status string, offset, limit int) ([]models.Book, int, error) {
	if m.GetUserBooksFunc != nil {
		return m.GetUserBooksFunc(ctx, status, offset, limit)
	}
	return nil, 0, nil
}

func (m *MockHardcoverProvider) UpdateBookStatus(ctx context.Context, bookID, status string, progress int, rating float64) error {
	if m.UpdateBookStatusFunc != nil {
		return m.UpdateBookStatusFunc(ctx, bookID, status, progress, rating)
	}
	return nil
}

func (m *MockHardcoverProvider) AddBook(ctx context.Context, isbn, title, author, status string) (string, error) {
	if m.AddBookFunc != nil {
		return m.AddBookFunc(ctx, isbn, title, author, status)
	}
	return "", nil
}

func (m *MockHardcoverProvider) GetReadingStats(ctx context.Context, year int) (*models.ReadingStats, error) {
	if m.GetReadingStatsFunc != nil {
		return m.GetReadingStatsFunc(ctx, year)
	}
	return nil, nil
}

// MockLibbyProvider is a configurable mock for testing.
type MockLibbyProvider struct {
	SearchFunc            func(ctx context.Context, query string, formats []string, available bool, offset, limit int) ([]models.Book, int, error)
	CheckAvailabilityFunc func(ctx context.Context, isbn, title, author string) (*models.LibraryAvailability, error)
	GetLoansFunc          func(ctx context.Context) ([]models.LibbyLoan, error)
	GetHoldsFunc          func(ctx context.Context) ([]models.LibbyHold, error)
	PlaceHoldFunc         func(ctx context.Context, mediaID, format string, autoBorrow bool) (string, error)
	GetHistoryFunc        func(ctx context.Context, offset, limit int) ([]models.LibbyHistoryItem, int, error)
	GetTagsFunc           func(ctx context.Context) (map[string][]string, error)
	AddTagFunc            func(ctx context.Context, mediaID, tag string) error
	RemoveTagFunc         func(ctx context.Context, mediaID, tag string) error
}

func (m *MockLibbyProvider) Search(ctx context.Context, query string, formats []string, available bool, offset, limit int) ([]models.Book, int, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, query, formats, available, offset, limit)
	}
	return nil, 0, nil
}

func (m *MockLibbyProvider) CheckAvailability(ctx context.Context, isbn, title, author string) (*models.LibraryAvailability, error) {
	if m.CheckAvailabilityFunc != nil {
		return m.CheckAvailabilityFunc(ctx, isbn, title, author)
	}
	return nil, nil
}

func (m *MockLibbyProvider) GetLoans(ctx context.Context) ([]models.LibbyLoan, error) {
	if m.GetLoansFunc != nil {
		return m.GetLoansFunc(ctx)
	}
	return nil, nil
}

func (m *MockLibbyProvider) GetHolds(ctx context.Context) ([]models.LibbyHold, error) {
	if m.GetHoldsFunc != nil {
		return m.GetHoldsFunc(ctx)
	}
	return nil, nil
}

func (m *MockLibbyProvider) PlaceHold(ctx context.Context, mediaID, format string, autoBorrow bool) (string, error) {
	if m.PlaceHoldFunc != nil {
		return m.PlaceHoldFunc(ctx, mediaID, format, autoBorrow)
	}
	return "", nil
}

func (m *MockLibbyProvider) GetHistory(ctx context.Context, offset, limit int) ([]models.LibbyHistoryItem, int, error) {
	if m.GetHistoryFunc != nil {
		return m.GetHistoryFunc(ctx, offset, limit)
	}
	return nil, 0, nil
}

func (m *MockLibbyProvider) GetTags(ctx context.Context) (map[string][]string, error) {
	if m.GetTagsFunc != nil {
		return m.GetTagsFunc(ctx)
	}
	return nil, nil
}

func (m *MockLibbyProvider) AddTag(ctx context.Context, mediaID, tag string) error {
	if m.AddTagFunc != nil {
		return m.AddTagFunc(ctx, mediaID, tag)
	}
	return nil
}

func (m *MockLibbyProvider) RemoveTag(ctx context.Context, mediaID, tag string) error {
	if m.RemoveTagFunc != nil {
		return m.RemoveTagFunc(ctx, mediaID, tag)
	}
	return nil
}

// MockOpenLibraryProvider is a configurable mock for testing.
type MockOpenLibraryProvider struct {
	GetByISBNFunc      func(ctx context.Context, isbn string) (*models.Book, error)
	SearchFunc         func(ctx context.Context, query string, offset, limit int) ([]models.Book, int, error)
	GetCoverURLFunc    func(isbn string, size string) string
	GetDescriptionFunc func(ctx context.Context, workID string) (string, error)
}

func (m *MockOpenLibraryProvider) GetByISBN(ctx context.Context, isbn string) (*models.Book, error) {
	if m.GetByISBNFunc != nil {
		return m.GetByISBNFunc(ctx, isbn)
	}
	return nil, nil
}

func (m *MockOpenLibraryProvider) Search(ctx context.Context, query string, offset, limit int) ([]models.Book, int, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, query, offset, limit)
	}
	return nil, 0, nil
}

func (m *MockOpenLibraryProvider) GetCoverURL(isbn string, size string) string {
	if m.GetCoverURLFunc != nil {
		return m.GetCoverURLFunc(isbn, size)
	}
	return ""
}

func (m *MockOpenLibraryProvider) GetDescription(ctx context.Context, workID string) (string, error) {
	if m.GetDescriptionFunc != nil {
		return m.GetDescriptionFunc(ctx, workID)
	}
	return "", nil
}
