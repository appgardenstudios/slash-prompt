package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerTools(mcpServer *server.MCPServer, data *ServerData) {
	slog.Info("Registering tools")

	// Register listErrors tool
	listErrorsTool := mcp.NewTool("listErrors",
		mcp.WithDescription("List all loading errors that occurred during startup"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
	)
	mcpServer.AddTool(listErrorsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleListErrorsRequest(ctx, request, data)
	})

	// Register listResources tool
	listResourcesTool := mcp.NewTool("listResources",
		mcp.WithDescription("List all available resources, optionally filtered by repository"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithString("repo",
			mcp.Description("Optional repository filter"),
		),
	)
	mcpServer.AddTool(listResourcesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleListResourcesRequest(ctx, request, data)
	})

	// Register getResource tool
	getResourceTool := mcp.NewTool("getResource",
		mcp.WithDescription("Get a specific resource by its URI"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithString("resource_uri",
			mcp.Required(),
			mcp.Description("The fully qualified resource URI"),
		),
	)

	mcpServer.AddTool(getResourceTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleGetResourceRequest(ctx, request, data)
	})

	slog.Info("Registered all tools", "error_tools", 1, "resource_tools", 2)
}

func handleListErrorsRequest(ctx context.Context, request mcp.CallToolRequest, data *ServerData) (*mcp.CallToolResult, error) {
	slog.Debug("Handling listErrors request")

	if len(data.LoadingErrors) == 0 {
		return mcp.NewToolResultText("No loading errors occurred.\n"), nil
	}

	// Convert errors to JSON for structured output
	errorsJSON, err := json.MarshalIndent(data.LoadingErrors, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal loading errors", "error", err)
		return mcp.NewToolResultError("Failed to format loading errors"), nil
	}

	result := "Loading errors occurred during startup:\n\n```json\n" + string(errorsJSON) + "\n```\n"
	return mcp.NewToolResultText(result), nil
}

func handleListResourcesRequest(_ context.Context, request mcp.CallToolRequest, data *ServerData) (*mcp.CallToolResult, error) {
	slog.Debug("Handling listResources request")

	// Get optional repo filter
	repoFilter := request.GetString("repo", "")

	var result strings.Builder
	result.WriteString("Available resources:\n\n")

	// Collect and sort resource keys
	var uris []string
	for uri, resource := range data.Resources {
		// Apply repo filter if specified
		if repoFilter != "" && resource.Repo != repoFilter {
			continue
		}
		uris = append(uris, uri)
	}
	sort.Strings(uris)

	for _, uri := range uris {
		resource := data.Resources[uri]
		result.WriteString(fmt.Sprintf("- **%s** (%s)\n", resource.Name, resource.Repo))
		result.WriteString(fmt.Sprintf("  - URI: `%s`\n", uri))
		result.WriteString(fmt.Sprintf("  - Path: `%s`\n", resource.Path))
		result.WriteString(fmt.Sprintf("  - MIME Type: `%s`\n", resource.MimeType))
		result.WriteString("\n")
	}

	if len(uris) == 0 {
		if repoFilter != "" {
			result.WriteString(fmt.Sprintf("No resources found for repository: %s\n", repoFilter))
		} else {
			result.WriteString("No resources available.\n")
		}
	} else {
		result.WriteString(fmt.Sprintf("Total: %d resources", len(uris)))
		if repoFilter != "" {
			result.WriteString(fmt.Sprintf(" (filtered by repo: %s)", repoFilter))
		}
		result.WriteString("\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

func handleGetResourceRequest(_ context.Context, request mcp.CallToolRequest, data *ServerData) (*mcp.CallToolResult, error) {
	slog.Debug("Handling getResource request")

	// Get required resource_uri parameter
	resourceURI, err := request.RequireString("resource_uri")
	if err != nil {
		return mcp.NewToolResultError("resource_uri parameter is required"), nil
	}

	// Find resource using the full URI as key
	resource, exists := data.Resources[resourceURI]
	if !exists {
		return mcp.NewToolResultError(fmt.Sprintf("Resource not found: %s", resourceURI)), nil
	}

	// Return resource information and content
	var result strings.Builder
	result.WriteString(fmt.Sprintf("# Resource: %s\n\n", resource.Name))
	result.WriteString(fmt.Sprintf("- **Repository:** %s\n", resource.Repo))
	result.WriteString(fmt.Sprintf("- **Path:** %s\n", resource.Path))
	result.WriteString(fmt.Sprintf("- **MIME Type:** %s\n", resource.MimeType))
	result.WriteString(fmt.Sprintf("- **URI:** %s\n", resourceURI))
	result.WriteString(fmt.Sprintf("- **Is Binary:** %t\n\n", resource.IsBlob))

	if resource.IsBlob {
		result.WriteString("**Content:** (binary data, base64 encoded)\n")
		result.WriteString("```\n")
		result.WriteString(resource.Content)
		result.WriteString("\n```\n")
	} else {
		result.WriteString("**Content:**\n")
		result.WriteString("```\n")
		result.WriteString(resource.Content)
		result.WriteString("\n```\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}
