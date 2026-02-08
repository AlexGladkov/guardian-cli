// Package engine implements the rule checking engine for Guardian. It provides
// a registry of rule checkers, a unified orchestrator, and diff parsing utilities.
package engine

// RuleChecker is the interface that all rule type checkers must implement.
type RuleChecker interface {
	// Check evaluates the rule against the given context and returns any violations found.
	Check(ctx *CheckContext) ([]Violation, error)
	// Type returns the unique identifier for this checker type (e.g., "imports_forbidden").
	Type() string
}

// CheckContext provides all the information a RuleChecker needs to evaluate a rule.
type CheckContext struct {
	ChangedFiles []string
	DiffContent  string
	RuleConfig   map[string]interface{}
	Severity     string
	RuleID       string
	RuleDesc     string
}

// Violation represents a single rule violation found during checking.
type Violation struct {
	RuleID         string `json:"rule_id"`
	Severity       string `json:"severity"`
	Description    string `json:"description"`
	FilePath       string `json:"file_path"`
	DiffSnippet    string `json:"diff_snippet"`
	LLMExplanation string `json:"llm_explanation"`
}

// Registry maps rule type names to their corresponding checkers.
var Registry = map[string]RuleChecker{}

// RegisterChecker registers a RuleChecker in the global registry by its type name.
func RegisterChecker(c RuleChecker) {
	Registry[c.Type()] = c
}

func init() {
	RegisterChecker(&ImportsForbiddenChecker{})
	RegisterChecker(&DiffPatternForbiddenChecker{})
	RegisterChecker(&DiffPatternRequiresChecker{})
}
