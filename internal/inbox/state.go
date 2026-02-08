package inbox

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// guardianDir is the local state directory name.
const guardianDir = ".guardian"

// stateFileName is the state file name within the guardian directory.
const stateFileName = "state.json"

// State represents the local guardian state persisted in .guardian/state.json.
type State struct {
	LastInboxCheck time.Time `json:"last_inbox_check"`
}

// LoadState reads .guardian/state.json from the repo root.
// If the file does not exist, it returns a zero-valued State (no error).
func LoadState(repoRoot string) (*State, error) {
	statePath := filepath.Join(repoRoot, guardianDir, stateFileName)

	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{}, nil
		}
		return nil, fmt.Errorf("reading state file %s: %w", statePath, err)
	}

	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing state file %s: %w", statePath, err)
	}

	return &s, nil
}

// SaveState writes .guardian/state.json to the repo root.
// It creates the .guardian/ directory if it does not exist.
func SaveState(repoRoot string, s *State) error {
	dir := filepath.Join(repoRoot, guardianDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating guardian directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	statePath := filepath.Join(dir, stateFileName)
	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return fmt.Errorf("writing state file %s: %w", statePath, err)
	}

	return nil
}
