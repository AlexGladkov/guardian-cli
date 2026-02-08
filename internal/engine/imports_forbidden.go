package engine

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ImportsForbiddenChecker checks that files matching from_globs do not contain
// imports from paths matching forbid_globs. It works by inspecting added lines
// in the diff for path segments derived from the forbidden glob patterns.
type ImportsForbiddenChecker struct{}

// Type returns the checker type identifier.
func (c *ImportsForbiddenChecker) Type() string {
	return "imports_forbidden"
}

// Check evaluates the imports_forbidden rule against the given context.
func (c *ImportsForbiddenChecker) Check(ctx *CheckContext) ([]Violation, error) {
	fromGlobs, err := getStringSlice(ctx.RuleConfig, "from_globs")
	if err != nil {
		return nil, fmt.Errorf("imports_forbidden: %w", err)
	}

	forbidGlobs, err := getStringSlice(ctx.RuleConfig, "forbid_globs")
	if err != nil {
		return nil, fmt.Errorf("imports_forbidden: %w", err)
	}

	// Parse the diff into per-file diffs.
	fileDiffs := ParseDiff(ctx.DiffContent)

	// Build a map of file path to its diff for quick lookup.
	diffMap := make(map[string]*FileDiff, len(fileDiffs))
	for i := range fileDiffs {
		diffMap[fileDiffs[i].Path] = &fileDiffs[i]
	}

	// Extract forbidden path segments from forbid_globs.
	// For example, "infra/**" yields "infra".
	forbidSegments := extractPathSegments(forbidGlobs)

	var violations []Violation

	for _, file := range ctx.ChangedFiles {
		if !matchesAnyGlob(file, fromGlobs) {
			continue
		}

		fd, ok := diffMap[file]
		if !ok {
			continue
		}

		for _, addedLine := range fd.AddedLines {
			for _, segment := range forbidSegments {
				if containsSegment(addedLine, segment) {
					violations = append(violations, Violation{
						RuleID:      ctx.RuleID,
						Severity:    ctx.Severity,
						Description: ctx.RuleDesc,
						FilePath:    file,
						DiffSnippet: "+" + addedLine,
					})
					break // one violation per line is enough
				}
			}
		}
	}

	return violations, nil
}

// extractPathSegments takes a list of glob patterns and extracts the leading
// directory names that can be used as search terms. For example, "infra/**"
// yields "infra", "data/models/*" yields "data/models".
func extractPathSegments(globs []string) []string {
	segments := make([]string, 0, len(globs))
	for _, g := range globs {
		// Remove trailing glob characters to get the directory part.
		s := strings.TrimRight(g, "*")
		s = strings.TrimRight(s, "/")
		if s != "" {
			segments = append(segments, s)
		}
	}
	return segments
}

// containsSegment checks if a line contains a forbidden path segment.
// It looks for the segment as a substring, checking for path-like boundaries
// (e.g., "/" or "." before/after the segment).
func containsSegment(line, segment string) bool {
	return strings.Contains(line, segment+"/") || strings.Contains(line, segment+".")
}

// matchesAnyGlob checks if a file path matches any of the given glob patterns.
func matchesAnyGlob(file string, globs []string) bool {
	for _, g := range globs {
		if globMatch(file, g) {
			return true
		}
	}
	return false
}

// globMatch performs glob matching that supports "**" for recursive directory matching.
// For patterns with "**", it splits on "**" and checks prefix/suffix or just presence
// of the parts. For simple patterns, it delegates to filepath.Match.
func globMatch(path, pattern string) bool {
	// Handle "**" patterns.
	if strings.Contains(pattern, "**") {
		parts := strings.SplitN(pattern, "**", 2)
		prefix := parts[0]
		suffix := ""
		if len(parts) > 1 {
			suffix = strings.TrimLeft(parts[1], "/")
		}

		// Check prefix.
		if prefix != "" {
			prefix = strings.TrimRight(prefix, "/")
			if !strings.HasPrefix(path, prefix) {
				return false
			}
		}

		// Check suffix using filepath.Match on remaining path parts.
		if suffix != "" {
			// Try matching the suffix against each possible sub-path.
			remaining := path
			if prefix != "" {
				remaining = strings.TrimPrefix(path, prefix)
				remaining = strings.TrimLeft(remaining, "/")
			}

			// Try matching the suffix against the file name or sub-paths.
			pathParts := strings.Split(remaining, "/")
			for i := range pathParts {
				sub := strings.Join(pathParts[i:], "/")
				if matched, _ := filepath.Match(suffix, sub); matched {
					return true
				}
			}
			return false
		}

		// Pattern is just "prefix/**" — anything under prefix matches.
		return true
	}

	// Simple glob without **.
	matched, _ := filepath.Match(pattern, path)
	return matched
}

// getStringSlice extracts a []string from a map[string]interface{} by key.
func getStringSlice(cfg map[string]interface{}, key string) ([]string, error) {
	val, ok := cfg[key]
	if !ok {
		return nil, fmt.Errorf("missing config key %q", key)
	}

	switch v := val.(type) {
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("config key %q: expected string elements, got %T", key, item)
			}
			result = append(result, s)
		}
		return result, nil
	case []string:
		return v, nil
	default:
		return nil, fmt.Errorf("config key %q: expected string slice, got %T", key, val)
	}
}
