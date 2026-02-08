package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AlexGladkov/guardian-cli/internal/config"
)

func TestNewClient_MissingAPIKey(t *testing.T) {
	t.Setenv("GUARDIAN_LLM_API_KEY", "")

	cfg := config.LLMConfig{
		Provider: ProviderOpenAI,
	}

	client, err := NewClient(cfg)
	assert.Nil(t, client)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "GUARDIAN_LLM_API_KEY")
}

func TestNewClient_ValidConfig(t *testing.T) {
	t.Setenv("GUARDIAN_LLM_API_KEY", "test-key-12345")

	tests := []struct {
		name             string
		cfg              config.LLMConfig
		wantProvider     string
		wantModel        string
		wantEndpointSub  string
	}{
		{
			name: "deepseek provider with defaults",
			cfg: config.LLMConfig{
				Provider: ProviderDeepSeek,
			},
			wantProvider:    ProviderDeepSeek,
			wantModel:       "deepseek-chat",
			wantEndpointSub: "deepseek.com",
		},
		{
			name: "openai provider with defaults",
			cfg: config.LLMConfig{
				Provider: ProviderOpenAI,
			},
			wantProvider:    ProviderOpenAI,
			wantModel:       "gpt-4o",
			wantEndpointSub: "openai.com",
		},
		{
			name: "claude provider with defaults",
			cfg: config.LLMConfig{
				Provider: ProviderClaude,
			},
			wantProvider:    ProviderClaude,
			wantModel:       "claude-sonnet-4-5-20250929",
			wantEndpointSub: "anthropic.com",
		},
		{
			name: "custom provider with endpoint",
			cfg: config.LLMConfig{
				Provider: ProviderCustom,
				Endpoint: "http://localhost:11434/v1",
				Model:    "llama3",
			},
			wantProvider:    ProviderCustom,
			wantModel:       "llama3",
			wantEndpointSub: "localhost",
		},
		{
			name: "model override",
			cfg: config.LLMConfig{
				Provider: ProviderOpenAI,
				Model:    "gpt-3.5-turbo",
			},
			wantProvider:    ProviderOpenAI,
			wantModel:       "gpt-3.5-turbo",
			wantEndpointSub: "openai.com",
		},
		{
			name: "empty provider defaults to deepseek",
			cfg:  config.LLMConfig{},
			wantProvider:    ProviderDeepSeek,
			wantModel:       "deepseek-chat",
			wantEndpointSub: "deepseek.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.cfg)
			require.NoError(t, err)
			require.NotNil(t, client)

			assert.Equal(t, tt.wantProvider, client.provider)
			assert.Equal(t, tt.wantModel, client.model)
			assert.Contains(t, client.endpoint, tt.wantEndpointSub)
			assert.Equal(t, "test-key-12345", client.apiKey)
		})
	}
}

func TestNewClient_CustomProviderNoEndpoint(t *testing.T) {
	t.Setenv("GUARDIAN_LLM_API_KEY", "test-key")

	cfg := config.LLMConfig{
		Provider: ProviderCustom,
	}

	client, err := NewClient(cfg)
	assert.Nil(t, client)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no endpoint configured")
}

func TestAnalyzeCheck_OpenAI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request format
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Parse request body
		var req openAIRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "gpt-4o", req.Model)
		assert.Len(t, req.Messages, 2)
		assert.Equal(t, "system", req.Messages[0].Role)
		assert.Equal(t, "user", req.Messages[1].Role)

		// Return response
		resp := openAIResponse{
			Choices: []struct {
				Message openAIMessage `json:"message"`
			}{
				{
					Message: openAIMessage{
						Role:    "assistant",
						Content: "[test_rule] This is a test explanation for the violation.",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	t.Setenv("GUARDIAN_LLM_API_KEY", "test-api-key")

	cfg := config.LLMConfig{
		Provider: ProviderOpenAI,
		Endpoint: server.URL + "/v1",
		Model:    "gpt-4o",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	rules := []config.Rule{
		{
			ID:          "test_rule",
			Description: "Test rule description",
			Type:        "diff_pattern_forbidden",
			Severity:    "error",
		},
	}

	violations := []Violation{
		{
			RuleID:      "test_rule",
			Severity:    "error",
			Description: "Test violation",
			FilePath:    "test.go",
			DiffSnippet: "+ badImport",
		},
	}

	analysis, err := client.AnalyzeCheck("diff content", rules, violations)
	require.NoError(t, err)
	require.NotNil(t, analysis)
	assert.Contains(t, analysis.Explanations, "test_rule")
	assert.Contains(t, analysis.Explanations["test_rule"], "test explanation")
}

func TestAnalyzeCheck_Claude(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Claude-specific request format
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/messages", r.URL.Path)
		assert.Equal(t, "test-claude-key", r.Header.Get("x-api-key"))
		assert.Equal(t, "2023-06-01", r.Header.Get("anthropic-version"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify no Bearer token for Claude
		assert.Empty(t, r.Header.Get("Authorization"))

		// Parse request body
		var req claudeRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "claude-sonnet-4-5-20250929", req.Model)
		assert.Equal(t, 4096, req.MaxTokens)
		assert.NotEmpty(t, req.System)
		assert.Len(t, req.Messages, 1)
		assert.Equal(t, "user", req.Messages[0].Role)

		// Return Claude-format response
		resp := claudeResponse{
			Content: []struct {
				Text string `json:"text"`
			}{
				{Text: "[claude_rule] Claude detected this violation."},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	t.Setenv("GUARDIAN_LLM_API_KEY", "test-claude-key")

	cfg := config.LLMConfig{
		Provider: ProviderClaude,
		Endpoint: server.URL + "/v1",
		Model:    "claude-sonnet-4-5-20250929",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	rules := []config.Rule{
		{
			ID:          "claude_rule",
			Description: "Claude test rule",
			Type:        "imports_forbidden",
			Severity:    "error",
		},
	}

	violations := []Violation{
		{
			RuleID:      "claude_rule",
			Severity:    "error",
			Description: "Import violation",
			FilePath:    "main.go",
			DiffSnippet: "+ import forbidden/pkg",
		},
	}

	analysis, err := client.AnalyzeCheck("diff", rules, violations)
	require.NoError(t, err)
	require.NotNil(t, analysis)
	assert.Contains(t, analysis.Explanations, "claude_rule")
}

func TestDraftProposal_OpenAI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)

		resp := openAIResponse{
			Choices: []struct {
				Message openAIMessage `json:"message"`
			}{
				{
					Message: openAIMessage{
						Role: "assistant",
						Content: `CHANGE_DESCRIPTION: Allow infra imports in adapters
CHANGE_DETAILS: Modify the domain_no_infra rule to exclude domain/adapters/ directory
REASON: Domain adapters need to implement infra interfaces by design
IMPACT: Files in domain/adapters/ can now import from infra/ package`,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	t.Setenv("GUARDIAN_LLM_API_KEY", "test-key")

	cfg := config.LLMConfig{
		Provider: ProviderOpenAI,
		Endpoint: server.URL + "/v1",
		Model:    "gpt-4o",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	rule := config.Rule{
		ID:          "domain_no_infra",
		Description: "Domain layer must not depend on infra",
		Type:        "imports_forbidden",
		Severity:    "error",
	}

	draft, err := client.DraftProposal(rule, "We need adapters to access infra")
	require.NoError(t, err)
	require.NotNil(t, draft)
	assert.Contains(t, draft.ChangeDescription, "infra imports")
	assert.Contains(t, draft.ChangeDetails, "domain_no_infra")
	assert.Contains(t, draft.Reason, "adapters")
	assert.Contains(t, draft.Impact, "domain/adapters/")
}

func TestDraftProposal_Claude(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/messages", r.URL.Path)
		assert.Equal(t, "test-claude-key", r.Header.Get("x-api-key"))

		resp := claudeResponse{
			Content: []struct {
				Text string `json:"text"`
			}{
				{
					Text: `CHANGE_DESCRIPTION: Relax float restriction for display-only values
CHANGE_DETAILS: Update money_minor_units rule to allow float in display layers
REASON: Display formatting often requires float conversion
IMPACT: Presentation layer files can use float types`,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	t.Setenv("GUARDIAN_LLM_API_KEY", "test-claude-key")

	cfg := config.LLMConfig{
		Provider: ProviderClaude,
		Endpoint: server.URL + "/v1",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	rule := config.Rule{
		ID:          "money_minor_units",
		Description: "Money must use int minor units",
		Severity:    "warning",
	}

	draft, err := client.DraftProposal(rule, "")
	require.NoError(t, err)
	require.NotNil(t, draft)
	assert.NotEmpty(t, draft.ChangeDescription)
	assert.NotEmpty(t, draft.Reason)
}

func TestAnalyzeCheck_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	t.Setenv("GUARDIAN_LLM_API_KEY", "test-key")

	cfg := config.LLMConfig{
		Provider: ProviderOpenAI,
		Endpoint: server.URL + "/v1",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	_, err = client.AnalyzeCheck("diff", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 500")
}

func TestAnalyzeCheck_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAIResponse{
			Error: &struct {
				Message string `json:"message"`
			}{
				Message: "Rate limit exceeded",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	t.Setenv("GUARDIAN_LLM_API_KEY", "test-key")

	cfg := config.LLMConfig{
		Provider: ProviderOpenAI,
		Endpoint: server.URL + "/v1",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	_, err = client.AnalyzeCheck("diff", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Rate limit exceeded")
}

func TestAnalyzeCheck_EmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAIResponse{
			Choices: []struct {
				Message openAIMessage `json:"message"`
			}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	t.Setenv("GUARDIAN_LLM_API_KEY", "test-key")

	cfg := config.LLMConfig{
		Provider: ProviderOpenAI,
		Endpoint: server.URL + "/v1",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	_, err = client.AnalyzeCheck("diff", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no choices")
}

func TestAnalyzeCheck_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("GUARDIAN_LLM_API_KEY", "test-key")

	cfg := config.LLMConfig{
		Provider: ProviderOpenAI,
		Endpoint: server.URL + "/v1",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	// Override the timeout to be short for testing
	client.httpClient.Timeout = 100 * time.Millisecond

	_, err = client.AnalyzeCheck("diff", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sending request")
}

func TestAnalyzeCheck_ClaudeAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := claudeResponse{
			Error: &struct {
				Message string `json:"message"`
			}{
				Message: "Authentication failed",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	t.Setenv("GUARDIAN_LLM_API_KEY", "test-key")

	cfg := config.LLMConfig{
		Provider: ProviderClaude,
		Endpoint: server.URL + "/v1",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	_, err = client.AnalyzeCheck("diff", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Authentication failed")
}

func TestAnalyzeCheck_ClaudeEmptyContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := claudeResponse{
			Content: []struct {
				Text string `json:"text"`
			}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	t.Setenv("GUARDIAN_LLM_API_KEY", "test-key")

	cfg := config.LLMConfig{
		Provider: ProviderClaude,
		Endpoint: server.URL + "/v1",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	_, err = client.AnalyzeCheck("diff", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no content")
}

func TestParseExplanations(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		violations []Violation
		wantKeys   []string
	}{
		{
			name:     "structured response with rule IDs",
			response: "[rule_a] This is the explanation for rule A.\n[rule_b] This is the explanation for rule B.",
			violations: []Violation{
				{RuleID: "rule_a"},
				{RuleID: "rule_b"},
			},
			wantKeys: []string{"rule_a", "rule_b"},
		},
		{
			name:     "unstructured response falls back to all violations",
			response: "General explanation for all violations.",
			violations: []Violation{
				{RuleID: "rule_x"},
			},
			wantKeys: []string{"rule_x"},
		},
		{
			name:       "empty violations returns empty map",
			response:   "Some response",
			violations: []Violation{},
			wantKeys:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseExplanations(tt.response, tt.violations)
			for _, key := range tt.wantKeys {
				assert.Contains(t, result, key)
				assert.NotEmpty(t, result[key])
			}
		})
	}
}

func TestParseProposalDraft(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     ProposalDraft
	}{
		{
			name: "structured sections",
			response: `CHANGE_DESCRIPTION: Allow float in display
CHANGE_DETAILS: Update rule config
REASON: Display needs float
IMPACT: Display files affected`,
			want: ProposalDraft{
				ChangeDescription: "Allow float in display",
				ChangeDetails:     "Update rule config",
				Reason:            "Display needs float",
				Impact:            "Display files affected",
			},
		},
		{
			name:     "unstructured falls back to description",
			response: "Just some plain text proposal",
			want: ProposalDraft{
				ChangeDescription: "Just some plain text proposal",
			},
		},
		{
			name: "multiline sections",
			response: `CHANGE_DESCRIPTION: Allow float
CHANGE_DETAILS: First line
Second line
Third line
REASON: Some reason
IMPACT: Some impact`,
			want: ProposalDraft{
				ChangeDescription: "Allow float",
				ChangeDetails:     "First line\nSecond line\nThird line",
				Reason:            "Some reason",
				Impact:            "Some impact",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			draft := parseProposalDraft(tt.response)
			require.NotNil(t, draft)
			assert.Equal(t, tt.want.ChangeDescription, draft.ChangeDescription)
			assert.Equal(t, tt.want.ChangeDetails, draft.ChangeDetails)
			assert.Equal(t, tt.want.Reason, draft.Reason)
			assert.Equal(t, tt.want.Impact, draft.Impact)
		})
	}
}

func TestNewClient_WithPromptOverrides(t *testing.T) {
	t.Setenv("GUARDIAN_LLM_API_KEY", "test-key")

	cfg := config.LLMConfig{
		Provider: ProviderOpenAI,
		Prompts: config.LLMPrompts{
			CheckSystem:   "Custom check prompt",
			ProposeSystem: "Custom propose prompt",
		},
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	assert.Equal(t, "Custom check prompt", client.prompts.CheckSystem)
	assert.Equal(t, "Custom propose prompt", client.prompts.ProposeSystem)
}

func TestAnalyzeCheck_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	t.Setenv("GUARDIAN_LLM_API_KEY", "test-key")

	cfg := config.LLMConfig{
		Provider: ProviderOpenAI,
		Endpoint: server.URL + "/v1",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	_, err = client.AnalyzeCheck("diff", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing response JSON")
}
