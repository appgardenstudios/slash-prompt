package internal

import (
	"context"
	"log/slog"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterPrompts(mcpServer *server.MCPServer, data *ServerData) {
	slog.Info("Registering prompts", "prompts", len(data.Prompts))

	for name, prompt := range data.Prompts {
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

		if resource.IsBlob {
			resourceContent = mcp.BlobResourceContents{
				URI:      resource.URI,
				MIMEType: resource.MimeType,
				Blob:     resource.Content,
			}
		} else {
			resourceContent = mcp.TextResourceContents{
				URI:      resource.URI,
				MIMEType: resource.MimeType,
				Text:     resource.Content,
			}
		}

		messages = append(messages, mcp.NewPromptMessage(
			mcp.RoleUser,
			mcp.NewEmbeddedResource(resourceContent),
		))
	}

	return mcp.NewGetPromptResult(prompt.Description, messages), nil
}
