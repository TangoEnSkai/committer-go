package ai

import "fmt"

// NewFromConfig returns an AIProvider based on the given configuration fields.
// Returns ErrProviderUnavailable if enabled is false or the provider is unsupported.
func NewFromConfig(enabled bool, provider, model, apiKeyEnv string) (AIProvider, error) {
	if !enabled {
		return nil, ErrProviderUnavailable
	}
	switch provider {
	case "anthropic":
		return NewAnthropicProvider(apiKeyEnv, model)
	default:
		return nil, fmt.Errorf("%w: unknown provider %q", ErrProviderUnavailable, provider)
	}
}
