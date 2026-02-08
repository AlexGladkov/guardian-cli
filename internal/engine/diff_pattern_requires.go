package engine

import (
	"fmt"
	"regexp"
)

// DiffPatternRequiresChecker checks that when files matching only_in_paths are
// changed, at least one of the required_regexes patterns appears somewhere in
// the added lines of the diff. If none match, a violation is returned.
type DiffPatternRequiresChecker struct{}

// Type returns the checker type identifier.
func (c *DiffPatternRequiresChecker) Type() string {
	return "diff_pattern_requires"
}

// Check evaluates the diff_pattern_requires rule against the given context.
func (c *DiffPatternRequiresChecker) Check(ctx *CheckContext) ([]Violation, error) {
	requiredRegexes, err := getStringSlice(ctx.RuleConfig, "required_regexes")
	if err != nil {
		return nil, fmt.Errorf("diff_pattern_requires: %w", err)
	}

	onlyInPaths, err := getStringSlice(ctx.RuleConfig, "only_in_paths")
	if err != nil {
		return nil, fmt.Errorf("diff_pattern_requires: %w", err)
	}

	// Filter changed files matching only_in_paths.
	matchedFiles := filterFilesByGlobs(ctx.ChangedFiles, onlyInPaths)
	if len(matchedFiles) == 0 {
		// Rule does not apply when no matching files are changed.
		return nil, nil
	}

	// Compile required regexes.
	compiled := make([]*regexp.Regexp, 0, len(requiredRegexes))
	for _, pattern := range requiredRegexes {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("diff_pattern_requires: invalid regex %q: %w", pattern, err)
		}
		compiled = append(compiled, re)
	}

	// Parse diff and collect all added lines.
	fileDiffs := ParseDiff(ctx.DiffContent)
	for _, fd := range fileDiffs {
		for _, addedLine := range fd.AddedLines {
			for _, re := range compiled {
				if re.MatchString(addedLine) {
					// Found a required pattern — no violation.
					return nil, nil
				}
			}
		}
	}

	// None of the required patterns were found.
	return []Violation{
		{
			RuleID:      ctx.RuleID,
			Severity:    ctx.Severity,
			Description: ctx.RuleDesc,
			FilePath:    matchedFiles[0],
			DiffSnippet: "",
		},
	}, nil
}
