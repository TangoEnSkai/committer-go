package ai

import (
	"context"
	"errors"
)

var (
	// ErrProviderUnavailable is returned when the AI provider is not configured or unavailable.
	ErrProviderUnavailable = errors.New("ai provider unavailable: set ANTHROPIC_API_KEY or configure ai.provider")
	// ErrRateLimited is returned when the AI provider rate-limits the request.
	ErrRateLimited = errors.New("ai provider rate limited")
	// ErrInvalidResponse is returned when the AI provider returns an unparseable suggestion.
	ErrInvalidResponse = errors.New("ai provider returned an invalid commit suggestion")
)

// CommitSuggestion is a structured Conventional Commits suggestion.
type CommitSuggestion struct {
	Type        string
	Scope       string
	Breaking    bool
	Description string
	Body        string
}

// Header returns the Conventional Commits header string.
func (s CommitSuggestion) Header() string {
	h := s.Type
	if s.Scope != "" {
		h += "(" + s.Scope + ")"
	}
	if s.Breaking {
		h += "!"
	}
	h += ": " + s.Description
	return h
}

// SuggestConfig carries validation rules and model preferences to the provider.
type SuggestConfig struct {
	AllowedTypes []string
	MaxHeaderLen int
	MaxDiffChars int
	Model        string
}

// AIProvider is the interface that AI backends must implement.
type AIProvider interface {
	// SuggestFromDiff generates a commit suggestion from a git diff.
	SuggestFromDiff(ctx context.Context, diff string, cfg SuggestConfig) (CommitSuggestion, error)
	// FixMessage rewrites a commit message that has lint violations.
	FixMessage(ctx context.Context, msg string, violations []string, cfg SuggestConfig) (CommitSuggestion, error)
}
