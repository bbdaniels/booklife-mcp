package hardcover

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hasura/go-graphql-client"
	"github.com/user/booklife-mcp/internal/models"
)

// Client is the Hardcover GraphQL API client
type Client struct {
	client   *graphql.Client
	endpoint string
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
	}, nil
}

// authTransport adds authorization header to requests
type authTransport struct {
	apiKey string
	base   http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.apiKey)
	req.Header.Set("Content-Type", "application/json")
	return t.base.RoundTrip(req)
}

// SearchBooks searches for books by query
func (c *Client) SearchBooks(ctx context.Context, query string, limit int) ([]models.Book, error) {
	// GraphQL query for book search
	var q struct {
		Search struct {
			Results []struct {
				Hit struct {
					ID              string   `graphql:"id"`
					Title           string   `graphql:"title"`
					Subtitle        string   `graphql:"subtitle"`
					ISBN10          string   `graphql:"isbn_10"`
					ISBN13          string   `graphql:"isbn_13"`
					Description     string   `graphql:"description"`
					PageCount       int      `graphql:"pages"`
					PublishedDate   string   `graphql:"release_date"`
					CoverURL        string   `graphql:"image"`
					Rating          float64  `graphql:"rating"`
					RatingsCount    int      `graphql:"ratings_count"`
					CachedContributors []struct {
						Author struct {
							Name string `graphql:"name"`
						} `graphql:"author"`
					} `graphql:"cached_contributors"`
				} `graphql:"hit"`
			} `graphql:"results"`
		} `graphql:"search(query: $query, query_type: \"Book\", per_page: $limit)"`
	}

	variables := map[string]interface{}{
		"query": query,
		"limit": limit,
	}

	err := c.client.Query(ctx, &q, variables)
	if err != nil {
		return nil, fmt.Errorf("search query failed: %w", err)
	}

	var books []models.Book
	for _, result := range q.Search.Results {
		hit := result.Hit
		book := models.Book{
			HardcoverID:     hit.ID,
			Title:           hit.Title,
			Subtitle:        hit.Subtitle,
			ISBN10:          hit.ISBN10,
			ISBN13:          hit.ISBN13,
			Description:     hit.Description,
			PageCount:       hit.PageCount,
			PublishedDate:   hit.PublishedDate,
			CoverURL:        hit.CoverURL,
			HardcoverRating: hit.Rating,
			HardcoverCount:  hit.RatingsCount,
		}

		// Extract authors
		for _, contrib := range hit.CachedContributors {
			book.Authors = append(book.Authors, models.Contributor{
				Name: contrib.Author.Name,
				Role: "author",
			})
		}

		books = append(books, book)
	}

	return books, nil
}

// GetBook retrieves a single book by ID
func (c *Client) GetBook(ctx context.Context, bookID string) (*models.Book, error) {
	var q struct {
		Book struct {
			ID              string  `graphql:"id"`
			Title           string  `graphql:"title"`
			Subtitle        string  `graphql:"subtitle"`
			ISBN10          string  `graphql:"isbn_10"`
			ISBN13          string  `graphql:"isbn_13"`
			Description     string  `graphql:"description"`
			PageCount       int     `graphql:"pages"`
			PublishedDate   string  `graphql:"release_date"`
			CoverURL        string  `graphql:"image"`
			Rating          float64 `graphql:"rating"`
			RatingsCount    int     `graphql:"ratings_count"`
			CachedContributors []struct {
				Author struct {
					Name string `graphql:"name"`
				} `graphql:"author"`
			} `graphql:"cached_contributors"`
		} `graphql:"book(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": bookID,
	}

	err := c.client.Query(ctx, &q, variables)
	if err != nil {
		return nil, fmt.Errorf("book query failed: %w", err)
	}

	book := &models.Book{
		HardcoverID:     q.Book.ID,
		Title:           q.Book.Title,
		Subtitle:        q.Book.Subtitle,
		ISBN10:          q.Book.ISBN10,
		ISBN13:          q.Book.ISBN13,
		Description:     q.Book.Description,
		PageCount:       q.Book.PageCount,
		PublishedDate:   q.Book.PublishedDate,
		CoverURL:        q.Book.CoverURL,
		HardcoverRating: q.Book.Rating,
		HardcoverCount:  q.Book.RatingsCount,
	}

	for _, contrib := range q.Book.CachedContributors {
		book.Authors = append(book.Authors, models.Contributor{
			Name: contrib.Author.Name,
			Role: "author",
		})
	}

	return book, nil
}

// GetUserBooks retrieves the user's books by status
func (c *Client) GetUserBooks(ctx context.Context, status string, limit int) ([]models.Book, error) {
	// Map status to Hardcover status_id
	statusID := getStatusID(status)

	var q struct {
		Me []struct {
			UserBooks []struct {
				ID         string  `graphql:"id"`
				Rating     float64 `graphql:"rating"`
				Progress   int     `graphql:"reading_progress_percent"`
				DateAdded  string  `graphql:"date_added"`
				DateFinished string `graphql:"finished_at"`
				Book       struct {
					ID            string  `graphql:"id"`
					Title         string  `graphql:"title"`
					ISBN13        string  `graphql:"isbn_13"`
					PageCount     int     `graphql:"pages"`
					CoverURL      string  `graphql:"image"`
					CachedContributors []struct {
						Author struct {
							Name string `graphql:"name"`
						} `graphql:"author"`
					} `graphql:"cached_contributors"`
				} `graphql:"book"`
			} `graphql:"user_books(where: {status_id: {_eq: $status}}, limit: $limit, order_by: {date_added: desc})"`
		} `graphql:"me"`
	}

	variables := map[string]interface{}{
		"status": statusID,
		"limit":  limit,
	}

	err := c.client.Query(ctx, &q, variables)
	if err != nil {
		return nil, fmt.Errorf("user books query failed: %w", err)
	}

	var books []models.Book
	if len(q.Me) > 0 {
		for _, ub := range q.Me[0].UserBooks {
			book := models.Book{
				ID:          ub.ID,
				HardcoverID: ub.Book.ID,
				Title:       ub.Book.Title,
				ISBN13:      ub.Book.ISBN13,
				PageCount:   ub.Book.PageCount,
				CoverURL:    ub.Book.CoverURL,
				UserStatus: &models.UserBookStatus{
					Status:   status,
					Rating:   ub.Rating,
					Progress: ub.Progress,
				},
			}

			for _, contrib := range ub.Book.CachedContributors {
				book.Authors = append(book.Authors, models.Contributor{
					Name: contrib.Author.Name,
					Role: "author",
				})
			}

			books = append(books, book)
		}
	}

	return books, nil
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
		books, err := c.SearchBooks(ctx, isbn, 1)
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
		books, err := c.SearchBooks(ctx, query, 1)
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
	books, err := c.GetUserBooks(ctx, "read", 500)
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
