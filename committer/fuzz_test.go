package committer

import (
	"strings"
	"testing"
)

func FuzzValidateMessage(f *testing.F) {
	// Seed corpus
	seeds := []string{
		"feat: add login",
		"fix(auth)!: remove legacy token",
		"docs: update README",
		"",
		"x",
		"feat: " + strings.Repeat("x", 100),
		"invalid commit message",
		"feat(scope): description\n\nBody line\n\nBREAKING CHANGE: something changed",
		"chore!: breaking without footer",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	cfg := DefaultConfig()
	f.Fuzz(func(t *testing.T, msg string) {
		// Must never panic
		_, _ = ValidateMessage(Message(msg), cfg)
		_ = Suggest(Message(msg), cfg)
		_, _ = ValidateFullMessage(Message(msg), cfg)
	})
}

func FuzzPatternMatch(f *testing.F) {
	f.Add("feat: add login")
	f.Add("fix(auth): fix bug")
	f.Add("feat!: breaking")
	f.Add("")
	f.Add(":")
	f.Add("feat:")
	f.Add("feat: ")

	f.Fuzz(func(t *testing.T, msg string) {
		_, _ = PatternMatch(Message(msg))
	})
}

func FuzzExtractType(f *testing.F) {
	f.Add("feat: add login")
	f.Add("fix(auth): fix")
	f.Add("")
	f.Add("feat")
	f.Add("(scope): no type")

	f.Fuzz(func(t *testing.T, msg string) {
		// extractType is unexported — test via CheckCommitTypeWithConfig
		cfg := DefaultConfig()
		_, _ = CheckCommitTypeWithConfig(Message(msg), cfg)
	})
}
