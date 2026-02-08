package engine

import (
	"fmt"
	"regexp"
)

// DiffPatternForbiddenChecker checks that added lines in the diff do not match
// any of the configured forbidden regular expressions. Optionally scoped to
// specific file path patterns via only_in_paths.
type DiffPatternForbiddenChecker struct{}

// Type returns the checker type identifier.
func (c *DiffPatternForbiddenChecker) Type() string {
	return "diff_pattern_forbidden"
}

// Check evaluates the diff_pattern_forbidden rule against the given context.
func (c *DiffPatternForbiddenChecker) Check(ctx *CheckContext) ([]Violation, error) {
	forbiddenRegexes, err := getStringSlice(ctx.RuleConfig, "forbidden_regexes")
	if err != nil {
		return nil, fmt.Errorf("diff_pattern_forbidden: %w", err)
	}

	// Compile all forbidden regexes.
	compiled := make([]*regexp.Regexp, 0, len(forbiddenRegexes))
	for _, pattern := range forbiddenRegexes {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("diff_pattern_forbidden: invalid regex %q: %w", pattern, err)
		}
		compiled = append(compiled, re)
	}

	// Determine which files to check.
	filesToCheck := ctx.ChangedFiles
	onlyInPaths, _ := getStringSlice(ctx.RuleConfig, "only_in_paths")
	if len(onlyInPaths) > 0 {
		filesToCheck = filterFilesByGlobs(ctx.ChangedFiles, onlyInPaths)
		if len(filesToCheck) == 0 {
			return nil, nil
		}
	}

	// Build a set for fast lookup.
	fileSet := make(map[string]bool, len(filesToCheck))
	for _, f := range filesToCheck {
		fileSet[f] = true
	}

	// Parse diff and check added lines.
	fileDiffs := ParseDiff(ctx.DiffContent)

	var violations []Violation
	for _, fd := range fileDiffs {
		if !fileSet[fd.Path] {
			continue
		}

		for _, addedLine := range fd.AddedLines {
			for _, re := range compiled {
				if re.MatchString(addedLine) {
					violations = append(violations, Violation{
						RuleID:      ctx.RuleID,
						Severity:    ctx.Severity,
						Description: ctx.RuleDesc,
						FilePath:    fd.Path,
						DiffSnippet: "+" + addedLine,
					})
					break // one violation per line is enough
				}
			}
		}
	}

	return violations, nil
}

// filterFilesByGlobs returns only the files that match at least one of the given globs.
func filterFilesByGlobs(files []string, globs []string) []string {
	var matched []string
	for _, f := range files {
		if matchesAnyGlob(f, globs) {
			matched = append(matched, f)
		}
	}
	return matched
}
