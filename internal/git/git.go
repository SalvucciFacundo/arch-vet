package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetModifiedFiles returns a list of modified, staged, or untracked Go files within rootDir.
func GetModifiedFiles(rootDir string) ([]string, error) {
	absRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve rootDir: %w", err)
	}

	// 1. Find git root
	rootCmd := exec.Command("git", "rev-parse", "--show-toplevel")
	rootCmd.Dir = absRootDir
	var rootOut, rootErr bytes.Buffer
	rootCmd.Stdout = &rootOut
	rootCmd.Stderr = &rootErr
	if err := rootCmd.Run(); err != nil {
		return nil, fmt.Errorf("not a git repository: %s (%w)", rootErr.String(), err)
	}
	gitRepoRoot := strings.TrimSpace(rootOut.String())

	// 2. Run git status --porcelain -uall
	statusCmd := exec.Command("git", "status", "--porcelain", "-uall")
	statusCmd.Dir = gitRepoRoot
	var statusOut, statusErr bytes.Buffer
	statusCmd.Stdout = &statusOut
	statusCmd.Stderr = &statusErr
	if err := statusCmd.Run(); err != nil {
		return nil, fmt.Errorf("git status failed: %s (%w)", statusErr.String(), err)
	}

	var files []string
	lines := strings.Split(statusOut.String(), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) < 4 {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		gitRelPath := parts[len(parts)-1]
		if !strings.HasSuffix(gitRelPath, ".go") {
			continue
		}

		// Calculate path relative to rootDir
		fullPath := filepath.Join(gitRepoRoot, gitRelPath)
		relToTarget, err := filepath.Rel(absRootDir, fullPath)
		if err == nil && !strings.HasPrefix(relToTarget, "..") {
			files = append(files, filepath.ToSlash(relToTarget))
		}
	}

	return files, nil
}
