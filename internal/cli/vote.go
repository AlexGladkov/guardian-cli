package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlexGladkov/guardian-cli/internal/config"
	"github.com/AlexGladkov/guardian-cli/internal/git"
	"github.com/AlexGladkov/guardian-cli/internal/governance"
)

const voteUsage = `Usage: guardian vote <proposal_id> --yes|--no [--comment "..."]

Vote on an active proposal.

Arguments:
  proposal_id    The ID of the proposal to vote on

Flags:
  --yes          Vote yes
  --no           Vote no
  --comment      Add a comment to the vote
  --help         Show this help message

Exit codes:
  0  Vote recorded successfully
  1  Not allowed (not a voter, self-approval, or duplicate vote)
  2  Error occurred
`

func runVote(args []string) int {
	fs := flag.NewFlagSet("vote", flag.ContinueOnError)
	voteYes := fs.Bool("yes", false, "Vote yes")
	voteNo := fs.Bool("no", false, "Vote no")
	comment := fs.String("comment", "", "Vote comment")
	fs.Usage = func() { fmt.Fprint(os.Stderr, voteUsage) }

	// Reorder args so flags work in any position (before or after proposal_id).
	if err := fs.Parse(reorderArgs(args)); err != nil {
		return 2
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: proposal_id argument is required")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprint(os.Stderr, voteUsage)
		return 2
	}

	proposalID := fs.Arg(0)

	// Validate vote decision.
	if !*voteYes && !*voteNo {
		fmt.Fprintln(os.Stderr, "Error: specify --yes or --no")
		return 2
	}
	if *voteYes && *voteNo {
		fmt.Fprintln(os.Stderr, "Error: cannot specify both --yes and --no")
		return 2
	}

	decision := "yes"
	if *voteNo {
		decision = "no"
	}

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
	proposal, _, err := findProposalByID(agreementsDir, proposalID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	// Verify proposal is still active.
	if proposal.Status != "proposed" {
		fmt.Fprintf(os.Stderr, "Error: proposal %q has status %q, not \"proposed\"\n", proposalID, proposal.Status)
		return 1
	}

	// Get voter email.
	email, err := git.GetUserEmail()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	// Check if voter is eligible.
	if !governance.IsVoter(constitution, email) {
		fmt.Fprintf(os.Stderr, "Error: %s is not an eligible voter\n", email)
		return 1
	}

	// Check forbid_self_approval: proposal author cannot vote yes.
	if constitution.Governance.ForbidSelfApproval && proposal.CreatedBy == email && decision == "yes" {
		fmt.Fprintln(os.Stderr, "Error: self-approval is forbidden; you cannot vote yes on your own proposal")
		return 1
	}

	// Check for existing vote.
	existingVotes, err := loadVotesForProposalFrom(agreementsDir, proposalID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: loading votes: %v\n", err)
		return 2
	}

	for _, v := range existingVotes {
		if v.VoterEmail == email {
			if !constitution.Governance.AllowVoteChange {
				fmt.Fprintln(os.Stderr, "Error: you have already voted on this proposal and vote changes are not allowed")
				return 1
			}
			// Allow vote change - will overwrite the file.
			break
		}
	}

	// Build vote.
	vote := &config.Vote{
		ProposalID: proposalID,
		VoterEmail: email,
		Decision:   decision,
		Comment:    *comment,
		VotedAt:    time.Now().UTC(),
	}

	// Validate vote.
	if err := config.ValidateVote(vote); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid vote: %v\n", err)
		return 2
	}

	// Save vote.
	// Sanitize email for filename: replace @ and . characters.
	sanitizedEmail := strings.ReplaceAll(email, "@", "_at_")
	sanitizedEmail = strings.ReplaceAll(sanitizedEmail, ".", "_")
	votePath := filepath.Join(agreementsDir, "votes", proposalID, sanitizedEmail+".yml")
	if err := config.SaveVote(votePath, vote); err != nil {
		fmt.Fprintf(os.Stderr, "Error: saving vote: %v\n", err)
		return 2
	}

	fmt.Fprintf(os.Stdout, "Vote recorded: %s on proposal %s\n", strings.ToUpper(decision), proposalID)
	fmt.Fprintf(os.Stdout, "File: %s\n", votePath)
	printGitHint(votePath)

	return 0
}
