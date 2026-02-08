package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsCloudProvider(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		want     bool
	}{
		{
			name:     "deepseek is cloud",
			provider: ProviderDeepSeek,
			want:     true,
		},
		{
			name:     "openai is cloud",
			provider: ProviderOpenAI,
			want:     true,
		},
		{
			name:     "claude is cloud",
			provider: ProviderClaude,
			want:     true,
		},
		{
			name:     "custom is not cloud",
			provider: ProviderCustom,
			want:     false,
		},
		{
			name:     "empty string is not cloud",
			provider: "",
			want:     false,
		},
		{
			name:     "unknown provider is not cloud",
			provider: "unknown",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCloudProvider(tt.provider)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidProviders(t *testing.T) {
	providers := ValidProviders()

	require.Len(t, providers, 4)
	assert.Contains(t, providers, ProviderDeepSeek)
	assert.Contains(t, providers, ProviderOpenAI)
	assert.Contains(t, providers, ProviderClaude)
	assert.Contains(t, providers, ProviderCustom)
}

func TestProviderEndpoints(t *testing.T) {
	tests := []struct {
		provider string
		want     string
	}{
		{ProviderDeepSeek, "https://api.deepseek.com/v1"},
		{ProviderOpenAI, "https://api.openai.com/v1"},
		{ProviderClaude, "https://api.anthropic.com/v1"},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			endpoint, ok := ProviderEndpoints[tt.provider]
			require.True(t, ok, "endpoint should exist for provider %s", tt.provider)
			assert.Equal(t, tt.want, endpoint)
		})
	}

	// Custom provider should NOT have an endpoint
	_, ok := ProviderEndpoints[ProviderCustom]
	assert.False(t, ok, "custom provider should not have a default endpoint")
}

func TestDefaultModels(t *testing.T) {
	tests := []struct {
		provider string
		want     string
	}{
		{ProviderDeepSeek, "deepseek-chat"},
		{ProviderOpenAI, "gpt-4o"},
		{ProviderClaude, "claude-sonnet-4-5-20250929"},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			model, ok := DefaultModels[tt.provider]
			require.True(t, ok, "default model should exist for provider %s", tt.provider)
			assert.Equal(t, tt.want, model)
		})
	}
}

func TestProviderConstants(t *testing.T) {
	assert.Equal(t, "deepseek", ProviderDeepSeek)
	assert.Equal(t, "openai", ProviderOpenAI)
	assert.Equal(t, "claude", ProviderClaude)
	assert.Equal(t, "custom", ProviderCustom)
}
