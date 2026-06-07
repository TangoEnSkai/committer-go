package committer_test

import (
	"fmt"
	"testing"

	"github.com/TangoEnSkai/committer-go/committer"
)

func TestCheckLength(t *testing.T) {
	const (
		minLength         = 10
		maxLength         = 60
		shortCommit       = "short"
		validLengthCommit = "the commit length is good enough"
		longCommit        = "this commit is toooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo long"
	)
	tests := []struct {
		name       string
		args       committer.Message
		wantErrStr string
		wantOk     bool
	}{
		{
			name:       "success/valid commit message",
			args:       validLengthCommit,
			wantErrStr: "",
			wantOk:     true,
		},
		{
			name:       "success/exactly minLength",
			args:       "1234567890", // exactly 10 chars
			wantErrStr: "",
			wantOk:     true,
		},
		{
			name:       "success/exactly maxLength",
			args:       "feat: exactly sixty chars: padding here xxxxxxxxxxxxxxxxxxxx", // 60 chars
			wantErrStr: "",
			wantOk:     true,
		},
		{
			name:       "failed/one above maxLength",
			args:       "feat: exactly sixty chars: padding here xxxxxxxxxxxxxxxxxxxxx", // 61 chars
			wantErrStr: fmt.Sprintf("\tthe commit `%s` is too long: got(%d), need <= (%d)", "feat: exactly sixty chars: padding here xxxxxxxxxxxxxxxxxxxxx", 61, maxLength),
			wantOk:     false,
		},
		{
			name:       "failed/too short",
			args:       shortCommit,
			wantErrStr: fmt.Sprintf("\tthe commit `%s` is too short: got(%d), need >= (%d)", shortCommit, len(shortCommit), minLength),
			wantOk:     false,
		},
		{
			name:       "failed/too long",
			args:       longCommit,
			wantErrStr: fmt.Sprintf("\tthe commit `%s` is too long: got(%d), need <= (%d)", longCommit, len(longCommit), maxLength),
			wantOk:     false,
		},
		{
			name:       "failed/one below minLength",
			args:       "123456789", // 9 chars
			wantErrStr: fmt.Sprintf("\tthe commit `%s` is too short: got(%d), need >= (%d)", "123456789", 9, minLength),
			wantOk:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotErrStr, gotOk := committer.CheckLength(tc.args)
			if gotErrStr != tc.wantErrStr {
				t.Errorf("CheckLength() gotErrStr = %v, want %v", gotErrStr, tc.wantErrStr)
			}
			if gotOk != tc.wantOk {
				t.Errorf("CheckLength() gotOk = %v, want %v", gotOk, tc.wantOk)
			}
		})
	}
}
