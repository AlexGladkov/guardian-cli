package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlexGladkov/guardian-cli/internal/config"
	"github.com/AlexGladkov/guardian-cli/internal/llm"
)

const llmUsage = `Usage: guardian llm configure

Configure the LLM provider for Guardian.

Subcommands:
  configure    Interactive LLM provider configuration

Flags:
  --help     Show this help message

Exit codes:
  0  Success
  2  Error occurred
`

func runLLM(args []string) int {
	fs := flag.NewFlagSet("llm", flag.ContinueOnError)
	fs.Usage = func() { fmt.Fprint(os.Stderr, llmUsage) }

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: subcommand required (configure)")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprint(os.Stderr, llmUsage)
		return 2
	}

	subcommand := fs.Arg(0)

	switch subcommand {
	case "configure":
		return runLLMConfigure()
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown llm subcommand %q; use configure\n", subcommand)
		return 2
	}
}

func runLLMConfigure() int {
	// Find .agreements directory.
	agreementsDir, err := findAgreementsDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	// Run interactive configuration.
	llmCfg, err := llm.RunConfigure(os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: LLM configuration failed: %v\n", err)
		return 2
	}

	// Load existing constitution.
	constitutionPath := filepath.Join(agreementsDir, "constitution.yml")
	constitution, err := config.LoadConstitution(constitutionPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: loading constitution: %v\n", err)
		return 2
	}

	// Update LLM config.
	constitution.LLM = *llmCfg

	// Save constitution.
	if err := config.SaveConstitution(constitutionPath, constitution); err != nil {
		fmt.Fprintf(os.Stderr, "Error: saving constitution: %v\n", err)
		return 2
	}

	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "LLM configuration saved to .agreements/constitution.yml")

	return 0
}
