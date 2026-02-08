package cli

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/AlexGladkov/guardian-cli/internal/config"
	"github.com/AlexGladkov/guardian-cli/internal/git"
	"github.com/AlexGladkov/guardian-cli/internal/inbox"
	"github.com/AlexGladkov/guardian-cli/internal/output"
)

const inboxUsage = `Usage: guardian inbox [flags]

Show proposals awaiting your vote.

Flags:
  --notify           Send OS notification for pending proposals
  --since-last-check Only show proposals since last inbox check
  --quiet            Minimal output (count only)
  --no-fetch         Skip git fetch before checking
  --json             Output results as JSON
  --help             Show this help message

Exit codes:
  0  Success
`

func runInbox(args []string) int {
	fs := flag.NewFlagSet("inbox", flag.ContinueOnError)
	notify := fs.Bool("notify", false, "Send OS notification")
	sinceLastCheck := fs.Bool("since-last-check", false, "Filter by last check time")
	quiet := fs.Bool("quiet", false, "Minimal output")
	noFetch := fs.Bool("no-fetch", false, "Skip git fetch")
	jsonOutput := fs.Bool("json", false, "Output results as JSON")
	fs.Usage = func() { fmt.Fprint(os.Stderr, inboxUsage) }

	if err := fs.Parse(args); err != nil {
		return 2
	}

	// Optionally fetch.
	if !*noFetch {
		if err := git.Fetch(); err != nil {
			// Non-fatal: warn and continue.
			fmt.Fprintf(os.Stderr, "Warning: git fetch failed: %v\n", err)
		}
	}

	// Find .agreements directory.
	agreementsDir, err := findAgreementsDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 0
	}

	root := repoRoot(agreementsDir)

	// Load constitution.
	constitution, err := loadConstitutionFrom(agreementsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: loading constitution: %v\n", err)
		return 0
	}

	// Get user email.
	email, err := git.GetUserEmail()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 0
	}

	// Load proposals.
	proposals, err := loadAllProposalsFrom(agreementsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: loading proposals: %v\n", err)
		return 0
	}

	// Load votes for each proposal.
	votesMap := make(map[string][]*config.Vote)
	for _, p := range proposals {
		votes, loadErr := loadVotesForProposalFrom(agreementsDir, p.ID)
		if loadErr != nil {
			continue
		}
		votesMap[p.ID] = votes
	}

	// Determine sinceLastCheck time.
	var sinceTime *time.Time
	if *sinceLastCheck {
		state, stateErr := inbox.LoadState(root)
		if stateErr == nil && !state.LastInboxCheck.IsZero() {
			sinceTime = &state.LastInboxCheck
		}
	}

	// Get inbox items.
	items, err := inbox.GetInbox(proposals, votesMap, constitution, email, sinceTime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: getting inbox: %v\n", err)
		return 0
	}

	// Update state.
	state := &inbox.State{
		LastInboxCheck: time.Now().UTC(),
	}
	if saveErr := inbox.SaveState(root, state); saveErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not save state: %v\n", saveErr)
	}

	// Send notification if requested.
	if *notify && len(items) > 0 {
		title := "Guardian"
		message := fmt.Sprintf("%d proposal(s) awaiting your vote", len(items))
		if notifyErr := inbox.SendNotification(title, message); notifyErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: notification failed: %v\n", notifyErr)
		}
	}

	// Quiet mode: just print count.
	if *quiet {
		if len(items) > 0 {
			fmt.Fprintf(os.Stdout, "%d\n", len(items))
		}
		return 0
	}

	// Build report.
	report := buildInboxReport(items)

	// Output.
	if *jsonOutput {
		if err := output.PrintInboxReportJSON(os.Stdout, report); err != nil {
			fmt.Fprintf(os.Stderr, "Error: writing JSON output: %v\n", err)
		}
	} else {
		output.PrintInboxReportHuman(os.Stdout, report)
	}

	return 0
}

// buildInboxReport converts inbox items to an output report.
func buildInboxReport(items []inbox.InboxItem) *output.InboxReport {
	report := &output.InboxReport{
		Total: len(items),
	}

	for _, item := range items {
		report.Items = append(report.Items, output.InboxItem{
			ProposalID:   item.Proposal.ID,
			RuleID:       item.Proposal.RuleID,
			ProposalType: item.Proposal.ProposalType,
			CreatedBy:    item.Proposal.CreatedBy,
			CreatedAt:    item.Proposal.CreatedAt.Format(time.RFC3339),
			Age:          formatDuration(item.Age),
		})
	}

	return report
}

// formatDuration formats a duration as a human-readable string.
func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24

	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}
	minutes := int(d.Minutes())
	if minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	return "just now"
}
