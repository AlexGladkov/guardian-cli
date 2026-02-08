package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/AlexGladkov/guardian-cli/internal/config"
	"github.com/AlexGladkov/guardian-cli/internal/governance"
)

const finalizeUsage = `Usage: guardian finalize <proposal_id>

Finalize an accepted proposal. Any eligible voter can finalize.

The proposal must have achieved ACCEPTED status via tally before it
can be finalized. Finalization updates the proposal status to "accepted"
and creates a history record.

Arguments:
  proposal_id    The ID of the proposal to finalize

Flags:
  --help     Show this help message

Exit codes:
  0  Proposal finalized successfully
  1  Proposal not accepted (cannot finalize)
  2  Error occurred
`

func runFinalize(args []string) int {
	fs := flag.NewFlagSet("finalize", flag.ContinueOnError)
	fs.Usage = func() { fmt.Fprint(os.Stderr, finalizeUsage) }

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: proposal_id argument is required")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprint(os.Stderr, finalizeUsage)
		return 2
	}

	proposalID := fs.Arg(0)

	// Find .agreements directory.
	agreementsDir, err := findAgreementsDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	// Load constitution.
	constitution, err := loadConstitutionFrom(agreementsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: loading constitution: %v\n", err)
		return 2
	}

	// Find proposal.
	proposal, proposalPath, err := findProposalByID(agreementsDir, proposalID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	// Check proposal is still active.
	if proposal.Status != "proposed" {
		fmt.Fprintf(os.Stderr, "Error: proposal %q has status %q, not \"proposed\"\n", proposalID, proposal.Status)
		return 1
	}

	// Load votes.
	votes, err := loadVotesForProposalFrom(agreementsDir, proposalID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: loading votes: %v\n", err)
		return 2
	}

	// Compute tally.
	tally := governance.ComputeTally(proposal, votes, constitution)

	// Only finalize if ACCEPTED.
	if tally.QuorumResult.Result != "ACCEPTED" {
		fmt.Fprintf(os.Stderr, "Error: proposal %q has tally result %q, not ACCEPTED\n", proposalID, tally.QuorumResult.Result)
		fmt.Fprintf(os.Stderr, "  Yes: %d / No: %d / Required: %d\n",
			tally.QuorumResult.YesVotes, tally.QuorumResult.NoVotes, tally.QuorumResult.Required)
		return 1
	}

	// Update proposal status.
	proposal.Status = "accepted"
	if err := saveProposalAtPath(proposalPath, proposal); err != nil {
		fmt.Fprintf(os.Stderr, "Error: saving proposal: %v\n", err)
		return 2
	}

	// Create history record.
	historyPath := filepath.Join(agreementsDir, "history", proposalID+".md")
	historyContent := buildHistoryContent(proposal, tally)
	if err := os.MkdirAll(filepath.Dir(historyPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: creating history directory: %v\n", err)
		return 2
	}
	if err := os.WriteFile(historyPath, []byte(historyContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error: writing history file: %v\n", err)
		return 2
	}

	fmt.Fprintf(os.Stdout, "Proposal %s finalized as ACCEPTED.\n", proposalID)
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "The proposal has been approved. You may now apply the changes:")

	switch proposal.ProposalType {
	case "modify":
		fmt.Fprintf(os.Stdout, "  Edit .agreements/rules.yml to modify rule %q as described in the proposal.\n", proposal.RuleID)
	case "add":
		fmt.Fprintf(os.Stdout, "  Add new rule %q to .agreements/rules.yml as described in the proposal.\n", proposal.RuleID)
	case "remove":
		fmt.Fprintf(os.Stdout, "  Remove rule %q from .agreements/rules.yml.\n", proposal.RuleID)
	}

	printGitHint(proposalPath, historyPath)

	return 0
}

// saveProposalAtPath saves a proposal to the given path using config.SaveProposal.
func saveProposalAtPath(path string, p *config.Proposal) error {
	return config.SaveProposal(path, p)
}

// buildHistoryContent creates a markdown history record for a finalized proposal.
func buildHistoryContent(p *config.Proposal, tally *governance.TallyResult) string {
	now := time.Now().UTC()
	content := fmt.Sprintf("# %s\n\n", p.ID)
	content += fmt.Sprintf("- **Rule:** %s\n", p.RuleID)
	content += fmt.Sprintf("- **Type:** %s\n", p.ProposalType)
	content += fmt.Sprintf("- **Created by:** %s\n", p.CreatedBy)
	content += fmt.Sprintf("- **Created at:** %s\n", p.CreatedAt.Format(time.RFC3339))
	content += fmt.Sprintf("- **Finalized at:** %s\n", now.Format(time.RFC3339))
	content += fmt.Sprintf("- **Result:** ACCEPTED (Yes: %d, No: %d, Required: %d)\n",
		tally.QuorumResult.YesVotes, tally.QuorumResult.NoVotes, tally.QuorumResult.Required)
	content += "\n## Change\n\n"
	content += fmt.Sprintf("%s\n", p.Change.Description)
	if p.Change.Details != "" {
		content += fmt.Sprintf("\n%s\n", p.Change.Details)
	}
	content += "\n## Reason\n\n"
	content += fmt.Sprintf("%s\n", p.Reason)
	if p.Impact != "" {
		content += "\n## Impact\n\n"
		content += fmt.Sprintf("%s\n", p.Impact)
	}
	return content
}
