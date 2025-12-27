package history

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/user/booklife-mcp/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

// Store manages local reading history using SQLite
type Store struct {
	db *sql.DB
	mu sync.RWMutex
}

// NewStore creates or opens a local history database
func NewStore(dataDir string) (*Store, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("creating data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "history.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	store := &Store{db: db}
	if err := store.init(); err != nil {
		db.Close()
		return nil, fmt.Errorf("initializing database: %w", err)
	}

	return store, nil
}

// init creates the database schema
func (s *Store) init() error {
	query := `
	CREATE TABLE IF NOT EXISTS history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title_id TEXT NOT NULL,
		title TEXT NOT NULL,
		author TEXT NOT NULL,
		publisher TEXT,
		isbn TEXT,
		timestamp INTEGER NOT NULL,
		activity TEXT NOT NULL,
		details TEXT,
		library TEXT NOT NULL,
		library_key TEXT NOT NULL,
		format TEXT NOT NULL,
		cover_url TEXT,
		color TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(title_id, timestamp, activity)
	);

	CREATE INDEX IF NOT EXISTS idx_timestamp ON history(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_title ON history(title);
	CREATE INDEX IF NOT EXISTS idx_author ON history(author);
	CREATE INDEX IF NOT EXISTS idx_library ON history(library);
	CREATE INDEX IF NOT EXISTS idx_activity ON history(activity);
	`

	_, err := s.db.Exec(query)
	return err
}

// ImportTimeline imports timeline data from Libby export
func (s *Store) ImportTimeline(timeline *models.TimelineResponse) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("beginning transaction: %w", err)
	}

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO history
		(title_id, title, author, publisher, isbn, timestamp, activity, details, library, library_key, format, cover_url, color)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	count := 0
	for _, entry := range timeline.Timeline {
		// Parse cover URL from nested structure
		coverURL := entry.CoverURL
		color := entry.Color

		_, err := stmt.Exec(
			entry.TitleID,
			entry.Title,
			entry.Author,
			entry.Publisher,
			entry.ISBN,
			entry.Timestamp,
			entry.Activity,
			entry.Details,
			entry.Library,
			entry.LibraryKey,
			entry.Format,
			coverURL,
			color,
		)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("inserting entry %s: %w", entry.Title, err)
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("committing transaction: %w", err)
	}

	return count, nil
}

// ImportCurrentLoan imports a current loan from Libby sync
func (s *Store) ImportCurrentLoan(loan models.LibbyLoan) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a "Borrowed" activity for current loans
	timestamp := loan.CheckoutDate.UnixMilli()

	query := `
		INSERT OR REPLACE INTO history
		(title_id, title, author, timestamp, activity, details, library, library_key, format, cover_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		loan.ID,
		loan.Title,
		loan.Author,
		timestamp,
		"Borrowed",
		fmt.Sprintf("%d days", int(time.Until(loan.DueDate).Hours()/24)),
		"Current Loan",
		"",
		loan.Format,
		loan.CoverURL,
	)

	return err
}

// GetHistory returns all history entries with pagination
func (s *Store) GetHistory(offset, limit int) ([]models.TimelineEntry, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get total count
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM history").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("counting entries: %w", err)
	}

	// Get paginated entries
	query := `
		SELECT title_id, title, author, publisher, isbn, timestamp, activity,
		       details, library, library_key, format, cover_url, color
		FROM history
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("querying history: %w", err)
	}
	defer rows.Close()

	entries := []models.TimelineEntry{}
	for rows.Next() {
		var e models.TimelineEntry
		err := rows.Scan(
			&e.TitleID, &e.Title, &e.Author, &e.Publisher, &e.ISBN,
			&e.Timestamp, &e.Activity, &e.Details, &e.Library,
			&e.LibraryKey, &e.Format, &e.CoverURL, &e.Color,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scanning entry: %w", err)
		}
		entries = append(entries, e)
	}

	return entries, total, nil
}

// SearchHistory searches history by title or author
func (s *Store) SearchHistory(query string, offset, limit int) ([]models.TimelineEntry, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get total count
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM history WHERE title LIKE ? OR author LIKE ?",
		"%"+query+"%", "%"+query+"%").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("counting entries: %w", err)
	}

	// Get paginated entries
	searchQuery := `
		SELECT title_id, title, author, publisher, isbn, timestamp, activity,
		       details, library, library_key, format, cover_url, color
		FROM history
		WHERE title LIKE ? OR author LIKE ?
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(searchQuery, "%"+query+"%", "%"+query+"%", limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("querying history: %w", err)
	}
	defer rows.Close()

	entries := []models.TimelineEntry{}
	for rows.Next() {
		var e models.TimelineEntry
		err := rows.Scan(
			&e.TitleID, &e.Title, &e.Author, &e.Publisher, &e.ISBN,
			&e.Timestamp, &e.Activity, &e.Details, &e.Library,
			&e.LibraryKey, &e.Format, &e.CoverURL, &e.Color,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scanning entry: %w", err)
		}
		entries = append(entries, e)
	}

	return entries, total, nil
}

// GetStats returns reading statistics
func (s *Store) GetStats() (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]interface{})

	// Total entries
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM history").Scan(&total)
	if err != nil {
		return nil, err
	}
	stats["total_entries"] = total

	// Books borrowed
	var borrowed int
	err = s.db.QueryRow("SELECT COUNT(DISTINCT title_id) FROM history WHERE activity = 'Borrowed'").Scan(&borrowed)
	if err != nil {
		return nil, err
	}
	stats["unique_borrows"] = borrowed

	// Format breakdown
	formatQuery := `
		SELECT format, COUNT(*) as count
		FROM history
		GROUP BY format
	`
	rows, err := s.db.Query(formatQuery)
	if err == nil {
		formats := make(map[string]int)
		for rows.Next() {
			var format string
			var count int
			rows.Scan(&format, &count)
			formats[format] = count
		}
		rows.Close()
		stats["by_format"] = formats
	}

	// Library breakdown
	libraryQuery := `
		SELECT library, COUNT(*) as count
		FROM history
		GROUP BY library
		ORDER BY count DESC
	`
	rows, err = s.db.Query(libraryQuery)
	if err == nil {
		libraries := make(map[string]int)
		for rows.Next() {
			var library string
			var count int
			rows.Scan(&library, &count)
			libraries[library] = count
		}
		rows.Close()
		stats["by_library"] = libraries
	}

	// Date range
	var firstTimestamp, lastTimestamp int64
	s.db.QueryRow("SELECT MIN(timestamp), MAX(timestamp) FROM history").Scan(&firstTimestamp, &lastTimestamp)
	if firstTimestamp > 0 {
		stats["first_activity"] = time.UnixMilli(firstTimestamp).Format("2006-01-02")
	}
	if lastTimestamp > 0 {
		stats["last_activity"] = time.UnixMilli(lastTimestamp).Format("2006-01-02")
	}

	return stats, nil
}

// GetYearlyStats returns statistics grouped by year
func (s *Store) GetYearlyStats() ([]map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
		SELECT
			IFNULL(strftime('%Y', datetime(timestamp/1000, 'unix')), 'Unknown') as year,
			COUNT(*) as total,
			SUM(CASE WHEN activity = 'Borrowed' THEN 1 ELSE 0 END) as borrowed
		FROM history
		WHERE timestamp IS NOT NULL AND timestamp > 0
		GROUP BY year
		ORDER BY year DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("querying yearly stats: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var year string
		var total, borrowed int
		err := rows.Scan(&year, &total, &borrowed)
		if err != nil {
			return nil, fmt.Errorf("scanning yearly stat: %w", err)
		}
		results = append(results, map[string]interface{}{
			"year":     year,
			"total":    total,
			"borrowed": borrowed,
		})
	}

	return results, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// ExportJSON exports all history as JSON
func (s *Store) ExportJSON() ([]byte, error) {
	entries, _, err := s.GetHistory(0, 10000) // Get all entries
	if err != nil {
		return nil, err
	}

	timeline := &models.TimelineResponse{
		Version:  1,
		Timeline: entries,
	}

	return json.MarshalIndent(timeline, "", "  ")
}
