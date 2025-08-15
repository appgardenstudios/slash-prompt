package main

import (
	"fmt"
	"io"
	"log/slog"
	"mime"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// File represents a file with lazy content loading
type File struct {
	Name    string
	Size    int64
	gitFile *object.File // Reference to the git file for lazy loading
	content []byte       // Cached content
	isRead  bool         // Track if content has been read
}

// Contents returns the file content, reading it lazily if needed
func (f *File) Contents() ([]byte, error) {
	if f.isRead {
		return f.content, nil
	}

	if f.gitFile == nil {
		return nil, fmt.Errorf("no git file reference for lazy loading")
	}

	// Read content from git file
	reader, err := f.gitFile.Reader()
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	defer reader.Close()

	content := make([]byte, f.gitFile.Size)
	_, err = io.ReadFull(reader, content)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	f.content = content
	f.isRead = true
	return content, nil
}

func filterFiles(files map[string]*File, filter *FileFilter) []*File {
	if filter == nil {
		return []*File{}
	}

	var result []*File

	for filePath, file := range files {
		slog.Debug("Checking file", "path", filePath)

		// Check include patterns
		included := false
		for _, pattern := range getFileFilterInclude(filter) {
			if doublestar.MatchUnvalidated(pattern, filePath) {
				slog.Debug("File matches include pattern", "pattern", pattern, "file", filePath)
				included = true
				break
			}
		}

		if !included {
			slog.Debug("File does not match any include patterns", "file", filePath)
			continue
		}

		// Check exclude patterns
		excluded := false
		for _, pattern := range filter.Exclude {
			if doublestar.MatchUnvalidated(pattern, filePath) {
				slog.Debug("File matches exclude pattern", "pattern", pattern, "file", filePath)
				excluded = true
				break
			}
		}

		if !excluded {
			slog.Debug("File included after filtering", "file", filePath)
			result = append(result, file)
		}
	}

	slog.Debug("Filtered files", "included", len(result))

	return result
}

func detectMIMEType(filePath string) string {
	ext := filepath.Ext(filePath)
	mimeType := mime.TypeByExtension(ext)

	if mimeType == "" {
		// Handle extensions not in the builtin list
		switch strings.ToLower(ext) {
		case ".md":
			return "text/markdown"
		case ".yml", ".yaml":
			return "application/x-yaml"
		case ".txt":
			return "text/plain"
		default:
			return "application/octet-stream"
		}
	}

	// Clean up charset from MIME type for consistency
	if semicolon := strings.Index(mimeType, ";"); semicolon != -1 {
		return mimeType[:semicolon]
	}

	return mimeType
}

func isTextMIMEType(mimeType string) bool {
	return strings.HasPrefix(mimeType, "text/") ||
		mimeType == "application/json" ||
		mimeType == "application/x-yaml" ||
		mimeType == "application/javascript" ||
		mimeType == "application/xml"
}
