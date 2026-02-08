package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// RulesFile represents the rules.yml configuration containing all rule definitions.
type RulesFile struct {
	Rules []Rule `yaml:"rules"`
}

// Rule defines a single rule that Guardian checks code changes against.
type Rule struct {
	ID          string                 `yaml:"id"`
	Description string                 `yaml:"description"`
	Type        string                 `yaml:"type"`
	Config      map[string]interface{} `yaml:"config"`
	Severity    string                 `yaml:"severity"`
}

// LoadRules reads and parses a rules.yml file from the given path.
func LoadRules(path string) (*RulesFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading rules file %s: %w", path, err)
	}

	var r RulesFile
	if err := yaml.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parsing rules file %s: %w", path, err)
	}

	return &r, nil
}
