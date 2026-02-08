package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// HookMarker is the comment marker used to identify guardian-managed hooks.
const HookMarker = "# GUARDIAN-MANAGED-HOOK"

// hookScript returns the content for a guardian-managed git hook.
func hookScript() string {
	return fmt.Sprintf(`#!/bin/sh
%s
nohup guardian inbox --notify --since-last-check --quiet &>/dev/null &
`, HookMarker)
}

// hookNames lists the git hooks that guardian manages.
var hookNames = []string{"post-merge", "post-checkout"}

// getHooksDir returns the path to the .git/hooks directory by looking for
// the .git directory starting from the current working directory.
func getHooksDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}

	gitDir := filepath.Join(cwd, ".git")
	info, err := os.Stat(gitDir)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("not a git repository (no .git directory found in %s)", cwd)
	}

	hooksDir := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return "", fmt.Errorf("creating hooks directory: %w", err)
	}

	return hooksDir, nil
}

// InstallHooks creates post-merge and post-checkout hooks that run
// guardian inbox in the background. Existing hooks that are not
// guardian-managed are left untouched and an error is returned.
func InstallHooks() error {
	hooksDir, err := getHooksDir()
	if err != nil {
		return err
	}

	script := hookScript()

	for _, name := range hookNames {
		hookPath := filepath.Join(hooksDir, name)

		// Check if a non-guardian hook already exists.
		if data, err := os.ReadFile(hookPath); err == nil {
			if !strings.Contains(string(data), HookMarker) {
				return fmt.Errorf("hook %s already exists and is not managed by guardian", name)
			}
		}

		if err := os.WriteFile(hookPath, []byte(script), 0755); err != nil {
			return fmt.Errorf("writing hook %s: %w", name, err)
		}
	}

	return nil
}

// UninstallHooks removes only guardian-managed hooks identified by the
// HookMarker comment. Non-guardian hooks are left untouched.
func UninstallHooks() error {
	hooksDir, err := getHooksDir()
	if err != nil {
		return err
	}

	for _, name := range hookNames {
		hookPath := filepath.Join(hooksDir, name)

		data, err := os.ReadFile(hookPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("reading hook %s: %w", name, err)
		}

		if strings.Contains(string(data), HookMarker) {
			if err := os.Remove(hookPath); err != nil {
				return fmt.Errorf("removing hook %s: %w", name, err)
			}
		}
	}

	return nil
}

// IsHookInstalled checks if a guardian-managed hook exists for the given hook name.
func IsHookInstalled(hookName string) (bool, error) {
	hooksDir, err := getHooksDir()
	if err != nil {
		return false, err
	}

	hookPath := filepath.Join(hooksDir, hookName)

	data, err := os.ReadFile(hookPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("reading hook %s: %w", hookName, err)
	}

	return strings.Contains(string(data), HookMarker), nil
}
