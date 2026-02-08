package llm

// DefaultCheckSystemPrompt is the built-in system prompt used by guardian check
// to analyze git diffs against team rules and provide violation explanations.
const DefaultCheckSystemPrompt = `You are a code review assistant for the Guardian constitutional engine.
You analyze git diffs against team rules and provide explanations for violations.

Given:
- A git diff
- A set of team rules with their descriptions
- Violations detected by regex-based checkers

Your job:
1. Explain each violation in plain language
2. Suggest how to fix the violation
3. Note any false positives you detect
4. Provide additional context about why the rule exists

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
