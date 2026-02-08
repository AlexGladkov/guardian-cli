package output

import (
	"fmt"
	"io"
	"strings"
)

// PrintCheckReportHuman writes a human-readable check report to the given writer.
func PrintCheckReportHuman(w io.Writer, r *CheckReport) {
	fmt.Fprintln(w, "Guardian Check Report")
	fmt.Fprintln(w, "=====================")
	fmt.Fprintln(w)

	if len(r.Violations) == 0 {
		fmt.Fprintln(w, "No violations found.")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Result: PASSED")
		return
	}

	for _, v := range r.Violations {
		label := "VIOLATION"
		if v.Severity == "warning" {
			label = "WARNING"
		}

		fmt.Fprintf(w, "%s [%s] %s\n", label, v.Severity, v.RuleID)
		fmt.Fprintf(w, "  %s\n", v.Description)

		if v.FilePath != "" {
			fmt.Fprintf(w, "  File: %s\n", v.FilePath)
		}

		if v.DiffSnippet != "" {
			fmt.Fprintln(w, "  Diff:")
			for _, line := range strings.Split(v.DiffSnippet, "\n") {
				fmt.Fprintf(w, "    %s\n", line)
			}
		}

		if v.LLMExplanation != "" {
			fmt.Fprintln(w, "  AI:", v.LLMExplanation)
		}

		fmt.Fprintln(w)
	}

	passedStr := "PASSED"
	if !r.Summary.Passed {
		passedStr = "FAILED"
	}

	fmt.Fprintf(w, "Result: %d error(s), %d warning(s) - %s\n",
		r.Summary.Errors, r.Summary.Warnings, passedStr)
}

// PrintTallyReportHuman writes a human-readable tally report to the given writer.
func PrintTallyReportHuman(w io.Writer, r *TallyReport) {
	fmt.Fprintln(w, "Guardian Tally Report")
	fmt.Fprintln(w, "=====================")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "Proposal: %s\n", r.ProposalID)
	fmt.Fprintf(w, "Rule:     %s\n", r.RuleID)
	fmt.Fprintln(w)

	fmt.Fprintf(w, "Eligible voters (%d):\n", len(r.EligibleVoters))
	for _, voter := range r.EligibleVoters {
		fmt.Fprintf(w, "  - %s\n", voter)
	}
	fmt.Fprintln(w)

	fmt.Fprintf(w, "Votes (%d):\n", len(r.Votes))
	if len(r.Votes) == 0 {
		fmt.Fprintln(w, "  (no votes yet)")
	} else {
		for _, vote := range r.Votes {
			comment := ""
			if vote.Comment != "" {
				comment = fmt.Sprintf(" - %q", vote.Comment)
			}
			fmt.Fprintf(w, "  %s: %s%s\n", vote.Email, strings.ToUpper(vote.Decision), comment)
		}
	}
	fmt.Fprintln(w)

	fmt.Fprintf(w, "Yes: %d / No: %d / Required: %d\n", r.YesCount, r.NoCount, r.Required)
	fmt.Fprintf(w, "Result: %s\n", r.Result)
}

// PrintInboxReportHuman writes a human-readable inbox report to the given writer.
func PrintInboxReportHuman(w io.Writer, r *InboxReport) {
	fmt.Fprintln(w, "Guardian Inbox")
	fmt.Fprintln(w, "==============")
	fmt.Fprintln(w)

	if r.Total == 0 {
		fmt.Fprintln(w, "No pending proposals require your vote.")
		return
	}

	fmt.Fprintf(w, "%d proposal(s) awaiting your vote:\n", r.Total)
	fmt.Fprintln(w)

	for _, item := range r.Items {
		fmt.Fprintf(w, "  [%s] %s (%s)\n", item.ProposalType, item.ProposalID, item.RuleID)
		fmt.Fprintf(w, "    Created by: %s\n", item.CreatedBy)
		fmt.Fprintf(w, "    Created at: %s (age: %s)\n", item.CreatedAt, item.Age)
		fmt.Fprintln(w)
	}
}

// PrintHistoryReportHuman writes a human-readable history report to the given writer.
func PrintHistoryReportHuman(w io.Writer, r *HistoryReport) {
	fmt.Fprintln(w, "Guardian History")
	fmt.Fprintln(w, "================")
	fmt.Fprintln(w)

	if len(r.Items) == 0 {
		fmt.Fprintln(w, "No finalized proposals.")
		return
	}

	for _, item := range r.Items {
		fmt.Fprintf(w, "  [%s] %s (%s)\n", item.ProposalType, item.ProposalID, item.RuleID)
		fmt.Fprintf(w, "    Finalized: %s\n", item.FinalizedAt)
		if item.Summary != "" {
			fmt.Fprintf(w, "    Summary: %s\n", item.Summary)
		}
		fmt.Fprintln(w)
	}
}
