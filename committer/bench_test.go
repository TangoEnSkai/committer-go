package committer

import (
	"strings"
	"testing"
)

var benchCfg = DefaultConfig()

func BenchmarkValidateMessage(b *testing.B) {
	msg := Message("feat(auth): add JWT refresh token rotation")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ValidateMessage(msg, benchCfg)
	}
}

func BenchmarkValidateMessage_Invalid(b *testing.B) {
	msg := Message("wip: stuff")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ValidateMessage(msg, benchCfg)
	}
}

func BenchmarkSuggest_ValidMessage(b *testing.B) {
	msg := Message("feat: add login")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Suggest(msg, benchCfg)
	}
}

func BenchmarkSuggest_InvalidType(b *testing.B) {
	msg := Message("wip: add login")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Suggest(msg, benchCfg)
	}
}

func BenchmarkCheckCommitTypeWithConfig_ExtraTypes(b *testing.B) {
	cfg := benchCfg
	cfg.Types.Extra = []string{"wip", "release", "hotfix", "spike", "chore2"}
	msg := Message("feat: add login")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = CheckCommitTypeWithConfig(msg, cfg)
	}
}

func BenchmarkValidateFullMessage_WithBody(b *testing.B) {
	msg := Message("feat(auth): add JWT refresh\n\nThis is the body line.\n\nRefs: #42")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ValidateFullMessage(msg, benchCfg)
	}
}

func BenchmarkFindTypeSuggestion(b *testing.B) {
	allTypes := append(validTypes, "wip", "release")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = findTypeSuggestion("fixx", allTypes)
	}
}

func BenchmarkPatternMatch(b *testing.B) {
	msg := Message("feat(auth)!: add breaking change")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = PatternMatch(msg)
	}
}

func BenchmarkValidateMessage_LongMessage(b *testing.B) {
	msg := Message("feat: " + strings.Repeat("x", 200))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ValidateMessage(msg, benchCfg)
	}
}
