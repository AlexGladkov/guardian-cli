package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// PrintCheckReportJSON writes a JSON-formatted check report to the given writer.
func PrintCheckReportJSON(w io.Writer, r *CheckReport) error {
	return writeJSON(w, r)
}

// PrintTallyReportJSON writes a JSON-formatted tally report to the given writer.
func PrintTallyReportJSON(w io.Writer, r *TallyReport) error {
	return writeJSON(w, r)
}

// PrintInboxReportJSON writes a JSON-formatted inbox report to the given writer.
func PrintInboxReportJSON(w io.Writer, r *InboxReport) error {
	return writeJSON(w, r)
}

// PrintHistoryReportJSON writes a JSON-formatted history report to the given writer.
func PrintHistoryReportJSON(w io.Writer, r *HistoryReport) error {
	return writeJSON(w, r)
}

// writeJSON encodes the given value as indented JSON and writes it to w.
func writeJSON(w io.Writer, v interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(v); err != nil {
		return fmt.Errorf("encoding JSON output: %w", err)
	}
	return nil
}
