package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

const constitutionUsage = `Usage: guardian constitution

Display the current constitution.

Exit codes:
  0  Success
  2  Error occurred
`

func runConstitution(args []string) int {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprint(os.Stdout, constitutionUsage)
		return 0
	}

	agreementsDir, err := findAgreementsDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	path := filepath.Join(agreementsDir, "constitution.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: reading constitution: %v\n", err)
		return 2
	}

	fmt.Fprint(os.Stdout, string(data))
	return 0
}
