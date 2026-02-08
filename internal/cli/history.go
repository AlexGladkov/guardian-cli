package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexGladkov/guardian-cli/internal/output"
)

const historyUsage = `Usage: guardian history [--json]

Show the history of finalized proposals.

Flags:
  --json     Output results as JSON
  --help     Show this help message

Exit codes:
  0  Success
`

func runHistory(args []string) int {
	fs := flag.NewFlagSet("history", flag.ContinueOnError)
	jsonOutput := fs.Bool("json", false, "Output results as JSON")
	fs.Usage = func() { fmt.Fprint(os.Stderr, historyUsage) }

	if err := fs.Parse(args); err != nil {
		return 2
	}

	// Find .agreements directory.
	agreementsDir, err := findAgreementsDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 0
	}

	// Read history directory.
	historyDir := filepath.Join(agreementsDir, "history")
	entries, err := os.ReadDir(historyDir)
	if err != nil {
		if os.IsNotExist(err) {
			entries = nil
		} else {
			fmt.Fprintf(os.Stderr, "Error: reading history directory: %v\n", err)
			return 0
		}
	}

	// Build history report.
	report := &output.HistoryReport{}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".md" {
			continue
		}

		filePath := filepath.Join(historyDir, entry.Name())
		data, readErr := os.ReadFile(filePath)
		if readErr != nil {
			continue
		}

		item := parseHistoryFile(entry.Name(), string(data))
		report.Items = append(report.Items, item)
	}

	// Output.
	if *jsonOutput {
		if err := output.PrintHistoryReportJSON(os.Stdout, report); err != nil {
			fmt.Fprintf(os.Stderr, "Error: writing JSON output: %v\n", err)
		}
	} else {
		output.PrintHistoryReportHuman(os.Stdout, report)
	}

	return 0
}

// parseHistoryFile extracts a HistoryItem from a markdown history file.
func parseHistoryFile(filename string, content string) output.HistoryItem {
	// Extract proposal ID from filename (remove .md extension).
	proposalID := strings.TrimSuffix(filename, ".md")

	item := output.HistoryItem{
		ProposalID: proposalID,
	}

	// Parse metadata from markdown content.
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- **Rule:**") {
			item.RuleID = strings.TrimSpace(strings.TrimPrefix(trimmed, "- **Rule:**"))
		} else if strings.HasPrefix(trimmed, "- **Type:**") {
			item.ProposalType = strings.TrimSpace(strings.TrimPrefix(trimmed, "- **Type:**"))
		} else if strings.HasPrefix(trimmed, "- **Finalized at:**") {
			item.FinalizedAt = strings.TrimSpace(strings.TrimPrefix(trimmed, "- **Finalized at:**"))
		} else if strings.HasPrefix(trimmed, "- **Result:**") {
			item.Summary = strings.TrimSpace(strings.TrimPrefix(trimmed, "- **Result:**"))
		}
	}

	return item
}
