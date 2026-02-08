package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexGladkov/guardian-cli/internal/config"
	"github.com/AlexGladkov/guardian-cli/internal/engine"
	"github.com/AlexGladkov/guardian-cli/internal/git"
	"github.com/AlexGladkov/guardian-cli/internal/llm"
	"github.com/AlexGladkov/guardian-cli/internal/output"
)

const checkUsage = `Usage: guardian check [base..head] [--json]

Check code changes against configured rules.

The diff range is determined by (in priority order):
  1. Explicit argument (e.g., HEAD~3..HEAD)
  2. CI auto-detection (GitHub Actions, GitLab CI)
  3. Default: origin/main..HEAD

Flags:
  --json     Output results as JSON
  --help     Show this help message

Exit codes:
  0  All checks passed
  1  Violations found
  2  Error occurred
`

func runCheck(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	jsonOutput := fs.Bool("json", false, "Output results as JSON")
	fs.Usage = func() { fmt.Fprint(os.Stderr, checkUsage) }

	if err := fs.Parse(args); err != nil {
		return 2
	}

	// Determine diff range.
	diffRange := determineDiffRange(fs.Args())

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

	// Load rules.
	rulesFile, err := loadRulesFrom(agreementsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: loading rules: %v\n", err)
		return 2
	}

	// Load exceptions.
	exceptions, err := loadAllExceptionsFrom(agreementsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: loading exceptions: %v\n", err)
		return 2
	}

	// Dereference exception pointers to values.
	exceptionValues := make([]config.Exception, 0, len(exceptions))
	for _, e := range exceptions {
		if e != nil {
			exceptionValues = append(exceptionValues, *e)
		}
	}

	// Get diff.
	diffResult, err := git.GetDiff(diffRange)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: getting diff: %v\n", err)
		return 2
	}

	// Handle empty diff.
	if len(diffResult.ChangedFiles) == 0 {
		fmt.Fprintln(os.Stdout, "No changes found. Try specifying a range: guardian check HEAD~3..HEAD")
		return 0
	}

	// Run engine checks.
	eng := engine.NewEngine(rulesFile.Rules, exceptionValues)
	engineResult, err := eng.Run(diffResult.ChangedFiles, diffResult.DiffContent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: running checks: %v\n", err)
		return 2
	}

	// Run meta check: look for protected .agreements/ file changes.
	metaChecker := &engine.MetaChecker{}
	proposalsDir := filepath.Join(agreementsDir, "proposals")
	metaViolations, err := metaChecker.Check(diffResult.ChangedFiles, proposalsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: running meta check: %v\n", err)
		return 2
	}

	// Merge meta violations into engine result.
	allViolations := append(engineResult.Violations, metaViolations...)
	for _, mv := range metaViolations {
		if mv.Severity == "error" {
			engineResult.Errors++
		} else {
			engineResult.Warnings++
		}
	}

	// Try LLM analysis for explanations (non-fatal if unavailable).
	llmExplanations := tryLLMAnalysis(constitution, rulesFile, diffResult.DiffContent, allViolations)

	// Build report.
	report := buildCheckReport(allViolations, engineResult.Errors, engineResult.Warnings, llmExplanations)

	// Output report.
	if *jsonOutput {
		if err := output.PrintCheckReportJSON(os.Stdout, report); err != nil {
			fmt.Fprintf(os.Stderr, "Error: writing JSON output: %v\n", err)
			return 2
		}
	} else {
		output.PrintCheckReportHuman(os.Stdout, report)
	}

	// Exit code based on results.
	if engineResult.Errors > 0 {
		return 1
	}
	return 0
}

// determineDiffRange resolves the diff range from arguments, CI, or default.
func determineDiffRange(positionalArgs []string) string {
	// 1. Explicit argument.
	if len(positionalArgs) > 0 && positionalArgs[0] != "" {
		return positionalArgs[0]
	}

	// 2. CI auto-detection.
	ci := git.DetectCI()
	if ci.Detected && ci.BaseRef != "" {
		head := ci.HeadRef
		if head == "" {
			head = "HEAD"
		}
		return fmt.Sprintf("origin/%s..%s", ci.BaseRef, head)
	}

	// 3. Default.
	return "origin/main..HEAD"
}

// tryLLMAnalysis attempts to get LLM explanations for violations.
// Returns a map of rule_id -> explanation. Returns nil on any error.
func tryLLMAnalysis(
	constitution *config.Constitution,
	rulesFile *config.RulesFile,
	diffContent string,
	violations []engine.Violation,
) map[string]string {
	if len(violations) == 0 {
		return nil
	}

	if constitution.LLM.Provider == "" {
		return nil
	}

	client, err := llm.NewClient(constitution.LLM)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: LLM unavailable: %v\n", err)
		return nil
	}

	// Convert engine violations to llm violations.
	llmViolations := make([]llm.Violation, 0, len(violations))
	for _, v := range violations {
		llmViolations = append(llmViolations, llm.Violation{
			RuleID:      v.RuleID,
			Severity:    v.Severity,
			Description: v.Description,
			FilePath:    v.FilePath,
			DiffSnippet: v.DiffSnippet,
		})
	}

	analysis, err := client.AnalyzeCheck(diffContent, rulesFile.Rules, llmViolations)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: LLM analysis failed: %v\n", err)
		return nil
	}

	return analysis.Explanations
}

// buildCheckReport converts engine violations into an output report.
func buildCheckReport(
	violations []engine.Violation,
	errorCount int,
	warningCount int,
	llmExplanations map[string]string,
) *output.CheckReport {
	report := &output.CheckReport{
		Summary: output.ReportSummary{
			Errors:   errorCount,
			Warnings: warningCount,
			Passed:   errorCount == 0,
		},
	}

	for _, v := range violations {
		vr := output.ViolationReport{
			RuleID:      v.RuleID,
			Severity:    v.Severity,
			Description: v.Description,
			FilePath:    v.FilePath,
			DiffSnippet: v.DiffSnippet,
		}

		// Add LLM explanation if available.
		if llmExplanations != nil {
			if explanation, ok := llmExplanations[v.RuleID]; ok {
				vr.LLMExplanation = explanation
			}
		}

		// Also use the engine's LLM explanation if it was set directly.
		if vr.LLMExplanation == "" && v.LLMExplanation != "" {
			vr.LLMExplanation = v.LLMExplanation
		}

		report.Violations = append(report.Violations, vr)
	}

	return report
}

// isAgreementsFile checks if a file path is within the .agreements directory.
func isAgreementsFile(path string) bool {
	normalized := filepath.ToSlash(path)
	return strings.HasPrefix(normalized, ".agreements/")
}
