package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImportsForbidden_BasicViolation(t *testing.T) {
	diff := `diff --git a/domain/service/UserService.kt b/domain/service/UserService.kt
--- a/domain/service/UserService.kt
+++ b/domain/service/UserService.kt
@@ -1,3 +1,4 @@
 package domain.service
+import com.myapp.infra.database.UserRepository

 class UserService {`

	checker := &ImportsForbiddenChecker{}
	ctx := &CheckContext{
		ChangedFiles: []string{"domain/service/UserService.kt"},
		DiffContent:  diff,
		RuleConfig: map[string]interface{}{
			"from_globs":  []interface{}{"domain/**"},
			"forbid_globs": []interface{}{"infra/**"},
		},
		Severity: "error",
		RuleID:   "domain_no_infra",
		RuleDesc: "Domain layer must not depend on infra",
	}

	violations, err := checker.Check(ctx)
	require.NoError(t, err)

	assert.Len(t, violations, 1)
	assert.Equal(t, "domain_no_infra", violations[0].RuleID)
	assert.Equal(t, "error", violations[0].Severity)
	assert.Equal(t, "domain/service/UserService.kt", violations[0].FilePath)
	assert.Contains(t, violations[0].DiffSnippet, "infra")
}

func TestImportsForbidden_NoViolation_FileNotInFromGlobs(t *testing.T) {
	diff := `diff --git a/app/service/AppService.kt b/app/service/AppService.kt
--- a/app/service/AppService.kt
+++ b/app/service/AppService.kt
@@ -1,3 +1,4 @@
 package app.service
+import com.myapp.infra.database.UserRepository

 class AppService {`

	checker := &ImportsForbiddenChecker{}
	ctx := &CheckContext{
		ChangedFiles: []string{"app/service/AppService.kt"},
		DiffContent:  diff,
		RuleConfig: map[string]interface{}{
			"from_globs":  []interface{}{"domain/**"},
			"forbid_globs": []interface{}{"infra/**"},
		},
		Severity: "error",
		RuleID:   "domain_no_infra",
		RuleDesc: "Domain layer must not depend on infra",
	}

	violations, err := checker.Check(ctx)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestImportsForbidden_NoViolation_NoForbiddenImport(t *testing.T) {
	diff := `diff --git a/domain/service/UserService.kt b/domain/service/UserService.kt
--- a/domain/service/UserService.kt
+++ b/domain/service/UserService.kt
@@ -1,3 +1,4 @@
 package domain.service
+import com.myapp.domain.model.User

 class UserService {`

	checker := &ImportsForbiddenChecker{}
	ctx := &CheckContext{
		ChangedFiles: []string{"domain/service/UserService.kt"},
		DiffContent:  diff,
		RuleConfig: map[string]interface{}{
			"from_globs":  []interface{}{"domain/**"},
			"forbid_globs": []interface{}{"infra/**"},
		},
		Severity: "error",
		RuleID:   "domain_no_infra",
		RuleDesc: "Domain layer must not depend on infra",
	}

	violations, err := checker.Check(ctx)
	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestImportsForbidden_MultipleForbiddenGlobs(t *testing.T) {
	diff := `diff --git a/domain/service/UserService.kt b/domain/service/UserService.kt
--- a/domain/service/UserService.kt
+++ b/domain/service/UserService.kt
@@ -1,4 +1,6 @@
 package domain.service
+import com.myapp.infra.database.UserRepository
+import com.myapp.data.cache.CacheManager

 class UserService {`

	checker := &ImportsForbiddenChecker{}
	ctx := &CheckContext{
		ChangedFiles: []string{"domain/service/UserService.kt"},
		DiffContent:  diff,
		RuleConfig: map[string]interface{}{
			"from_globs":  []interface{}{"domain/**"},
			"forbid_globs": []interface{}{"infra/**", "data/**"},
		},
		Severity: "error",
		RuleID:   "domain_no_infra",
		RuleDesc: "Domain layer must not depend on infra or data",
	}

	violations, err := checker.Check(ctx)
	require.NoError(t, err)

	assert.Len(t, violations, 2)
}

func TestImportsForbidden_MultipleFiles(t *testing.T) {
	diff := `diff --git a/domain/service/UserService.kt b/domain/service/UserService.kt
--- a/domain/service/UserService.kt
+++ b/domain/service/UserService.kt
@@ -1,3 +1,4 @@
 package domain.service
+import com.myapp.infra.database.UserRepository

 class UserService {
diff --git a/domain/model/Order.kt b/domain/model/Order.kt
--- a/domain/model/Order.kt
+++ b/domain/model/Order.kt
@@ -1,3 +1,4 @@
 package domain.model
+import com.myapp.infra.messaging.Queue

 data class Order(`

	checker := &ImportsForbiddenChecker{}
	ctx := &CheckContext{
		ChangedFiles: []string{
			"domain/service/UserService.kt",
			"domain/model/Order.kt",
		},
		DiffContent: diff,
		RuleConfig: map[string]interface{}{
			"from_globs":  []interface{}{"domain/**"},
			"forbid_globs": []interface{}{"infra/**"},
		},
		Severity: "error",
		RuleID:   "domain_no_infra",
		RuleDesc: "Domain layer must not depend on infra",
	}

	violations, err := checker.Check(ctx)
	require.NoError(t, err)

	assert.Len(t, violations, 2)
	assert.Equal(t, "domain/service/UserService.kt", violations[0].FilePath)
	assert.Equal(t, "domain/model/Order.kt", violations[1].FilePath)
}

func TestImportsForbidden_MissingConfig(t *testing.T) {
	checker := &ImportsForbiddenChecker{}
	ctx := &CheckContext{
		ChangedFiles: []string{"domain/test.kt"},
		DiffContent:  "",
		RuleConfig:   map[string]interface{}{},
		Severity:     "error",
		RuleID:       "test",
		RuleDesc:     "test",
	}

	_, err := checker.Check(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "from_globs")
}

func TestImportsForbidden_Type(t *testing.T) {
	checker := &ImportsForbiddenChecker{}
	assert.Equal(t, "imports_forbidden", checker.Type())
}

func TestImportsForbidden_EmptyDiff(t *testing.T) {
	checker := &ImportsForbiddenChecker{}
	ctx := &CheckContext{
		ChangedFiles: []string{"domain/service/UserService.kt"},
		DiffContent:  "",
		RuleConfig: map[string]interface{}{
			"from_globs":  []interface{}{"domain/**"},
			"forbid_globs": []interface{}{"infra/**"},
		},
		Severity: "error",
		RuleID:   "test",
		RuleDesc: "test",
	}

	violations, err := checker.Check(ctx)
	require.NoError(t, err)
	assert.Empty(t, violations)
}
