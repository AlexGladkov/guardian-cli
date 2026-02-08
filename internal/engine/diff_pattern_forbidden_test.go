package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiffPatternForbidden_BasicViolation(t *testing.T) {
	diff := `diff --git a/domain/model/Price.kt b/domain/model/Price.kt
--- a/domain/model/Price.kt
+++ b/domain/model/Price.kt
@@ -1,3 +1,4 @@
 package domain.model
+val amount: Double = 0.0

 data class Price(`

	checker := &DiffPatternForbiddenChecker{}
	ctx := &CheckContext{
		ChangedFiles: []string{"domain/model/Price.kt"},
		DiffContent:  diff,
		RuleConfig: map[string]interface{}{
			"forbidden_regexes": []interface{}{`\bDouble\b`, `\bfloat\b`},
			"only_in_paths":     []interface{}{"**/*.kt"},
		},
		Severity: "warning",
		RuleID:   "money_minor_units",
		RuleDesc: "Money must use int minor units, not float/double",
	}

	violations, err := checker.Check(ctx)
	require.NoError(t, err)

	assert.Len(t, violations, 1)
	assert.Equal(t, "money_minor_units", violations[0].RuleID)
	assert.Equal(t, "warning", violations[0].Severity)
	assert.Contains(t, violations[0].DiffSnippet, "Double")
}

func TestDiffPatternForbidden_NoViolation_FileNotInPaths(t *testing.T) {
	diff := `diff --git a/readme.md b/readme.md
--- a/readme.md
+++ b/readme.md
@@ -1,2 +1,3 @@
 # Readme
+Double precision is used here`

	checker := &DiffPatternForbiddenChecker{}
	ctx := &CheckContext{
		ChangedFiles: []string{"readme.md"},
		DiffContent:  diff,
		RuleConfig: map[string]interface{}{
			"forbidden_regexes": []interface{}{`\bDouble\b`},
			"only_in_paths":     []interface{}{"**/*.kt", "**/*.java"},
		},
		Severity: "warning",
		RuleID:   "money_minor_units",
		RuleDesc: "test",
	}

	violations, err := checker.Check(ctx)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestDiffPatternForbidden_NoViolation_NoMatchingRegex(t *testing.T) {
	diff := `diff --git a/domain/model/Price.kt b/domain/model/Price.kt
--- a/domain/model/Price.kt
+++ b/domain/model/Price.kt
@@ -1,3 +1,4 @@
 package domain.model
+val amount: Long = 0L

 data class Price(`

	checker := &DiffPatternForbiddenChecker{}
	ctx := &CheckContext{
		ChangedFiles: []string{"domain/model/Price.kt"},
		DiffContent:  diff,
		RuleConfig: map[string]interface{}{
			"forbidden_regexes": []interface{}{`\bDouble\b`, `\bfloat\b`},
			"only_in_paths":     []interface{}{"**/*.kt"},
		},
		Severity: "warning",
		RuleID:   "money_minor_units",
		RuleDesc: "test",
	}

	violations, err := checker.Check(ctx)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestDiffPatternForbidden_WithoutOnlyInPaths(t *testing.T) {
	diff := `diff --git a/config.yml b/config.yml
--- a/config.yml
+++ b/config.yml
@@ -1,2 +1,3 @@
 settings:
+  password: secret123`

	checker := &DiffPatternForbiddenChecker{}
	ctx := &CheckContext{
		ChangedFiles: []string{"config.yml"},
		DiffContent:  diff,
		RuleConfig: map[string]interface{}{
			"forbidden_regexes": []interface{}{`password:\s*\S+`},
		},
		Severity: "error",
		RuleID:   "no_passwords",
		RuleDesc: "No hardcoded passwords",
	}

	violations, err := checker.Check(ctx)
	require.NoError(t, err)

	assert.Len(t, violations, 1)
	assert.Equal(t, "no_passwords", violations[0].RuleID)
}

func TestDiffPatternForbidden_MultipleViolationsInOneFile(t *testing.T) {
	diff := `diff --git a/domain/model/Price.kt b/domain/model/Price.kt
--- a/domain/model/Price.kt
+++ b/domain/model/Price.kt
@@ -1,3 +1,5 @@
 package domain.model
+val amount: Double = 0.0
+val tax: float = 0.0f

 data class Price(`

	checker := &DiffPatternForbiddenChecker{}
	ctx := &CheckContext{
		ChangedFiles: []string{"domain/model/Price.kt"},
		DiffContent:  diff,
		RuleConfig: map[string]interface{}{
			"forbidden_regexes": []interface{}{`\bDouble\b`, `\bfloat\b`},
			"only_in_paths":     []interface{}{"**/*.kt"},
		},
		Severity: "warning",
		RuleID:   "money_minor_units",
		RuleDesc: "test",
	}

	violations, err := checker.Check(ctx)
	require.NoError(t, err)

	assert.Len(t, violations, 2)
}

func TestDiffPatternForbidden_InvalidRegex(t *testing.T) {
	checker := &DiffPatternForbiddenChecker{}
	ctx := &CheckContext{
		ChangedFiles: []string{"test.kt"},
		DiffContent:  "diff content",
		RuleConfig: map[string]interface{}{
			"forbidden_regexes": []interface{}{"[invalid"},
		},
		Severity: "error",
		RuleID:   "test",
		RuleDesc: "test",
	}

	_, err := checker.Check(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid regex")
}

func TestDiffPatternForbidden_Type(t *testing.T) {
	checker := &DiffPatternForbiddenChecker{}
	assert.Equal(t, "diff_pattern_forbidden", checker.Type())
}

func TestDiffPatternForbidden_MissingConfig(t *testing.T) {
	checker := &DiffPatternForbiddenChecker{}
	ctx := &CheckContext{
		ChangedFiles: []string{"test.kt"},
		DiffContent:  "",
		RuleConfig:   map[string]interface{}{},
		Severity:     "error",
		RuleID:       "test",
		RuleDesc:     "test",
	}

	_, err := checker.Check(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden_regexes")
}

func TestDiffPatternForbidden_MultipleFilesOnlyOneMatches(t *testing.T) {
	diff := `diff --git a/src/model.kt b/src/model.kt
--- a/src/model.kt
+++ b/src/model.kt
@@ -1,2 +1,3 @@
 package src
+val x: Double = 1.0
diff --git a/docs/readme.md b/docs/readme.md
--- a/docs/readme.md
+++ b/docs/readme.md
@@ -1,2 +1,3 @@
 # Docs
+Double check this`

	checker := &DiffPatternForbiddenChecker{}
	ctx := &CheckContext{
		ChangedFiles: []string{"src/model.kt", "docs/readme.md"},
		DiffContent:  diff,
		RuleConfig: map[string]interface{}{
			"forbidden_regexes": []interface{}{`\bDouble\b`},
			"only_in_paths":     []interface{}{"**/*.kt"},
		},
		Severity: "warning",
		RuleID:   "test",
		RuleDesc: "test",
	}

	violations, err := checker.Check(ctx)
	require.NoError(t, err)

	// Only the .kt file should have a violation, not the .md file.
	assert.Len(t, violations, 1)
	assert.Equal(t, "src/model.kt", violations[0].FilePath)
}
