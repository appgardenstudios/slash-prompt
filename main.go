package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/mark3labs/mcp-go/server"
)

var Version = "development"

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	slog.Info("Starting slash-prompt MCP server", "version", Version)

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Load all data
	serverData := loadAllData(config)

	// Create MCP server
	mcpServer := server.NewMCPServer(
		"slash-prompt",
		Version,
		server.WithToolCapabilities(false),
		server.WithResourceCapabilities(true, false),
		server.WithPromptCapabilities(true),
		server.WithRecovery(),
	)

	// Register tools and prompts
	registerTools(mcpServer, serverData)
	registerPrompts(mcpServer, serverData)
	registerResources(mcpServer, serverData)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigChan
		slog.Info("Received shutdown signal, gracefully stopping...")
		os.Exit(0)
	}()

	// Start stdio server
	slog.Info("Starting MCP server on stdio")
	if err := server.ServeStdio(mcpServer); err != nil {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}
}

