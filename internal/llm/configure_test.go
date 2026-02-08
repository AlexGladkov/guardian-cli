package llm

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunConfigure_DeepSeek(t *testing.T) {
	input := strings.NewReader("1\n")
	output := &bytes.Buffer{}

	cfg, err := RunConfigure(input, output)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, ProviderDeepSeek, cfg.Provider)
	assert.Empty(t, cfg.Endpoint)
	assert.Equal(t, "deepseek-chat", cfg.Model)

	// Should show privacy warning for cloud provider
	assert.Contains(t, output.String(), "WARNING")
	assert.Contains(t, output.String(), "cloud provider")
	assert.Contains(t, output.String(), "GUARDIAN_LLM_API_KEY")
}

func TestRunConfigure_OpenAI(t *testing.T) {
	input := strings.NewReader("2\n")
	output := &bytes.Buffer{}

	cfg, err := RunConfigure(input, output)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, ProviderOpenAI, cfg.Provider)
	assert.Empty(t, cfg.Endpoint)
	assert.Equal(t, "gpt-4o", cfg.Model)

	// Should show privacy warning
	assert.Contains(t, output.String(), "WARNING")
}

func TestRunConfigure_Claude(t *testing.T) {
	input := strings.NewReader("3\n")
	output := &bytes.Buffer{}

	cfg, err := RunConfigure(input, output)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, ProviderClaude, cfg.Provider)
	assert.Empty(t, cfg.Endpoint)
	assert.Equal(t, "claude-sonnet-4-5-20250929", cfg.Model)

	// Should show privacy warning
	assert.Contains(t, output.String(), "WARNING")
}

func TestRunConfigure_Custom(t *testing.T) {
	input := strings.NewReader("4\nhttp://localhost:11434/v1\n")
	output := &bytes.Buffer{}

	cfg, err := RunConfigure(input, output)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, ProviderCustom, cfg.Provider)
	assert.Equal(t, "http://localhost:11434/v1", cfg.Endpoint)

	// Should NOT show privacy warning for custom
	assert.NotContains(t, output.String(), "WARNING")
	// Should show API key instructions
	assert.Contains(t, output.String(), "GUARDIAN_LLM_API_KEY")
}

func TestRunConfigure_CustomEmptyEndpoint(t *testing.T) {
	input := strings.NewReader("4\n\n")
	output := &bytes.Buffer{}

	cfg, err := RunConfigure(input, output)
	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint URL is required")
}

func TestRunConfigure_InvalidSelection(t *testing.T) {
	input := strings.NewReader("5\n")
	output := &bytes.Buffer{}

	cfg, err := RunConfigure(input, output)
	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid selection")
}

func TestRunConfigure_EmptyInput(t *testing.T) {
	input := strings.NewReader("")
	output := &bytes.Buffer{}

	cfg, err := RunConfigure(input, output)
	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read")
}

func TestRunConfigure_ShowsProviderOptions(t *testing.T) {
	input := strings.NewReader("1\n")
	output := &bytes.Buffer{}

	_, err := RunConfigure(input, output)
	require.NoError(t, err)

	out := output.String()
	assert.Contains(t, out, "Select LLM provider:")
	assert.Contains(t, out, "1. DeepSeek")
	assert.Contains(t, out, "2. OpenAI")
	assert.Contains(t, out, "3. Claude")
	assert.Contains(t, out, "4. Custom")
}

func TestRunConfigure_NonNumericInput(t *testing.T) {
	input := strings.NewReader("abc\n")
	output := &bytes.Buffer{}

	cfg, err := RunConfigure(input, output)
	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid selection")
}

func TestRunConfigure_CustomEndpointEOF(t *testing.T) {
	// Select custom but then EOF before endpoint
	input := strings.NewReader("4\n")
	output := &bytes.Buffer{}

	// The scanner will read "4", then when trying to read endpoint,
	// there is nothing left but Scan returns false (EOF)
	cfg, err := RunConfigure(input, output)
	// The newline after "4" is consumed by the first Scan.
	// The second Scan will return false because nothing is left.
	assert.Nil(t, cfg)
	require.Error(t, err)
}

func TestRunConfigure_ShowsCompletionMessage(t *testing.T) {
	input := strings.NewReader("1\n")
	output := &bytes.Buffer{}

	_, err := RunConfigure(input, output)
	require.NoError(t, err)

	assert.Contains(t, output.String(), "LLM configuration complete")
}
