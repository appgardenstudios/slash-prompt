package internal

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	giturls "github.com/whilp/git-urls"
	"gopkg.in/yaml.v3"
)

var (
	repoIDCharactersRegex = regexp.MustCompile(`[A-Za-z0-9_-]`)
	repoIDRegex           = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
)

type Config struct {
	Repos []RepoConfig `yaml:"repos"`
}

type RepoConfig struct {
	Repo      string      `yaml:"repo"`
	ID        string      `yaml:"id,omitempty"`
	Ref       string      `yaml:"ref,omitempty"`
	Prompts   *FileFilter `yaml:"prompts,omitempty"`
	Resources *FileFilter `yaml:"resources,omitempty"`
	Auth      *AuthConfig `yaml:"auth,omitempty"`
}

type FileFilter struct {
	Include []string `yaml:"include,omitempty"`
	Exclude []string `yaml:"exclude,omitempty"`
}

type AuthConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func LoadConfig() (*Config, error) {
	configPath := os.Getenv("SLASH_PROMPT_CONFIG_PATH")
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("could not determine home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, ".prompt.yml")
	}

	slog.Info("Loading configuration", "path", configPath)

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("could not get absolute path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	// Replace environment variables
	data = []byte(os.ExpandEnv(string(data)))

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("could not parse YAML config: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	slog.Info("Configuration loaded successfully", "repos", len(config.Repos))
	return &config, nil
}

func validateConfig(config *Config) error {
	if len(config.Repos) == 0 {
		return errors.New("no repositories configured")
	}

	seenIDs := make(map[string]bool)

	for i, repo := range config.Repos {
		// Ensure repo is specified
		if repo.Repo == "" {
			return fmt.Errorf("repo %d: repo field is required", i)
		}

		repoID, err := getRepoID(repo)
		if err != nil {
			return fmt.Errorf("repo %d: %w", i, err)
		}

		if !repoIDRegex.MatchString(repoID) {
			return fmt.Errorf("repo %d: invalid repo name format %q, must match %s", i, repoID, repoIDRegex.String())
		}

		// Check for duplicate IDs
		if seenIDs[repoID] {
			return fmt.Errorf("duplicate repo ID: %s", repoID)
		}
		seenIDs[repoID] = true

		// Ensure either prompts or resources is set
		if repo.Prompts == nil && repo.Resources == nil {
			return fmt.Errorf("repo %s: either prompts or resources must be configured", repoID)
		}

		// Validate prompts configuration
		if repo.Prompts != nil {
			if err := validateFileFilter(repo.Prompts, "prompts"); err != nil {
				return fmt.Errorf("repo %s: %w", repoID, err)
			}
		}

		// Validate resources configuration
		if repo.Resources != nil {
			if err := validateFileFilter(repo.Resources, "resources"); err != nil {
				return fmt.Errorf("repo %s: %w", repoID, err)
			}
		}

		// Validate auth configuration
		if repo.Auth != nil {
			if repo.Auth.Username == "" {
				return fmt.Errorf("repo %s: auth username is required", repoID)
			}
			if repo.Auth.Password == "" {
				return fmt.Errorf("repo %s: auth password is required", repoID)
			}
		}
	}

	return nil
}

func validateFileFilter(filter *FileFilter, filterType string) error {
	// Get effective values for validation (without modifying the original spec)
	effectiveInclude := getFileFilterInclude(filter)

	// Validate include patterns
	for _, pattern := range effectiveInclude {
		if !doublestar.ValidatePattern(pattern) {
			return fmt.Errorf("invalid %s include pattern: %s", filterType, pattern)
		}
	}

	// Validate exclude patterns
	for _, pattern := range filter.Exclude {
		if !doublestar.ValidatePattern(pattern) {
			return fmt.Errorf("invalid %s exclude pattern: %s", filterType, pattern)
		}
	}

	return nil
}

func getRepoID(repo RepoConfig) (string, error) {
	if repo.ID != "" {
		return repo.ID, nil
	}
	return extractRepoName(repo.Repo)
}

func getFileFilterInclude(filter *FileFilter) []string {
	if filter == nil || len(filter.Include) == 0 {
		return []string{"**/*.md"}
	}
	return filter.Include
}

func extractRepoName(repoURL string) (string, error) {
	u, err := giturls.Parse(repoURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse git URL %q: %w", repoURL, err)
	}

	// Extract repository name from the path
	path := strings.TrimPrefix(u.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid git URL path: %s", repoURL)
	}

	repoName := parts[len(parts)-1]
	// Remove .git suffix if present
	repoName = strings.TrimSuffix(repoName, ".git")

	if repoName == "" {
		return "", fmt.Errorf("empty repository name from URL: %s", repoURL)
	}

	// Replace invalid characters with underscores
	var result []rune
	for _, char := range repoName {
		if repoIDCharactersRegex.MatchString(string(char)) {
			result = append(result, char)
		} else {
			result = append(result, '_')
		}
	}

	return string(result), nil
}
