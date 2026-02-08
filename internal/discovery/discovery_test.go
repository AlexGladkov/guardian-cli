package discovery

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindAgreementsDirFrom_InCurrentDir(t *testing.T) {
	dir := t.TempDir()

	agreementsDir := filepath.Join(dir, ".agreements")
	err := os.Mkdir(agreementsDir, 0755)
	require.NoError(t, err)

	found, err := FindAgreementsDirFrom(dir)
	require.NoError(t, err)
	assert.Equal(t, agreementsDir, found)
}

func TestFindAgreementsDirFrom_InParentDir(t *testing.T) {
	// Create: root/.agreements/ and root/subdir/
	root := t.TempDir()

	agreementsDir := filepath.Join(root, ".agreements")
	err := os.Mkdir(agreementsDir, 0755)
	require.NoError(t, err)

	subdir := filepath.Join(root, "subdir")
	err = os.Mkdir(subdir, 0755)
	require.NoError(t, err)

	found, err := FindAgreementsDirFrom(subdir)
	require.NoError(t, err)
	assert.Equal(t, agreementsDir, found)
}

func TestFindAgreementsDirFrom_InGrandparentDir(t *testing.T) {
	// Create: root/.agreements/ and root/sub1/sub2/
	root := t.TempDir()

	agreementsDir := filepath.Join(root, ".agreements")
	err := os.Mkdir(agreementsDir, 0755)
	require.NoError(t, err)

	deepDir := filepath.Join(root, "sub1", "sub2")
	err = os.MkdirAll(deepDir, 0755)
	require.NoError(t, err)

	found, err := FindAgreementsDirFrom(deepDir)
	require.NoError(t, err)
	assert.Equal(t, agreementsDir, found)
}

func TestFindAgreementsDirFrom_DeeplyNested(t *testing.T) {
	// Create: root/.agreements/ and root/a/b/c/d/e/
	root := t.TempDir()

	agreementsDir := filepath.Join(root, ".agreements")
	err := os.Mkdir(agreementsDir, 0755)
	require.NoError(t, err)

	deepDir := filepath.Join(root, "a", "b", "c", "d", "e")
	err = os.MkdirAll(deepDir, 0755)
	require.NoError(t, err)

	found, err := FindAgreementsDirFrom(deepDir)
	require.NoError(t, err)
	assert.Equal(t, agreementsDir, found)
}

func TestFindAgreementsDirFrom_NotFound(t *testing.T) {
	// Use a temp directory with no .agreements anywhere
	dir := t.TempDir()

	// Create a nested dir without .agreements
	nested := filepath.Join(dir, "some", "nested", "path")
	err := os.MkdirAll(nested, 0755)
	require.NoError(t, err)

	_, err = FindAgreementsDirFrom(nested)
	require.Error(t, err)
	assert.Contains(t, err.Error(), ".agreements directory not found")
}

func TestFindAgreementsDirFrom_FileNotDirectory(t *testing.T) {
	// Create a file named .agreements (not a directory)
	root := t.TempDir()

	agreementsFile := filepath.Join(root, ".agreements")
	err := os.WriteFile(agreementsFile, []byte("not a dir"), 0644)
	require.NoError(t, err)

	_, err = FindAgreementsDirFrom(root)
	require.Error(t, err)
	assert.Contains(t, err.Error(), ".agreements directory not found")
}

func TestFindAgreementsDirFrom_ClosestWins(t *testing.T) {
	// Create two .agreements dirs at different levels
	root := t.TempDir()

	// Parent .agreements
	parentAgreements := filepath.Join(root, ".agreements")
	err := os.Mkdir(parentAgreements, 0755)
	require.NoError(t, err)

	// Child .agreements (closer to start)
	childDir := filepath.Join(root, "child")
	err = os.Mkdir(childDir, 0755)
	require.NoError(t, err)

	childAgreements := filepath.Join(childDir, ".agreements")
	err = os.Mkdir(childAgreements, 0755)
	require.NoError(t, err)

	// Search from child dir should find child's .agreements
	found, err := FindAgreementsDirFrom(childDir)
	require.NoError(t, err)
	assert.Equal(t, childAgreements, found)
}

func TestFindAgreementsDirFrom_RelativePath(t *testing.T) {
	// The function should handle relative paths by converting to absolute
	root := t.TempDir()

	// Resolve symlinks so that macOS /var -> /private/var is handled
	root, err := filepath.EvalSymlinks(root)
	require.NoError(t, err)

	agreementsDir := filepath.Join(root, ".agreements")
	err = os.Mkdir(agreementsDir, 0755)
	require.NoError(t, err)

	// Change CWD temporarily for this test
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(origDir)
		require.NoError(t, err)
	}()

	err = os.Chdir(root)
	require.NoError(t, err)

	found, err := FindAgreementsDirFrom(".")
	require.NoError(t, err)
	assert.Equal(t, agreementsDir, found)
}

func TestFindAgreementsDir_UsesCWD(t *testing.T) {
	root := t.TempDir()

	// Resolve symlinks so that macOS /var -> /private/var is handled
	root, err := filepath.EvalSymlinks(root)
	require.NoError(t, err)

	agreementsDir := filepath.Join(root, ".agreements")
	err = os.Mkdir(agreementsDir, 0755)
	require.NoError(t, err)

	// Change CWD temporarily for this test
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(origDir)
		require.NoError(t, err)
	}()

	err = os.Chdir(root)
	require.NoError(t, err)

	found, err := FindAgreementsDir()
	require.NoError(t, err)
	assert.Equal(t, agreementsDir, found)
}

func TestFindAgreementsDirFrom_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	_, err := FindAgreementsDirFrom(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), ".agreements directory not found")
}

func TestFindAgreementsDirFrom_SymlinkToAgreements(t *testing.T) {
	// Create actual .agreements dir and a symlink
	root := t.TempDir()

	realDir := filepath.Join(root, "real_agreements")
	err := os.Mkdir(realDir, 0755)
	require.NoError(t, err)

	symlinkDir := filepath.Join(root, ".agreements")
	err = os.Symlink(realDir, symlinkDir)
	if err != nil {
		t.Skip("symlinks not supported on this platform")
	}

	found, err := FindAgreementsDirFrom(root)
	require.NoError(t, err)
	assert.Equal(t, symlinkDir, found)
}
