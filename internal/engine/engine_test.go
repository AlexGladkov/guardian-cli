package engine

import (
	"testing"
	"time"

	"github.com/AlexGladkov/guardian-cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngine_RunWithNoRules(t *testing.T) {
	e := NewEngine(nil, nil)
	result, err := e.Run([]string{"test.go"}, "")
	require.NoError(t, err)
	assert.Empty(t, result.Violations)
	assert.Equal(t, 0, result.Errors)
	assert.Equal(t, 0, result.Warnings)
}

func TestEngine_RunWithImportsForbidden(t *testing.T) {
	diff := `diff --git a/domain/service/UserService.kt b/domain/service/UserService.kt
--- a/domain/service/UserService.kt
+++ b/domain/service/UserService.kt
@@ -1,3 +1,4 @@
 package domain.service
+import com.myapp.infra.database.UserRepository

 class UserService {`

	rules := []config.Rule{
		{
			ID:          "domain_no_infra",
			Description: "Domain layer must not depend on infra",
			Type:        "imports_forbidden",
			Config: map[string]interface{}{
				"from_globs":  []interface{}{"domain/**"},
				"forbid_globs": []interface{}{"infra/**"},
			},
			Severity: "error",
		},
	}

	e := NewEngine(rules, nil)
	result, err := e.Run([]string{"domain/service/UserService.kt"}, diff)
	require.NoError(t, err)

	assert.Len(t, result.Violations, 1)
	assert.Equal(t, 1, result.Errors)
	assert.Equal(t, 0, result.Warnings)
}

func TestEngine_RunWithDiffPatternForbidden(t *testing.T) {
	diff := `diff --git a/domain/model/Price.kt b/domain/model/Price.kt
--- a/domain/model/Price.kt
+++ b/domain/model/Price.kt
@@ -1,3 +1,4 @@
 package domain.model
+val amount: Double = 0.0

 data class Price(`

	rules := []config.Rule{
		{
			ID:          "money_minor_units",
			Description: "Money must use int minor units, not float/double",
			Type:        "diff_pattern_forbidden",
			Config: map[string]interface{}{
				"forbidden_regexes": []interface{}{`\bDouble\b`, `\bfloat\b`},
				"only_in_paths":     []interface{}{"**/*.kt"},
			},
			Severity: "warning",
		},
	}

	e := NewEngine(rules, nil)
	result, err := e.Run([]string{"domain/model/Price.kt"}, diff)
	require.NoError(t, err)

	assert.Len(t, result.Violations, 1)
	assert.Equal(t, 0, result.Errors)
	assert.Equal(t, 1, result.Warnings)
}

func TestEngine_RunWithMultipleRules(t *testing.T) {
	diff := `diff --git a/domain/service/PaymentService.kt b/domain/service/PaymentService.kt
--- a/domain/service/PaymentService.kt
+++ b/domain/service/PaymentService.kt
@@ -1,3 +1,5 @@
 package domain.service
+import com.myapp.infra.database.PaymentRepo
+val total: Double = 0.0

 class PaymentService {`

	rules := []config.Rule{
		{
			ID:          "domain_no_infra",
			Description: "Domain layer must not depend on infra",
			Type:        "imports_forbidden",
			Config: map[string]interface{}{
				"from_globs":  []interface{}{"domain/**"},
				"forbid_globs": []interface{}{"infra/**"},
			},
			Severity: "error",
		},
		{
			ID:          "money_minor_units",
			Description: "Money must use int minor units",
			Type:        "diff_pattern_forbidden",
			Config: map[string]interface{}{
				"forbidden_regexes": []interface{}{`\bDouble\b`},
				"only_in_paths":     []interface{}{"**/*.kt"},
			},
			Severity: "warning",
		},
	}

	e := NewEngine(rules, nil)
	result, err := e.Run([]string{"domain/service/PaymentService.kt"}, diff)
	require.NoError(t, err)

	assert.Len(t, result.Violations, 2)
	assert.Equal(t, 1, result.Errors)
	assert.Equal(t, 1, result.Warnings)
}

func TestEngine_ExceptionFiltersByRuleIDAndPath(t *testing.T) {
	diff := `diff --git a/domain/legacy/OldAdapter.kt b/domain/legacy/OldAdapter.kt
--- a/domain/legacy/OldAdapter.kt
+++ b/domain/legacy/OldAdapter.kt
@@ -1,3 +1,4 @@
 package domain.legacy
+import com.myapp.infra.database.LegacyRepo

 class OldAdapter {`

	rules := []config.Rule{
		{
			ID:          "domain_no_infra",
			Description: "Domain layer must not depend on infra",
			Type:        "imports_forbidden",
			Config: map[string]interface{}{
				"from_globs":  []interface{}{"domain/**"},
				"forbid_globs": []interface{}{"infra/**"},
			},
			Severity: "error",
		},
	}

	exceptions := []config.Exception{
		{
			ID:     "exc-legacy",
			RuleID: "domain_no_infra",
			Paths:  []string{"domain/legacy/OldAdapter.kt"},
			Reason: "Legacy code",
		},
	}

	e := NewEngine(rules, exceptions)
	result, err := e.Run([]string{"domain/legacy/OldAdapter.kt"}, diff)
	require.NoError(t, err)

	assert.Empty(t, result.Violations, "violation should be filtered by exception")
	assert.Equal(t, 0, result.Errors)
}

func TestEngine_ExceptionDoesNotFilterDifferentRule(t *testing.T) {
	diff := `diff --git a/domain/legacy/OldAdapter.kt b/domain/legacy/OldAdapter.kt
--- a/domain/legacy/OldAdapter.kt
+++ b/domain/legacy/OldAdapter.kt
@@ -1,3 +1,4 @@
 package domain.legacy
+import com.myapp.infra.database.LegacyRepo

 class OldAdapter {`

	rules := []config.Rule{
		{
			ID:          "domain_no_infra",
			Description: "Domain layer must not depend on infra",
			Type:        "imports_forbidden",
			Config: map[string]interface{}{
				"from_globs":  []interface{}{"domain/**"},
				"forbid_globs": []interface{}{"infra/**"},
			},
			Severity: "error",
		},
	}

	// Exception is for a different rule.
	exceptions := []config.Exception{
		{
			ID:     "exc-other",
			RuleID: "other_rule",
			Paths:  []string{"domain/legacy/OldAdapter.kt"},
			Reason: "Different rule",
		},
	}

	e := NewEngine(rules, exceptions)
	result, err := e.Run([]string{"domain/legacy/OldAdapter.kt"}, diff)
	require.NoError(t, err)

	assert.Len(t, result.Violations, 1, "exception for different rule should not filter")
}

func TestEngine_ExpiredExceptionIsIgnored(t *testing.T) {
	diff := `diff --git a/domain/legacy/OldAdapter.kt b/domain/legacy/OldAdapter.kt
--- a/domain/legacy/OldAdapter.kt
+++ b/domain/legacy/OldAdapter.kt
@@ -1,3 +1,4 @@
 package domain.legacy
+import com.myapp.infra.database.LegacyRepo

 class OldAdapter {`

	rules := []config.Rule{
		{
			ID:          "domain_no_infra",
			Description: "Domain layer must not depend on infra",
			Type:        "imports_forbidden",
			Config: map[string]interface{}{
				"from_globs":  []interface{}{"domain/**"},
				"forbid_globs": []interface{}{"infra/**"},
			},
			Severity: "error",
		},
	}

	expiredTime := time.Now().Add(-24 * time.Hour)
	exceptions := []config.Exception{
		{
			ID:        "exc-expired",
			RuleID:    "domain_no_infra",
			Paths:     []string{"domain/legacy/OldAdapter.kt"},
			Reason:    "Legacy code",
			ExpiresAt: &expiredTime,
		},
	}

	e := NewEngine(rules, exceptions)
	result, err := e.Run([]string{"domain/legacy/OldAdapter.kt"}, diff)
	require.NoError(t, err)

	assert.Len(t, result.Violations, 1, "expired exception should not filter violation")
}

func TestEngine_ActiveExceptionFiltersViolation(t *testing.T) {
	diff := `diff --git a/domain/legacy/OldAdapter.kt b/domain/legacy/OldAdapter.kt
--- a/domain/legacy/OldAdapter.kt
+++ b/domain/legacy/OldAdapter.kt
@@ -1,3 +1,4 @@
 package domain.legacy
+import com.myapp.infra.database.LegacyRepo

 class OldAdapter {`

	rules := []config.Rule{
		{
			ID:          "domain_no_infra",
			Description: "Domain layer must not depend on infra",
			Type:        "imports_forbidden",
			Config: map[string]interface{}{
				"from_globs":  []interface{}{"domain/**"},
				"forbid_globs": []interface{}{"infra/**"},
			},
			Severity: "error",
		},
	}

	futureTime := time.Now().Add(30 * 24 * time.Hour)
	exceptions := []config.Exception{
		{
			ID:        "exc-active",
			RuleID:    "domain_no_infra",
			Paths:     []string{"domain/legacy/OldAdapter.kt"},
			Reason:    "Legacy code, migrating in Q2",
			ExpiresAt: &futureTime,
		},
	}

	e := NewEngine(rules, exceptions)
	result, err := e.Run([]string{"domain/legacy/OldAdapter.kt"}, diff)
	require.NoError(t, err)

	assert.Empty(t, result.Violations)
}

func TestEngine_PermanentException(t *testing.T) {
	diff := `diff --git a/domain/legacy/OldAdapter.kt b/domain/legacy/OldAdapter.kt
--- a/domain/legacy/OldAdapter.kt
+++ b/domain/legacy/OldAdapter.kt
@@ -1,3 +1,4 @@
 package domain.legacy
+import com.myapp.infra.database.LegacyRepo

 class OldAdapter {`

	rules := []config.Rule{
		{
			ID:          "domain_no_infra",
			Description: "Domain layer must not depend on infra",
			Type:        "imports_forbidden",
			Config: map[string]interface{}{
				"from_globs":  []interface{}{"domain/**"},
				"forbid_globs": []interface{}{"infra/**"},
			},
			Severity: "error",
		},
	}

	// ExpiresAt is nil = permanent.
	exceptions := []config.Exception{
		{
			ID:        "exc-permanent",
			RuleID:    "domain_no_infra",
			Paths:     []string{"domain/legacy/OldAdapter.kt"},
			Reason:    "Permanent exception for legacy code",
			ExpiresAt: nil,
		},
	}

	e := NewEngine(rules, exceptions)
	result, err := e.Run([]string{"domain/legacy/OldAdapter.kt"}, diff)
	require.NoError(t, err)

	assert.Empty(t, result.Violations, "permanent exception should filter violation")
}

func TestEngine_UnknownRuleType(t *testing.T) {
	rules := []config.Rule{
		{
			ID:       "unknown",
			Type:     "nonexistent_checker",
			Config:   map[string]interface{}{},
			Severity: "error",
		},
	}

	e := NewEngine(rules, nil)
	_, err := e.Run([]string{"test.go"}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown rule type")
}

func TestEngine_ExceptionWithGlobPath(t *testing.T) {
	diff := `diff --git a/domain/legacy/OldAdapter.kt b/domain/legacy/OldAdapter.kt
--- a/domain/legacy/OldAdapter.kt
+++ b/domain/legacy/OldAdapter.kt
@@ -1,3 +1,4 @@
 package domain.legacy
+import com.myapp.infra.database.LegacyRepo

 class OldAdapter {`

	rules := []config.Rule{
		{
			ID:          "domain_no_infra",
			Description: "Domain must not depend on infra",
			Type:        "imports_forbidden",
			Config: map[string]interface{}{
				"from_globs":  []interface{}{"domain/**"},
				"forbid_globs": []interface{}{"infra/**"},
			},
			Severity: "error",
		},
	}

	// Glob exception pattern: Note filepath.Match does not support **,
	// so we use a direct pattern that filepath.Match supports.
	exceptions := []config.Exception{
		{
			ID:     "exc-glob",
			RuleID: "domain_no_infra",
			Paths:  []string{"domain/legacy/*.kt"},
			Reason: "Legacy directory",
		},
	}

	e := NewEngine(rules, exceptions)
	result, err := e.Run([]string{"domain/legacy/OldAdapter.kt"}, diff)
	require.NoError(t, err)

	assert.Empty(t, result.Violations, "glob pattern exception should match")
}
