// Package git provides utilities for interacting with the git command-line tool.
// It wraps common git operations such as diff, identity lookup, hook management,
// fetch, and CI environment detection.
package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// DiffResult holds the result of a git diff operation.
type DiffResult struct {
	ChangedFiles []string
	DiffContent  string
}

// GetDiff runs git diff for the given range and returns changed files along with
// the full unified diff content. The diffRange should be in the form "base..head".
func GetDiff(diffRange string) (*DiffResult, error) {
	files, err := GetDiffNameOnly(diffRange)
	if err != nil {
		return nil, fmt.Errorf("getting changed files: %w", err)
	}

	cmd := exec.Command("git", "diff", diffRange)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("running git diff %s: %w", diffRange, err)
	}

	return &DiffResult{
		ChangedFiles: files,
		DiffContent:  string(out),
	}, nil
}

// GetDiffNameOnly runs git diff --name-only for the given range and returns
// the list of changed file paths.
func GetDiffNameOnly(diffRange string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", diffRange)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("running git diff --name-only %s: %w", diffRange, err)
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}

	lines := strings.Split(raw, "\n")
	files := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			files = append(files, trimmed)
		}
	}

	return files, nil
}
