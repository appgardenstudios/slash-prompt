package e2e

import (
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestToolsListResources(t *testing.T) {
	request := mcp.CallToolRequest{}
	request.Params.Name = "listResources"

	goldenPath := "./_golden/tools-list-resources.json"
	outputPath := "./_output/tools-list-resources-" + time.Now().Format("20060102150405") + ".json"

	response := callMCPTool(t, "./_input/test-config.yml", request)

	outputJSON(t, response, outputPath)

	if *update {
		updateGolden(goldenPath, outputPath, t)
	}

	compareFiles(goldenPath, outputPath, t)
}

func TestToolsListResourcesWithRepoFilter(t *testing.T) {
	request := mcp.CallToolRequest{}
	request.Params.Name = "listResources"
	request.Params.Arguments = map[string]any{
		"repo": "repo-1",
	}

	goldenPath := "./_golden/tools-list-resources-with-repo-filter.json"
	outputPath := "./_output/tools-list-resources-with-repo-filter-" + time.Now().Format("20060102150405") + ".json"

	response := callMCPTool(t, "./_input/test-config.yml", request)

	outputJSON(t, response, outputPath)

	if *update {
		updateGolden(goldenPath, outputPath, t)
	}

	compareFiles(goldenPath, outputPath, t)
}

func TestToolsGetResource(t *testing.T) {
	request := mcp.CallToolRequest{}
	request.Params.Name = "getResource"
	request.Params.Arguments = map[string]any{
		"resource_uri": "file://repo-1/resources/context.template.md",
	}

	goldenPath := "./_golden/tools-get-resource.json"
	outputPath := "./_output/tools-get-resource-" + time.Now().Format("20060102150405") + ".json"

	response := callMCPTool(t, "./_input/test-config.yml", request)

	outputJSON(t, response, outputPath)

	if *update {
		updateGolden(goldenPath, outputPath, t)
	}

	compareFiles(goldenPath, outputPath, t)
}

func TestToolsListErrors(t *testing.T) {
	request := mcp.CallToolRequest{}
	request.Params.Name = "listErrors"

	goldenPath := "./_golden/tools-list-errors.json"
	outputPath := "./_output/tools-list-errors-" + time.Now().Format("20060102150405") + ".json"

	response := callMCPTool(t, "./_input/test-config.yml", request)

	outputJSON(t, response, outputPath)

	if *update {
		updateGolden(goldenPath, outputPath, t)
	}

	compareFiles(goldenPath, outputPath, t)
}

func TestToolsGetResourceNotFound(t *testing.T) {
	request := mcp.CallToolRequest{}
	request.Params.Name = "getResource"
	request.Params.Arguments = map[string]any{
		"resource_uri": "file://repo-1/nonexistent.md",
	}

	response := callMCPTool(t, "./_input/test-config.yml", request)

	if !response.IsError {
		t.Fatal("expected result to be an error for non-existent resource")
	}

	goldenPath := "./_golden/tools-get-resource-not-found.json"
	outputPath := "./_output/tools-get-resource-not-found-" + time.Now().Format("20060102150405") + ".json"

	outputJSON(t, response, outputPath)

	if *update {
		updateGolden(goldenPath, outputPath, t)
	}

	compareFiles(goldenPath, outputPath, t)
}
