package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type GitHubCloner struct {
	token string
}

func NewGitHubCloner(token string) *GitHubCloner {
	if token == "" {
		fmt.Printf("GITHUB_TOKEN environment variable is not set")
		os.Exit(1)
	}
	if !isGitAvailable() {
		fmt.Printf("git is not installed or not available in PATH")
		os.Exit(1)
	}
	return &GitHubCloner{token: token}
}

func (gc *GitHubCloner) Clone(owner, repo, destination string) error {
	if owner == "" || repo == "" || destination == "" {
		return fmt.Errorf("owner, repo, and destination are required")
	}

	repoPath := filepath.Join(destination, repo)
	cloneURL := fmt.Sprintf("https://%s@github.com/%s/%s.git", gc.token, owner, repo)

	// Remove existing directory
	if _, err := os.Stat(repoPath); !os.IsNotExist(err) {
		if err := os.RemoveAll(repoPath); err != nil {
			return fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	// Create destination directory
	if err := os.MkdirAll(destination, 0755); err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}

	// Clone repository
	cmd := exec.Command("git", "clone", cloneURL, repoPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	return nil
}

func isGitAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}
