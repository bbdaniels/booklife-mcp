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

	"github.com/user/booklife-mcp/internal/dirs"
	"github.com/user/booklife-mcp/internal/models"
)

// Client is the Libby/OverDrive API client
// Based on reverse-engineered API from libby-calibre-plugin
type Client struct {
	httpClient *http.Client
	identity   *Identity
	libraries  []Library
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

// Endpoints
const (
	sentryReadURL = "https://sentry-read.svc.overdrive.com"
	vandalURL     = "https://vandal.svc.overdrive.com"
	thunderURL    = "https://thunder.api.overdrive.com"
)

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

	c := &Client{
		httpClient: &http.Client{},
	}

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
	
	// This is a simplified version - real implementation needs proper
	// device fingerprinting and headers
	resp, err := c.post(endpoint, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Identity Identity `json:"identity"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decoding identity: %w", err)
	}

	c.identity = &result.Identity
	return nil
}

// syncLibraries syncs the user's library cards
func (c *Client) syncLibraries() error {
	endpoint := fmt.Sprintf("%s/chip", sentryReadURL)
	
	resp, err := c.get(endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Cards []Library `json:"cards"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decoding libraries: %w", err)
	}

	c.libraries = result.Cards
	return nil
}

// Search searches the library catalog
func (c *Client) Search(ctx context.Context, query string, formats []string, available bool) ([]models.Book, error) {
	if len(c.libraries) == 0 {
		return nil, fmt.Errorf("no libraries linked")
	}

	lib := c.libraries[0] // Use first library for now
	
	endpoint := fmt.Sprintf("%s/v2/libraries/%s/media", thunderURL, lib.ID)
	
	params := url.Values{}
	params.Set("query", query)
	params.Set("perPage", "20")
	if available {
		params.Set("availability", "available")
	}

	resp, err := c.getWithParams(endpoint, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Items []struct {
			ID          string `json:"id"`
			Title       string `json:"title"`
			Subtitle    string `json:"subtitle"`
			FirstCreator struct {
				Name string `json:"name"`
			} `json:"firstCreatorName"`
			Formats []struct {
				ID           string `json:"id"`
				Name         string `json:"name"`
				IsAudiobook  bool   `json:"isAudiobook"`
				IsEbook      bool   `json:"isEbook"`
				Available    bool   `json:"isAvailable"`
				OwnedCopies  int    `json:"ownedCopies"`
				HoldsCount   int    `json:"holdsCount"`
			} `json:"formats"`
			Cover struct {
				URL string `json:"href"`
			} `json:"covers"`
		} `json:"items"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding search results: %w", err)
	}

	var books []models.Book
	for _, item := range result.Items {
		book := models.Book{
			OverdriveID: item.ID,
			Title:       item.Title,
			Subtitle:    item.Subtitle,
			Authors:     []models.Contributor{{Name: item.FirstCreator.Name, Role: "author"}},
			CoverURL:    item.Cover.URL,
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

	return books, nil
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

	books, err := c.Search(ctx, query, nil, false)
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
	endpoint := fmt.Sprintf("%s/chip/loans", sentryReadURL)
	
	resp, err := c.get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Loans []struct {
			ID        string `json:"id"`
			MediaID   string `json:"mediaId"`
			Title     string `json:"title"`
			FirstCreator string `json:"firstCreatorName"`
			Cover     struct {
				URL string `json:"href"`
			} `json:"covers"`
			Type       string `json:"type"`
			ExpireDate string `json:"expireDate"`
			Progress   float64 `json:"readingProgress"`
		} `json:"loans"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding loans: %w", err)
	}

	var loans []models.LibbyLoan
	for _, loan := range result.Loans {
		l := models.LibbyLoan{
			ID:       loan.ID,
			MediaID:  loan.MediaID,
			Title:    loan.Title,
			Author:   loan.FirstCreator,
			CoverURL: loan.Cover.URL,
			Format:   loan.Type,
			Progress: loan.Progress,
		}
		// Parse dates...
		loans = append(loans, l)
	}

	return loans, nil
}

// GetHolds returns current holds
func (c *Client) GetHolds(ctx context.Context) ([]models.LibbyHold, error) {
	endpoint := fmt.Sprintf("%s/chip/holds", sentryReadURL)
	
	resp, err := c.get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Holds []struct {
			ID           string `json:"id"`
			MediaID      string `json:"mediaId"`
			Title        string `json:"title"`
			FirstCreator string `json:"firstCreatorName"`
			Cover        struct {
				URL string `json:"href"`
			} `json:"covers"`
			Type           string `json:"type"`
			HoldListPosition int `json:"holdListPosition"`
			EstimatedWait    int `json:"estimatedWaitDays"`
			IsAvailable      bool `json:"isAvailable"`
			AutoBorrow       bool `json:"autoBorrow"`
		} `json:"holds"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding holds: %w", err)
	}

	var holds []models.LibbyHold
	for _, hold := range result.Holds {
		h := models.LibbyHold{
			ID:                hold.ID,
			MediaID:           hold.MediaID,
			Title:             hold.Title,
			Author:            hold.FirstCreator,
			CoverURL:          hold.Cover.URL,
			Format:            hold.Type,
			QueuePosition:     hold.HoldListPosition,
			EstimatedWaitDays: hold.EstimatedWait,
			IsReady:           hold.IsAvailable,
			AutoBorrow:        hold.AutoBorrow,
		}
		holds = append(holds, h)
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

	resp, err := c.post(endpoint, payload)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding hold response: %w", err)
	}

	return result.ID, nil
}

// HTTP helper methods

func (c *Client) get(endpoint string) (*http.Response, error) {
	return c.getWithParams(endpoint, nil)
}

func (c *Client) getWithParams(endpoint string, params url.Values) (*http.Response, error) {
	if params != nil {
		endpoint = endpoint + "?" + params.Encode()
	}
	
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	
	c.addHeaders(req)
	return c.httpClient.Do(req)
}

func (c *Client) post(endpoint string, payload interface{}) (*http.Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	
	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return nil, err
	}
	
	req.Body = nil // Would set body here
	_ = body // Use body
	
	c.addHeaders(req)
	req.Header.Set("Content-Type", "application/json")
	
	return c.httpClient.Do(req)
}

func (c *Client) addHeaders(req *http.Request) {
	setLibbyHeaders(req)
	if c.identity != nil {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.identity.ChipKey))
	}
}

// Identity persistence functions

const identityFile = "libby-identity.json"

// identityPath returns the full path to the identity file in the platform-specific config directory.
// Falls back to legacy ~/.booklife location if migration hasn't occurred.
func identityPath() (string, error) {
	// Check for legacy identity first for backward compatibility
	legacyDir, err := dirs.LegacyDir()
	if err == nil {
		legacyPath := filepath.Join(legacyDir, identityFile)
		if _, err := os.Stat(legacyPath); err == nil {
			return legacyPath, nil
		}
	}

	// Use platform-specific config directory
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

	fmt.Fprintf(os.Stderr, "   [DEBUG] Step 1: POST %s\n", chipEndpoint)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("getting chip: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("failed to get chip (HTTP %d)", resp.StatusCode)
	}

	// Read full response body for debugging
	chipBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("reading chip response: %w", err)
	}
	fmt.Fprintf(os.Stderr, "   [DEBUG] Chip response: %s\n", string(chipBody))

	var chipResult struct {
		Identity string `json:"identity"`
	}
	if err := json.Unmarshal(chipBody, &chipResult); err != nil {
		return nil, nil, fmt.Errorf("decoding chip: %w", err)
	}

	initialToken := chipResult.Identity
	fmt.Fprintf(os.Stderr, "   [DEBUG] Got initial token: %s...\n", initialToken[:20])

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

	fmt.Fprintf(os.Stderr, "   [DEBUG] Step 2: POST %s (code=%s)\n", cloneEndpoint, code)
	resp, err = httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("connecting to Libby: %w", err)
	}
	defer resp.Body.Close()

	// Read full response for debugging
	cloneBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("reading clone response: %w", err)
	}
	fmt.Fprintf(os.Stderr, "   [DEBUG] Clone response (HTTP %d): %s\n", resp.StatusCode, string(cloneBody))

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("clone code rejected (HTTP %d) - code may have expired", resp.StatusCode)
	}

	var cloneResult struct {
		Result   string `json:"result"`
		Identity string `json:"identity"`
	}
	if err := json.Unmarshal(cloneBody, &cloneResult); err != nil {
		return nil, nil, fmt.Errorf("decoding clone result: %w", err)
	}
	fmt.Fprintf(os.Stderr, "   [DEBUG] Clone result: %s, identity: %s...\n", cloneResult.Result, cloneResult.Identity[:min(20, len(cloneResult.Identity))])

	// Step 3: Sync to get library cards
	// According to libby-calibre-plugin, we should use the cloned identity token
	syncEndpoint := fmt.Sprintf("%s/chip/sync", sentryReadURL)

	req, err = http.NewRequest("GET", syncEndpoint, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("creating sync request: %w", err)
	}

	setLibbyHeaders(req)
	// Try using the cloned identity - this is what clone_by_code returns
	syncToken := cloneResult.Identity
	if syncToken == "" {
		// Fall back to initial token if clone didn't return one
		syncToken = initialToken
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", syncToken))

	fmt.Fprintf(os.Stderr, "   [DEBUG] Step 3: GET %s (token: %s...)\n", syncEndpoint, syncToken[:min(20, len(syncToken))])

	// Print cookies being sent
	parsedURL, _ := url.Parse(syncEndpoint)
	cookies := jar.Cookies(parsedURL)
	fmt.Fprintf(os.Stderr, "   [DEBUG] Cookies for sync: %d cookies\n", len(cookies))
	for _, c := range cookies {
		fmt.Fprintf(os.Stderr, "   [DEBUG]   - %s=%s...\n", c.Name, c.Value[:min(10, len(c.Value))])
	}

	resp, err = httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("syncing libraries: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for debugging
	syncBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("reading sync response: %w", err)
	}
	fmt.Fprintf(os.Stderr, "   [DEBUG] Sync response (HTTP %d): %s\n", resp.StatusCode, string(syncBody)[:min(500, len(syncBody))])

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("failed to sync libraries (HTTP %d)", resp.StatusCode)
	}

	var syncResult struct {
		Identity string    `json:"identity"`
		Chip     string    `json:"chip"`
		Cards    []Library `json:"cards"`
	}
	if err := json.Unmarshal(syncBody, &syncResult); err != nil {
		return nil, nil, fmt.Errorf("decoding libraries: %w", err)
	}

	identity := &Identity{
		ChipKey: syncResult.Identity,
		Clone:   syncResult.Chip,
	}

	return identity, syncResult.Cards, nil
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
	identity, err := LoadIdentity()
	if err != nil {
		return nil, err
	}

	c := &Client{
		httpClient: &http.Client{},
		identity:   identity,
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
