package internal

import (
	"fmt"
	"log/slog"
)

type LoadingError struct {
	Type   string `json:"type"` // "repo", "resource", "prompt"
	Repo   string `json:"repo"`
	RepoID string `json:"repoID"`
	Path   string `json:"path,omitempty"`
	Error  string `json:"error"`
}

type ServerData struct {
	Prompts       map[string]ParsedPrompt
	Resources     map[string]Resource
	LoadingErrors []LoadingError
}

func LoadAllData(config *Config) *ServerData {
	data := &ServerData{
		Prompts:       make(map[string]ParsedPrompt),
		Resources:     make(map[string]Resource),
		LoadingErrors: make([]LoadingError, 0),
	}

	for _, repoConfig := range config.Repos {
		loadRepoData(repoConfig, data)
	}

	slog.Info("Data loading completed",
		"prompts", len(data.Prompts),
		"resources", len(data.Resources),
		"errors", len(data.LoadingErrors))

	return data
}

func loadRepoData(repoConfig RepoConfig, data *ServerData) {
	repoID, _ := getRepoID(repoConfig)

	// Clone repository
	repo, err := cloneRepository(repoConfig)
	if err != nil {
		slog.Error("Failed to clone repository", "repo", repoID, "error", err)
		data.LoadingErrors = append(data.LoadingErrors, LoadingError{
			Type:   "repo",
			Repo:   repoConfig.Repo,
			RepoID: repoID,
			Error:  fmt.Sprintf("Failed to clone repository: %v", err),
		})
		return
	}

	// Get all files from repository
	allFiles, err := getRepositoryFiles(repo)
	if err != nil {
		slog.Error("Failed to get repository files", "repo", repoID, "error", err)
		data.LoadingErrors = append(data.LoadingErrors, LoadingError{
			Type:   "repo",
			Repo:   repoConfig.Repo,
			RepoID: repoID,
			Error:  fmt.Sprintf("Failed to get repository files: %v", err),
		})
		return
	}

	// Load prompts if configured
	if repoConfig.Prompts != nil {
		loadPrompts(repoConfig, repoID, allFiles, data)
	}

	// Load resources if configured
	if repoConfig.Resources != nil {
		loadResources(repoConfig, repoID, allFiles, data)
	}
}

func loadPrompts(repoConfig RepoConfig, repoID string, allFiles map[string]*File, data *ServerData) {
	// Filter files for prompts
	filteredFiles := filterFiles(allFiles, repoConfig.Prompts)

	if len(filteredFiles) == 0 {
		slog.Debug("No prompt files found", "repo", repoID)
		return
	}

	promptCount := 0
	for _, file := range filteredFiles {
		prompt, err := parsePrompt(file.Path, file, repoID, allFiles, data)
		if err != nil {
			slog.Error("Failed to parse prompt", "file", file.Path, "repo", repoID, "error", err)
			data.LoadingErrors = append(data.LoadingErrors, LoadingError{
				Type:   "prompt",
				Repo:   repoConfig.Repo,
				RepoID: repoID,
				Path:   file.Path,
				Error:  fmt.Sprintf("Failed to parse prompt: %v", err),
			})
			continue
		}

		// Store prompt both with and without repo prefix
		data.Prompts[prompt.Name] = *prompt
		data.Prompts[repoID+":"+prompt.Name] = *prompt

		// Add prompt resources to global resources map
		for _, resource := range prompt.Resources {
			if _, exists := data.Resources[resource.URI]; !exists {
				data.Resources[resource.URI] = resource
			}
		}

		promptCount++
	}

	slog.Info("Loaded prompts", "repo", repoID, "count", promptCount)
}

func loadResources(repoConfig RepoConfig, repoID string, allFiles map[string]*File, data *ServerData) {
	// Filter files for resources
	filteredFiles := filterFiles(allFiles, repoConfig.Resources)

	if len(filteredFiles) == 0 {
		slog.Debug("No resource files found", "repo", repoID)
		return
	}

	// Parse resources
	resources := parseResources(filteredFiles, repoID, data)

	// Add to global resources map
	// Check if resource already exists to avoid re-reading
	for uri, resource := range resources {
		if _, exists := data.Resources[uri]; !exists {
			data.Resources[uri] = resource
		}
	}

	slog.Info("Loaded resources", "repo", repoID, "count", len(resources))
}
