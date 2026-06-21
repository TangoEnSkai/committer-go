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
		return nil, fmt.Errorf("%w: anthropic provider not yet wired — set %s and install the anthropic-sdk-go", ErrProviderUnavailable, apiKeyEnv)
	default:
		return nil, fmt.Errorf("%w: unknown provider %q", ErrProviderUnavailable, provider)
	}
}
