package committer_test

import (
	"testing"

	"github.com/TangoEnSkai/committer-go/committer"
)

func TestValidateFullMessage(t *testing.T) {
	cfg := committer.DefaultConfig()

	tests := []struct {
		name      string
		msg       committer.Message
		wantValid bool
		wantCodes []committer.ErrorCode
	}{
		{
			name:      "valid single line",
			msg:       "feat: add login page",
			wantValid: true,
		},
		{
			name:      "valid multi-line with blank separator",
			msg:       "feat: add login page\n\nThis adds the login page with OAuth support.",
			wantValid: true,
		},
		{
			name:      "missing blank separator on line 2",
			msg:       "feat: add login page\nThis is wrong",
			wantValid: false,
			wantCodes: []committer.ErrorCode{"INVALID_BODY_SEPARATOR"},
		},
		{
			name:      "body line too long",
			msg:       committer.Message("feat: add login page\n\n" + string(make([]byte, 80))),
			wantValid: false,
			wantCodes: []committer.ErrorCode{"BODY_LINE_TOO_LONG"},
		},
		{
			name:      "breaking change bang without footer is flagged",
			msg:       "feat!: remove old API\n\nThis removes the deprecated API.",
			wantValid: false,
			wantCodes: []committer.ErrorCode{"INVALID_FORMAT"},
		},
		{
			name:      "breaking change bang with footer is valid",
			msg:       "feat!: remove old API\n\nThis removes the deprecated API.\n\nBREAKING CHANGE: old API is gone",
			wantValid: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			errs, ok := committer.ValidateFullMessage(tc.msg, cfg)
			if ok != tc.wantValid {
				t.Errorf("ValidateFullMessage(%q) valid = %v, want %v; errors: %v", tc.msg, ok, tc.wantValid, errs)
			}
			for _, wantCode := range tc.wantCodes {
				found := false
				for _, e := range errs {
					if e.Code == wantCode {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error code %q, got: %v", wantCode, errs)
				}
			}
		})
	}
}
