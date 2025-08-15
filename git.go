package main

import (
	"fmt"
	"log/slog"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
)

func cloneRepository(repoConfig RepoConfig) (*git.Repository, error) {
	repoID, _ := getRepoID(repoConfig)
	slog.Info("Cloning repository", "repo", repoConfig.Repo, "id", repoID, "ref", repoConfig.Ref)

	cloneOptions := &git.CloneOptions{
		URL:      repoConfig.Repo,
		Progress: nil, // Suppress progress output
	}

	// Set reference if specified
	if repoConfig.Ref != "" {
		cloneOptions.ReferenceName = plumbing.ReferenceName(repoConfig.Ref)
	}

	// Configure authentication if provided
	if repoConfig.Auth != nil {
		slog.Info("Authenticating with repository", "username", repoConfig.Auth.Username)
		cloneOptions.Auth = &http.BasicAuth{
			Username: repoConfig.Auth.Username,
			Password: repoConfig.Auth.Password,
		}
	}

	repo, err := git.Clone(memory.NewStorage(), nil, cloneOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	slog.Debug("Repository cloned successfully", "repo", repoID)
	return repo, nil
}

func getRepositoryFiles(repo *git.Repository) (map[string]*File, error) {
	// Use HEAD (which is already at the correct ref from clone)
	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository head: %w", err)
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit object: %w", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}

	files := make(map[string]*File)
	err = tree.Files().ForEach(func(gitFile *object.File) error {
		files[gitFile.Name] = &File{
			Name:    gitFile.Name,
			Size:    gitFile.Size,
			gitFile: gitFile,
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate files: %w", err)
	}

	slog.Debug("Retrieved repository files", "count", len(files))
	return files, nil
}
