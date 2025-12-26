// Package dirs provides platform-specific directory paths for BookLife configuration and data.
package dirs

import (
	"os"
	"path/filepath"
	"runtime"
)

// ConfigDir returns the platform-specific configuration directory.
// - Linux/Unix: ~/.config/booklife
// - macOS: ~/.config/booklife
// - Windows: %APPDATA%\BookLife
func ConfigDir() (string, error) {
	if runtime.GOOS == "windows" {
		// Windows: %APPDATA%\BookLife
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "BookLife"), nil
	}

	// Linux/macOS: follow XDG Base Directory Specification
	// Prefer XDG_CONFIG_HOME if set, otherwise use ~/.config/booklife
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig != "" {
		return filepath.Join(xdgConfig, "booklife"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "booklife"), nil
}

// DataDir returns the platform-specific data directory.
// - Linux/Unix: ~/.local/share/booklife (XDG_DATA_HOME)
// - macOS: ~/.local/share/booklife
// - Windows: %LOCALAPPDATA%\BookLife
func DataDir() (string, error) {
	if runtime.GOOS == "windows" {
		// Windows: %LOCALAPPDATA%\BookLife
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, "BookLife"), nil
	}

	// Linux/macOS: follow XDG Base Directory Specification
	// Prefer XDG_DATA_HOME if set, otherwise use ~/.local/share/booklife
	xdgData := os.Getenv("XDG_DATA_HOME")
	if xdgData != "" {
		return filepath.Join(xdgData, "booklife"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "booklife"), nil
}

// StateDir returns the platform-specific state directory for transient data.
// - Linux: ~/.local/state/booklife (XDG_STATE_HOME)
// - macOS: ~/.local/state/booklife
// - Windows: %LOCALAPPDATA%\BookLife\state
func StateDir() (string, error) {
	if runtime.GOOS == "windows" {
		dataDir, err := DataDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(dataDir, "state"), nil
	}

	// Linux/macOS: follow XDG Base Directory Specification
	// Prefer XDG_STATE_HOME if set, otherwise use ~/.local/state/booklife
	xdgState := os.Getenv("XDG_STATE_HOME")
	if xdgState != "" {
		return filepath.Join(xdgState, "booklife"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "state", "booklife"), nil
}

// CacheDir returns the platform-specific cache directory.
// - Linux: ~/.cache/booklife (XDG_CACHE_HOME)
// - macOS: ~/Library/Caches/booklife
// - Windows: %LOCALAPPDATA%\BookLife\cache
func CacheDir() (string, error) {
	if runtime.GOOS == "darwin" {
		// macOS: ~/Library/Caches/booklife
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Caches", "booklife"), nil
	}

	if runtime.GOOS == "windows" {
		// Windows: %LOCALAPPDATA%\BookLife\cache
		dataDir, err := DataDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(dataDir, "cache"), nil
	}

	// Linux: follow XDG Base Directory Specification
	// Prefer XDG_CACHE_HOME if set, otherwise use ~/.cache/booklife
	xdgCache := os.Getenv("XDG_CACHE_HOME")
	if xdgCache != "" {
		return filepath.Join(xdgCache, "booklife"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cache", "booklife"), nil
}

// LegacyDir returns the legacy ~/.booklife directory for backward compatibility.
// This is deprecated but checked for migrating existing data.
func LegacyDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".booklife"), nil
}

// EnsureAll creates all platform-specific directories with appropriate permissions.
func EnsureAll() error {
	dirs := []func() (string, error){
		ConfigDir,
		DataDir,
		StateDir,
		CacheDir,
	}

	for _, getDir := range dirs {
		dir, err := getDir()
		if err != nil {
			return err
		}
		// Create with owner-only permissions (0700)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}
	return nil
}

// MigrateFromLegacy migrates data from ~/.booklife to platform-specific directories.
// Returns true if migration occurred.
func MigrateFromLegacy() (bool, error) {
	legacyDir, err := LegacyDir()
	if err != nil {
		return false, err
	}

	// Check if legacy directory exists
	if _, err := os.Stat(legacyDir); os.IsNotExist(err) {
		return false, nil
	}

	// Check if migration already done (migration marker file)
	dataDir, err := DataDir()
	if err != nil {
		return false, err
	}
	markerFile := filepath.Join(dataDir, ".migrated-from-legacy")
	if _, err := os.Stat(markerFile); err == nil {
		return false, nil
	}

	// Perform migration
	configDir, err := ConfigDir()
	if err != nil {
		return false, err
	}

	// Move libby-identity.json to config dir
	legacyIdentity := filepath.Join(legacyDir, "libby-identity.json")
	newIdentity := filepath.Join(configDir, "libby-identity.json")
	if _, err := os.Stat(legacyIdentity); err == nil {
		// Ensure target directory exists
		if err := os.MkdirAll(configDir, 0700); err != nil {
			return false, err
		}
		if err := os.Rename(legacyIdentity, newIdentity); err != nil {
			return false, err
		}
	}

	// Create migration marker
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return true, err
	}
	if err := os.WriteFile(markerFile, []byte("migrated"), 0600); err != nil {
		return true, err
	}

	// Remove legacy directory if empty
	os.Remove(legacyDir)

	return true, nil
}
