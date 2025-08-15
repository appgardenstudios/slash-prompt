package main

import (
	"testing"
)

func TestDetectMIMEType(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{
			name:     "Markdown file",
			filePath: "prompt.md",
			expected: "text/markdown",
		},
		{
			name:     "YAML file",
			filePath: "config.yml",
			expected: "application/x-yaml",
		},
		{
			name:     "JSON file",
			filePath: "data.json",
			expected: "application/json",
		},
		{
			name:     "Text file",
			filePath: "readme.txt",
			expected: "text/plain",
		},
		{
			name:     "Unknown extension",
			filePath: "file.unknown",
			expected: "application/octet-stream",
		},
		{
			name:     "No extension",
			filePath: "README",
			expected: "application/octet-stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectMIMEType(tt.filePath)
			if result != tt.expected {
				t.Errorf("detectMIMEType(%q) = %q, want %q", tt.filePath, result, tt.expected)
			}
		})
	}
}

func TestIsTextMIMEType(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		expected bool
	}{
		{
			name:     "Text plain",
			mimeType: "text/plain",
			expected: true,
		},
		{
			name:     "Text markdown",
			mimeType: "text/markdown",
			expected: true,
		},
		{
			name:     "Application JSON",
			mimeType: "application/json",
			expected: true,
		},
		{
			name:     "Application YAML",
			mimeType: "application/x-yaml",
			expected: true,
		},
		{
			name:     "Application JavaScript",
			mimeType: "application/javascript",
			expected: true,
		},
		{
			name:     "Binary data",
			mimeType: "application/octet-stream",
			expected: false,
		},
		{
			name:     "Image PNG",
			mimeType: "image/png",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTextMIMEType(tt.mimeType)
			if result != tt.expected {
				t.Errorf("isTextMIMEType(%q) = %v, want %v", tt.mimeType, result, tt.expected)
			}
		})
	}
}