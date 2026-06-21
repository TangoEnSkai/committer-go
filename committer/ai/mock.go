package ai

import "context"

// MockProvider is a test double for AIProvider.
type MockProvider struct {
	SuggestResult CommitSuggestion
	SuggestErr    error
	FixResult     CommitSuggestion
	FixErr        error
}

func (m *MockProvider) SuggestFromDiff(_ context.Context, _ string, _ SuggestConfig) (CommitSuggestion, error) {
	return m.SuggestResult, m.SuggestErr
}

func (m *MockProvider) FixMessage(_ context.Context, _ string, _ []string, _ SuggestConfig) (CommitSuggestion, error) {
	return m.FixResult, m.FixErr
}
