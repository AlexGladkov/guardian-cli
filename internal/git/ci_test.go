package git

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectCI_GitHubActions(t *testing.T) {
	// Save and restore env.
	origBase := os.Getenv("GITHUB_BASE_REF")
	origHead := os.Getenv("GITHUB_HEAD_REF")
	origGitLab := os.Getenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME")

	t.Cleanup(func() {
		os.Setenv("GITHUB_BASE_REF", origBase)
		os.Setenv("GITHUB_HEAD_REF", origHead)
		os.Setenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME", origGitLab)
	})

	// Clear GitLab vars.
	os.Unsetenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME")

	require.NoError(t, os.Setenv("GITHUB_BASE_REF", "main"))
	require.NoError(t, os.Setenv("GITHUB_HEAD_REF", "feature/test"))

	ci := DetectCI()

	assert.True(t, ci.Detected)
	assert.Equal(t, "github", ci.System)
	assert.Equal(t, "main", ci.BaseRef)
	assert.Equal(t, "feature/test", ci.HeadRef)
}

func TestDetectCI_GitLabCI(t *testing.T) {
	origGitHub := os.Getenv("GITHUB_BASE_REF")
	origTarget := os.Getenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME")
	origSource := os.Getenv("CI_MERGE_REQUEST_SOURCE_BRANCH_NAME")

	t.Cleanup(func() {
		os.Setenv("GITHUB_BASE_REF", origGitHub)
		os.Setenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME", origTarget)
		os.Setenv("CI_MERGE_REQUEST_SOURCE_BRANCH_NAME", origSource)
	})

	// Clear GitHub vars.
	os.Unsetenv("GITHUB_BASE_REF")

	require.NoError(t, os.Setenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME", "develop"))
	require.NoError(t, os.Setenv("CI_MERGE_REQUEST_SOURCE_BRANCH_NAME", "feature/gitlab-test"))

	ci := DetectCI()

	assert.True(t, ci.Detected)
	assert.Equal(t, "gitlab", ci.System)
	assert.Equal(t, "develop", ci.BaseRef)
	assert.Equal(t, "feature/gitlab-test", ci.HeadRef)
}

func TestDetectCI_NoCIEnvironment(t *testing.T) {
	origGitHub := os.Getenv("GITHUB_BASE_REF")
	origGitLab := os.Getenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME")

	t.Cleanup(func() {
		os.Setenv("GITHUB_BASE_REF", origGitHub)
		os.Setenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME", origGitLab)
	})

	os.Unsetenv("GITHUB_BASE_REF")
	os.Unsetenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME")

	ci := DetectCI()

	assert.False(t, ci.Detected)
	assert.Equal(t, "", ci.System)
	assert.Equal(t, "", ci.BaseRef)
	assert.Equal(t, "", ci.HeadRef)
}

func TestDetectCI_GitHubTakesPrecedenceOverGitLab(t *testing.T) {
	origGitHubBase := os.Getenv("GITHUB_BASE_REF")
	origGitHubHead := os.Getenv("GITHUB_HEAD_REF")
	origGitLabTarget := os.Getenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME")
	origGitLabSource := os.Getenv("CI_MERGE_REQUEST_SOURCE_BRANCH_NAME")

	t.Cleanup(func() {
		os.Setenv("GITHUB_BASE_REF", origGitHubBase)
		os.Setenv("GITHUB_HEAD_REF", origGitHubHead)
		os.Setenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME", origGitLabTarget)
		os.Setenv("CI_MERGE_REQUEST_SOURCE_BRANCH_NAME", origGitLabSource)
	})

	require.NoError(t, os.Setenv("GITHUB_BASE_REF", "main"))
	require.NoError(t, os.Setenv("GITHUB_HEAD_REF", "gh-feature"))
	require.NoError(t, os.Setenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME", "develop"))
	require.NoError(t, os.Setenv("CI_MERGE_REQUEST_SOURCE_BRANCH_NAME", "gl-feature"))

	ci := DetectCI()

	assert.True(t, ci.Detected)
	assert.Equal(t, "github", ci.System, "GitHub should take precedence over GitLab")
	assert.Equal(t, "main", ci.BaseRef)
	assert.Equal(t, "gh-feature", ci.HeadRef)
}

func TestDetectCI_GitHubWithEmptyHead(t *testing.T) {
	origBase := os.Getenv("GITHUB_BASE_REF")
	origHead := os.Getenv("GITHUB_HEAD_REF")
	origGitLab := os.Getenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME")

	t.Cleanup(func() {
		os.Setenv("GITHUB_BASE_REF", origBase)
		os.Setenv("GITHUB_HEAD_REF", origHead)
		os.Setenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME", origGitLab)
	})

	os.Unsetenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME")
	require.NoError(t, os.Setenv("GITHUB_BASE_REF", "main"))
	os.Unsetenv("GITHUB_HEAD_REF")

	ci := DetectCI()

	assert.True(t, ci.Detected)
	assert.Equal(t, "github", ci.System)
	assert.Equal(t, "main", ci.BaseRef)
	assert.Equal(t, "", ci.HeadRef)
}
