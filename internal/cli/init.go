package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlexGladkov/guardian-cli/internal/config"
	"github.com/AlexGladkov/guardian-cli/internal/llm"
)

const initUsage = `Usage: guardian init [--force]

Initialize a new .agreements/ directory in the current working directory.

Creates:
  .agreements/constitution.yml  - Team governance configuration
  .agreements/rules.yml         - Rule definitions
  .agreements/proposals/        - Proposal storage
  .agreements/votes/            - Vote storage
  .agreements/history/          - History storage
  .agreements/exceptions/       - Exception storage

Flags:
  --force    Overwrite existing .agreements/ directory
  --help     Show this help message
`

func newTemplateConstitution() *config.Constitution {
	return &config.Constitution{
		Governance: config.Governance{
			Voters: []config.VoterRef{
				{Role: "maintainers"},
			},
			Quorum: config.QuorumConfig{
				Type: "majority",
			},
			ForbidSelfApproval: true,
			AllowVoteChange:    false,
			ProposalTTLDays:    30,
			Exceptions: config.ExceptionPolicy{
				RequireApproval: false,
			},
		},
		Identity: config.Identity{
			AllowedDomains:       []string{"example.com"},
			RequireSignedCommits: false,
		},
		Roles: map[string]config.Role{
			"maintainers": {
				Members: []config.RoleMember{
					{Email: "you@example.com"},
				},
			},
		},
		LLM: config.LLMConfig{},
	}
}

func newTemplateRules() *config.RulesFile {
	return &config.RulesFile{
		Rules: []config.Rule{
			{
				ID:          "example_rule",
				Description: "Example: forbid TODO comments in production code",
				Type:        "diff_pattern_forbidden",
				Config: map[string]interface{}{
					"forbidden_regexes": []interface{}{"TODO"},
					"only_in_paths":     []interface{}{"src/**"},
				},
				Severity: "warning",
			},
		},
	}
}

func runInit(args []string) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	force := fs.Bool("force", false, "Overwrite existing .agreements/ directory")
	fs.Usage = func() { fmt.Fprint(os.Stderr, initUsage) }

	if err := fs.Parse(args); err != nil {
		return 2
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	agreementsDir := filepath.Join(cwd, ".agreements")

	// Check if directory already exists.
	if info, statErr := os.Stat(agreementsDir); statErr == nil && info.IsDir() {
		if !*force {
			fmt.Fprintln(os.Stderr, "Error: .agreements/ directory already exists. Use --force to overwrite.")
			return 2
		}
	}

	// Create directory structure.
	subdirs := []string{
		"proposals",
		"votes",
		"history",
		"exceptions",
	}

	if err := os.MkdirAll(agreementsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: creating .agreements directory: %v\n", err)
		return 2
	}

	for _, sub := range subdirs {
		dir := filepath.Join(agreementsDir, sub)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error: creating %s directory: %v\n", sub, err)
			return 2
		}
	}

	// Write template constitution.
	constitutionPath := filepath.Join(agreementsDir, "constitution.yml")
	tmplConstitution := newTemplateConstitution()
	if err := config.SaveConstitution(constitutionPath, tmplConstitution); err != nil {
		fmt.Fprintf(os.Stderr, "Error: writing constitution.yml: %v\n", err)
		return 2
	}

	// Write template rules.
	rulesPath := filepath.Join(agreementsDir, "rules.yml")
	rulesData, err := marshalToYAML(newTemplateRules())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: marshaling rules.yml: %v\n", err)
		return 2
	}
	if err := os.WriteFile(rulesPath, rulesData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error: writing rules.yml: %v\n", err)
		return 2
	}

	fmt.Fprintln(os.Stdout, "Initialized .agreements/ directory.")
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "Created:")
	fmt.Fprintln(os.Stdout, "  .agreements/constitution.yml")
	fmt.Fprintln(os.Stdout, "  .agreements/rules.yml")
	fmt.Fprintln(os.Stdout, "  .agreements/proposals/")
	fmt.Fprintln(os.Stdout, "  .agreements/votes/")
	fmt.Fprintln(os.Stdout, "  .agreements/history/")
	fmt.Fprintln(os.Stdout, "  .agreements/exceptions/")
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "Next steps:")
	fmt.Fprintln(os.Stdout, "  1. Edit .agreements/constitution.yml to configure governance")
	fmt.Fprintln(os.Stdout, "  2. Edit .agreements/rules.yml to define your rules")
	fmt.Fprintln(os.Stdout, "  3. Run 'guardian hooks install' to set up git hooks")

	// Prompt for LLM configuration if not already configured.
	if tmplConstitution.LLM.Provider == "" {
		fmt.Fprintln(os.Stdout, "")
		fmt.Fprintln(os.Stdout, "LLM is not configured. Setting up LLM provider...")
		fmt.Fprintln(os.Stdout, "")

		llmCfg, llmErr := llm.RunConfigure(os.Stdin, os.Stdout)
		if llmErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: LLM configuration skipped: %v\n", llmErr)
			fmt.Fprintln(os.Stderr, "You can configure it later with: guardian llm configure")
			return 0
		}

		// Update constitution with LLM config.
		constitution, loadErr := config.LoadConstitution(constitutionPath)
		if loadErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not reload constitution: %v\n", loadErr)
			return 0
		}
		constitution.LLM = *llmCfg
		if saveErr := config.SaveConstitution(constitutionPath, constitution); saveErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not save LLM config: %v\n", saveErr)
		}
	}

	return 0
}
