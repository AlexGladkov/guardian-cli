package llm

// DefaultCheckSystemPrompt is the built-in system prompt used by guardian check
// to analyze git diffs against team rules and provide violation explanations.
const DefaultCheckSystemPrompt = `You are a code review assistant for the Guardian constitutional governance engine.
You analyze git diffs against team rules and provide explanations for violations.

Guardian enforces team agreements through a governance lifecycle:
- Rules: active checks enforced on every commit/PR
- Proposals: suggested changes to rules (add, modify, remove)
- Votes: team members vote on proposals
- Finalization: accepted proposals become rule changes

Given:
- A git diff
- A set of team rules with their descriptions
- Violations detected by regex-based checkers
- Governance context: proposals that may affect how violations should be interpreted

Your job:
1. Explain each violation in plain language
2. Suggest how to fix the violation
3. Note any false positives you detect
4. Provide additional context about why the rule exists
5. If a relevant proposal is accepted and pending finalization, note that the rule may change soon
6. If a relevant proposal is under review, mention it as additional context

Be concise. Focus on actionable advice. Format each violation explanation as a short paragraph.`

// DefaultProposeSystemPrompt is the built-in system prompt used by guardian propose
// to generate draft proposal text for rule changes.
const DefaultProposeSystemPrompt = `You are an assistant helping draft a proposal to change a team rule.
Given a rule description and context, generate:
1. A clear description of the proposed change
2. The reason for the change
3. The expected impact

Be specific and actionable. Use professional technical language.`

// GetCheckPrompt returns the check system prompt. If override is non-empty,
// it is used instead of the built-in default prompt.
func GetCheckPrompt(override string) string {
	if override != "" {
		return override
	}
	return DefaultCheckSystemPrompt
}

// GetProposePrompt returns the propose system prompt. If override is non-empty,
// it is used instead of the built-in default prompt.
func GetProposePrompt(override string) string {
	if override != "" {
		return override
	}
	return DefaultProposeSystemPrompt
}
