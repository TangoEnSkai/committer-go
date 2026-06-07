package committer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TangoEnSkai/committer-go/committer"
)

func TestRead(t *testing.T) {
	t.Run("success/reads file content", func(t *testing.T) {
		f, err := os.CreateTemp(t.TempDir(), "commit-msg-*")
		if err != nil {
			t.Fatal(err)
		}
		content := "feat: add something"
		if _, err := f.WriteString(content); err != nil {
			t.Fatal(err)
		}
		f.Close()

		got, err := committer.Read(f.Name())
		if err != nil {
			t.Fatalf("Read() unexpected error: %v", err)
		}
		if string(got) != content {
			t.Errorf("Read() = %q, want %q", got, content)
		}
	})

	t.Run("failed/file not found", func(t *testing.T) {
		_, err := committer.Read(filepath.Join(t.TempDir(), "nonexistent"))
		if err == nil {
			t.Error("Read() expected error for missing file, got nil")
		}
	})
}

func TestNewFirstLine(t *testing.T) {
	tests := []struct {
		name    string
		input   committer.Message
		want    committer.Message
		wantErr bool
	}{
		{
			name:  "success/single line",
			input: "feat: add feature",
			want:  "feat: add feature",
		},
		{
			name:  "success/multi-line picks first",
			input: "feat: add feature\n\nThis is the body.\n\nCloses #123",
			want:  "feat: add feature",
		},
		{
			name:    "failed/empty message",
			input:   "",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := committer.NewFirstLine(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("NewFirstLine() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr && got != tc.want {
				t.Errorf("NewFirstLine() = %q, want %q", got, tc.want)
			}
		})
	}
}
