package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/booklife-mcp/internal/config"
	"github.com/user/booklife-mcp/internal/dirs"
	"github.com/user/booklife-mcp/internal/providers/libby"
	"github.com/user/booklife-mcp/internal/server"
)

// Version information set by ldflags during build
var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "serve":
		cmdServe()
	case "libby-connect":
		cmdLibbyConnect()
	case "version", "-v", "--version":
		cmdVersion()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func cmdVersion() {
	fmt.Printf("BookLife MCP Server v%s (built %s)\n", Version, BuildTime)
}

func printUsage() {
	defaultPath, _ := config.DefaultPath()
	configDir, _ := dirs.ConfigDir()

	fmt.Println(`BookLife MCP Server - Personal Reading Life Assistant

Usage:
  booklife <command> [options]

Commands:
  serve            Start the MCP server
  libby-connect    Connect a Libby clone code (you have 40 seconds!)
  version          Show version information
  help             Show this help message

Serve Options:
  --config PATH    Path to configuration file
                   (default: ` + defaultPath + `)

Config directories:
  Linux/macOS:     ` + configDir + `
  Windows:         %APPDATA%\BookLife

Libby Connect:
  booklife libby-connect <8-digit-code> [--skip-tls-verify]

  To get a clone code:
    1. Open Libby app on your phone
    2. Go to Settings > Copy To Another Device
    3. Tap "Sonos Speakers" or "Android Automotive"
    4. Run this command immediately with the displayed code
    5. You have ~40 seconds before the code expires!

Options:
  --skip-tls-verify  Skip TLS certificate verification (insecure)`)
}

func cmdServe() {
	// Use default config path if not specified
	configPath, _ := config.DefaultPath()

	// Parse serve-specific flags
	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--config", "-c":
			if i+1 < len(args) {
				configPath = args[i+1]
				i++
			}
		}
	}

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals (silently - MCP uses stdio)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Create and run the MCP server
	srv, err := server.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create server: %v\n", err)
		os.Exit(1)
	}

	// Note: No startup message - MCP uses stdio for communication
	if err := srv.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func cmdLibbyConnect() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Error: clone code required")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Usage: booklife libby-connect <8-digit-code> [--skip-tls-verify]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Options:")
		fmt.Fprintln(os.Stderr, "  --skip-tls-verify  Skip TLS certificate verification (use if OverDrive cert is broken)")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "To get a clone code:")
		fmt.Fprintln(os.Stderr, "  1. Open Libby app on your phone")
		fmt.Fprintln(os.Stderr, "  2. Go to Settings > Copy To Another Device")
		fmt.Fprintln(os.Stderr, "  3. Tap \"Sonos Speakers\" or \"Android Automotive\"")
		fmt.Fprintln(os.Stderr, "  4. Run this command immediately with the displayed code")
		os.Exit(1)
	}

	code := os.Args[2]
	skipTLSVerify := false
	for _, arg := range os.Args[3:] {
		if arg == "--skip-tls-verify" {
			skipTLSVerify = true
		}
	}

	// Validate code format
	if len(code) != 8 {
		fmt.Fprintf(os.Stderr, "Error: clone code must be exactly 8 digits (got %d)\n", len(code))
		os.Exit(1)
	}

	for _, c := range code {
		if c < '0' || c > '9' {
			fmt.Fprintf(os.Stderr, "Error: clone code must contain only digits\n")
			os.Exit(1)
		}
	}

	fmt.Println("⏱️  Connecting to Libby...")
	fmt.Printf("   Code: %s\n", code)
	if skipTLSVerify {
		fmt.Println("   ⚠️  TLS verification disabled (insecure)")
	}

	// Connect and save identity
	identity, libraries, err := libby.ConnectWithOptions(code, skipTLSVerify)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Failed to connect: %v\n", err)
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Common issues:")
		fmt.Fprintln(os.Stderr, "  - Code expired (you only have ~40 seconds)")
		fmt.Fprintln(os.Stderr, "  - Typo in the code")
		fmt.Fprintln(os.Stderr, "  - Network connectivity issue")
		os.Exit(1)
	}

	// Save identity to disk
	if err := libby.SaveIdentity(identity); err != nil {
		fmt.Fprintf(os.Stderr, "\n⚠️  Connected but failed to save identity: %v\n", err)
		os.Exit(1)
	}

	// Get the actual path where identity was saved
	configDir, _ := dirs.ConfigDir()
	identityPath := configDir + "/libby-identity.json"

	fmt.Println("\n✅ Successfully connected to Libby!")
	fmt.Println("")
	fmt.Printf("   Linked libraries (%d):\n", len(libraries))
	for _, lib := range libraries {
		fmt.Printf("     • %s\n", lib.Name)
	}
	fmt.Println("")
	fmt.Printf("   Identity saved to %s\n", identityPath)
	fmt.Println("   You can now use BookLife with your Libby account!")
}
