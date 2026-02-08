// Package governance implements the voting, quorum, and role logic for Guardian's
// constitutional governance system.
package governance

import (
	"github.com/AlexGladkov/guardian-cli/internal/config"
)

// GetEligibleVoters returns a deduplicated list of email addresses from all
// roles that are designated as voters in the constitution. The deduplication
// ensures that a person appearing in multiple voter roles is counted only once.
func GetEligibleVoters(constitution *config.Constitution) []string {
	seen := make(map[string]bool)
	var voters []string

	for _, voterRef := range constitution.Governance.Voters {
		role, ok := constitution.Roles[voterRef.Role]
		if !ok {
			continue
		}
		for _, member := range role.Members {
			if !seen[member.Email] {
				seen[member.Email] = true
				voters = append(voters, member.Email)
			}
		}
	}

	return voters
}

// GetUserRoles returns the list of role names that a given email address
// belongs to. It checks all roles defined in the constitution, not just
// voter roles.
func GetUserRoles(constitution *config.Constitution, email string) []string {
	var roles []string
	for roleName, role := range constitution.Roles {
		for _, member := range role.Members {
			if member.Email == email {
				roles = append(roles, roleName)
				break
			}
		}
	}
	return roles
}

// IsVoter checks if the given email belongs to any of the voter roles
// defined in the constitution.
func IsVoter(constitution *config.Constitution, email string) bool {
	for _, voterRef := range constitution.Governance.Voters {
		role, ok := constitution.Roles[voterRef.Role]
		if !ok {
			continue
		}
		for _, member := range role.Members {
			if member.Email == email {
				return true
			}
		}
	}
	return false
}
