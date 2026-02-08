package git

import "os"

// CIInfo holds detected CI environment information.
type CIInfo struct {
	Detected bool
	System   string // "github", "gitlab", ""
	BaseRef  string
	HeadRef  string
}

// DetectCI reads CI environment variables to determine if the tool is running
// inside a CI system. It supports GitHub Actions and GitLab CI.
func DetectCI() *CIInfo {
	// Check GitHub Actions.
	if base := os.Getenv("GITHUB_BASE_REF"); base != "" {
		return &CIInfo{
			Detected: true,
			System:   "github",
			BaseRef:  base,
			HeadRef:  os.Getenv("GITHUB_HEAD_REF"),
		}
	}

	// Check GitLab CI.
	if base := os.Getenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME"); base != "" {
		return &CIInfo{
			Detected: true,
			System:   "gitlab",
			BaseRef:  base,
			HeadRef:  os.Getenv("CI_MERGE_REQUEST_SOURCE_BRANCH_NAME"),
		}
	}

	return &CIInfo{
		Detected: false,
		System:   "",
		BaseRef:  "",
		HeadRef:  "",
	}
}
