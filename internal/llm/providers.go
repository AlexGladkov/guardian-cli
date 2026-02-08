// Package llm provides LLM integration for the Guardian CLI tool.
// It supports multiple LLM providers (DeepSeek, OpenAI, Claude, Custom)
// and handles API communication, prompt management, and interactive configuration.
package llm

// Provider constants identify supported LLM providers.
const (
	ProviderDeepSeek = "deepseek"
	ProviderOpenAI   = "openai"
	ProviderClaude   = "claude"
	ProviderCustom   = "custom"
)

// ProviderEndpoints maps known providers to their API endpoints.
var ProviderEndpoints = map[string]string{
	ProviderDeepSeek: "https://api.deepseek.com/v1",
	ProviderOpenAI:   "https://api.openai.com/v1",
	ProviderClaude:   "https://api.anthropic.com/v1",
}

// DefaultModels maps known providers to default model names.
var DefaultModels = map[string]string{
	ProviderDeepSeek: "deepseek-chat",
	ProviderOpenAI:   "gpt-4o",
	ProviderClaude:   "claude-sonnet-4-5-20250929",
}

// IsCloudProvider returns true for non-local providers that send data
// to external APIs (deepseek, openai, claude).
func IsCloudProvider(provider string) bool {
	return provider == ProviderDeepSeek || provider == ProviderOpenAI || provider == ProviderClaude
}

// ValidProviders returns the list of all valid provider names.
func ValidProviders() []string {
	return []string{ProviderDeepSeek, ProviderOpenAI, ProviderClaude, ProviderCustom}
}
