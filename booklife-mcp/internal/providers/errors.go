// Package providers defines interfaces and errors for external book service providers.
package providers

import "fmt"

// BookNotFoundError indicates a book was not found in the specified source.
type BookNotFoundError struct {
	Query  string // The search query or identifier used
	Source string // The provider that was searched (e.g., "hardcover", "openlibrary")
}

func (e *BookNotFoundError) Error() string {
	return fmt.Sprintf("book not found in %s: %s", e.Source, e.Query)
}

// ProviderError wraps provider-specific errors with context.
type ProviderError struct {
	Provider string // The provider name (e.g., "hardcover", "libby")
	Op       string // The operation that failed (e.g., "SearchBooks", "GetLoans")
	Err      error  // The underlying error
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("%s.%s: %v", e.Provider, e.Op, e.Err)
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}

// AuthenticationError indicates authentication or authorization failure.
type AuthenticationError struct {
	Provider string // The provider that rejected authentication
	Reason   string // Human-readable reason for the failure
}

func (e *AuthenticationError) Error() string {
	return fmt.Sprintf("%s authentication failed: %s", e.Provider, e.Reason)
}

// RateLimitError indicates the provider's rate limit was exceeded.
type RateLimitError struct {
	Provider   string // The rate-limited provider
	RetryAfter int    // Seconds to wait before retrying (0 if unknown)
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("%s rate limited, retry after %ds", e.Provider, e.RetryAfter)
	}
	return fmt.Sprintf("%s rate limited", e.Provider)
}

// ConfigurationError indicates a provider is not properly configured.
type ConfigurationError struct {
	Provider string // The misconfigured provider
	Missing  string // What's missing or invalid
}

func (e *ConfigurationError) Error() string {
	return fmt.Sprintf("%s not configured: %s", e.Provider, e.Missing)
}

// NetworkError indicates a network-level failure communicating with a provider.
type NetworkError struct {
	Provider string // The provider being contacted
	Err      error  // The underlying network error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error communicating with %s: %v", e.Provider, e.Err)
}

func (e *NetworkError) Unwrap() error {
	return e.Err
}
