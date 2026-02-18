package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	kibble "github.com/thinkscotty/kibble"
	"github.com/thinkscotty/kibble/internal/ai"
	"github.com/thinkscotty/kibble/internal/config"
	"github.com/thinkscotty/kibble/internal/database"
	"github.com/thinkscotty/kibble/internal/scheduler"
	"github.com/thinkscotty/kibble/internal/scraper"
	"github.com/thinkscotty/kibble/internal/server"
	"github.com/thinkscotty/kibble/internal/similarity"
	"github.com/thinkscotty/kibble/internal/updater"
	"github.com/thinkscotty/kibble/internal/wikipedia"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	themesPath := flag.String("themes", "themes.yaml", "Path to themes file")
	showVersion := flag.Bool("version", false, "Show version and exit")
	doUpdate := flag.Bool("update", false, "Check for updates and install if available")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Kibble %s (built %s)\n", version, buildTime)
		os.Exit(0)
	}

	if *doUpdate {
		runUpdate(version)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Initialize logger
	var logLevel slog.Level
	switch strings.ToLower(cfg.Logging.Level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))

	slog.Info("Starting Kibble", "version", version)

	// Initialize database
	db, err := database.New(cfg.Database.Path)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	slog.Info("Database initialized", "path", cfg.Database.Path)

	// Load color themes
	themes, err := config.LoadThemes(*themesPath, kibble.ThemesYAML)
	if err != nil {
		slog.Error("Failed to load themes", "error", err)
		os.Exit(1)
	}
	slog.Info("Loaded themes", "count", len(themes))

	// Initialize services
	wikiClient := wikipedia.New()
	aiClient := ai.NewClient(db, wikiClient)
	sim := similarity.New(cfg.Similarity.Threshold, cfg.Similarity.NGramSize)
	sc := scraper.New()
	sched := scheduler.New(db, aiClient, sim, sc)

	// Build HTTP server
	srv := server.New(cfg, db, aiClient, sim, sched, themes, version, buildTime)

	// Start scheduler in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go sched.Run(ctx)

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		slog.Info("Shutting down...")
		cancel()
		srv.Shutdown(context.Background())
	}()

	// Start serving
	slog.Info("Server listening", "addr", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))
	if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}
}

func runUpdate(currentVersion string) {
	fmt.Printf("Kibble %s — checking for updates...\n", currentVersion)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	info, err := updater.CheckForUpdate(ctx, currentVersion)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Update check failed: %s\n", err)
		os.Exit(1)
	}
	if info == nil {
		fmt.Println("Already running the latest version.")
		return
	}

	fmt.Printf("Update available: %s → %s\n", currentVersion, info.TagName)
	fmt.Printf("Binary: %s (%s)\n", info.AssetName, updater.FormatBytes(info.AssetSize))
	fmt.Printf("Downloading...\n")

	result, err := updater.DownloadAndInstall(ctx, info, currentVersion)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Installation failed: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Updated %s → %s successfully.\n", result.OldVersion, result.NewVersion)
	fmt.Println("Restart the service to use the new version:")
	fmt.Println("  sudo systemctl restart kibble")
}
