package engine

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/AlexGladkov/guardian-cli/internal/config"
)

// Engine orchestrates rule checking by running all configured rules against
// the given diff and filtering out exceptions.
type Engine struct {
	Rules      []config.Rule
	Exceptions []config.Exception
}

// EngineResult holds the aggregated results of running all rules.
type EngineResult struct {
	Violations []Violation
	Errors     int
	Warnings   int
}

// NewEngine creates a new Engine with the given rules and exceptions.
func NewEngine(rules []config.Rule, exceptions []config.Exception) *Engine {
	return &Engine{
		Rules:      rules,
		Exceptions: exceptions,
	}
}

// Run executes all registered rule checkers against the changed files and diff
// content, filters out exceptions, and returns the aggregated result.
func (e *Engine) Run(changedFiles []string, diffContent string) (*EngineResult, error) {
	var allViolations []Violation

	for _, rule := range e.Rules {
		checker, ok := Registry[rule.Type]
		if !ok {
			return nil, fmt.Errorf("unknown rule type %q for rule %q", rule.Type, rule.ID)
		}

		ctx := &CheckContext{
			ChangedFiles: changedFiles,
			DiffContent:  diffContent,
			RuleConfig:   rule.Config,
			Severity:     rule.Severity,
			RuleID:       rule.ID,
			RuleDesc:     rule.Description,
		}

		violations, err := checker.Check(ctx)
		if err != nil {
			return nil, fmt.Errorf("checking rule %q: %w", rule.ID, err)
		}

		allViolations = append(allViolations, violations...)
	}

	// Filter out exceptions.
	filtered := e.applyExceptions(allViolations)

	// Count errors and warnings.
	result := &EngineResult{
		Violations: filtered,
	}
	for _, v := range filtered {
		switch v.Severity {
		case "error":
			result.Errors++
		case "warning":
			result.Warnings++
		}
	}

	return result, nil
}

// applyExceptions removes violations that are covered by non-expired exceptions.
func (e *Engine) applyExceptions(violations []Violation) []Violation {
	now := time.Now()

	// Build a list of active exceptions.
	activeExceptions := make([]config.Exception, 0, len(e.Exceptions))
	for _, exc := range e.Exceptions {
		if exc.ExpiresAt != nil && exc.ExpiresAt.Before(now) {
			continue // expired
		}
		activeExceptions = append(activeExceptions, exc)
	}

	if len(activeExceptions) == 0 {
		return violations
	}

	filtered := make([]Violation, 0, len(violations))
	for _, v := range violations {
		if isExcepted(v, activeExceptions) {
			continue
		}
		filtered = append(filtered, v)
	}

	return filtered
}

// isExcepted checks if a violation is covered by any of the given exceptions.
// An exception matches if its RuleID matches the violation's RuleID and at
// least one of its Paths glob-matches the violation's FilePath.
func isExcepted(v Violation, exceptions []config.Exception) bool {
	for _, exc := range exceptions {
		if exc.RuleID != v.RuleID {
			continue
		}

		for _, pattern := range exc.Paths {
			matched, err := filepath.Match(pattern, v.FilePath)
			if err != nil {
				continue
			}
			if matched {
				return true
			}
		}
	}
	return false
}
