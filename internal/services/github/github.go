package github

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type GitHubCloner struct {
	token string
}

type CloneOptions struct {
	Branch      string // Specific branch to clone (optional)
	Depth       int    // Shallow clone depth (0 = full clone)
	Destination string // Destination directory (required)
	UseSSH      bool   // Use SSH URL instead of HTTPS
	Recursive   bool   // Clone submodules recursively
}

// NewGitHubCloner creates a new GitHub cloner
func NewGitHubCloner(token string) *GitHubCloner {
	if token == "" {
		fmt.Printf("GITHUB_TOKEN environment variable is not set")
		os.Exit(1)
	}
	if !IsGitAvailable() {
		fmt.Printf("git is not installed or not available in PATH")
		os.Exit(1)
	}
	return &GitHubCloner{
		token: token,
	}
}

// buildCloneURL constructs the appropriate clone URL
func (gc *GitHubCloner) buildCloneURL(owner, repo string, useSSH bool) string {
	if useSSH {
		return fmt.Sprintf("git@github.com:%s/%s.git", owner, repo)
	}

	if gc.token != "" {
		return fmt.Sprintf("https://%s@github.com/%s/%s.git", gc.token, owner, repo)
	}

	return fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
}

// Clone clones a GitHub repository
func (gc *GitHubCloner) Clone(owner, repo string, opts *CloneOptions) error {
	if opts == nil {
		opts = &CloneOptions{}
	}

	// Validate required parameters
	if owner == "" || repo == "" {
		return fmt.Errorf("owner and repo are required")
	}

	if opts.Destination == "" {
		return fmt.Errorf("destination directory is required")
	}

	// Build clone URL
	cloneURL := gc.buildCloneURL(owner, repo, opts.UseSSH)

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(opts.Destination, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Build git clone command
	args := []string{"clone"}

	// Add branch option
	if opts.Branch != "" {
		args = append(args, "--branch", opts.Branch)
	}

	// Add depth option for shallow clone
	if opts.Depth > 0 {
		args = append(args, "--depth", fmt.Sprintf("%d", opts.Depth))
	}

	// Add recursive option for submodules
	if opts.Recursive {
		args = append(args, "--recursive")
	}

	// Add URL and destination
	repoPath := filepath.Join(opts.Destination, repo)
	args = append(args, cloneURL, repoPath)

	if _, err := os.Stat(repoPath); !os.IsNotExist(err) {
		if err := os.RemoveAll(repoPath); err != nil {
			return fmt.Errorf("failed to remove existing repository directory: %w", err)
		}
	}

	// Execute git clone
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	fmt.Printf("Successfully cloned %s/%s to %s\n", owner, repo, repoPath)
	return nil
}

// QuickClone is a convenience method for simple cloning
func (gc *GitHubCloner) QuickClone(repoURL, destination string) error {
	// Parse repo URL to extract owner/repo
	repoURL = strings.TrimSuffix(repoURL, ".git")
	parts := strings.Split(repoURL, "/")

	if len(parts) < 2 {
		return fmt.Errorf("invalid repository URL format %s", repoURL)
	}

	owner := parts[len(parts)-2]
	repo := parts[len(parts)-1]

	return gc.Clone(owner, repo, &CloneOptions{
		Destination: destination,
	})
}

// Helper function to check if git is available
func IsGitAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}
