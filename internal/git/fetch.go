package git

import (
	"fmt"
	"os/exec"
)

// Fetch runs git fetch to update remote-tracking branches.
func Fetch() error {
	cmd := exec.Command("git", "fetch")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("running git fetch: %s: %w", string(out), err)
	}
	return nil
}
