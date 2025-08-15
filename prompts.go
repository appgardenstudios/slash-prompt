package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerPrompts(mcpServer *server.MCPServer, data *ServerData) {
	slog.Info("Registering prompts", "prompts", len(data.Prompts))

	for name, prompt := range data.Prompts {
		// Create MCP prompt with the actual name being used
		mcpPrompt := mcp.NewPrompt(name, createPromptOptions(prompt)...)

		// Register prompt handler
		mcpServer.AddPrompt(mcpPrompt, func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return handlePromptRequest(ctx, request, prompt)
		})

		slog.Debug("Registered prompt", "name", name, "repo", prompt.Repo)
	}
}

func createPromptOptions(prompt ParsedPrompt) []mcp.PromptOption {
	// Create prompt options
	opts := []mcp.PromptOption{mcp.WithPromptDescription(prompt.Description)}

	// Convert arguments
	for _, arg := range prompt.Arguments {
		if arg.Required {
			opts = append(opts, mcp.WithArgument(
				arg.Name,
				mcp.ArgumentDescription(arg.Description),
				mcp.RequiredArgument(),
			))
		} else {
			opts = append(opts, mcp.WithArgument(
				arg.Name,
				mcp.ArgumentDescription(arg.Description),
			))
		}
	}

	return opts
}

func handlePromptRequest(_ context.Context, request mcp.GetPromptRequest, prompt ParsedPrompt) (*mcp.GetPromptResult, error) {
	slog.Debug("Handling prompt request", "name", prompt.Name, "args", len(request.Params.Arguments))

	// Process arguments and substitute in content
	content := prompt.Content
	for argName, argValue := range request.Params.Arguments {
		placeholder := "${" + argName + "}"
		content = strings.ReplaceAll(content, placeholder, argValue)
	}

	// Create messages
	messages := []mcp.PromptMessage{
		mcp.NewPromptMessage(
			mcp.RoleUser,
			mcp.NewTextContent(content),
		),
	}

	// Add embedded resources
	for _, resource := range prompt.Resources {
		var resourceContent mcp.ResourceContents
		uri := BuildResourceURI(resource.Repo, resource.Path)

		if resource.IsBlob {
			resourceContent = mcp.BlobResourceContents{
				URI:      uri,
				MIMEType: resource.MimeType,
				Blob:     resource.Content,
			}
		} else {
			resourceContent = mcp.TextResourceContents{
				URI:      uri,
				MIMEType: resource.MimeType,
				Text:     resource.Content,
			}
		}

		messages = append(messages, mcp.NewPromptMessage(
			mcp.RoleUser,
			mcp.NewEmbeddedResource(resourceContent),
		))
	}

	description := prompt.Description
	if description == "" {
		description = fmt.Sprintf("Prompt from %s repository", prompt.Repo)
	}

	return mcp.NewGetPromptResult(description, messages), nil
}
