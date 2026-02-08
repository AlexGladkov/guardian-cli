package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Vote represents a voter's decision on a proposal.
type Vote struct {
	ProposalID string    `yaml:"proposal_id"`
	VoterEmail string    `yaml:"voter_email"`
	Decision   string    `yaml:"decision"` // yes|no
	Comment    string    `yaml:"comment"`
	VotedAt    time.Time `yaml:"voted_at"`
}

// LoadVote reads and parses a vote YAML file from the given path.
func LoadVote(path string) (*Vote, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading vote file %s: %w", path, err)
	}

	var v Vote
	if err := yaml.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("parsing vote file %s: %w", path, err)
	}

	return &v, nil
}

// SaveVote writes a Vote to the given path as YAML.
func SaveVote(path string, v *Vote) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshaling vote: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating vote directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing vote file %s: %w", path, err)
	}

	return nil
}

// LoadVotesForProposal reads all vote files from a proposal-specific
// subdirectory within the votes directory.
// It expects the layout: <dir>/<proposalID>/<voter_email>.yml
func LoadVotesForProposal(dir string, proposalID string) ([]*Vote, error) {
	proposalVotesDir := filepath.Join(dir, proposalID)

	entries, err := os.ReadDir(proposalVotesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading votes directory %s: %w", proposalVotesDir, err)
	}

	var votes []*Vote
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".yml" && ext != ".yaml" {
			continue
		}

		v, err := LoadVote(filepath.Join(proposalVotesDir, name))
		if err != nil {
			return nil, fmt.Errorf("loading vote %s: %w", name, err)
		}
		votes = append(votes, v)
	}

	return votes, nil
}
