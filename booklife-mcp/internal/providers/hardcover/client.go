package hardcover

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/hasura/go-graphql-client"
	"github.com/user/booklife-mcp/internal/models"
)

// Client is the Hardcover GraphQL API client
type Client struct {
	client   *graphql.Client
	endpoint string
	apiKey   string
}

// NewClient creates a new Hardcover API client
func NewClient(endpoint, apiKey string) (*Client, error) {
	if endpoint == "" {
		endpoint = "https://api.hardcover.app/v1/graphql"
	}

	httpClient := &http.Client{
		Transport: &authTransport{
			apiKey: apiKey,
			base:   http.DefaultTransport,
		},
	}

	client := graphql.NewClient(endpoint, httpClient)

	return &Client{
		client:   client,
		endpoint: endpoint,
		apiKey:   apiKey,
	}, nil
}

// authTransport adds authorization header to requests and logs responses
type authTransport struct {
	apiKey string
	base   http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.base.RoundTrip(req)
	if err == nil && resp != nil && resp.ContentLength > 0 && resp.ContentLength < 1<<20 {
		// Only capture response body for debugging if content length is known and reasonable
		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr == nil && len(body) > 0 {
			os.WriteFile("/tmp/hardcover-debug.log", body, 0644)
		}
		resp.Body = io.NopCloser(bytes.NewReader(body))
	}
	return resp, err
}

// SearchBooks searches for books by query with pagination support
func (c *Client) SearchBooks(ctx context.Context, query string, offset, limit int) ([]models.Book, int, error) {
	// Make a direct HTTP POST request to get the raw JSONB results
	queryStr := fmt.Sprintf(`{
		search(query: %q, query_type: "Book", per_page: %d) {
			ids
			results
		}
	}`, query, offset+limit) // Request enough to get offset+limit items

	reqBody := map[string]string{
		"query": queryStr,
	}

	reqBodyJSON, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", c.endpoint, bytes.NewReader(reqBodyJSON))
	if err != nil {
		return nil, 0, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, 0, fmt.Errorf("API returned HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			Search struct {
				IDS     []string          `json:"ids"`
				Results json.RawMessage   `json:"results"`
			} `json:"search"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, 0, fmt.Errorf("parsing response: %w", err)
	}

	if len(result.Errors) > 0 {
		return nil, 0, fmt.Errorf("API errors: %v", result.Errors)
	}

	totalCount := len(result.Data.Search.IDS)

	// Parse the JSONB results
	var searchResults struct {
		Hits []struct {
			Document struct {
				ID           string   `json:"id"`
				Title        string   `json:"title"`
				Subtitle     string   `json:"subtitle"`
				Description  string   `json:"description"`
				Pages        int      `json:"pages"`
				ReleaseDate  string   `json:"release_date"`
				Rating       float64  `json:"rating"`
				RatingsCount int      `json:"ratings_count"`
				AuthorNames  []string `json:"author_names"`
				ISBNs        []string `json:"isbns"`
				Image        struct {
					URL string `json:"url"`
				} `json:"image"`
			} `json:"document"`
		} `json:"hits"`
	}

	if err := json.Unmarshal(result.Data.Search.Results, &searchResults); err != nil {
		return nil, totalCount, fmt.Errorf("parsing search results: %w", err)
	}

	// Apply offset to get the correct page
	startIdx := offset
	if startIdx >= len(searchResults.Hits) {
		return []models.Book{}, totalCount, nil
	}
	if startIdx+limit > len(searchResults.Hits) {
		limit = len(searchResults.Hits) - startIdx
	}

	var books []models.Book
	for _, hit := range searchResults.Hits[startIdx : startIdx+limit] {
		doc := hit.Document
		book := models.Book{
			HardcoverID:     doc.ID,
			Title:           doc.Title,
			Subtitle:        doc.Subtitle,
			Description:     doc.Description,
			PageCount:       doc.Pages,
			PublishedDate:   doc.ReleaseDate,
			CoverURL:        doc.Image.URL,
			HardcoverRating: doc.Rating,
			HardcoverCount:  doc.RatingsCount,
		}

		// Extract ISBNs
		for _, isbn := range doc.ISBNs {
			if len(isbn) == 10 {
				book.ISBN10 = isbn
			} else if len(isbn) == 13 {
				book.ISBN13 = isbn
			}
		}

		// Extract authors
		for _, authorName := range doc.AuthorNames {
			book.Authors = append(book.Authors, models.Contributor{
				Name: authorName,
				Role: "author",
			})
		}

		books = append(books, book)
	}

	return books, totalCount, nil
}

// getBookByID retrieves a single book by string ID
func (c *Client) getBookByID(ctx context.Context, idStr string) (*models.Book, error) {
	var q struct {
		Books []struct {
			ID              int     `graphql:"id"`
			Title           string  `graphql:"title"`
			Subtitle        string  `graphql:"subtitle"`
			Description     string  `graphql:"description"`
			PageCount       int     `graphql:"pages"`
			PublishedDate   string  `graphql:"release_date"`
			Rating          float64 `graphql:"rating"`
			RatingsCount    int     `graphql:"ratings_count"`
			// ISBNs might not be available in the books table
			// We'll get them from the search results or look them up separately
		} `graphql:"books(where: {id: {_eq: $id}}, limit: 1)"`
	}

	// Convert string ID to int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid book ID: %w", err)
	}

	variables := map[string]interface{}{
		"id": id,
	}

	err = c.client.Query(ctx, &q, variables)
	if err != nil {
		return nil, fmt.Errorf("book query failed: %w", err)
	}

	if len(q.Books) == 0 {
		return nil, fmt.Errorf("book not found")
	}

	b := q.Books[0]
	book := &models.Book{
		HardcoverID:     idStr, // Keep string ID for consistency
		Title:           b.Title,
		Subtitle:        b.Subtitle,
		Description:     b.Description,
		PageCount:       b.PageCount,
		PublishedDate:   b.PublishedDate,
		HardcoverRating: b.Rating,
		HardcoverCount:  b.RatingsCount,
	}

	return book, nil
}

// GetBook retrieves a single book by ID (public API)
func (c *Client) GetBook(ctx context.Context, bookID string) (*models.Book, error) {
	return c.getBookByID(ctx, bookID)
}

// GetUserBooks retrieves the user's books by status with pagination support
func (c *Client) GetUserBooks(ctx context.Context, status string, offset, limit int) ([]models.Book, int, error) {
	// Map status to Hardcover status_id
	statusID := getStatusID(status)

	// First, get total count
	var countQ struct {
		Me []struct {
			UserBooksAggregate struct {
				Aggregate struct {
					Count int `graphql:"count"`
				} `graphql:"aggregate"`
			} `graphql:"user_books_aggregate(where: {status_id: {_eq: $status}})"`
		} `graphql:"me"`
	}

	countVars := map[string]interface{}{
		"status": statusID,
	}

	err := c.client.Query(ctx, &countQ, countVars)
	if err != nil {
		return nil, 0, fmt.Errorf("user books count query failed: %w", err)
	}

	totalCount := 0
	if len(countQ.Me) > 0 {
		totalCount = countQ.Me[0].UserBooksAggregate.Aggregate.Count
	}

	// Then get the paginated results
	var q struct {
		Me []struct {
			UserBooks []struct {
				ID        int     `graphql:"id"`
				Rating    float64 `graphql:"rating"`
				DateAdded string  `graphql:"date_added"`
				Book      struct {
					ID        int    `graphql:"id"`
					Title     string `graphql:"title"`
					PageCount int    `graphql:"pages"`
				} `graphql:"book"`
			} `graphql:"user_books(where: {status_id: {_eq: $status}}, limit: $limit, offset: $offset, order_by: {date_added: desc})"`
		} `graphql:"me"`
	}

	variables := map[string]interface{}{
		"status": statusID,
		"limit":  limit,
		"offset": offset,
	}

	err = c.client.Query(ctx, &q, variables)
	if err != nil {
		return nil, totalCount, fmt.Errorf("user books query failed: %w", err)
	}

	var books []models.Book
	if len(q.Me) > 0 {
		for _, ub := range q.Me[0].UserBooks {
			book := models.Book{
				ID:          fmt.Sprintf("%d", ub.ID),
				HardcoverID: fmt.Sprintf("%d", ub.Book.ID),
				Title:       ub.Book.Title,
				PageCount:   ub.Book.PageCount,
				UserStatus: &models.UserBookStatus{
					Status:   status,
					Rating:   ub.Rating,
					Progress: 0, // Progress field no longer available in API schema
				},
			}

			books = append(books, book)
		}
	}

	return books, totalCount, nil
}

// UpdateBookStatus updates a book's status in the user's library
func (c *Client) UpdateBookStatus(ctx context.Context, bookID, status string, progress int, rating float64) error {
	statusID := getStatusID(status)

	var m struct {
		UpdateUserBook struct {
			ID string `graphql:"id"`
		} `graphql:"update_user_books(where: {book_id: {_eq: $bookID}}, _set: {status_id: $status, reading_progress_percent: $progress, rating: $rating})"`
	}

	variables := map[string]interface{}{
		"bookID":   bookID,
		"status":   statusID,
		"progress": progress,
		"rating":   rating,
	}

	err := c.client.Mutate(ctx, &m, variables)
	if err != nil {
		return fmt.Errorf("update status mutation failed: %w", err)
	}

	return nil
}

// AddBook adds a book to the user's library
func (c *Client) AddBook(ctx context.Context, isbn, title, author, status string) (string, error) {
	// First, find the book
	var bookID string
	if isbn != "" {
		books, _, err := c.SearchBooks(ctx, isbn, 0, 1)
		if err != nil {
			return "", err
		}
		if len(books) > 0 {
			bookID = books[0].HardcoverID
		}
	}

	if bookID == "" && title != "" {
		query := title
		if author != "" {
			query += " " + author
		}
		books, _, err := c.SearchBooks(ctx, query, 0, 1)
		if err != nil {
			return "", err
		}
		if len(books) > 0 {
			bookID = books[0].HardcoverID
		}
	}

	if bookID == "" {
		return "", fmt.Errorf("book not found")
	}

	// Add to library
	statusID := getStatusID(status)

	var m struct {
		InsertUserBook struct {
			ID string `graphql:"id"`
		} `graphql:"insert_user_books_one(object: {book_id: $bookID, status_id: $status})"`
	}

	variables := map[string]interface{}{
		"bookID": bookID,
		"status": statusID,
	}

	err := c.client.Mutate(ctx, &m, variables)
	if err != nil {
		return "", fmt.Errorf("add book mutation failed: %w", err)
	}

	return m.InsertUserBook.ID, nil
}

// GetReadingStats retrieves reading statistics for a year
func (c *Client) GetReadingStats(ctx context.Context, year int) (*models.ReadingStats, error) {
	// Simplified - would need proper date filtering
	books, _, err := c.GetUserBooks(ctx, "read", 0, 500)
	if err != nil {
		return nil, err
	}

	stats := &models.ReadingStats{
		Year:           year,
		BooksRead:      len(books),
		GenreBreakdown: make(map[string]int),
	}

	var totalRating float64
	var ratingCount int
	for _, book := range books {
		stats.PagesRead += book.PageCount
		
		if book.UserStatus != nil && book.UserStatus.Rating > 0 {
			totalRating += book.UserStatus.Rating
			ratingCount++
		}

		for _, genre := range book.Genres {
			stats.GenreBreakdown[genre]++
		}
	}

	if ratingCount > 0 {
		stats.AverageRating = totalRating / float64(ratingCount)
	}

	return stats, nil
}

// Helper to convert status string to Hardcover status_id
func getStatusID(status string) int {
	switch status {
	case "want-to-read":
		return 1
	case "reading":
		return 2
	case "read":
		return 3
	case "dnf":
		return 5
	default:
		return 0 // all
	}
}
