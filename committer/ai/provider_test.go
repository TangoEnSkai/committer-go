package ai_test

import (
	"context"
	"errors"
	"testing"

	"github.com/TangoEnSkai/committer-go/committer/ai"
)

func TestMockProvider_SuggestFromDiff(t *testing.T) {
	want := ai.CommitSuggestion{Type: "feat", Description: "add login"}
	m := &ai.MockProvider{SuggestResult: want}
	got, err := m.SuggestFromDiff(context.Background(), "diff", ai.SuggestConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMockProvider_FixMessage(t *testing.T) {
	want := ai.CommitSuggestion{Type: "fix", Description: "correct message"}
	m := &ai.MockProvider{FixResult: want}
	got, err := m.FixMessage(context.Background(), "bad msg", []string{"type must be lowercase"}, ai.SuggestConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestCommitSuggestionHeader(t *testing.T) {
	tests := []struct {
		s    ai.CommitSuggestion
		want string
	}{
		{ai.CommitSuggestion{Type: "feat", Description: "add login"}, "feat: add login"},
		{ai.CommitSuggestion{Type: "fix", Scope: "auth", Description: "fix token"}, "fix(auth): fix token"},
		{ai.CommitSuggestion{Type: "feat", Breaking: true, Description: "remove api"}, "feat!: remove api"},
		{ai.CommitSuggestion{Type: "fix", Scope: "db", Breaking: true, Description: "drop column"}, "fix(db)!: drop column"},
	}
	for _, tt := range tests {
		if got := tt.s.Header(); got != tt.want {
			t.Errorf("Header() = %q, want %q", got, tt.want)
		}
	}
}

func TestNewFromConfig_Disabled(t *testing.T) {
	_, err := ai.NewFromConfig(false, "anthropic", "claude-haiku-4-5-20251001", "ANTHROPIC_API_KEY")
	if err == nil {
		t.Fatal("expected error when AI disabled")
	}
	if !errors.Is(err, ai.ErrProviderUnavailable) {
		t.Errorf("got %v, want ErrProviderUnavailable", err)
	}
}

func TestNewFromConfig_AnthropicNotWired(t *testing.T) {
	_, err := ai.NewFromConfig(true, "anthropic", "claude-haiku-4-5-20251001", "ANTHROPIC_API_KEY")
	if err == nil {
		t.Fatal("expected error for not-yet-wired anthropic provider")
	}
	if !errors.Is(err, ai.ErrProviderUnavailable) {
		t.Errorf("got %v, want ErrProviderUnavailable", err)
	}
}

func TestNewFromConfig_UnknownProvider(t *testing.T) {
	_, err := ai.NewFromConfig(true, "openai", "", "")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
	if !errors.Is(err, ai.ErrProviderUnavailable) {
		t.Errorf("got %v, want ErrProviderUnavailable", err)
	}
}
