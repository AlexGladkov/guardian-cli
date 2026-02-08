package config

import (
	"errors"
	"fmt"
	"strings"
)

// validQuorumTypes is the set of valid quorum type values.
var validQuorumTypes = map[string]bool{
	"majority":   true,
	"two_thirds": true,
	"unanimous":  true,
	"custom":     true,
}

// validSeverities is the set of valid severity values for rules.
var validSeverities = map[string]bool{
	"error":   true,
	"warning": true,
}

// validProposalTypes is the set of valid proposal type values.
var validProposalTypes = map[string]bool{
	"modify": true,
	"add":    true,
	"remove": true,
}

// validProposalStatuses is the set of valid proposal status values.
var validProposalStatuses = map[string]bool{
	"proposed":  true,
	"accepted":  true,
	"rejected":  true,
	"withdrawn": true,
	"expired":   true,
}

// validDecisions is the set of valid vote decision values.
var validDecisions = map[string]bool{
	"yes": true,
	"no":  true,
}

// validLLMProviders is the set of valid LLM provider values.
var validLLMProviders = map[string]bool{
	"deepseek": true,
	"openai":   true,
	"claude":   true,
	"custom":   true,
}

// ValidateConstitution validates a Constitution for required fields and correct enum values.
func ValidateConstitution(c *Constitution) error {
	if c == nil {
		return errors.New("constitution is nil")
	}

	var errs []string

	// Validate governance
	if len(c.Governance.Voters) == 0 {
		errs = append(errs, "governance.voters must not be empty")
	}
	for i, v := range c.Governance.Voters {
		if v.Role == "" {
			errs = append(errs, fmt.Sprintf("governance.voters[%d].role must not be empty", i))
		}
	}

	// Validate quorum
	if c.Governance.Quorum.Type == "" {
		errs = append(errs, "governance.quorum.type must not be empty")
	} else if !validQuorumTypes[c.Governance.Quorum.Type] {
		errs = append(errs, fmt.Sprintf("governance.quorum.type %q is invalid; must be one of: majority, two_thirds, unanimous, custom", c.Governance.Quorum.Type))
	}

	if c.Governance.Quorum.Type == "custom" {
		if c.Governance.Quorum.Threshold <= 0 || c.Governance.Quorum.Threshold > 1 {
			errs = append(errs, "governance.quorum.threshold must be between 0 (exclusive) and 1 (inclusive) for custom quorum type")
		}
	}

	if c.Governance.ProposalTTLDays < 0 {
		errs = append(errs, "governance.proposal_ttl_days must not be negative")
	}

	// Validate per-rule overrides
	for ruleID, override := range c.Governance.PerRuleOverrides {
		if override.Quorum.Type == "" {
			errs = append(errs, fmt.Sprintf("governance.per_rule_overrides[%s].quorum.type must not be empty", ruleID))
		} else if !validQuorumTypes[override.Quorum.Type] {
			errs = append(errs, fmt.Sprintf("governance.per_rule_overrides[%s].quorum.type %q is invalid", ruleID, override.Quorum.Type))
		}
		if override.Quorum.Type == "custom" {
			if override.Quorum.Threshold <= 0 || override.Quorum.Threshold > 1 {
				errs = append(errs, fmt.Sprintf("governance.per_rule_overrides[%s].quorum.threshold must be between 0 (exclusive) and 1 (inclusive)", ruleID))
			}
		}
	}

	// Validate roles
	if len(c.Roles) == 0 {
		errs = append(errs, "roles must not be empty")
	}
	for name, role := range c.Roles {
		if len(role.Members) == 0 {
			errs = append(errs, fmt.Sprintf("roles[%s].members must not be empty", name))
		}
		for i, member := range role.Members {
			if member.Email == "" {
				errs = append(errs, fmt.Sprintf("roles[%s].members[%d].email must not be empty", name, i))
			}
		}
	}

	// Validate voter roles reference existing roles
	for _, v := range c.Governance.Voters {
		if v.Role != "" {
			if _, exists := c.Roles[v.Role]; !exists {
				errs = append(errs, fmt.Sprintf("governance.voters references role %q which is not defined in roles", v.Role))
			}
		}
	}

	// Validate LLM config
	if c.LLM.Provider != "" && !validLLMProviders[c.LLM.Provider] {
		errs = append(errs, fmt.Sprintf("llm.provider %q is invalid; must be one of: deepseek, openai, claude, custom", c.LLM.Provider))
	}

	if c.LLM.Provider == "custom" && c.LLM.Endpoint == "" {
		errs = append(errs, "llm.endpoint is required when provider is custom")
	}

	if len(errs) > 0 {
		return fmt.Errorf("constitution validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}

	return nil
}

// ValidateRules validates a RulesFile for required fields and correct values.
func ValidateRules(r *RulesFile) error {
	if r == nil {
		return errors.New("rules file is nil")
	}

	var errs []string

	seenIDs := make(map[string]bool)
	for i, rule := range r.Rules {
		if rule.ID == "" {
			errs = append(errs, fmt.Sprintf("rules[%d].id must not be empty", i))
		} else if seenIDs[rule.ID] {
			errs = append(errs, fmt.Sprintf("rules[%d].id %q is duplicated", i, rule.ID))
		} else {
			seenIDs[rule.ID] = true
		}

		if rule.Description == "" {
			errs = append(errs, fmt.Sprintf("rules[%d].description must not be empty", i))
		}

		if rule.Type == "" {
			errs = append(errs, fmt.Sprintf("rules[%d].type must not be empty", i))
		}

		if rule.Severity == "" {
			errs = append(errs, fmt.Sprintf("rules[%d].severity must not be empty", i))
		} else if !validSeverities[rule.Severity] {
			errs = append(errs, fmt.Sprintf("rules[%d].severity %q is invalid; must be one of: error, warning", i, rule.Severity))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("rules validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}

	return nil
}

// ValidateProposal validates a Proposal for required fields and correct enum values.
func ValidateProposal(p *Proposal) error {
	if p == nil {
		return errors.New("proposal is nil")
	}

	var errs []string

	if p.ID == "" {
		errs = append(errs, "id must not be empty")
	}

	if p.RuleID == "" {
		errs = append(errs, "rule_id must not be empty")
	}

	if p.ProposalType == "" {
		errs = append(errs, "proposal_type must not be empty")
	} else if !validProposalTypes[p.ProposalType] {
		errs = append(errs, fmt.Sprintf("proposal_type %q is invalid; must be one of: modify, add, remove", p.ProposalType))
	}

	if p.Change.Description == "" {
		errs = append(errs, "change.description must not be empty")
	}

	if p.Reason == "" {
		errs = append(errs, "reason must not be empty")
	}

	if p.CreatedBy == "" {
		errs = append(errs, "created_by must not be empty")
	}

	if p.CreatedAt.IsZero() {
		errs = append(errs, "created_at must not be zero")
	}

	if p.Status == "" {
		errs = append(errs, "status must not be empty")
	} else if !validProposalStatuses[p.Status] {
		errs = append(errs, fmt.Sprintf("status %q is invalid; must be one of: proposed, accepted, rejected, withdrawn, expired", p.Status))
	}

	if len(errs) > 0 {
		return fmt.Errorf("proposal validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}

	return nil
}

// ValidateVote validates a Vote for required fields and correct enum values.
func ValidateVote(v *Vote) error {
	if v == nil {
		return errors.New("vote is nil")
	}

	var errs []string

	if v.ProposalID == "" {
		errs = append(errs, "proposal_id must not be empty")
	}

	if v.VoterEmail == "" {
		errs = append(errs, "voter_email must not be empty")
	}

	if v.Decision == "" {
		errs = append(errs, "decision must not be empty")
	} else if !validDecisions[v.Decision] {
		errs = append(errs, fmt.Sprintf("decision %q is invalid; must be one of: yes, no", v.Decision))
	}

	if v.VotedAt.IsZero() {
		errs = append(errs, "voted_at must not be zero")
	}

	if len(errs) > 0 {
		return fmt.Errorf("vote validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}

	return nil
}

// ValidateException validates an Exception for required fields.
func ValidateException(e *Exception) error {
	if e == nil {
		return errors.New("exception is nil")
	}

	var errs []string

	if e.ID == "" {
		errs = append(errs, "id must not be empty")
	}

	if e.RuleID == "" {
		errs = append(errs, "rule_id must not be empty")
	}

	if len(e.Paths) == 0 {
		errs = append(errs, "paths must not be empty")
	}

	if e.Reason == "" {
		errs = append(errs, "reason must not be empty")
	}

	if e.CreatedBy == "" {
		errs = append(errs, "created_by must not be empty")
	}

	if e.CreatedAt.IsZero() {
		errs = append(errs, "created_at must not be zero")
	}

	if e.ExpiresAt != nil && !e.ExpiresAt.IsZero() && e.ExpiresAt.Before(e.CreatedAt) {
		errs = append(errs, "expires_at must not be before created_at")
	}

	if len(errs) > 0 {
		return fmt.Errorf("exception validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}

	return nil
}
