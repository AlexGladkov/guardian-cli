package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Exception represents a temporary or permanent exemption from a rule
// for specific file paths.
type Exception struct {
	ID        string     `yaml:"id"`
	RuleID    string     `yaml:"rule_id"`
	Paths     []string   `yaml:"paths"`
	Reason    string     `yaml:"reason"`
	CreatedBy string     `yaml:"created_by"`
	CreatedAt time.Time  `yaml:"created_at"`
	ExpiresAt *time.Time `yaml:"expires_at,omitempty"`
}

// LoadException reads and parses an exception YAML file from the given path.
func LoadException(path string) (*Exception, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading exception file %s: %w", path, err)
	}

	var e Exception
	if err := yaml.Unmarshal(data, &e); err != nil {
		return nil, fmt.Errorf("parsing exception file %s: %w", path, err)
	}

	return &e, nil
}

// SaveException writes an Exception to the given path as YAML.
func SaveException(path string, e *Exception) error {
	data, err := yaml.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshaling exception: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating exception directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing exception file %s: %w", path, err)
	}

	return nil
}

// LoadAllExceptions reads all YAML exception files from the given directory.
func LoadAllExceptions(dir string) ([]*Exception, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading exceptions directory %s: %w", dir, err)
	}

	var exceptions []*Exception
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".yml" && ext != ".yaml" {
			continue
		}

		e, err := LoadException(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("loading exception %s: %w", name, err)
		}
		exceptions = append(exceptions, e)
	}

	return exceptions, nil
}
