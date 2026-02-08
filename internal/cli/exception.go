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
)

const exceptionUsage = `Usage: guardian exception create <rule_id>

Create a rule exception for specific file paths.

Subcommands:
  create <rule_id>    Create an exception for a rule

Flags:
  --help     Show this help message

Exit codes:
  0  Exception created successfully
  2  Error occurred
`

func runException(args []string) int {
	fs := flag.NewFlagSet("exception", flag.ContinueOnError)
	fs.Usage = func() { fmt.Fprint(os.Stderr, exceptionUsage) }

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: subcommand required (create)")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprint(os.Stderr, exceptionUsage)
		return 2
	}

	subcommand := fs.Arg(0)

	switch subcommand {
	case "create":
		return runExceptionCreate(fs.Args()[1:])
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown exception subcommand %q; use create\n", subcommand)
		return 2
	}
}

func runExceptionCreate(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: rule_id argument is required")
		fmt.Fprintln(os.Stderr, "Usage: guardian exception create <rule_id>")
		return 2
	}

	ruleID := args[0]

	// Find .agreements directory.
	agreementsDir, err := findAgreementsDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	// Load constitution for exception policy check.
	constitution, err := loadConstitutionFrom(agreementsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: loading constitution: %v\n", err)
		return 2
	}

	// Get user email.
	email, err := git.GetUserEmail()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	// Interactive prompts.
	pathsStr, err := promptLine("File paths (comma-separated globs, e.g., src/legacy/*.go): ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	paths := splitAndTrim(pathsStr, ",")
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "Error: at least one path is required")
		return 2
	}

	reason, err := promptLine("Reason for exception: ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	expiresStr, err := promptLine("Expiration date (YYYY-MM-DD, or empty for no expiration): ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	var expiresAt *time.Time
	if expiresStr != "" {
		t, parseErr := time.Parse("2006-01-02", expiresStr)
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid date format %q; use YYYY-MM-DD\n", expiresStr)
			return 2
		}
		expiresAt = &t
	}

	// Build exception.
	now := time.Now().UTC()
	exceptionID := fmt.Sprintf("%s-%s-%s", now.Format("20060102"), ruleID, sanitizeForFilename(email))

	exception := &config.Exception{
		ID:        exceptionID,
		RuleID:    ruleID,
		Paths:     paths,
		Reason:    reason,
		CreatedBy: email,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}

	// Validate exception.
	if err := config.ValidateException(exception); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid exception: %v\n", err)
		return 2
	}

	// Save exception.
	exceptionPath := filepath.Join(agreementsDir, "exceptions", exceptionID+".yml")
	if err := config.SaveException(exceptionPath, exception); err != nil {
		fmt.Fprintf(os.Stderr, "Error: saving exception: %v\n", err)
		return 2
	}

	fmt.Fprintf(os.Stdout, "Exception created: %s\n", exceptionID)
	fmt.Fprintf(os.Stdout, "File: %s\n", exceptionPath)

	// Warn if require_approval is enabled.
	if constitution.Governance.Exceptions.RequireApproval {
		fmt.Fprintln(os.Stdout, "")
		fmt.Fprintln(os.Stdout, "NOTE: Exception approval is required by the constitution.")
		fmt.Fprintln(os.Stdout, "This exception may need additional review before taking effect.")
	}

	printGitHint(exceptionPath)

	return 0
}

// splitAndTrim splits a string by separator and trims whitespace from each part.
func splitAndTrim(s string, sep string) []string {
	parts := strings.Split(s, sep)
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// sanitizeForFilename replaces characters that are problematic in filenames.
func sanitizeForFilename(s string) string {
	s = strings.ReplaceAll(s, "@", "_at_")
	s = strings.ReplaceAll(s, ".", "_")
	return s
}
