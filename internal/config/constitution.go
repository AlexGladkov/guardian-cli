// Package config provides YAML configuration parsing and validation
// for the Guardian CLI tool. It handles reading, writing, and validating
// the various configuration files stored in the .agreements/ directory.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Constitution represents the top-level constitution.yml configuration file.
type Constitution struct {
	Governance Governance      `yaml:"governance"`
	Identity   Identity        `yaml:"identity"`
	Roles      map[string]Role `yaml:"roles"`
	LLM        LLMConfig       `yaml:"llm"`
}

// Governance defines the voting and proposal governance rules.
type Governance struct {
	Voters             []VoterRef                `yaml:"voters"`
	Quorum             QuorumConfig              `yaml:"quorum"`
	ForbidSelfApproval bool                      `yaml:"forbid_self_approval"`
	AllowVoteChange    bool                      `yaml:"allow_vote_change"`
	ProposalTTLDays    int                       `yaml:"proposal_ttl_days"`
	PerRuleOverrides   map[string]RuleOverride   `yaml:"per_rule_overrides"`
	Exceptions         ExceptionPolicy           `yaml:"exceptions"`
}

// VoterRef references a role that is eligible to vote.
type VoterRef struct {
	Role string `yaml:"role"`
}

// QuorumConfig defines the quorum calculation method.
type QuorumConfig struct {
	Type      string  `yaml:"type"`      // majority|two_thirds|unanimous|custom
	Threshold float64 `yaml:"threshold"` // for custom
}

// RuleOverride allows per-rule quorum overrides.
type RuleOverride struct {
	Quorum QuorumConfig `yaml:"quorum"`
}

// ExceptionPolicy configures how exceptions are handled.
type ExceptionPolicy struct {
	RequireApproval bool `yaml:"require_approval"`
}

// Identity configures identity verification settings.
type Identity struct {
	AllowedDomains       []string `yaml:"allowed_domains"`
	RequireSignedCommits bool     `yaml:"require_signed_commits"`
}

// Role defines a named role and its members.
type Role struct {
	Members []RoleMember `yaml:"members"`
}

// RoleMember represents a single member within a role.
type RoleMember struct {
	Email string `yaml:"email"`
}

// LLMConfig configures the LLM provider used by Guardian.
type LLMConfig struct {
	Provider string     `yaml:"provider"` // deepseek|openai|claude|custom
	Endpoint string     `yaml:"endpoint"`
	Model    string     `yaml:"model"`
	Prompts  LLMPrompts `yaml:"prompts"`
}

// LLMPrompts allows overriding the built-in LLM system prompts.
type LLMPrompts struct {
	CheckSystem   string `yaml:"check_system"`
	ProposeSystem string `yaml:"propose_system"`
}

// LoadConstitution reads and parses a constitution.yml file from the given path.
func LoadConstitution(path string) (*Constitution, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading constitution file %s: %w", path, err)
	}

	var c Constitution
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parsing constitution file %s: %w", path, err)
	}

	return &c, nil
}

// SaveConstitution writes a Constitution to the given path as YAML.
func SaveConstitution(path string, c *Constitution) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling constitution: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing constitution file %s: %w", path, err)
	}

	return nil
}
