package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/AlexGladkov/guardian-cli/internal/config"
	"github.com/AlexGladkov/guardian-cli/internal/git"
)

const withdrawUsage = `Usage: guardian withdraw <proposal_id>

Withdraw a proposal. Only the original author can withdraw their proposal.

Arguments:
  proposal_id    The ID of the proposal to withdraw

Flags:
  --help     Show this help message

Exit codes:
  0  Proposal withdrawn successfully
  1  Not the author (cannot withdraw)
  2  Error occurred
`

func runWithdraw(args []string) int {
	fs := flag.NewFlagSet("withdraw", flag.ContinueOnError)
	fs.Usage = func() { fmt.Fprint(os.Stderr, withdrawUsage) }

	if err := fs.Parse(reorderArgs(args)); err != nil {
		return 2
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: proposal_id argument is required")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprint(os.Stderr, withdrawUsage)
		return 2
	}

	proposalID := fs.Arg(0)

	// Find .agreements directory.
	agreementsDir, err := findAgreementsDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
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

	// Get user email.
	email, err := git.GetUserEmail()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	// Only the author can withdraw.
	if proposal.CreatedBy != email {
		fmt.Fprintf(os.Stderr, "Error: only the author (%s) can withdraw this proposal\n", proposal.CreatedBy)
		return 1
	}

	// Update proposal status.
	proposal.Status = "withdrawn"
	if err := config.SaveProposal(proposalPath, proposal); err != nil {
		fmt.Fprintf(os.Stderr, "Error: saving proposal: %v\n", err)
		return 2
	}

	fmt.Fprintf(os.Stdout, "Proposal %s withdrawn.\n", proposalID)
	printGitHint(proposalPath)

	return 0
}
