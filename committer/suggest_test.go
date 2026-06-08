package committer

import (
	"strings"
	"testing"
)

func TestSuggest_InvalidType(t *testing.T) {
	cfg := DefaultConfig()
	// "feature" is close to "feat" via Jaro-Winkler
	suggestions := Suggest(Message("feature: add login page"), cfg)
	if len(suggestions) == 0 {
		t.Fatal("expected at least one suggestion for invalid type")
	}
	got := suggestions[0].Message
	if !strings.HasPrefix(got, "feat:") {
		t.Errorf("expected suggestion to start with 'feat:', got %q", got)
	}
}

func TestSuggest_TooLong(t *testing.T) {
	cfg := DefaultConfig()
	long := "feat: " + strings.Repeat("a", 60) // total > 60
	suggestions := Suggest(Message(long), cfg)
	if len(suggestions) == 0 {
		t.Fatal("expected at least one suggestion for too-long message")
	}
	got := suggestions[0].Message
	if len(got) > cfg.Length.Max {
		t.Errorf("suggestion %q still exceeds max length %d", got, cfg.Length.Max)
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("expected suggestion to end with '...', got %q", got)
	}
}

func TestSuggest_TooShort(t *testing.T) {
	cfg := DefaultConfig()
	suggestions := Suggest(Message("feat: hi"), cfg) // 8 chars, < 10
	if len(suggestions) == 0 {
		t.Fatal("expected at least one suggestion for too-short message")
	}
}

func TestSuggest_ValidMessage(t *testing.T) {
	cfg := DefaultConfig()
	suggestions := Suggest(Message("feat: add login page"), cfg)
	if len(suggestions) != 0 {
		t.Errorf("expected no suggestions for valid message, got %v", suggestions)
	}
}

func TestSuggest_FormatWrong(t *testing.T) {
	cfg := DefaultConfig()
	// Valid type, valid length, bad format (no colon-space)
	suggestions := Suggest(Message("feat add login page here now"), cfg)
	if len(suggestions) == 0 {
		t.Fatal("expected at least one suggestion for bad format")
	}
}

func TestFormatSuggestions_Empty(t *testing.T) {
	out := FormatSuggestions(nil)
	if out != "" {
		t.Errorf("expected empty string for nil suggestions, got %q", out)
	}
}

func TestFormatSuggestions_NonEmpty(t *testing.T) {
	suggestions := []Suggestion{{Message: "feat: add login page", Reason: "test"}}
	out := FormatSuggestions(suggestions)
	if !strings.Contains(out, "feat: add login page") {
		t.Errorf("expected output to contain suggestion message, got %q", out)
	}
	if !strings.Contains(out, "Suggestions") {
		t.Errorf("expected output to contain 'Suggestions', got %q", out)
	}
}
