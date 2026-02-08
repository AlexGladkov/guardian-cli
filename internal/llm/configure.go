package llm

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/AlexGladkov/guardian-cli/internal/config"
)

// providerOption represents a selectable provider in the interactive flow.
type providerOption struct {
	number   int
	name     string
	provider string
}

var providerOptions = []providerOption{
	{1, "DeepSeek", ProviderDeepSeek},
	{2, "OpenAI", ProviderOpenAI},
	{3, "Claude (Anthropic)", ProviderClaude},
	{4, "Custom (OpenAI-compatible)", ProviderCustom},
}

// RunConfigure runs the interactive LLM configuration flow.
// It reads from r and writes prompts/output to w.
// Returns the updated LLMConfig based on user selections.
func RunConfigure(r io.Reader, w io.Writer) (*config.LLMConfig, error) {
	scanner := bufio.NewScanner(r)

	// Step 1: Select provider
	fmt.Fprintln(w, "Select LLM provider:")
	for _, opt := range providerOptions {
		fmt.Fprintf(w, "  %d. %s\n", opt.number, opt.name)
	}
	fmt.Fprint(w, "Enter selection (1-4): ")

	if !scanner.Scan() {
		return nil, fmt.Errorf("failed to read provider selection: %w", scanner.Err())
	}
	selection := strings.TrimSpace(scanner.Text())

	var selectedProvider string
	var selectedEndpoint string

	switch selection {
	case "1":
		selectedProvider = ProviderDeepSeek
	case "2":
		selectedProvider = ProviderOpenAI
	case "3":
		selectedProvider = ProviderClaude
	case "4":
		selectedProvider = ProviderCustom
	default:
		return nil, fmt.Errorf("invalid selection %q; please enter 1, 2, 3, or 4", selection)
	}

	// Step 2: If custom, ask for endpoint URL
	if selectedProvider == ProviderCustom {
		fmt.Fprint(w, "Enter custom endpoint URL (e.g., http://localhost:11434/v1): ")
		if !scanner.Scan() {
			return nil, fmt.Errorf("failed to read endpoint URL: %w", scanner.Err())
		}
		selectedEndpoint = strings.TrimSpace(scanner.Text())
		if selectedEndpoint == "" {
			return nil, fmt.Errorf("endpoint URL is required for custom provider")
		}
	}

	// Step 3: Privacy warning for cloud providers
	if IsCloudProvider(selectedProvider) {
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "WARNING: You selected a cloud provider. Code diffs will be sent to an external API.")
		fmt.Fprintln(w, "For confidential code, consider using a custom provider with a local endpoint")
		fmt.Fprintln(w, "(e.g., Ollama at http://localhost:11434/v1).")
	}

	// Step 4: API key instructions
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "Set the environment variable %s with your API key:\n", apiKeyEnvVar)
	fmt.Fprintf(w, "  export %s=<your-api-key>\n", apiKeyEnvVar)
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "LLM configuration complete.")

	// Build config
	cfg := &config.LLMConfig{
		Provider: selectedProvider,
		Endpoint: selectedEndpoint,
	}

	// Set default endpoint for known providers (leave empty, will be resolved at runtime)
	// Set model default
	if model, ok := DefaultModels[selectedProvider]; ok {
		cfg.Model = model
	}

	return cfg, nil
}
