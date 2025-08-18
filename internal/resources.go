package internal

import (
	"context"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterResources(mcpServer *server.MCPServer, data *ServerData) {
	slog.Info("Registering resources", "count", len(data.Resources))

	// Register each resource as a static resource
	for uri, resource := range data.Resources {
		// Create MCP resource
		mcpResource := mcp.NewResource(
			uri,
			resource.Name,
			mcp.WithMIMEType(resource.MimeType),
		)

		// Register resource handler
		mcpServer.AddResource(mcpResource, func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			return handleResourceRequest(ctx, request, resource)
		})

		slog.Debug("Registered resource", "uri", uri, "name", resource.Name, "repo", resource.Repo)
	}
}

func handleResourceRequest(_ context.Context, request mcp.ReadResourceRequest, resource Resource) ([]mcp.ResourceContents, error) {
	slog.Debug("Handling resource request", "uri", request.Params.URI, "name", resource.Name)

	var content mcp.ResourceContents

	if resource.IsBlob {
		content = mcp.BlobResourceContents{
			URI:      request.Params.URI,
			MIMEType: resource.MimeType,
			Blob:     resource.Content,
		}
	} else {
		content = mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: resource.MimeType,
			Text:     resource.Content,
		}
	}

	return []mcp.ResourceContents{content}, nil
}
