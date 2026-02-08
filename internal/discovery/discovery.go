// Package discovery provides functionality to locate the .agreements/
// directory by traversing upward from a starting directory, similar to
// how git searches for the .git/ directory.
package discovery

import (
	"fmt"
	"os"
	"path/filepath"
)

const agreementsDirName = ".agreements"

// FindAgreementsDir searches for the .agreements/ directory starting from
// the current working directory and traversing upward toward the filesystem root.
func FindAgreementsDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting current working directory: %w", err)
	}
	return FindAgreementsDirFrom(cwd)
}

// FindAgreementsDirFrom searches for the .agreements/ directory starting from
// the given directory and traversing upward toward the filesystem root.
// It returns the absolute path to the .agreements/ directory if found.
func FindAgreementsDirFrom(startDir string) (string, error) {
	absStart, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("resolving absolute path for %s: %w", startDir, err)
	}

	current := absStart
	for {
		candidate := filepath.Join(current, agreementsDirName)
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return candidate, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			// Reached filesystem root without finding .agreements/
			return "", fmt.Errorf(".agreements directory not found (searched from %s to filesystem root)", absStart)
		}
		current = parent
	}
}
