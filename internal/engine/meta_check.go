package engine

import (
	"os"
	"path/filepath"
	"strings"
)

// protectedFiles lists the .agreements files that require an accepted proposal
// before they can be modified.
var protectedFiles = []string{
	".agreements/constitution.yml",
	".agreements/rules.yml",
}

// MetaChecker detects unauthorized changes to governance files
// (.agreements/constitution.yml and .agreements/rules.yml). Changes to these
// files without a corresponding accepted proposal produce a violation.
type MetaChecker struct{}

// Check inspects the changed files for protected governance files and verifies
// that a corresponding accepted proposal exists in the proposals directory.
func (m *MetaChecker) Check(changedFiles []string, proposalsDir string) ([]Violation, error) {
	var violations []Violation

	for _, file := range changedFiles {
		if !isProtectedFile(file) {
			continue
		}

		hasProposal, err := hasAcceptedProposal(proposalsDir)
		if err != nil {
			return nil, err
		}

		if !hasProposal {
			violations = append(violations, Violation{
				RuleID:      "meta_check",
				Severity:    "error",
				Description: "Changes to " + file + " require an accepted proposal",
				FilePath:    file,
				DiffSnippet: "",
			})
		}
	}

	return violations, nil
}

// isProtectedFile checks if the given file path is one of the protected
// governance files.
func isProtectedFile(file string) bool {
	normalized := filepath.ToSlash(file)
	for _, p := range protectedFiles {
		if normalized == p {
			return true
		}
	}
	return false
}

// hasAcceptedProposal checks if there is at least one accepted proposal
// in the proposals directory. It reads YAML files and looks for
// "status: accepted" in their content.
func hasAcceptedProposal(proposalsDir string) (bool, error) {
	entries, err := os.ReadDir(proposalsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".yml" && ext != ".yaml" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(proposalsDir, entry.Name()))
		if err != nil {
			continue
		}

		content := string(data)
		if strings.Contains(content, "status: accepted") {
			return true, nil
		}
	}

	return false, nil
}
