package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadRules_FullFile(t *testing.T) {
	content := `
rules:
  - id: domain_no_infra
    description: Domain layer must not depend on infra
    type: imports_forbidden
    config:
      from_globs: ["domain/**"]
      forbid_globs: ["infra/**"]
    severity: error
  - id: money_minor_units
    description: Money must use int minor units
    type: diff_pattern_forbidden
    config:
      forbidden_regexes:
        - "\\bDouble\\b"
        - "\\bfloat\\b"
      only_in_paths: ["**/*.kt", "**/*.java"]
    severity: warning
  - id: public_api_stability
    description: Public API changes require RFC tag
    type: diff_pattern_requires
    config:
      required_regexes:
        - "RFC:"
      only_in_paths: ["sdk/public/**"]
    severity: error
`
	path := writeTestFile(t, "rules.yml", content)

	r, err := LoadRules(path)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Len(t, r.Rules, 3)

	// First rule
	assert.Equal(t, "domain_no_infra", r.Rules[0].ID)
	assert.Equal(t, "Domain layer must not depend on infra", r.Rules[0].Description)
	assert.Equal(t, "imports_forbidden", r.Rules[0].Type)
	assert.Equal(t, "error", r.Rules[0].Severity)
	assert.NotNil(t, r.Rules[0].Config)
	fromGlobs, ok := r.Rules[0].Config["from_globs"]
	require.True(t, ok)
	assert.NotNil(t, fromGlobs)

	// Second rule
	assert.Equal(t, "money_minor_units", r.Rules[1].ID)
	assert.Equal(t, "diff_pattern_forbidden", r.Rules[1].Type)
	assert.Equal(t, "warning", r.Rules[1].Severity)

	// Third rule
	assert.Equal(t, "public_api_stability", r.Rules[2].ID)
	assert.Equal(t, "diff_pattern_requires", r.Rules[2].Type)
}

func TestLoadRules_EmptyRules(t *testing.T) {
	content := `rules: []`
	path := writeTestFile(t, "rules.yml", content)

	r, err := LoadRules(path)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Empty(t, r.Rules)
}

func TestLoadRules_NoRulesKey(t *testing.T) {
	content := `# empty file with no rules key`
	path := writeTestFile(t, "rules.yml", content)

	r, err := LoadRules(path)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Nil(t, r.Rules)
}

func TestLoadRules_FileNotFound(t *testing.T) {
	_, err := LoadRules("/nonexistent/rules.yml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading rules file")
}

func TestLoadRules_InvalidYAML(t *testing.T) {
	path := writeTestFile(t, "bad_rules.yml", "{{invalid yaml")

	_, err := LoadRules(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing rules file")
}

func TestLoadRules_SingleRule(t *testing.T) {
	content := `
rules:
  - id: single_rule
    description: A single test rule
    type: diff_pattern_forbidden
    config:
      forbidden_regexes: ["TODO"]
    severity: warning
`
	path := writeTestFile(t, "rules.yml", content)

	r, err := LoadRules(path)
	require.NoError(t, err)
	require.Len(t, r.Rules, 1)
	assert.Equal(t, "single_rule", r.Rules[0].ID)
}

func TestLoadRules_RuleWithoutConfig(t *testing.T) {
	content := `
rules:
  - id: no_config_rule
    description: Rule with no config
    type: custom_check
    severity: error
`
	path := writeTestFile(t, "rules.yml", content)

	r, err := LoadRules(path)
	require.NoError(t, err)
	require.Len(t, r.Rules, 1)
	assert.Nil(t, r.Rules[0].Config)
}

func TestLoadRules_ConfigValues(t *testing.T) {
	content := `
rules:
  - id: test_rule
    description: Test config parsing
    type: diff_pattern_forbidden
    config:
      forbidden_regexes:
        - "pattern1"
        - "pattern2"
      only_in_paths:
        - "src/**"
      max_violations: 5
    severity: error
`
	path := writeTestFile(t, "rules.yml", content)

	r, err := LoadRules(path)
	require.NoError(t, err)
	require.Len(t, r.Rules, 1)

	cfg := r.Rules[0].Config
	require.NotNil(t, cfg)

	regexes, ok := cfg["forbidden_regexes"]
	require.True(t, ok)
	regexSlice, ok := regexes.([]interface{})
	require.True(t, ok)
	assert.Len(t, regexSlice, 2)
}
