package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadException_FullFile(t *testing.T) {
	content := `
id: "exc-2024-01-20-domain_no_infra"
rule_id: domain_no_infra
paths:
  - "domain/legacy/old_adapter.kt"
  - "domain/legacy/another.kt"
reason: "Legacy code, will be migrated in Q2"
created_by: ivan@company.com
created_at: "2024-01-20T09:00:00Z"
expires_at: "2024-06-30T00:00:00Z"
`
	path := writeTestFile(t, "exception.yml", content)

	e, err := LoadException(path)
	require.NoError(t, err)
	require.NotNil(t, e)

	assert.Equal(t, "exc-2024-01-20-domain_no_infra", e.ID)
	assert.Equal(t, "domain_no_infra", e.RuleID)
	assert.Equal(t, []string{"domain/legacy/old_adapter.kt", "domain/legacy/another.kt"}, e.Paths)
	assert.Equal(t, "Legacy code, will be migrated in Q2", e.Reason)
	assert.Equal(t, "ivan@company.com", e.CreatedBy)
	assert.False(t, e.CreatedAt.IsZero())
	require.NotNil(t, e.ExpiresAt)
	assert.False(t, e.ExpiresAt.IsZero())
}

func TestLoadException_PermanentNoExpiry(t *testing.T) {
	content := `
id: "exc-permanent"
rule_id: some_rule
paths:
  - "path/to/file.go"
reason: "Permanent exception"
created_by: admin@test.com
created_at: "2024-01-01T00:00:00Z"
`
	path := writeTestFile(t, "exception.yml", content)

	e, err := LoadException(path)
	require.NoError(t, err)
	require.NotNil(t, e)
	assert.Nil(t, e.ExpiresAt)
}

func TestLoadException_FileNotFound(t *testing.T) {
	_, err := LoadException("/nonexistent/exception.yml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading exception file")
}

func TestLoadException_InvalidYAML(t *testing.T) {
	path := writeTestFile(t, "bad.yml", "{{invalid yaml")
	_, err := LoadException(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing exception file")
}

func TestSaveException(t *testing.T) {
	dir := t.TempDir()
	exceptionsDir := filepath.Join(dir, "exceptions")
	path := filepath.Join(exceptionsDir, "exc-test.yml")

	now := time.Now().UTC().Truncate(time.Second)
	expiresAt := now.Add(90 * 24 * time.Hour)

	e := &Exception{
		ID:        "exc-test",
		RuleID:    "test_rule",
		Paths:     []string{"src/legacy/*.go"},
		Reason:    "Legacy code migration in progress",
		CreatedBy: "dev@test.com",
		CreatedAt: now,
		ExpiresAt: &expiresAt,
	}

	err := SaveException(path, e)
	require.NoError(t, err)

	// Verify directory was created
	_, err = os.Stat(exceptionsDir)
	require.NoError(t, err)

	// Round-trip
	loaded, err := LoadException(path)
	require.NoError(t, err)
	assert.Equal(t, e.ID, loaded.ID)
	assert.Equal(t, e.RuleID, loaded.RuleID)
	assert.Equal(t, e.Paths, loaded.Paths)
	assert.Equal(t, e.Reason, loaded.Reason)
	assert.Equal(t, e.CreatedBy, loaded.CreatedBy)
	assert.True(t, e.CreatedAt.Equal(loaded.CreatedAt))
	require.NotNil(t, loaded.ExpiresAt)
	assert.True(t, e.ExpiresAt.Equal(*loaded.ExpiresAt))
}

func TestSaveException_NilExpiresAt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "exc-perm.yml")

	now := time.Now().UTC().Truncate(time.Second)
	e := &Exception{
		ID:        "exc-perm",
		RuleID:    "perm_rule",
		Paths:     []string{"path/to/file"},
		Reason:    "Permanent exception",
		CreatedBy: "admin@test.com",
		CreatedAt: now,
		ExpiresAt: nil,
	}

	err := SaveException(path, e)
	require.NoError(t, err)

	loaded, err := LoadException(path)
	require.NoError(t, err)
	assert.Nil(t, loaded.ExpiresAt)
}

func TestLoadAllExceptions(t *testing.T) {
	dir := t.TempDir()

	now := time.Now().UTC().Truncate(time.Second)

	exceptions := []*Exception{
		{
			ID: "exc-1", RuleID: "rule1", Paths: []string{"path1"},
			Reason: "R1", CreatedBy: "a@test.com", CreatedAt: now,
		},
		{
			ID: "exc-2", RuleID: "rule2", Paths: []string{"path2"},
			Reason: "R2", CreatedBy: "b@test.com", CreatedAt: now,
		},
	}

	for _, e := range exceptions {
		err := SaveException(filepath.Join(dir, e.ID+".yml"), e)
		require.NoError(t, err)
	}

	// Non-YAML file
	err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("skip"), 0644)
	require.NoError(t, err)

	loaded, err := LoadAllExceptions(dir)
	require.NoError(t, err)
	assert.Len(t, loaded, 2)

	ids := make(map[string]bool)
	for _, e := range loaded {
		ids[e.ID] = true
	}
	assert.True(t, ids["exc-1"])
	assert.True(t, ids["exc-2"])
}

func TestLoadAllExceptions_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	loaded, err := LoadAllExceptions(dir)
	require.NoError(t, err)
	assert.Empty(t, loaded)
}

func TestLoadAllExceptions_NonexistentDir(t *testing.T) {
	loaded, err := LoadAllExceptions("/nonexistent/exceptions")
	require.NoError(t, err)
	assert.Nil(t, loaded)
}

func TestLoadAllExceptions_SkipsSubdirectories(t *testing.T) {
	dir := t.TempDir()

	now := time.Now().UTC().Truncate(time.Second)
	e := &Exception{
		ID: "exc-1", RuleID: "rule1", Paths: []string{"path1"},
		Reason: "R1", CreatedBy: "a@test.com", CreatedAt: now,
	}
	err := SaveException(filepath.Join(dir, "exc-1.yml"), e)
	require.NoError(t, err)

	err = os.MkdirAll(filepath.Join(dir, "subdir"), 0755)
	require.NoError(t, err)

	loaded, err := LoadAllExceptions(dir)
	require.NoError(t, err)
	assert.Len(t, loaded, 1)
}

func TestLoadAllExceptions_BothYmlAndYaml(t *testing.T) {
	dir := t.TempDir()

	now := time.Now().UTC().Truncate(time.Second)

	e1 := &Exception{
		ID: "exc-1", RuleID: "rule1", Paths: []string{"path1"},
		Reason: "R1", CreatedBy: "a@test.com", CreatedAt: now,
	}
	err := SaveException(filepath.Join(dir, "exc-1.yml"), e1)
	require.NoError(t, err)

	e2 := &Exception{
		ID: "exc-2", RuleID: "rule2", Paths: []string{"path2"},
		Reason: "R2", CreatedBy: "b@test.com", CreatedAt: now,
	}
	err = SaveException(filepath.Join(dir, "exc-2.yaml"), e2)
	require.NoError(t, err)

	loaded, err := LoadAllExceptions(dir)
	require.NoError(t, err)
	assert.Len(t, loaded, 2)
}

func TestLoadException_MultiplePaths(t *testing.T) {
	content := `
id: "exc-multi"
rule_id: multi_rule
paths:
  - "src/legacy/**"
  - "src/deprecated/old_file.go"
  - "vendor/"
reason: "Multiple legacy paths"
created_by: dev@test.com
created_at: "2024-03-01T00:00:00Z"
`
	path := writeTestFile(t, "exception.yml", content)

	e, err := LoadException(path)
	require.NoError(t, err)
	assert.Len(t, e.Paths, 3)
	assert.Equal(t, "src/legacy/**", e.Paths[0])
	assert.Equal(t, "src/deprecated/old_file.go", e.Paths[1])
	assert.Equal(t, "vendor/", e.Paths[2])
}
