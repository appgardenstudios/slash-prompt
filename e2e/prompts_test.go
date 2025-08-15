package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestPromptsList(t *testing.T) {
	client := setupMCPClient(t, "./_input/test-config.yml")
	ctx := context.Background()

	prompts, err := client.ListPrompts(ctx, mcp.ListPromptsRequest{})
	if err != nil {
		t.Fatalf("expected to list prompts successfully: %v", err)
	}

	goldenPath := "./_golden/prompts-list.json"
	outputPath := "./_output/prompts-list-" + time.Now().Format("20060102150405") + ".json"

	outputJSON(t, prompts, outputPath)

	if *update {
		updateGolden(goldenPath, outputPath, t)
	}

	compareFiles(goldenPath, outputPath, t)
}

func TestPromptsGetPrompt(t *testing.T) {
	request := mcp.GetPromptRequest{}
	request.Params.Name = "analyze"
	request.Params.Arguments = map[string]string{
		"language": "Go",
		"focus":    "performance and security",
	}

	goldenPath := "./_golden/prompts-get-prompt.json"
	outputPath := "./_output/prompts-get-prompt-" + time.Now().Format("20060102150405") + ".json"

	response := callMCPPrompt(t, "./_input/test-config.yml", request)

	outputJSON(t, response, outputPath)

	if *update {
		updateGolden(goldenPath, outputPath, t)
	}

	compareFiles(goldenPath, outputPath, t)
}

func TestPromptsGetPromptWithRepoPrefix(t *testing.T) {
	request := mcp.GetPromptRequest{}
	request.Params.Name = "repo-1:analyze"
	request.Params.Arguments = map[string]string{
		"language": "JavaScript",
		"focus":    "maintainability",
	}

	goldenPath := "./_golden/prompts-get-prompt-with-repo-prefix.json"
	outputPath := "./_output/prompts-get-prompt-with-repo-prefix-" + time.Now().Format("20060102150405") + ".json"

	response := callMCPPrompt(t, "./_input/test-config.yml", request)

	outputJSON(t, response, outputPath)

	if *update {
		updateGolden(goldenPath, outputPath, t)
	}

	compareFiles(goldenPath, outputPath, t)
}
