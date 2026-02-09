package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/AlexGladkov/guardian-cli/internal/governance"
	"github.com/AlexGladkov/guardian-cli/internal/output"
)

const tallyUsage = `Usage: guardian tally <proposal_id> [--json]

Show the voting tally for a proposal.

Arguments:
  proposal_id    The ID of the proposal

Flags:
  --json     Output results as JSON
  --help     Show this help message

Exit codes:
  0  Success
  2  Error occurred
`

func runTally(args []string) int {
	fs := flag.NewFlagSet("tally", flag.ContinueOnError)
	jsonOutput := fs.Bool("json", false, "Output results as JSON")
	fs.Usage = func() { fmt.Fprint(os.Stderr, tallyUsage) }

	if err := fs.Parse(reorderArgs(args)); err != nil {
		return 2
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: proposal_id argument is required")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprint(os.Stderr, tallyUsage)
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
	proposal, _, err := findProposalByID(agreementsDir, proposalID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	// Load votes.
	votes, err := loadVotesForProposalFrom(agreementsDir, proposalID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: loading votes: %v\n", err)
		return 2
	}

	// Compute tally.
	tally := governance.ComputeTally(proposal, votes, constitution)

	// Build report.
	report := buildTallyReport(tally)

	// Output.
	if *jsonOutput {
		if err := output.PrintTallyReportJSON(os.Stdout, report); err != nil {
			fmt.Fprintf(os.Stderr, "Error: writing JSON output: %v\n", err)
			return 2
		}
	} else {
		output.PrintTallyReportHuman(os.Stdout, report)
	}

	return 0
}

// buildTallyReport converts a governance TallyResult into an output TallyReport.
func buildTallyReport(tally *governance.TallyResult) *output.TallyReport {
	report := &output.TallyReport{
		ProposalID:     tally.ProposalID,
		RuleID:         tally.RuleID,
		EligibleVoters: tally.EligibleVoters,
		Result:         tally.QuorumResult.Result,
		YesCount:       tally.QuorumResult.YesVotes,
		NoCount:        tally.QuorumResult.NoVotes,
		Required:       tally.QuorumResult.Required,
	}

	for _, v := range tally.Votes {
		report.Votes = append(report.Votes, output.VoteEntry{
			Email:    v.VoterEmail,
			Decision: v.Decision,
			Comment:  v.Comment,
		})
	}

	return report
}
