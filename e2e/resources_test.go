package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestResourcesList(t *testing.T) {
	client := setupMCPClient(t, "./_input/test-config.yml")
	ctx := context.Background()

	request := mcp.ListResourcesRequest{}
	response, err := client.ListResources(ctx, request)
	if err != nil {
		t.Fatalf("expected to list resources successfully: %v", err)
	}

	goldenPath := "./_golden/resources-list.json"
	outputPath := "./_output/resources-list-" + time.Now().Format("20060102150405") + ".json"

	outputJSON(t, response, outputPath)

	if *update {
		updateGolden(goldenPath, outputPath, t)
	}

	compareFiles(goldenPath, outputPath, t)
}

func TestResourcesRead(t *testing.T) {
	client := setupMCPClient(t, "./_input/test-config.yml")
	ctx := context.Background()

	request := mcp.ReadResourceRequest{}
	request.Params.URI = "file://repo-1/resources/context.template.md"

	response, err := client.ReadResource(ctx, request)
	if err != nil {
		t.Fatalf("expected to read resource successfully: %v", err)
	}

	if len(response.Contents) == 0 {
		t.Fatal("expected at least one resource content")
	}

	goldenPath := "./_golden/resources-read.json"
	outputPath := "./_output/resources-read-" + time.Now().Format("20060102150405") + ".json"

	outputJSON(t, response, outputPath)

	if *update {
		updateGolden(goldenPath, outputPath, t)
	}

	compareFiles(goldenPath, outputPath, t)
}
