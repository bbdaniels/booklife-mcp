package history

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/user/booklife-mcp/internal/models"
)

// Importer handles importing timeline data from Libby exports
type Importer struct {
	store *Store
}

// NewImporter creates a new timeline importer
func NewImporter(store *Store) *Importer {
	return &Importer{store: store}
}

// FetchTimeline fetches timeline data from a Libby share URL
func (im *Importer) FetchTimeline(url string) (*models.TimelineResponse, error) {
	client := &http.Client{Timeout: 60 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching timeline: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	// Use ParseTimeline which handles nested objects correctly
	return ParseTimeline(data)
}

// ImportTimeline fetches and imports timeline data from a URL
func (im *Importer) ImportTimeline(url string) (int, error) {
	// Fetch timeline
	timeline, err := im.FetchTimeline(url)
	if err != nil {
		return 0, err
	}

	// Import to store
	count, err := im.store.ImportTimeline(timeline)
	if err != nil {
		return 0, fmt.Errorf("importing to store: %w", err)
	}

	return count, nil
}

// ParseTimeline parses timeline JSON data
func ParseTimeline(data []byte) (*models.TimelineResponse, error) {
	var raw struct {
		Version  int                      `json:"version"`
		Timeline []map[string]interface{} `json:"timeline"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	entries := make([]models.TimelineEntry, 0, len(raw.Timeline))
	for _, item := range raw.Timeline {
		entry := models.TimelineEntry{
			Activity:  "Borrowed",
			Format:    "ebook",
			Timestamp: time.Now().UnixMilli(),
		}

		// Parse title
		if title, ok := item["title"].(map[string]interface{}); ok {
			if text, ok := title["text"].(string); ok {
				entry.Title = text
			}
			if titleID, ok := title["titleId"].(string); ok {
				entry.TitleID = titleID
			}
		}

		// Simple fields
		if v, ok := item["author"].(string); ok {
			entry.Author = v
		}
		if v, ok := item["publisher"].(string); ok {
			entry.Publisher = v
		}
		if v, ok := item["isbn"].(string); ok {
			entry.ISBN = v
		}
		if v, ok := item["activity"].(string); ok {
			entry.Activity = v
		}
		if v, ok := item["details"].(string); ok {
			entry.Details = v
		}
		if v, ok := item["color"].(string); ok {
			entry.Color = v
		}

		// Timestamp
		if v, ok := item["timestamp"].(float64); ok {
			entry.Timestamp = int64(v)
		}

		// Library
		if lib, ok := item["library"].(map[string]interface{}); ok {
			if text, ok := lib["text"].(string); ok {
				entry.Library = text
			}
			if key, ok := lib["key"].(string); ok {
				entry.LibraryKey = key
			}
		}

		// Cover
		if cover, ok := item["cover"].(map[string]interface{}); ok {
			if format, ok := cover["format"].(string); ok {
				entry.Format = format
			}
			if url, ok := cover["url"].(string); ok {
				entry.CoverURL = url
			}
		}

		entries = append(entries, entry)
	}

	return &models.TimelineResponse{
		Version:  raw.Version,
		Timeline: entries,
	}, nil
}

// ImportTimelineBytes imports timeline from raw JSON bytes
func (im *Importer) ImportTimelineBytes(data []byte) (int, error) {
	timeline, err := ParseTimeline(data)
	if err != nil {
		return 0, err
	}

	return im.store.ImportTimeline(timeline)
}
