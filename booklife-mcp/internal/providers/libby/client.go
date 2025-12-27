package libby

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/user/booklife-mcp/internal/dirs"
	"github.com/user/booklife-mcp/internal/models"
)

// Client is the Libby/OverDrive API client
// Based on reverse-engineered API from libby-calibre-plugin
type Client struct {
	identity       *Identity
	libraries      []Library
	skipTLSVerify  bool
}

// Identity represents the Libby device/user identity
type Identity struct {
	Clone     string `json:"clone"`
	ChipKey   string `json:"chip_key"`
	ChipCode  string `json:"chip_code"`
	DeviceID  string `json:"device_id"`
}

// Library represents a linked library card
type Library struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	WebsiteURL   string `json:"website_url"`
	CardID       string `json:"card_id"`
	Advantagekey string `json:"advantagekey"`
}

// apiCard represents the raw card structure from Libby API
type apiCard struct {
	CardID       string `json:"cardId"`
	CardName     string `json:"cardName"`
	AdvantageKey string `json:"advantageKey"`
	Library      struct {
		WebsiteID string `json:"websiteId"`
		Name      string `json:"name"`
	} `json:"library"`
}

// Endpoints
const (
	sentryReadURL = "https://sentry-read.svc.overdrive.com"
	vandalURL     = "https://vandal.svc.overdrive.com"
	thunderURL    = "https://thunder.api.overdrive.com"
)

// TLS ServerNames for each endpoint (cert is valid for *.odrsre.overdrive.com)
var tlsServerNames = map[string]string{
	sentryReadURL: "sentry-read.odrsre.overdrive.com",
	vandalURL:     "vandal.odrsre.overdrive.com",
	thunderURL:    "thunder.odrsre.overdrive.com",
}

// doRequest centralizes all HTTP requests with proper TLS, headers, and logging
func (c *Client) doRequest(method, endpoint string, body io.Reader) ([]byte, error) {
	// Determine TLS server name from endpoint
	var serverName string
	for url, name := range tlsServerNames {
		if strings.HasPrefix(endpoint, url) {
			serverName = name
			break
		}
	}
	if serverName == "" {
		serverName = "odrsre.overdrive.com" // fallback
	}

	// Create request
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		return nil, err
	}

	// Add headers
	setLibbyHeaders(req)
	if c.identity != nil {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.identity.ChipKey))
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Create HTTP client with proper TLS
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				ServerName:         serverName,
				InsecureSkipVerify: c.skipTLSVerify,
			},
		},
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Log response
	os.WriteFile("/tmp/libby-debug.log", respBody, 0644)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned HTTP %d: %s", resp.StatusCode, string(respBody[:min(len(respBody), 200)]))
	}

	return respBody, nil
}

// Browser-like user agent (matches libby-calibre-plugin)
const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.2 Safari/605.1.15"

// setLibbyHeaders sets browser-like headers for Libby API requests
func setLibbyHeaders(req *http.Request) {
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://libbyapp.com/")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Origin", "https://libbyapp.com")
}

// NewClient creates a new Libby client using a clone code
func NewClient(cloneCode string) (*Client, error) {
	if len(cloneCode) != 8 {
		return nil, fmt.Errorf("clone code must be 8 digits")
	}

	c := &Client{}

	// Clone the identity from an existing Libby app
	if err := c.cloneIdentity(cloneCode); err != nil {
		return nil, fmt.Errorf("failed to clone identity: %w", err)
	}

	// Sync library cards
	if err := c.syncLibraries(); err != nil {
		return nil, fmt.Errorf("failed to sync libraries: %w", err)
	}

	return c, nil
}

// cloneIdentity clones the identity from an existing Libby installation
func (c *Client) cloneIdentity(code string) error {
	// POST to sentry-read to exchange clone code for identity
	endpoint := fmt.Sprintf("%s/chip/clone/code", sentryReadURL)

	payload := map[string]string{
		"code": code,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	respBody, err := c.doRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}

	var result struct {
		Identity Identity `json:"identity"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("decoding identity: %w", err)
	}

	c.identity = &result.Identity
	return nil
}

// syncLibraries syncs the user's library cards
func (c *Client) syncLibraries() error {
	endpoint := fmt.Sprintf("%s/chip/sync", sentryReadURL)

	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}

	var result struct {
		Cards []apiCard `json:"cards"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("decoding libraries: %w", err)
	}

	// Convert API cards to our Library type
	c.libraries = make([]Library, 0, len(result.Cards))
	for _, card := range result.Cards {
		c.libraries = append(c.libraries, Library{
			ID:           card.Library.WebsiteID,
			Name:         card.Library.Name,
			CardID:       card.CardID,
			Advantagekey: card.AdvantageKey,
		})
	}
	return nil
}

// Search searches the library catalog with pagination support
func (c *Client) Search(ctx context.Context, query string, formats []string, available bool, offset, limit int) ([]models.Book, int, error) {
	if len(c.libraries) == 0 {
		return nil, 0, fmt.Errorf("no libraries linked")
	}

	lib := c.libraries[0] // Use first library for now

	// Thunder API uses the library's advantageKey (slug), not websiteId
	libraryKey := lib.Advantagekey
	if libraryKey == "" {
		libraryKey = lib.ID // Fall back to websiteId if no advantageKey
	}

	endpoint := fmt.Sprintf("%s/v2/libraries/%s/media", thunderURL, libraryKey)

	params := url.Values{}
	params.Set("query", query)
	params.Set("perPage", fmt.Sprintf("%d", limit))
	params.Set("page", fmt.Sprintf("%d", (offset/limit)+1))
	if available {
		params.Set("availability", "available")
	}

	respBody, err := c.doRequest("GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, 0, err
	}

	var result struct {
		Items []struct {
			ID            string `json:"id"`
			Title         string `json:"title"`
			Subtitle      string `json:"subtitle"`
			FirstCreator  string `json:"firstCreatorName"` // This is a string, not an object
			Formats []struct {
				ID           string `json:"id"`
				Name         string `json:"name"`
				IsAudiobook  bool   `json:"isAudiobook"`
				IsEbook      bool   `json:"isEbook"`
				Available    bool   `json:"isAvailable"`
				OwnedCopies  int    `json:"ownedCopies"`
				HoldsCount   int    `json:"holdsCount"`
			} `json:"formats"`
			Covers struct {
				Cover300Wide struct {
					Href string `json:"href"`
				} `json:"cover300Wide"`
			} `json:"covers"`
		} `json:"items"`
		TotalItems int `json:"totalItems"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, 0, fmt.Errorf("decoding search results: %w", err)
	}

	totalCount := result.TotalItems

	var books []models.Book
	for _, item := range result.Items {
		book := models.Book{
			OverdriveID: item.ID,
			Title:       item.Title,
			Subtitle:    item.Subtitle,
			Authors:     []models.Contributor{{Name: item.FirstCreator, Role: "author"}},
			CoverURL:    item.Covers.Cover300Wide.Href,
		}

		// Add availability info
		avail := &models.LibraryAvailability{
			LibraryName: lib.Name,
			MediaID:     item.ID,
		}

		for _, format := range item.Formats {
			if format.IsEbook {
				avail.Formats = append(avail.Formats, "ebook")
				avail.EbookAvailable = format.Available
				avail.EbookCopies = format.OwnedCopies
				avail.EbookWaitlistSize = format.HoldsCount
			}
			if format.IsAudiobook {
				avail.Formats = append(avail.Formats, "audiobook")
				avail.AudiobookAvailable = format.Available
				avail.AudiobookCopies = format.OwnedCopies
				avail.AudiobookWaitlistSize = format.HoldsCount
			}
		}

		book.LibraryAvailability = avail
		books = append(books, book)
	}

	return books, totalCount, nil
}

// CheckAvailability checks if a specific book is available
func (c *Client) CheckAvailability(ctx context.Context, isbn, title, author string) (*models.LibraryAvailability, error) {
	query := isbn
	if query == "" {
		query = title
		if author != "" {
			query += " " + author
		}
	}

	books, _, err := c.Search(ctx, query, nil, false, 0, 1)
	if err != nil {
		return nil, err
	}

	if len(books) == 0 {
		return nil, nil
	}

	return books[0].LibraryAvailability, nil
}

// GetLoans returns current loans
func (c *Client) GetLoans(ctx context.Context) ([]models.LibbyLoan, error) {
	endpoint := fmt.Sprintf("%s/chip", sentryReadURL)

	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Parse response to check for loans field
	var result struct {
		Loans []struct {
			CheckoutID      int64  `json:"checkoutId"`
			ID              string `json:"id"`
			Title           string `json:"title"`
			FirstCreator    string `json:"firstCreatorName"`
			Cover300WideURL string `json:"covers.cover300Wide.href"`
			TypeName        string `json:"type.name"`
			Expires         string `json:"expires"`
			CheckoutDate    string `json:"checkoutDate"`
		} `json:"loans"`
		Cards []Library `json:"cards"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decoding loans: %w", err)
	}

	var loans []models.LibbyLoan
	for _, loan := range result.Loans {
		// Parse dates
		var dueDate, checkoutTime time.Time
		if loan.Expires != "" {
			dueDate, _ = time.Parse(time.RFC3339, loan.Expires)
		}
		if loan.CheckoutDate != "" {
			checkoutTime, _ = time.Parse(time.RFC3339, loan.CheckoutDate)
		}

		l := models.LibbyLoan{
			ID:           loan.ID,
			MediaID:      loan.ID, // mediaId is same as id
			Title:        loan.Title,
			Author:       loan.FirstCreator,
			CoverURL:     loan.Cover300WideURL,
			Format:       loan.TypeName,
			CheckoutDate: checkoutTime,
			DueDate:      dueDate,
			Progress:     0, // Not in this response
		}
		loans = append(loans, l)
	}

	// Update libraries if present
	if len(result.Cards) > 0 {
		c.libraries = result.Cards
	}

	return loans, nil
}

// GetHistory returns past loans (returned books) with pagination support
//
// NOTE: Libby's API does not expose checkout history. The /chip/sync endpoint
// only returns current loans and holds. Reading history is only available within
// the Libby mobile app itself, not through the web API.
//
// This function returns an empty result for compatibility with the interface.
func (c *Client) GetHistory(ctx context.Context, offset, limit int) ([]models.LibbyHistoryItem, int, error) {
	// Libby's API doesn't provide checkout history - only active loans and holds
	// History is only tracked in the Libby mobile app, not exposed via the API
	return []models.LibbyHistoryItem{}, 0, nil
}

// GetTags returns user's book tags
// Note: Libby's API may not expose tags through the web API.
// This attempts to fetch from /chip/sync which includes tag data if available.
func (c *Client) GetTags(ctx context.Context) (map[string][]string, error) {
	// Tags are embedded in the sync response, not a separate endpoint
	endpoint := fmt.Sprintf("%s/chip/sync", sentryReadURL)

	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		// Return empty map on error - tags may not be supported
		return make(map[string][]string), nil
	}

	// Parse response to check for tags field
	var result struct {
		Tags []struct {
			Name     string   `json:"name"`
			MediaIDs []string `json:"titleIds"`
		} `json:"tags"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		// If tags field doesn't exist or is empty, return empty map
		return make(map[string][]string), nil
	}

	// Convert to map format
	tags := make(map[string][]string)
	for _, tag := range result.Tags {
		if tag.Name != "" {
			tags[tag.Name] = tag.MediaIDs
		}
	}

	return tags, nil
}

// AddTag adds a tag to a media item
// Note: Tag modification may not be supported in the Libby web API.
func (c *Client) AddTag(ctx context.Context, mediaID, tag string) error {
	// The Libby web API doesn't expose tag modification endpoints.
	// Tags are managed within the mobile app only.
	return fmt.Errorf("tag modification not supported via Libby web API")
}

// RemoveTag removes a tag from a media item
// Note: Tag modification may not be supported in the Libby web API.
func (c *Client) RemoveTag(ctx context.Context, mediaID, tag string) error {
	// The Libby web API doesn't expose tag modification endpoints.
	// Tags are managed within the mobile app only.
	return fmt.Errorf("tag modification not supported via Libby web API")
}

// GetHolds returns current holds
func (c *Client) GetHolds(ctx context.Context) ([]models.LibbyHold, error) {
	endpoint := fmt.Sprintf("%s/chip", sentryReadURL)

	respBody, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Parse response to check for holds field
	var result struct {
		Holds []struct {
			ID                  string `json:"id"`
			Title               string `json:"title"`
			FirstCreator        string `json:"firstCreatorName"`
			Cover300WideURL     string `json:"covers.cover300Wide.href"`
			TypeName            string `json:"type.name"`
			HoldListPosition    int    `json:"holdListPosition"`
			EstimatedWaitDays   int    `json:"estimatedWaitDays"`
			IsAvailable         bool   `json:"isAvailable"`
			AutoBorrow          bool   `json:"autoBorrow"`
			PlacedDate          string `json:"placedDate"`
		} `json:"holds"`
		Cards []Library `json:"cards"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decoding holds: %w", err)
	}

	var holds []models.LibbyHold
	for _, hold := range result.Holds {
		// Parse placed date
		var placedDate time.Time
		if hold.PlacedDate != "" {
			placedDate, _ = time.Parse(time.RFC3339, hold.PlacedDate)
		}

		h := models.LibbyHold{
			ID:                hold.ID,
			MediaID:           hold.ID, // mediaId is same as id
			Title:             hold.Title,
			Author:            hold.FirstCreator,
			CoverURL:          hold.Cover300WideURL,
			Format:            hold.TypeName,
			QueuePosition:     hold.HoldListPosition,
			EstimatedWaitDays: hold.EstimatedWaitDays,
			IsReady:           hold.IsAvailable,
			AutoBorrow:        hold.AutoBorrow,
			HoldPlacedDate:    placedDate,
		}
		holds = append(holds, h)
	}

	// Update libraries if present
	if len(result.Cards) > 0 {
		c.libraries = result.Cards
	}

	return holds, nil
}

// PlaceHold places a hold on a media item
func (c *Client) PlaceHold(ctx context.Context, mediaID, format string, autoBorrow bool) (string, error) {
	if len(c.libraries) == 0 {
		return "", fmt.Errorf("no libraries linked")
	}

	lib := c.libraries[0]

	endpoint := fmt.Sprintf("%s/chip/holds", sentryReadURL)

	payload := map[string]interface{}{
		"mediaId":    mediaID,
		"cardId":     lib.CardID,
		"autoBorrow": autoBorrow,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	respBody, err := c.doRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("decoding hold response: %w", err)
	}

	return result.ID, nil
}

// Identity persistence functions

const identityFile = "libby-identity.json"

// identityPath returns the full path to the identity file in the platform-specific config directory.
func identityPath() (string, error) {
	configDir, err := dirs.ConfigDir()
	if err != nil {
		return "", fmt.Errorf("getting config directory: %w", err)
	}
	return filepath.Join(configDir, identityFile), nil
}

// Connect exchanges a clone code for a Libby identity and returns it along with linked libraries.
// This is the fast path for the CLI - it does the minimum work needed to authenticate.
func Connect(code string) (*Identity, []Library, error) {
	return ConnectWithOptions(code, false)
}

// ConnectWithOptions exchanges a clone code with optional TLS verification skip.
// Use skipTLSVerify=true if OverDrive's certificate is misconfigured (temporary workaround).
func ConnectWithOptions(code string, skipTLSVerify bool) (*Identity, []Library, error) {
	if len(code) != 8 {
		return nil, nil, fmt.Errorf("clone code must be 8 digits")
	}

	// Create cookie jar for session persistence (like libby-calibre-plugin)
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, nil, fmt.Errorf("creating cookie jar: %w", err)
	}

	httpClient := &http.Client{Jar: jar}
	if skipTLSVerify {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	// Step 1: Get initial chip identity token
	chipEndpoint := fmt.Sprintf("%s/chip?client=dewey", sentryReadURL)

	req, err := http.NewRequest("POST", chipEndpoint, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("creating chip request: %w", err)
	}
	setLibbyHeaders(req)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("getting chip: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("failed to get chip (HTTP %d)", resp.StatusCode)
	}

	var chipResult struct {
		Identity string `json:"identity"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&chipResult); err != nil {
		return nil, nil, fmt.Errorf("decoding chip: %w", err)
	}
	initialToken := chipResult.Identity

	// Step 2: Exchange clone code for full identity (form-encoded, not JSON)
	cloneEndpoint := fmt.Sprintf("%s/chip/clone/code", sentryReadURL)

	formData := url.Values{}
	formData.Set("code", code)

	req, err = http.NewRequest("POST", cloneEndpoint, bytes.NewReader([]byte(formData.Encode())))
	if err != nil {
		return nil, nil, fmt.Errorf("creating clone request: %w", err)
	}

	setLibbyHeaders(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", initialToken))

	resp, err = httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("connecting to Libby: %w", err)
	}
	defer resp.Body.Close()

	// Read full response - we need it for both cloneResult and cloneChip parsing
	cloneBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("reading clone response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("clone code rejected (HTTP %d) - code may have expired", resp.StatusCode)
	}

	var cloneResult struct {
		Result   string `json:"result"`
		Identity string `json:"identity"`
		Chip     string `json:"chip"`
	}
	if err := json.Unmarshal(cloneBody, &cloneResult); err != nil {
		return nil, nil, fmt.Errorf("decoding clone result: %w", err)
	}

	// Step 3: Sync to get library cards
	syncEndpoint := fmt.Sprintf("%s/chip/sync", sentryReadURL)

	req, err = http.NewRequest("GET", syncEndpoint, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("creating sync request: %w", err)
	}

	setLibbyHeaders(req)
	// Use the cloned identity token, fall back to initial if not provided
	syncToken := cloneResult.Identity
	if syncToken == "" {
		syncToken = initialToken
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", syncToken))

	resp, err = httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("syncing libraries: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("failed to sync libraries (HTTP %d)", resp.StatusCode)
	}

	var syncResult struct {
		Result string    `json:"result"`
		Cards  []apiCard `json:"cards"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&syncResult); err != nil {
		return nil, nil, fmt.Errorf("decoding libraries: %w", err)
	}

	// Convert API cards to our Library type
	libraries := make([]Library, 0, len(syncResult.Cards))
	for _, card := range syncResult.Cards {
		libraries = append(libraries, Library{
			ID:           card.Library.WebsiteID,
			Name:         card.Library.Name,
			CardID:       card.CardID,
			Advantagekey: card.AdvantageKey,
		})
	}

	// Build identity from the data we collected
	identity := &Identity{
		ChipKey:  syncToken,
		DeviceID: cloneResult.Chip,
	}

	return identity, libraries, nil
}

// SaveIdentity saves the Libby identity to disk for future use
func SaveIdentity(identity *Identity) error {
	path, err := identityPath()
	if err != nil {
		return err
	}

	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	data, err := json.MarshalIndent(identity, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding identity: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing identity file: %w", err)
	}

	return nil
}

// LoadIdentity loads a previously saved Libby identity from disk
func LoadIdentity() (*Identity, error) {
	path, err := identityPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no saved identity found - run 'booklife libby-connect' first")
		}
		return nil, fmt.Errorf("reading identity file: %w", err)
	}

	var identity Identity
	if err := json.Unmarshal(data, &identity); err != nil {
		return nil, fmt.Errorf("decoding identity: %w", err)
	}

	return &identity, nil
}

// NewClientFromSavedIdentity creates a new Libby client using a previously saved identity
func NewClientFromSavedIdentity() (*Client, error) {
	return NewClientFromSavedIdentityWithOptions(false)
}

// NewClientFromSavedIdentityWithOptions creates a new Libby client with TLS options
func NewClientFromSavedIdentityWithOptions(skipTLSVerify bool) (*Client, error) {
	identity, err := LoadIdentity()
	if err != nil {
		return nil, err
	}

	c := &Client{
		identity:      identity,
		skipTLSVerify: skipTLSVerify,
	}

	// Sync library cards
	if err := c.syncLibraries(); err != nil {
		return nil, fmt.Errorf("syncing libraries: %w", err)
	}

	return c, nil
}

// HasSavedIdentity checks if a saved identity exists
func HasSavedIdentity() bool {
	path, err := identityPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}
