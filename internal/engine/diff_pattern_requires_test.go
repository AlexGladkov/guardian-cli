package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiffPatternRequires_PatternPresent(t *testing.T) {
	diff := `diff --git a/sdk/public/api.go b/sdk/public/api.go
--- a/sdk/public/api.go
+++ b/sdk/public/api.go
@@ -1,3 +1,5 @@
 package api
+// RFC: Adding new endpoint for user profile
+func GetProfile() {}`

	checker := &DiffPatternRequiresChecker{}
	ctx := &CheckContext{
		ChangedFiles: []string{"sdk/public/api.go"},
		DiffContent:  diff,
		RuleConfig: map[string]interface{}{
			"required_regexes": []interface{}{"RFC:"},
			"only_in_paths":    []interface{}{"sdk/public/**"},
		},
		Severity: "error",
		RuleID:   "public_api_stability",
		RuleDesc: "Public API changes require RFC tag",
	}

	violations, err := checker.Check(ctx)
	require.NoError(t, err)
	assert.Empty(t, violations, "should have no violations when required pattern is present")
}

func TestDiffPatternRequires_PatternMissing(t *testing.T) {
	diff := `diff --git a/sdk/public/api.go b/sdk/public/api.go
--- a/sdk/public/api.go
+++ b/sdk/public/api.go
@@ -1,3 +1,4 @@
 package api
+func GetProfile() {}

 func main() {}`

	checker := &DiffPatternRequiresChecker{}
	ctx := &CheckContext{
		ChangedFiles: []string{"sdk/public/api.go"},
		DiffContent:  diff,
		RuleConfig: map[string]interface{}{
			"required_regexes": []interface{}{"RFC:"},
			"only_in_paths":    []interface{}{"sdk/public/**"},
		},
		Severity: "error",
		RuleID:   "public_api_stability",
		RuleDesc: "Public API changes require RFC tag",
	}

	violations, err := checker.Check(ctx)
	require.NoError(t, err)

	assert.Len(t, violations, 1)
	assert.Equal(t, "public_api_stability", violations[0].RuleID)
	assert.Equal(t, "error", violations[0].Severity)
}

func TestDiffPatternRequires_NoMatchingFiles(t *testing.T) {
	diff := `diff --git a/internal/service.go b/internal/service.go
--- a/internal/service.go
+++ b/internal/service.go
@@ -1,2 +1,3 @@
 package internal
+func DoSomething() {}`

	checker := &DiffPatternRequiresChecker{}
	ctx := &CheckContext{
		ChangedFiles: []string{"internal/service.go"},
		DiffContent:  diff,
		RuleConfig: map[string]interface{}{
			"required_regexes": []interface{}{"RFC:"},
			"only_in_paths":    []interface{}{"sdk/public/**"},
		},
		Severity: "error",
		RuleID:   "public_api_stability",
		RuleDesc: "Public API changes require RFC tag",
	}

	violations, err := checker.Check(ctx)
	require.NoError(t, err)
	assert.Empty(t, violations, "should have no violations when no files match only_in_paths")
}

func TestDiffPatternRequires_MultipleRequiredRegexes(t *testing.T) {
	diff := `diff --git a/sdk/public/api.go b/sdk/public/api.go
--- a/sdk/public/api.go
+++ b/sdk/public/api.go
@@ -1,3 +1,4 @@
 package api
+// BREAKING: removing old endpoint

 func main() {}`

	checker := &DiffPatternRequiresChecker{}
	ctx := &CheckContext{
		ChangedFiles: []string{"sdk/public/api.go"},
		DiffContent:  diff,
		RuleConfig: map[string]interface{}{
			"required_regexes": []interface{}{"RFC:", "BREAKING:"},
			"only_in_paths":    []interface{}{"sdk/public/**"},
		},
		Severity: "error",
		RuleID:   "public_api_stability",
		RuleDesc: "Public API changes require RFC or BREAKING tag",
	}

	violations, err := checker.Check(ctx)
	require.NoError(t, err)
	assert.Empty(t, violations, "should pass because BREAKING: matches one of the required regexes")
}

func TestDiffPatternRequires_InvalidRegex(t *testing.T) {
	checker := &DiffPatternRequiresChecker{}
	ctx := &CheckContext{
		ChangedFiles: []string{"sdk/public/api.go"},
		DiffContent: `diff --git a/sdk/public/api.go b/sdk/public/api.go
--- a/sdk/public/api.go
+++ b/sdk/public/api.go
@@ -1,2 +1,3 @@
 package api
+func Test() {}`,
		RuleConfig: map[string]interface{}{
			"required_regexes": []interface{}{"[invalid"},
			"only_in_paths":    []interface{}{"sdk/public/**"},
		},
		Severity: "error",
		RuleID:   "test",
		RuleDesc: "test",
	}

	_, err := checker.Check(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid regex")
}

func TestDiffPatternRequires_Type(t *testing.T) {
	checker := &DiffPatternRequiresChecker{}
	assert.Equal(t, "diff_pattern_requires", checker.Type())
}

func TestDiffPatternRequires_MissingRequiredRegexes(t *testing.T) {
	checker := &DiffPatternRequiresChecker{}
	ctx := &CheckContext{
		ChangedFiles: []string{"sdk/public/api.go"},
		DiffContent:  "",
		RuleConfig: map[string]interface{}{
			"only_in_paths": []interface{}{"sdk/public/**"},
		},
		Severity: "error",
		RuleID:   "test",
		RuleDesc: "test",
	}

	_, err := checker.Check(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required_regexes")
}

func TestDiffPatternRequires_MissingOnlyInPaths(t *testing.T) {
	checker := &DiffPatternRequiresChecker{}
	ctx := &CheckContext{
		ChangedFiles: []string{"test.go"},
		DiffContent:  "",
		RuleConfig: map[string]interface{}{
			"required_regexes": []interface{}{"RFC:"},
		},
		Severity: "error",
		RuleID:   "test",
		RuleDesc: "test",
	}

	_, err := checker.Check(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only_in_paths")
}

func TestDiffPatternRequires_PatternInDifferentFile(t *testing.T) {
	// The required pattern is found in any file's added lines, not just the
	// matching files. This tests the spec: "Parse all added lines in diff"
	// and "Check if at least one required_regex matches somewhere in the diff".
	diff := `diff --git a/sdk/public/api.go b/sdk/public/api.go
--- a/sdk/public/api.go
+++ b/sdk/public/api.go
@@ -1,2 +1,3 @@
 package api
+func NewEndpoint() {}
diff --git a/docs/changelog.md b/docs/changelog.md
--- a/docs/changelog.md
+++ b/docs/changelog.md
@@ -1,2 +1,3 @@
 # Changelog
+RFC: Added new endpoint`

	checker := &DiffPatternRequiresChecker{}
	ctx := &CheckContext{
		ChangedFiles: []string{"sdk/public/api.go", "docs/changelog.md"},
		DiffContent:  diff,
		RuleConfig: map[string]interface{}{
			"required_regexes": []interface{}{"RFC:"},
			"only_in_paths":    []interface{}{"sdk/public/**"},
		},
		Severity: "error",
		RuleID:   "public_api_stability",
		RuleDesc: "test",
	}

	violations, err := checker.Check(ctx)
	require.NoError(t, err)
	assert.Empty(t, violations, "RFC: found in diff (in changelog), so no violation")
}
