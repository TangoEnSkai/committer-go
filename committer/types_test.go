package committer_test

import (
	"testing"

	"github.com/TangoEnSkai/committer-go/committer"
)

func TestCheckCommitType(t *testing.T) {
	tests := []struct {
		name       string
		input      committer.Message
		wantOk     bool
	}{
		// valid types
		{"success/feat", "feat: add new feature", true},
		{"success/fix", "fix: resolve null pointer", true},
		{"success/docs", "docs: update readme", true},
		{"success/style", "style: reformat code", true},
		{"success/refactor", "refactor: extract method", true},
		{"success/perf", "perf: improve query speed", true},
		{"success/test", "test: add unit tests", true},
		{"success/build", "build: update deps", true},
		{"success/ci", "ci: add lint step", true},
		{"success/chore", "chore: remove unused file", true},
		{"success/revert", "revert: undo last commit", true},
		{"success/deps", "deps: upgrade lodash", true},
		{"success/feat with scope", "feat(api): add endpoint", true},
		{"success/feat breaking", "feat!: remove old API", true},
		// invalid types
		{"failed/unknown type", "unknown: something", false},
		{"failed/performance", "performance: too verbose", false},
		{"failed/empty", ":", false},
		{"failed/feature", "feature: not valid", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, gotOk := committer.CheckCommitType(tc.input)
			if gotOk != tc.wantOk {
				t.Errorf("CheckCommitType(%q) ok = %v, want %v", tc.input, gotOk, tc.wantOk)
			}
		})
	}
}
