// Package cli implements the command-line interface for the Guardian CLI tool.
// It uses Go's standard os.Args and flag packages for argument parsing and
// routes commands to their respective handlers.
package cli

import (
	"fmt"
	"os"
)

const version = "0.1.0"

const usageText = `Guardian - Constitutional governance for code

Usage:
  guardian <command> [arguments]

Commands:
  init             Initialize .agreements/ directory
  check            Check code changes against rules
  propose          Create a proposal to change a rule
  vote             Vote on a proposal
  tally            Show voting tally for a proposal
  finalize         Finalize an accepted proposal
  withdraw         Withdraw a proposal
  inbox            Show proposals awaiting your vote
  hooks            Install or uninstall git hooks
  history          Show finalized proposal history
  exception        Manage rule exceptions
  llm              LLM configuration management

Flags:
  --help           Show this help message
  --version        Show version information

Run 'guardian <command> --help' for more information on a command.
`

// Run parses the command-line arguments and dispatches to the appropriate
// handler. It returns an exit code: 0 for success, 1 for logical failure
// (e.g., violations found, not allowed), and 2 for errors.
func Run(args []string) int {
	if len(args) == 0 {
		fmt.Fprint(os.Stdout, usageText)
		return 0
	}

	command := args[0]
	commandArgs := args[1:]

	switch command {
	case "--help", "-h", "help":
		fmt.Fprint(os.Stdout, usageText)
		return 0
	case "--version", "-v", "version":
		fmt.Fprintf(os.Stdout, "guardian %s\n", version)
		return 0
	case "init":
		return runInit(commandArgs)
	case "check":
		return runCheck(commandArgs)
	case "propose":
		return runPropose(commandArgs)
	case "vote":
		return runVote(commandArgs)
	case "tally":
		return runTally(commandArgs)
	case "finalize":
		return runFinalize(commandArgs)
	case "withdraw":
		return runWithdraw(commandArgs)
	case "inbox":
		return runInbox(commandArgs)
	case "hooks":
		return runHooks(commandArgs)
	case "history":
		return runHistory(commandArgs)
	case "exception":
		return runException(commandArgs)
	case "llm":
		return runLLM(commandArgs)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command %q\n\n", command)
		fmt.Fprint(os.Stderr, usageText)
		return 2
	}
}
