package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/AlexGladkov/guardian-cli/internal/git"
)

const hooksUsage = `Usage: guardian hooks <install|uninstall>

Manage Guardian git hooks. Installs or uninstalls post-merge and
post-checkout hooks that trigger 'guardian inbox --notify'.

Subcommands:
  install      Install guardian git hooks
  uninstall    Uninstall guardian git hooks

Flags:
  --help     Show this help message

Exit codes:
  0  Success
  2  Error occurred
`

func runHooks(args []string) int {
	fs := flag.NewFlagSet("hooks", flag.ContinueOnError)
	fs.Usage = func() { fmt.Fprint(os.Stderr, hooksUsage) }

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: subcommand required (install or uninstall)")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprint(os.Stderr, hooksUsage)
		return 2
	}

	subcommand := fs.Arg(0)

	switch subcommand {
	case "install":
		if err := git.InstallHooks(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: installing hooks: %v\n", err)
			return 2
		}
		fmt.Fprintln(os.Stdout, "Guardian hooks installed (post-merge, post-checkout).")
		return 0

	case "uninstall":
		if err := git.UninstallHooks(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: uninstalling hooks: %v\n", err)
			return 2
		}
		fmt.Fprintln(os.Stdout, "Guardian hooks uninstalled.")
		return 0

	default:
		fmt.Fprintf(os.Stderr, "Error: unknown hooks subcommand %q; use install or uninstall\n", subcommand)
		return 2
	}
}
