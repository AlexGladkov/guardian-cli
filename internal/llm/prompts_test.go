package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCheckPrompt(t *testing.T) {
	tests := []struct {
		name     string
		override string
		want     string
	}{
		{
			name:     "returns default when override is empty",
			override: "",
			want:     DefaultCheckSystemPrompt,
		},
		{
			name:     "returns override when provided",
			override: "Custom check prompt",
			want:     "Custom check prompt",
		},
		{
			name:     "returns override even if whitespace-only",
			override: "   ",
			want:     "   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCheckPrompt(tt.override)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetProposePrompt(t *testing.T) {
	tests := []struct {
		name     string
		override string
		want     string
	}{
		{
			name:     "returns default when override is empty",
			override: "",
			want:     DefaultProposeSystemPrompt,
		},
		{
			name:     "returns override when provided",
			override: "Custom propose prompt",
			want:     "Custom propose prompt",
		},
		{
			name:     "returns override even if whitespace-only",
			override: "   ",
			want:     "   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetProposePrompt(tt.override)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDefaultPromptsAreNonEmpty(t *testing.T) {
	assert.NotEmpty(t, DefaultCheckSystemPrompt, "default check prompt should not be empty")
	assert.NotEmpty(t, DefaultProposeSystemPrompt, "default propose prompt should not be empty")
}
