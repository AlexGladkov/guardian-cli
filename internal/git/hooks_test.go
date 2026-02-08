package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTempGitDir creates a temporary directory with a .git/hooks structure
// and changes the working directory to it. It returns a cleanup function
// that restores the original working directory.
func setupTempGitDir(t *testing.T) (string, func()) {
	t.Helper()

	origDir, err := os.Getwd()
	require.NoError(t, err)

	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	hooksDir := filepath.Join(gitDir, "hooks")
	require.NoError(t, os.MkdirAll(hooksDir, 0755))

	require.NoError(t, os.Chdir(tmpDir))

	return tmpDir, func() {
		os.Chdir(origDir)
	}
}

func TestInstallHooks_CreatesHookFiles(t *testing.T) {
	tmpDir, cleanup := setupTempGitDir(t)
	defer cleanup()

	err := InstallHooks()
	require.NoError(t, err)

	hooksDir := filepath.Join(tmpDir, ".git", "hooks")

	// Verify post-merge hook.
	postMerge, err := os.ReadFile(filepath.Join(hooksDir, "post-merge"))
	require.NoError(t, err)
	assert.Contains(t, string(postMerge), HookMarker)
	assert.Contains(t, string(postMerge), "guardian inbox")
	assert.Contains(t, string(postMerge), "#!/bin/sh")

	// Verify post-checkout hook.
	postCheckout, err := os.ReadFile(filepath.Join(hooksDir, "post-checkout"))
	require.NoError(t, err)
	assert.Contains(t, string(postCheckout), HookMarker)
	assert.Contains(t, string(postCheckout), "guardian inbox")

	// Verify hooks are executable.
	info, err := os.Stat(filepath.Join(hooksDir, "post-merge"))
	require.NoError(t, err)
	assert.NotZero(t, info.Mode()&0100, "post-merge should be executable")

	info, err = os.Stat(filepath.Join(hooksDir, "post-checkout"))
	require.NoError(t, err)
	assert.NotZero(t, info.Mode()&0100, "post-checkout should be executable")
}

func TestInstallHooks_IdempotentReinstall(t *testing.T) {
	_, cleanup := setupTempGitDir(t)
	defer cleanup()

	require.NoError(t, InstallHooks())
	// Installing again should not error since they are guardian-managed.
	require.NoError(t, InstallHooks())
}

func TestInstallHooks_ErrorOnExistingNonGuardianHook(t *testing.T) {
	tmpDir, cleanup := setupTempGitDir(t)
	defer cleanup()

	hooksDir := filepath.Join(tmpDir, ".git", "hooks")
	existingHook := filepath.Join(hooksDir, "post-merge")
	require.NoError(t, os.WriteFile(existingHook, []byte("#!/bin/sh\necho custom hook\n"), 0755))

	err := InstallHooks()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not managed by guardian")
}

func TestUninstallHooks_RemovesGuardianHooks(t *testing.T) {
	tmpDir, cleanup := setupTempGitDir(t)
	defer cleanup()

	require.NoError(t, InstallHooks())

	hooksDir := filepath.Join(tmpDir, ".git", "hooks")

	// Verify hooks exist.
	_, err := os.Stat(filepath.Join(hooksDir, "post-merge"))
	require.NoError(t, err)

	// Uninstall.
	require.NoError(t, UninstallHooks())

	// Verify hooks are gone.
	_, err = os.Stat(filepath.Join(hooksDir, "post-merge"))
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat(filepath.Join(hooksDir, "post-checkout"))
	assert.True(t, os.IsNotExist(err))
}

func TestUninstallHooks_LeavesNonGuardianHooks(t *testing.T) {
	tmpDir, cleanup := setupTempGitDir(t)
	defer cleanup()

	hooksDir := filepath.Join(tmpDir, ".git", "hooks")
	customContent := "#!/bin/sh\necho custom hook\n"
	require.NoError(t, os.WriteFile(filepath.Join(hooksDir, "post-merge"), []byte(customContent), 0755))

	require.NoError(t, UninstallHooks())

	// Custom hook should still be there.
	data, err := os.ReadFile(filepath.Join(hooksDir, "post-merge"))
	require.NoError(t, err)
	assert.Equal(t, customContent, string(data))
}

func TestUninstallHooks_NoErrorWhenNoHooksExist(t *testing.T) {
	_, cleanup := setupTempGitDir(t)
	defer cleanup()

	err := UninstallHooks()
	assert.NoError(t, err)
}

func TestIsHookInstalled_ReturnsTrueForGuardianHook(t *testing.T) {
	_, cleanup := setupTempGitDir(t)
	defer cleanup()

	require.NoError(t, InstallHooks())

	installed, err := IsHookInstalled("post-merge")
	require.NoError(t, err)
	assert.True(t, installed)

	installed, err = IsHookInstalled("post-checkout")
	require.NoError(t, err)
	assert.True(t, installed)
}

func TestIsHookInstalled_ReturnsFalseForMissingHook(t *testing.T) {
	_, cleanup := setupTempGitDir(t)
	defer cleanup()

	installed, err := IsHookInstalled("post-merge")
	require.NoError(t, err)
	assert.False(t, installed)
}

func TestIsHookInstalled_ReturnsFalseForNonGuardianHook(t *testing.T) {
	tmpDir, cleanup := setupTempGitDir(t)
	defer cleanup()

	hooksDir := filepath.Join(tmpDir, ".git", "hooks")
	require.NoError(t, os.WriteFile(
		filepath.Join(hooksDir, "post-merge"),
		[]byte("#!/bin/sh\necho non-guardian\n"),
		0755,
	))

	installed, err := IsHookInstalled("post-merge")
	require.NoError(t, err)
	assert.False(t, installed)
}

func TestInstallHooks_NoGitDirectory(t *testing.T) {
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)

	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))

	err = InstallHooks()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a git repository")
}
