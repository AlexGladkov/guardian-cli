package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetUserEmail returns the current git user.email by running git config user.email.
func GetUserEmail() (string, error) {
	cmd := exec.Command("git", "config", "user.email")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("running git config user.email: %w", err)
	}

	email := strings.TrimSpace(string(out))
	if email == "" {
		return "", fmt.Errorf("git user.email is not configured")
	}

	return email, nil
}
