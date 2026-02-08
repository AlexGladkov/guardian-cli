package inbox

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadState_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()

	state, err := LoadState(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, state)
	assert.True(t, state.LastInboxCheck.IsZero(), "last_inbox_check should be zero for non-existent state")
}

func TestSaveState_CreatesDirAndFile(t *testing.T) {
	tmpDir := t.TempDir()
	now := time.Now().Truncate(time.Second)

	state := &State{
		LastInboxCheck: now,
	}

	err := SaveState(tmpDir, state)
	require.NoError(t, err)

	// Verify directory was created
	guardianPath := filepath.Join(tmpDir, ".guardian")
	info, err := os.Stat(guardianPath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Verify file was created
	statePath := filepath.Join(guardianPath, "state.json")
	_, err = os.Stat(statePath)
	require.NoError(t, err)
}

func TestSaveState_LoadState_Roundtrip(t *testing.T) {
	tmpDir := t.TempDir()
	now := time.Now().Truncate(time.Second)

	original := &State{
		LastInboxCheck: now,
	}

	// Save
	err := SaveState(tmpDir, original)
	require.NoError(t, err)

	// Load
	loaded, err := LoadState(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, loaded)

	// Compare (truncate to second to avoid nanosecond differences in JSON)
	assert.Equal(t, original.LastInboxCheck.Unix(), loaded.LastInboxCheck.Unix())
}

func TestSaveState_OverwritesExisting(t *testing.T) {
	tmpDir := t.TempDir()

	// Save first state
	first := &State{
		LastInboxCheck: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	err := SaveState(tmpDir, first)
	require.NoError(t, err)

	// Save second state
	second := &State{
		LastInboxCheck: time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC),
	}
	err = SaveState(tmpDir, second)
	require.NoError(t, err)

	// Load should return second state
	loaded, err := LoadState(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, second.LastInboxCheck.Unix(), loaded.LastInboxCheck.Unix())
}

func TestSaveState_ExistingDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Pre-create the .guardian directory
	guardianPath := filepath.Join(tmpDir, ".guardian")
	err := os.MkdirAll(guardianPath, 0755)
	require.NoError(t, err)

	state := &State{
		LastInboxCheck: time.Now().Truncate(time.Second),
	}

	// Should not error when directory already exists
	err = SaveState(tmpDir, state)
	require.NoError(t, err)

	// Verify roundtrip
	loaded, err := LoadState(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, state.LastInboxCheck.Unix(), loaded.LastInboxCheck.Unix())
}

func TestLoadState_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	guardianPath := filepath.Join(tmpDir, ".guardian")
	err := os.MkdirAll(guardianPath, 0755)
	require.NoError(t, err)

	// Write invalid JSON
	statePath := filepath.Join(guardianPath, "state.json")
	err = os.WriteFile(statePath, []byte(`{invalid json}`), 0644)
	require.NoError(t, err)

	state, err := LoadState(tmpDir)
	require.Error(t, err)
	assert.Nil(t, state)
	assert.Contains(t, err.Error(), "parsing state file")
}

func TestLoadState_ZeroState(t *testing.T) {
	tmpDir := t.TempDir()

	state, err := LoadState(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, state)

	// Zero state should have zero time
	assert.True(t, state.LastInboxCheck.IsZero())
}

func TestSaveState_ZeroTime(t *testing.T) {
	tmpDir := t.TempDir()

	// Save state with zero time
	state := &State{}
	err := SaveState(tmpDir, state)
	require.NoError(t, err)

	// Load it back
	loaded, err := LoadState(tmpDir)
	require.NoError(t, err)
	assert.True(t, loaded.LastInboxCheck.IsZero())
}

func TestStateJSON_Format(t *testing.T) {
	tmpDir := t.TempDir()
	ts := time.Date(2024, 1, 16, 14, 0, 0, 0, time.UTC)

	state := &State{
		LastInboxCheck: ts,
	}
	err := SaveState(tmpDir, state)
	require.NoError(t, err)

	// Read the raw JSON to verify format
	data, err := os.ReadFile(filepath.Join(tmpDir, ".guardian", "state.json"))
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "last_inbox_check")
	assert.Contains(t, content, "2024-01-16")
}
