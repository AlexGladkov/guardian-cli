package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexGladkov/guardian-cli/internal/config"
	"github.com/AlexGladkov/guardian-cli/internal/discovery"
	"gopkg.in/yaml.v3"
)

// findAgreementsDir locates the .agreements/ directory or prints an error.
func findAgreementsDir() (string, error) {
	dir, err := discovery.FindAgreementsDir()
	if err != nil {
		return "", fmt.Errorf(".agreements directory not found. Run 'guardian init' first")
	}
	return dir, nil
}

// loadConstitutionFrom loads the constitution from the agreements directory.
func loadConstitutionFrom(agreementsDir string) (*config.Constitution, error) {
	path := filepath.Join(agreementsDir, "constitution.yml")
	return config.LoadConstitution(path)
}

// loadRulesFrom loads the rules from the agreements directory.
func loadRulesFrom(agreementsDir string) (*config.RulesFile, error) {
	path := filepath.Join(agreementsDir, "rules.yml")
	return config.LoadRules(path)
}

// loadAllProposalsFrom loads all proposals from the agreements directory.
func loadAllProposalsFrom(agreementsDir string) ([]*config.Proposal, error) {
	dir := filepath.Join(agreementsDir, "proposals")
	return config.LoadAllProposals(dir)
}

// loadAllExceptionsFrom loads all exceptions from the agreements directory.
func loadAllExceptionsFrom(agreementsDir string) ([]*config.Exception, error) {
	dir := filepath.Join(agreementsDir, "exceptions")
	return config.LoadAllExceptions(dir)
}

// loadVotesForProposalFrom loads all votes for a proposal from the agreements directory.
func loadVotesForProposalFrom(agreementsDir string, proposalID string) ([]*config.Vote, error) {
	dir := filepath.Join(agreementsDir, "votes")
	return config.LoadVotesForProposal(dir, proposalID)
}

// findProposalByID searches all proposals for one matching the given ID.
func findProposalByID(agreementsDir string, proposalID string) (*config.Proposal, string, error) {
	proposalsDir := filepath.Join(agreementsDir, "proposals")
	entries, err := os.ReadDir(proposalsDir)
	if err != nil {
		return nil, "", fmt.Errorf("reading proposals directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".yml" && ext != ".yaml" {
			continue
		}
		path := filepath.Join(proposalsDir, entry.Name())
		p, loadErr := config.LoadProposal(path)
		if loadErr != nil {
			continue
		}
		if p.ID == proposalID {
			return p, path, nil
		}
	}

	return nil, "", fmt.Errorf("proposal %q not found", proposalID)
}

// marshalToYAML marshals a value to YAML bytes.
func marshalToYAML(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

// promptLine reads a single line of input from stdin with a prompt.
func promptLine(prompt string) (string, error) {
	fmt.Fprint(os.Stdout, prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("reading input: %w", err)
		}
		return "", fmt.Errorf("no input received")
	}
	return strings.TrimSpace(scanner.Text()), nil
}

// promptMultiLine reads multiple lines until an empty line.
func promptMultiLine(prompt string) (string, error) {
	fmt.Fprintln(os.Stdout, prompt)
	fmt.Fprintln(os.Stdout, "(Enter an empty line to finish)")
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			break
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading input: %w", err)
	}
	return strings.Join(lines, "\n"), nil
}

// repoRoot returns the parent directory of the .agreements/ directory,
// which is the repository root.
func repoRoot(agreementsDir string) string {
	return filepath.Dir(agreementsDir)
}

// printGitHint prints a hint to add and commit files.
func printGitHint(paths ...string) {
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "Next steps:")
	for _, p := range paths {
		fmt.Fprintf(os.Stdout, "  git add %s\n", p)
	}
	fmt.Fprintln(os.Stdout, "  git commit -m \"guardian: update agreements\"")
}
