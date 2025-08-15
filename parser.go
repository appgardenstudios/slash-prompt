package main

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type PromptMetadata struct {
	Name        string           `yaml:"name,omitempty"`
	Description string           `yaml:"description,omitempty"`
	Arguments   []PromptArgument `yaml:"arguments,omitempty"`
	Resources   []PromptResource `yaml:"resources,omitempty"`
}

type PromptArgument struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Required    bool   `yaml:"required,omitempty"`
}

type PromptResource struct {
	Path     string `yaml:"path"`
	Name     string `yaml:"name,omitempty"`
	MimeType string `yaml:"mimeType,omitempty"`
}

type ParsedPrompt struct {
	Name        string
	Repo        string
	Path        string
	Content     string
	Description string
	Arguments   []PromptArgument
	Resources   []Resource
}

type Resource struct {
	Path     string
	Name     string
	Repo     string
	Content  string
	IsBlob   bool
	MimeType string
}

func parsePrompt(filePath string, file *File, repoID string, allFiles map[string]*File, data *ServerData) (*ParsedPrompt, error) {
	content, err := file.Contents()
	if err != nil {
		return nil, fmt.Errorf("failed to read file contents: %w", err)
	}

	// Extract frontmatter and content
	frontmatter, promptContent, err := extractFrontmatter(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to extract frontmatter: %w", err)
	}

	// Parse frontmatter
	var metadata PromptMetadata
	if frontmatter != "" {
		if err := yaml.Unmarshal([]byte(frontmatter), &metadata); err != nil {
			return nil, fmt.Errorf("failed to parse frontmatter YAML: %w", err)
		}
	}

	// Set default name from filename if not specified
	promptName := metadata.Name
	if promptName == "" {
		promptName = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	}

	// Build parsed prompt
	parsed := &ParsedPrompt{
		Name:        promptName,
		Repo:        repoID,
		Path:        filePath,
		Content:     promptContent,
		Description: metadata.Description,
		Arguments:   metadata.Arguments,
		Resources:   make([]Resource, 0),
	}

	// Process declared resources
	for _, resourceSpec := range metadata.Resources {
		resource, err := getPromptResource(resourceSpec, filePath, repoID, allFiles, data)
		if err != nil {
			return nil, fmt.Errorf("failed to load prompt resource '%s' for prompt '%s': %w", resourceSpec.Path, promptName, err)
		}
		parsed.Resources = append(parsed.Resources, *resource)
	}

	slog.Info("Parsed prompt",
		"name", promptName,
		"repo", repoID,
		"resources", len(parsed.Resources),
		"arguments", len(parsed.Arguments))

	return parsed, nil
}

func extractFrontmatter(content string) (frontmatter, body string, err error) {
	lines := strings.Split(content, "\n")

	if len(lines) < 3 || lines[0] != "---" {
		// No frontmatter
		return "", content, nil
	}

	// Find end of frontmatter
	endIndex := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			endIndex = i
			break
		}
	}

	if endIndex == -1 {
		return "", content, fmt.Errorf("unclosed frontmatter")
	}

	frontmatter = strings.Join(lines[1:endIndex], "\n")

	if endIndex+1 < len(lines) {
		body = strings.Join(lines[endIndex+1:], "\n")
	} else {
		body = ""
	}

	return frontmatter, body, nil
}

func getPromptResource(spec PromptResource, promptPath, repoID string, allFiles map[string]*File, data *ServerData) (*Resource, error) {
	// Resolve resource path relative to prompt path
	resourcePath := spec.Path
	if !filepath.IsAbs(resourcePath) {
		resourcePath = filepath.Join(filepath.Dir(promptPath), resourcePath)
		resourcePath = filepath.Clean(resourcePath)
	}

	// Check if resource already exists in data.Resources
	resourceKey := filepath.Join(repoID, resourcePath)
	if existingResource, exists := data.Resources[resourceKey]; exists {
		return &existingResource, nil
	}

	// Set resource name
	resourceName := spec.Name
	if resourceName == "" {
		resourceName = filepath.Base(resourcePath)
	}

	// Find resource file
	file, exists := allFiles[resourcePath]
	if !exists {
		return nil, fmt.Errorf("resource not found: %s", resourcePath)
	}

	return getResource(file, resourcePath, resourceName, repoID, spec.MimeType)
}

func parseResources(files []*File, repoID string, data *ServerData) map[string]Resource {
	resources := make(map[string]Resource)

	for _, file := range files {
		filePath := file.Name
		resourceKey := BuildResourceURI(repoID, filePath)

		// Check if resource already exists
		if _, exists := data.Resources[resourceKey]; exists {
			continue
		}

		resource, err := getResource(file, filePath, filepath.Base(filePath), repoID, "")
		if err != nil {
			slog.Warn("Failed to read resource file", "path", filePath, "error", err)
			continue
		}

		resources[resourceKey] = *resource
	}

	slog.Debug("Parsed resources", "repo", repoID, "count", len(resources))
	return resources
}

func getResource(file *File, path, name, repo, mimeType string) (*Resource, error) {
	// Get file content
	content, err := file.Contents()
	if err != nil {
		return nil, fmt.Errorf("failed to read resource contents: %w", err)
	}

	// Determine MIME type if not provided
	if mimeType == "" {
		mimeType = detectMIMEType(path)
	}

	// Determine if content should be base64 encoded
	isBlob := !isTextMIMEType(mimeType)
	resourceContent := string(content)
	if isBlob {
		resourceContent = base64.StdEncoding.EncodeToString(content)
	}

	return &Resource{
		Path:     path,
		Name:     name,
		Repo:     repo,
		Content:  resourceContent,
		IsBlob:   isBlob,
		MimeType: mimeType,
	}, nil
}
