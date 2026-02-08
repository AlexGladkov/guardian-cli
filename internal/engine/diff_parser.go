package engine

import (
	"strings"
)

// FileDiff represents the diff for a single file.
type FileDiff struct {
	Path       string
	AddedLines []string // lines starting with "+" (without the leading "+")
}

// ParseDiff parses unified diff content into per-file diffs. Each entry in the
// returned slice corresponds to one file in the diff and contains the file path
// and all added lines (lines prefixed with "+", excluding the "+++ b/" header).
func ParseDiff(diffContent string) []FileDiff {
	if diffContent == "" {
		return nil
	}

	var result []FileDiff
	var current *FileDiff

	lines := strings.Split(diffContent, "\n")
	for _, line := range lines {
		// Detect new file in diff.
		if strings.HasPrefix(line, "diff --git ") {
			if current != nil {
				result = append(result, *current)
			}
			current = &FileDiff{}
			continue
		}

		// Extract file path from the "+++ b/..." line.
		if strings.HasPrefix(line, "+++ b/") {
			if current != nil {
				current.Path = strings.TrimPrefix(line, "+++ b/")
			}
			continue
		}

		// Skip the "--- a/..." header line.
		if strings.HasPrefix(line, "--- ") {
			continue
		}

		// Skip hunk headers.
		if strings.HasPrefix(line, "@@") {
			continue
		}

		// Collect added lines (but not the "+++ b/" header which we already handled).
		if strings.HasPrefix(line, "+") && current != nil {
			// Store the line without the leading "+"
			current.AddedLines = append(current.AddedLines, line[1:])
		}
	}

	if current != nil {
		result = append(result, *current)
	}

	return result
}
