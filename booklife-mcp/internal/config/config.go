package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sblinch/kdl-go"
)

// Config represents the complete BookLife configuration
type Config struct {
	Server    ServerConfig    `kdl:"server"`
	User      UserConfig      `kdl:"user"`
	Providers ProvidersConfig `kdl:"providers"`
	Cache     CacheConfig     `kdl:"cache"`
	Features  FeaturesConfig  `kdl:"features"`
}

type ServerConfig struct {
	Name      string `kdl:"name"`
	Version   string `kdl:"version"`
	Transport string `kdl:"transport"`
}

type UserConfig struct {
	Name        string            `kdl:"name"`
	Timezone    string            `kdl:"timezone"`
	Preferences PreferencesConfig `kdl:"preferences"`
}

type PreferencesConfig struct {
	Genres           []string `kdl:"genres"`
	AvoidGenres      []string `kdl:"avoid-genres"`
	PreferredFormats []string `kdl:"preferred-formats"`
	MaxTBRSize       int      `kdl:"max-tbr-size"`
}

type ProvidersConfig struct {
	Hardcover      HardcoverConfig      `kdl:"hardcover"`
	Libby          LibbyConfig          `kdl:"libby"`
	OpenLibrary    OpenLibraryConfig    `kdl:"open-library"`
	Wikidata       WikidataConfig       `kdl:"wikidata"`
	YouTube        YouTubeConfig        `kdl:"youtube"`
	TikTok         TikTokConfig         `kdl:"tiktok"`
	LocalBookstores LocalBookstoresConfig `kdl:"local-bookstores"`
}

type HardcoverConfig struct {
	Enabled  bool   `kdl:"enabled"`
	APIKey   string `kdl:"api-key"`
	Endpoint string `kdl:"endpoint"`
	Sync     struct {
		AutoImportLibby bool   `kdl:"auto-import-libby"`
		DefaultStatus   string `kdl:"default-status"`
	} `kdl:"sync"`
}

type LibbyConfig struct {
	Enabled       bool `kdl:"enabled"`
	SkipTLSVerify bool `kdl:"skip-tls-verify"`
	// Libraries are synced from Libby automatically via 'booklife libby-connect'
	Notifications struct {
		HoldAvailable bool `kdl:"hold-available"`
		DueSoonDays   int  `kdl:"due-soon-days"`
	} `kdl:"notifications"`
	// TimelineURL is the optional Libby timeline export URL for importing reading history
	// Format: https://share.libbyapp.com/data/{uuid}/libbytimeline-all-loans.json
	TimelineURL string `kdl:"timeline-url"`
}

type OpenLibraryConfig struct {
	Enabled        bool   `kdl:"enabled"`
	Endpoint       string `kdl:"endpoint"`
	CoversEndpoint string `kdl:"covers-endpoint"`
	RateLimitMS    int    `kdl:"rate-limit-ms"`
}

type WikidataConfig struct {
	Enabled       bool   `kdl:"enabled"`
	SPARQLEndpoint string `kdl:"sparql-endpoint"`
}

type YouTubeConfig struct {
	Enabled          bool     `kdl:"enabled"`
	APIKey           string   `kdl:"api-key"`
	BookTubeChannels []string `kdl:"booktube-channels"`
}

type TikTokConfig struct {
	Enabled    bool     `kdl:"enabled"`
	ScraperAPI string   `kdl:"scraper-api"`
	Hashtags   []string `kdl:"hashtags"`
}

type LocalBookstoresConfig struct {
	Enabled bool          `kdl:"enabled"`
	Stores  []StoreConfig `kdl:"store,multiple"`
}

type StoreConfig struct {
	ID            string   `kdl:",arg"`
	Name          string   `kdl:"name"`
	Website       string   `kdl:"website"`
	Location      string   `kdl:"location"`
	Phone         string   `kdl:"phone"`
	EventsURL     string   `kdl:"events-url"`
	SearchEnabled bool     `kdl:"search-enabled"`
	Specialties   []string `kdl:"specialties"`
}

type CacheConfig struct {
	Path           string           `kdl:"path"`
	BookMetadata   bool             `kdl:"book-metadata"`
	CoverImages    bool             `kdl:"cover-images"`
	ReadingHistory bool             `kdl:"reading-history"`
	Embeddings     EmbeddingsConfig `kdl:"embeddings"`
}

type EmbeddingsConfig struct {
	Enabled   bool   `kdl:"enabled"`
	Model     string `kdl:"model"`
	IndexPath string `kdl:"index-path"`
}

type FeaturesConfig struct {
	SemanticSearch bool `kdl:"semantic-search"`
	AutoHold       bool `kdl:"auto-hold"`
	MoodTracking   bool `kdl:"mood-tracking"`
}

// DefaultPath returns the platform-specific default configuration file path.
// Checks in order:
// 1. ./booklife.kdl (current directory, for development)
// 2. ~/.config/booklife/booklife.kdl (platform-specific config dir)
func DefaultPath() (string, error) {
	// Check current directory first (for development)
	localPath := "booklife.kdl"
	if _, err := os.Stat(localPath); err == nil {
		return localPath, nil
	}

	// Use platform-specific config directory (or XDG_CONFIG_HOME)
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig == "" {
		xdgConfig = filepath.Join(home, ".config")
	}
	return filepath.Join(xdgConfig, "booklife", "booklife.kdl"), nil
}

// Load reads and parses the KDL configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := kdl.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	// Resolve environment variables
	resolveEnvVars(&cfg)

	// Validate configuration
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}

// resolveEnvVars replaces env="VAR_NAME" patterns with actual environment values
func resolveEnvVars(cfg *Config) {
	cfg.Providers.Hardcover.APIKey = resolveEnv(cfg.Providers.Hardcover.APIKey)
	cfg.Providers.YouTube.APIKey = resolveEnv(cfg.Providers.YouTube.APIKey)
	cfg.Providers.TikTok.ScraperAPI = resolveEnv(cfg.Providers.TikTok.ScraperAPI)
}

func resolveEnv(value string) string {
	if strings.HasPrefix(value, "env=") {
		envVar := strings.TrimPrefix(value, "env=")
		envVar = strings.Trim(envVar, "\"")
		return os.Getenv(envVar)
	}
	return value
}

func validate(cfg *Config) error {
	if cfg.Server.Name == "" {
		return fmt.Errorf("server name is required")
	}

	if cfg.Providers.Hardcover.Enabled && cfg.Providers.Hardcover.APIKey == "" {
		return fmt.Errorf("Hardcover API key required when Hardcover is enabled")
	}

	// Note: Libby no longer requires clone code in config.
	// Identity is stored in ~/.config/booklife/libby-identity.json
	// via 'booklife libby-connect <code>'

	return nil
}
