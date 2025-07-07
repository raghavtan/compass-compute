package services

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

type GitHubCloner struct {
	token string
}

type GitInfo struct {
	Owner    string
	Repo     string
	Path     string
	Branch   string
	IsSSH    bool
	IsGitURL bool
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

// SetupMetricDirectory handles METRIC_DIR environment variable
// Returns true if catalog repo should be skipped, false otherwise
func (gc *GitHubCloner) SetupMetricDirectory(verbose bool) (bool, error) {
	metricDir := os.Getenv("METRIC_DIR")
	if metricDir == "" {
		if verbose {
			fmt.Println("METRIC_DIR not set, using default catalog repository")
		}
		return false, nil // Don't skip catalog repo
	}

	targetPath := filepath.Join(LocalBasePath, "metrics")

	if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
		if info, err := os.Stat(targetPath); err == nil && info.IsDir() {
			entries, err := os.ReadDir(targetPath)
			if err != nil {
				return false, fmt.Errorf("failed to read target directory: %w", err)
			}
			if len(entries) > 0 {
				if verbose {
					fmt.Printf("Target directory '%s' already exists and is not empty, skipping metric directory setup\n", targetPath)
				}
				return true, nil // Skip catalog repo
			}
		} else {
			return false, fmt.Errorf("target path '%s' exists but is not a directory", targetPath)
		}
	}

	// Check if it's a local path
	if isLocalPath(metricDir) {
		if verbose {
			fmt.Printf("Using local metric directory: %s\n", metricDir)
		}
		return gc.copyLocalDirectory(metricDir, targetPath, verbose)
	}

	// Check if it's a git URL
	gitInfo, err := parseGitURL(metricDir)
	if err == nil {
		if verbose {
			fmt.Printf("Using git metric directory: %s\n", metricDir)
		}
		return gc.cloneAndExtractPath(gitInfo, targetPath, verbose)
	}

	return false, fmt.Errorf("invalid METRIC_DIR format: %s", metricDir)
}

func isLocalPath(path string) bool {
	// Check if path exists locally
	_, err := os.Stat(path)
	return err == nil
}

func (gc *GitHubCloner) copyLocalDirectory(src, dst string, verbose bool) (bool, error) {
	if verbose {
		fmt.Printf("Copying local directory from %s to %s\n", src, dst)
	}

	// Create destination directory
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return false, fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Copy directory
	if err := copyDir(src, dst); err != nil {
		return false, fmt.Errorf("failed to copy directory: %w", err)
	}

	return true, nil // Skip catalog repo
}

func (gc *GitHubCloner) cloneAndExtractPath(gitInfo *GitInfo, targetPath string, verbose bool) (bool, error) {
	tempDir := filepath.Join(LocalBasePath, "temp-"+gitInfo.Repo)

	// Remove temp directory if it exists
	if err := os.RemoveAll(tempDir); err != nil {
		return false, fmt.Errorf("failed to remove temp directory: %w", err)
	}

	// Clone the repository
	var cloneURL string
	if gitInfo.IsSSH {
		cloneURL = fmt.Sprintf("git@github.com:%s/%s.git", gitInfo.Owner, gitInfo.Repo)
	} else {
		cloneURL = fmt.Sprintf("https://%s@github.com/%s/%s.git", gc.token, gitInfo.Owner, gitInfo.Repo)
	}

	if verbose {
		fmt.Printf("Cloning repository: %s\n", cloneURL)
	}

	cmd := exec.Command("git", "clone", cloneURL, tempDir)
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("git clone failed: %w", err)
	}

	// Extract the specific path
	sourcePath := filepath.Join(tempDir, gitInfo.Path)
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return false, fmt.Errorf("path '%s' not found in repository", gitInfo.Path)
	}

	if verbose {
		fmt.Printf("Extracting path %s to %s\n", sourcePath, targetPath)
	}

	// Create target directory
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return false, fmt.Errorf("failed to create target directory: %w", err)
	}

	// Copy the extracted path
	if err := copyDir(sourcePath, targetPath); err != nil {
		return false, fmt.Errorf("failed to copy extracted path: %w", err)
	}

	// Clean up temp directory
	if err := os.RemoveAll(tempDir); err != nil {
		if verbose {
			fmt.Printf("Warning: failed to remove temp directory: %v\n", err)
		}
	}

	return true, nil // Skip catalog repo
}

func parseGitURL(gitURL string) (*GitInfo, error) {
	gitInfo := &GitInfo{Branch: "main"} // Default branch

	// GitHub tree URL: https://github.com/owner/repo/tree/branch/path/to/dir
	treeRegex := regexp.MustCompile(`^https://github\.com/([^/]+)/([^/]+)/tree/([^/]+)/(.+)$`)
	if matches := treeRegex.FindStringSubmatch(gitURL); matches != nil {
		gitInfo.Owner = matches[1]
		gitInfo.Repo = matches[2]
		gitInfo.Branch = matches[3]
		gitInfo.Path = matches[4]
		gitInfo.IsGitURL = true
		return gitInfo, nil
	}

	// Git HTTPS URL with path: https://github.com/owner/repo.git/path/to/dir
	httpsRegex := regexp.MustCompile(`^https://github\.com/([^/]+)/([^/]+)\.git/(.+)$`)
	if matches := httpsRegex.FindStringSubmatch(gitURL); matches != nil {
		gitInfo.Owner = matches[1]
		gitInfo.Repo = matches[2]
		gitInfo.Path = matches[3]
		gitInfo.IsGitURL = true
		return gitInfo, nil
	}

	// Git SSH URL with path: git@github.com:owner/repo.git/path/to/dir
	sshRegex := regexp.MustCompile(`^git@github\.com:([^/]+)/([^/]+)\.git/(.+)$`)
	if matches := sshRegex.FindStringSubmatch(gitURL); matches != nil {
		gitInfo.Owner = matches[1]
		gitInfo.Repo = matches[2]
		gitInfo.Path = matches[3]
		gitInfo.IsSSH = true
		gitInfo.IsGitURL = true
		return gitInfo, nil
	}

	return nil, fmt.Errorf("unsupported git URL format")
}

func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(srcFile *os.File) {
		err := srcFile.Close()
		if err != nil {
			fmt.Printf("Warning: failed to close source file %s: %v\n", src, err)
		}
	}(srcFile)

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(dstFile *os.File) {
		err := dstFile.Close()
		if err != nil {
			fmt.Printf("Warning: failed to close destination file %s: %v\n", dst, err)
		}
	}(dstFile)

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func isGitAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}
