package e2e

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	mcpClient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

var update = flag.Bool("update", false, "update golden files")

// setupMCPClient creates and initializes an MCP client for testing
func setupMCPClient(t *testing.T, configPath string) *mcpClient.Client {
	// Get absolute path to config
	absConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		t.Fatalf("expected to get absolute path for config: %v", err)
	}

	// Build the binary path
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("expected to get current working directory: %v", err)
	}
	binaryPath := filepath.Join(dir, "../slash-prompt-e2e")

	// Create MCP client using stdio transport
	args := []string{}

	// Set environment variables for config path and test repo path
	os.Setenv("SLASH_PROMPT_CONFIG_PATH", absConfigPath)
	testRepoPath, err := filepath.Abs(filepath.Join(dir, "../e2e/_input/test-repos"))
	if err != nil {
		t.Fatalf("expected to get absolute path for test repos: %v", err)
	}
	os.Setenv("E2E_TEST_REPO_PATH", testRepoPath)

	t.Logf("Starting MCP client with config: %v", absConfigPath)
	client, err := mcpClient.NewStdioMCPClient(binaryPath, nil, args...)
	if err != nil {
		t.Fatalf("expected to create MCP client successfully: %v", err)
	}

	t.Cleanup(func() {
		// Ignore broken pipe errors during cleanup as they're expected
		_ = client.Close()
	})

	// Initialize the client
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	request := mcp.InitializeRequest{}
	request.Params.ProtocolVersion = "2024-11-05"
	request.Params.ClientInfo = mcp.Implementation{
		Name:    "slash-prompt-e2e-test-client",
		Version: "0.0.1",
	}

	result, err := client.Initialize(ctx, request)
	if err != nil {
		t.Fatalf("failed to initialize MCP client: %v", err)
	}
	if result.ServerInfo.Name != "slash-prompt" {
		t.Fatalf("unexpected server name, got %s, expected slash-prompt", result.ServerInfo.Name)
	}

	return client
}

// callMCPTool calls an MCP tool and returns the raw response
func callMCPTool(t *testing.T, configPath string, request mcp.CallToolRequest) *mcp.CallToolResult {
	client := setupMCPClient(t, configPath)
	ctx := context.Background()

	response, err := client.CallTool(ctx, request)
	if err != nil {
		t.Fatalf("expected to call '%s' tool successfully: %v", request.Params.Name, err)
	}

	return response
}

// callMCPPrompt calls an MCP prompt and returns the raw response
func callMCPPrompt(t *testing.T, configPath string, request mcp.GetPromptRequest) *mcp.GetPromptResult {
	client := setupMCPClient(t, configPath)
	ctx := context.Background()

	response, err := client.GetPrompt(ctx, request)
	if err != nil {
		t.Fatalf("expected to get '%s' prompt successfully: %v", request.Params.Name, err)
	}

	return response
}

// outputJSON marshals data to JSON and writes it to the specified output file
func outputJSON(t *testing.T, data any, outputPath string) {
	// Output the data as JSON-encoded
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("expected to marshal data to JSON: %v", err)
	}

	err = os.WriteFile(outputPath, jsonData, 0644)
	if err != nil {
		t.Fatalf("expected to write output file: %v", err)
	}
}

// updateGolden updates golden file with output content
func updateGolden(goldenPath, outputPath string, t *testing.T) {
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	err = os.WriteFile(goldenPath, content, 0644)
	if err != nil {
		t.Fatalf("failed to update golden file: %v", err)
	}
	t.Logf("Updated golden file: %s", goldenPath)
}

// compareFiles compares output file with golden file
func compareFiles(goldenPath, outputPath string, t *testing.T) {
	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}

	output, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if string(golden) != string(output) {
		t.Errorf("output doesn't match golden file.\nExpected:\n%s\n\nGot:\n%s", string(golden), string(output))
	}
}
