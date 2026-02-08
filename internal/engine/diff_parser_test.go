package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDiff_EmptyContent(t *testing.T) {
	result := ParseDiff("")
	assert.Nil(t, result)
}

func TestParseDiff_SingleFile(t *testing.T) {
	diff := `diff --git a/domain/service/UserService.kt b/domain/service/UserService.kt
--- a/domain/service/UserService.kt
+++ b/domain/service/UserService.kt
@@ -1,3 +1,4 @@
 package domain.service
+import com.myapp.infra.database.UserRepository

 class UserService {`

	result := ParseDiff(diff)

	assert.Len(t, result, 1)
	assert.Equal(t, "domain/service/UserService.kt", result[0].Path)
	assert.Len(t, result[0].AddedLines, 1)
	assert.Equal(t, "import com.myapp.infra.database.UserRepository", result[0].AddedLines[0])
}

func TestParseDiff_MultipleFiles(t *testing.T) {
	diff := `diff --git a/file1.go b/file1.go
--- a/file1.go
+++ b/file1.go
@@ -1,2 +1,3 @@
 package main
+import "fmt"
diff --git a/file2.go b/file2.go
--- a/file2.go
+++ b/file2.go
@@ -1,2 +1,3 @@
 package util
+import "os"
+import "io"`

	result := ParseDiff(diff)

	assert.Len(t, result, 2)

	assert.Equal(t, "file1.go", result[0].Path)
	assert.Len(t, result[0].AddedLines, 1)
	assert.Equal(t, `import "fmt"`, result[0].AddedLines[0])

	assert.Equal(t, "file2.go", result[1].Path)
	assert.Len(t, result[1].AddedLines, 2)
	assert.Equal(t, `import "os"`, result[1].AddedLines[0])
	assert.Equal(t, `import "io"`, result[1].AddedLines[1])
}

func TestParseDiff_ContextLinesIgnored(t *testing.T) {
	diff := `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,5 +1,6 @@
 package main

-import "old"
+import "new"

 func main() {`

	result := ParseDiff(diff)

	assert.Len(t, result, 1)
	assert.Equal(t, "main.go", result[0].Path)
	// Only added lines (with +) are captured.
	assert.Len(t, result[0].AddedLines, 1)
	assert.Equal(t, `import "new"`, result[0].AddedLines[0])
}

func TestParseDiff_NewFile(t *testing.T) {
	diff := `diff --git a/newfile.go b/newfile.go
--- /dev/null
+++ b/newfile.go
@@ -0,0 +1,3 @@
+package newpkg
+
+func Hello() string { return "hello" }`

	result := ParseDiff(diff)

	assert.Len(t, result, 1)
	assert.Equal(t, "newfile.go", result[0].Path)
	assert.Len(t, result[0].AddedLines, 3)
	assert.Equal(t, "package newpkg", result[0].AddedLines[0])
}

func TestParseDiff_MultipleHunks(t *testing.T) {
	diff := `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main
+import "fmt"

 func main() {
@@ -10,3 +11,4 @@
 }

+func helper() {}`

	result := ParseDiff(diff)

	assert.Len(t, result, 1)
	assert.Equal(t, "main.go", result[0].Path)
	assert.Len(t, result[0].AddedLines, 2)
	assert.Equal(t, `import "fmt"`, result[0].AddedLines[0])
	assert.Equal(t, "func helper() {}", result[0].AddedLines[1])
}

func TestParseDiff_NoAddedLines(t *testing.T) {
	diff := `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,3 +1,2 @@
 package main
-import "fmt"

 func main() {}`

	result := ParseDiff(diff)

	assert.Len(t, result, 1)
	assert.Equal(t, "main.go", result[0].Path)
	assert.Empty(t, result[0].AddedLines)
}
