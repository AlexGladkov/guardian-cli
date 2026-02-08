package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/AlexGladkov/guardian-cli/internal/config"
	"github.com/AlexGladkov/guardian-cli/internal/git"
	"github.com/AlexGladkov/guardian-cli/internal/llm"
)

const proposeUsage = `Usage: guardian propose <rule_id> [--llm]

Create a proposal to modify, add, or remove a rule.

Arguments:
  rule_id    The ID of the rule to propose changes for

Flags:
  --llm      Use LLM to draft proposal text
  --help     Show this help message

Exit codes:
  0  Proposal created successfully
  1  Blocked (active proposal already exists for rule)
  2  Error occurred
`

func runPropose(args []string) int {
	fs := flag.NewFlagSet("propose", flag.ContinueOnError)
	useLLM := fs.Bool("llm", false, "Use LLM to draft proposal text")
	fs.Usage = func() { fmt.Fprint(os.Stderr, proposeUsage) }

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: rule_id argument is required")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprint(os.Stderr, proposeUsage)
		return 2
	}

	ruleID := fs.Arg(0)

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

	// Load rules to verify rule_id exists (unless it's a new rule being added).
	rulesFile, err := loadRulesFrom(agreementsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: loading rules: %v\n", err)
		return 2
	}

	// Check for existing active proposal for the same rule.
	proposals, err := loadAllProposalsFrom(agreementsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: loading proposals: %v\n", err)
		return 2
	}

	for _, p := range proposals {
		if p.RuleID == ruleID && p.Status == "proposed" {
			fmt.Fprintf(os.Stderr, "Error: an active proposal already exists for rule %q (proposal %s)\n", ruleID, p.ID)
			return 1
		}
	}

	// Get user email.
	email, err := git.GetUserEmail()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	var proposalType, changeDesc, changeDetails, reason, impact string

	if *useLLM {
		// Find the rule for LLM context.
		var targetRule *config.Rule
		for i := range rulesFile.Rules {
			if rulesFile.Rules[i].ID == ruleID {
				targetRule = &rulesFile.Rules[i]
				break
			}
		}

		if targetRule == nil {
			fmt.Fprintf(os.Stderr, "Error: rule %q not found in rules.yml (use interactive mode for new rules)\n", ruleID)
			return 2
		}

		client, llmErr := llm.NewClient(constitution.LLM)
		if llmErr != nil {
			fmt.Fprintf(os.Stderr, "Error: LLM unavailable: %v\n", llmErr)
			return 2
		}

		// Prompt for proposal type even in LLM mode.
		proposalType, err = promptLine("Proposal type (modify/add/remove): ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 2
		}

		context, _ := promptLine("Additional context for the LLM (optional): ")

		fmt.Fprintln(os.Stdout, "Generating proposal draft with LLM...")
		draft, draftErr := client.DraftProposal(*targetRule, context)
		if draftErr != nil {
			fmt.Fprintf(os.Stderr, "Error: LLM draft failed: %v\n", draftErr)
			return 2
		}

		changeDesc = draft.ChangeDescription
		changeDetails = draft.ChangeDetails
		reason = draft.Reason
		impact = draft.Impact

		fmt.Fprintln(os.Stdout, "")
		fmt.Fprintln(os.Stdout, "LLM draft:")
		fmt.Fprintf(os.Stdout, "  Change: %s\n", changeDesc)
		fmt.Fprintf(os.Stdout, "  Details: %s\n", changeDetails)
		fmt.Fprintf(os.Stdout, "  Reason: %s\n", reason)
		fmt.Fprintf(os.Stdout, "  Impact: %s\n", impact)
		fmt.Fprintln(os.Stdout, "")

		confirm, _ := promptLine("Accept this draft? (y/n): ")
		if confirm != "y" && confirm != "Y" && confirm != "yes" {
			fmt.Fprintln(os.Stdout, "Draft rejected. Falling back to manual input.")
			changeDesc = ""
			changeDetails = ""
			reason = ""
			impact = ""
		}
	}

	// Interactive prompts for anything not filled by LLM.
	if proposalType == "" {
		proposalType, err = promptLine("Proposal type (modify/add/remove): ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 2
		}
	}

	if proposalType != "modify" && proposalType != "add" && proposalType != "remove" {
		fmt.Fprintf(os.Stderr, "Error: invalid proposal type %q; must be modify, add, or remove\n", proposalType)
		return 2
	}

	if changeDesc == "" {
		changeDesc, err = promptLine("Change description: ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 2
		}
	}

	if changeDetails == "" {
		changeDetails, err = promptMultiLine("Change details:")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 2
		}
	}

	if reason == "" {
		reason, err = promptLine("Reason for change: ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 2
		}
	}

	if impact == "" {
		impact, err = promptLine("Expected impact: ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 2
		}
	}

	// Build proposal.
	now := time.Now().UTC()
	proposalID := fmt.Sprintf("%s-%s", now.Format("2006-01-02"), ruleID)

	proposal := &config.Proposal{
		ID:           proposalID,
		RuleID:       ruleID,
		ProposalType: proposalType,
		Change: config.ProposalChange{
			Description: changeDesc,
			Details:     changeDetails,
		},
		Reason:    reason,
		Impact:    impact,
		CreatedBy: email,
		CreatedAt: now,
		Status:    "proposed",
	}

	// Validate proposal.
	if err := config.ValidateProposal(proposal); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid proposal: %v\n", err)
		return 2
	}

	// Save proposal.
	proposalPath := filepath.Join(agreementsDir, "proposals", proposalID+".yml")
	if err := config.SaveProposal(proposalPath, proposal); err != nil {
		fmt.Fprintf(os.Stderr, "Error: saving proposal: %v\n", err)
		return 2
	}

	fmt.Fprintf(os.Stdout, "Proposal created: %s\n", proposalID)
	fmt.Fprintf(os.Stdout, "File: %s\n", proposalPath)
	printGitHint(proposalPath)

	return 0
}
