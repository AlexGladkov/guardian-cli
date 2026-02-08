package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Proposal represents a proposal to modify, add, or remove a rule.
type Proposal struct {
	ID           string         `yaml:"id"`
	RuleID       string         `yaml:"rule_id"`
	ProposalType string         `yaml:"proposal_type"` // modify|add|remove
	Change       ProposalChange `yaml:"change"`
	Reason       string         `yaml:"reason"`
	Impact       string         `yaml:"impact"`
	CreatedBy    string         `yaml:"created_by"`
	CreatedAt    time.Time      `yaml:"created_at"`
	Status       string         `yaml:"status"` // proposed|accepted|rejected|withdrawn|expired
}

// ProposalChange describes the proposed change.
type ProposalChange struct {
	Description string `yaml:"description"`
	Details     string `yaml:"details"`
}

// LoadProposal reads and parses a proposal YAML file from the given path.
func LoadProposal(path string) (*Proposal, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading proposal file %s: %w", path, err)
	}

	var p Proposal
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parsing proposal file %s: %w", path, err)
	}

	return &p, nil
}

// SaveProposal writes a Proposal to the given path as YAML.
func SaveProposal(path string, p *Proposal) error {
	data, err := yaml.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshaling proposal: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating proposal directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing proposal file %s: %w", path, err)
	}

	return nil
}

// LoadAllProposals reads all YAML proposal files from the given directory.
// It returns only files with .yml or .yaml extensions.
func LoadAllProposals(dir string) ([]*Proposal, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading proposals directory %s: %w", dir, err)
	}

	var proposals []*Proposal
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".yml" && ext != ".yaml" {
			continue
		}

		p, err := LoadProposal(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("loading proposal %s: %w", name, err)
		}
		proposals = append(proposals, p)
	}

	return proposals, nil
}
