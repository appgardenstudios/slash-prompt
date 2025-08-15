package main

import (
	"testing"
)

func TestExtractRepoName(t *testing.T) {
	tests := []struct {
		name     string
		repoURL  string
		expected string
	}{
		{
			name:     "SSH GitHub URL",
			repoURL:  "git@github.com:owner/repo.git",
			expected: "repo",
		},
		{
			name:     "HTTPS GitHub URL",
			repoURL:  "https://github.com/owner/repo.git",
			expected: "repo",
		},
		{
			name:     "HTTPS URL without .git",
			repoURL:  "https://github.com/owner/repo",
			expected: "repo",
		},
		{
			name:     "Local path",
			repoURL:  "./test-repo",
			expected: "test_repo",
		},
		{
			name:     "Local absolute path",
			repoURL:  "/path/to/my-repo",
			expected: "my_repo",
		},
		{
			name:     "Complex repo name with hyphens",
			repoURL:  "git@github.com:owner/my-awesome-repo.git",
			expected: "my_awesome_repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractRepoName(tt.repoURL)
			if err != nil {
				t.Errorf("extractRepoName(%q) returned an error: %v", tt.repoURL, err)
			}
			if result != tt.expected {
				t.Errorf("extractRepoName(%q) = %q, want %q", tt.repoURL, result, tt.expected)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		expectErr bool
	}{
		{
			name: "Valid config with prompts",
			config: Config{
				Repos: []RepoConfig{
					{
						Repo: "git@github.com:owner/repo.git",
						ID:   "test-repo",
						Prompts: &FileFilter{
							Include: []string{"**/*.md"},
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "Valid config with resources",
			config: Config{
				Repos: []RepoConfig{
					{
						Repo: "https://github.com/owner/repo.git",
						Resources: &FileFilter{
							Include: []string{"**/*.template.md"},
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "Invalid config - no repos",
			config: Config{
				Repos: []RepoConfig{},
			},
			expectErr: true,
		},
		{
			name: "Invalid config - missing repo field",
			config: Config{
				Repos: []RepoConfig{
					{
						Prompts: &FileFilter{},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Invalid config - neither prompts nor resources",
			config: Config{
				Repos: []RepoConfig{
					{
						Repo: "git@github.com:owner/repo.git",
					},
				},
			},
			expectErr: true,
		},
		{
			name: "Invalid config - duplicate IDs",
			config: Config{
				Repos: []RepoConfig{
					{
						Repo: "git@github.com:owner/repo1.git",
						ID:   "same-id",
						Prompts: &FileFilter{
							Include: []string{"**/*.md"},
						},
					},
					{
						Repo: "git@github.com:owner/repo2.git",
						ID:   "same-id",
						Prompts: &FileFilter{
							Include: []string{"**/*.md"},
						},
					},
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(&tt.config)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestValidateFileFilter(t *testing.T) {
	tests := []struct {
		name      string
		filter    *FileFilter
		expectErr bool
	}{
		{
			name: "Valid filter",
			filter: &FileFilter{
				Include: []string{"**/*.md", "*.txt"},
				Exclude: []string{"**/draft.md"},
			},
			expectErr: false,
		},
		{
			name:      "Valid filter with defaults",
			filter:    &FileFilter{},
			expectErr: false,
		},
		{
			name: "Invalid include pattern",
			filter: &FileFilter{
				Include: []string{"[invalid"},
			},
			expectErr: true,
		},
		{
			name: "Invalid exclude pattern",
			filter: &FileFilter{
				Include: []string{"**/*.md"},
				Exclude: []string{"[invalid"},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFileFilter(tt.filter, "test")
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
