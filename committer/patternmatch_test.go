package committer_test

import (
	"fmt"
	"testing"

	"github.com/TangoEnSkai/committer-go/committer"
)

func TestPatternMatch(t *testing.T) {
	const pattern = `^(BREAKING CHANGE|build|chore|ci|deps|docs|feat|fix|perf|refactor|revert|style|test)(\([a-z0-9 \-]+\))?!?: [\w \-]+$`

	errorMessage := fmt.Sprintf("invalid commit message. \n\tmust follow this rule: %v\n\t\t", pattern)

	tests := []struct {
		name       string
		input      committer.Message
		wantErrMsg string
		wantOk     bool
	}{
		// valid
		{"success/simple", "perf: optimise pattern matching", "", true},
		{"success/with scope", "feat(api): add new endpoint", "", true},
		{"success/with numeric scope", "fix(v2): correct parser", "", true},
		{"success/breaking change bang", "feat!: remove deprecated API", "", true},
		{"success/scope and breaking", "feat(auth)!: require 2FA", "", true},
		{"success/revert", "revert: undo login change", "", true},
		{"success/deps", "deps: upgrade go-chi", "", true},
		// invalid
		{"failed/unknown type", "performance: is not valid type", errorMessage, false},
		{"failed/missing colon space", "feat add feature", errorMessage, false},
		{"failed/no space after colon", "fix:fix the bug", errorMessage, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotErrMsg, gotOk := committer.PatternMatch(tc.input)
			if gotErrMsg != tc.wantErrMsg {
				t.Errorf("PatternMatch() gotErrMsg = %q, want %q", gotErrMsg, tc.wantErrMsg)
			}
			if gotOk != tc.wantOk {
				t.Errorf("PatternMatch() gotOk = %v, want %v", gotOk, tc.wantOk)
			}
		})
	}
}
