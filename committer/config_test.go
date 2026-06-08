package committer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TangoEnSkai/committer-go/committer"
)

func TestLoadConfig_NoFile(t *testing.T) {
	// Change to a temp dir that has no .committer.yaml
	tmp := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	cfg, err := committer.LoadConfig()
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}

	def := committer.DefaultConfig()
	if cfg.Length.Min != def.Length.Min {
		t.Errorf("Min: got %d, want %d", cfg.Length.Min, def.Length.Min)
	}
	if cfg.Length.Max != def.Length.Max {
		t.Errorf("Max: got %d, want %d", cfg.Length.Max, def.Length.Max)
	}
	if len(cfg.Types.Extra) != 0 {
		t.Errorf("Extra: got %v, want empty", cfg.Types.Extra)
	}
}

func TestLoadConfig_CustomValues(t *testing.T) {
	tmp := t.TempDir()
	content := []byte("length:\n  min: 5\n  max: 100\ntypes:\n  extra:\n    - mytype\n    - hotfix\n")
	if err := os.WriteFile(filepath.Join(tmp, ".committer.yaml"), content, 0o644); err != nil {
		t.Fatal(err)
	}

	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	cfg, err := committer.LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Length.Min != 5 {
		t.Errorf("Min: got %d, want 5", cfg.Length.Min)
	}
	if cfg.Length.Max != 100 {
		t.Errorf("Max: got %d, want 100", cfg.Length.Max)
	}
	if len(cfg.Types.Extra) != 2 || cfg.Types.Extra[0] != "mytype" || cfg.Types.Extra[1] != "hotfix" {
		t.Errorf("Extra: got %v, want [mytype hotfix]", cfg.Types.Extra)
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, ".committer.yaml"), []byte(":::invalid"), 0o644); err != nil {
		t.Fatal(err)
	}

	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	_, err := committer.LoadConfig()
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestCheckLengthWithConfig(t *testing.T) {
	cfg := committer.Config{
		Length: committer.LengthConfig{Min: 5, Max: 20},
	}

	tests := []struct {
		name   string
		msg    committer.Message
		wantOk bool
	}{
		{"too short", "hi", false},
		{"exactly min", "hello", true},
		{"valid", "feat: do thing", true},
		{"exactly max", "12345678901234567890", true},
		{"too long", "123456789012345678901", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, ok := committer.CheckLengthWithConfig(tc.msg, cfg)
			if ok != tc.wantOk {
				t.Errorf("CheckLengthWithConfig(%q) ok=%v, want %v", tc.msg, ok, tc.wantOk)
			}
		})
	}
}

func TestLoadConfig_Extends_LocalFile(t *testing.T) {
	tmp := t.TempDir()

	// Write the base config.
	baseContent := []byte("length:\n  min: 5\n  max: 80\ntypes:\n  extra:\n    - hotfix\n")
	basePath := filepath.Join(tmp, "base.committer.yaml")
	if err := os.WriteFile(basePath, baseContent, 0o644); err != nil {
		t.Fatal(err)
	}

	// Write the local config that extends base.
	localContent := []byte("extends: " + basePath + "\nlength:\n  max: 100\ntypes:\n  extra:\n    - wip\n")
	if err := os.WriteFile(filepath.Join(tmp, ".committer.yaml"), localContent, 0o644); err != nil {
		t.Fatal(err)
	}

	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	cfg, err := committer.LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Local overrides base for max.
	if cfg.Length.Max != 100 {
		t.Errorf("Length.Max: got %d, want 100", cfg.Length.Max)
	}
	// Base value for min is inherited.
	if cfg.Length.Min != 5 {
		t.Errorf("Length.Min: got %d, want 5", cfg.Length.Min)
	}
	// Extra types from both base and local are present.
	found := map[string]bool{}
	for _, e := range cfg.Types.Extra {
		found[e] = true
	}
	if !found["hotfix"] {
		t.Error("expected base extra type 'hotfix' in merged config")
	}
	if !found["wip"] {
		t.Error("expected local extra type 'wip' in merged config")
	}
}

func TestLoadConfig_Extends_NotFound(t *testing.T) {
	tmp := t.TempDir()
	localContent := []byte("extends: /nonexistent/path/.committer.yaml\n")
	if err := os.WriteFile(filepath.Join(tmp, ".committer.yaml"), localContent, 0o644); err != nil {
		t.Fatal(err)
	}

	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	_, err := committer.LoadConfig()
	if err == nil {
		t.Error("expected error when extends file not found, got nil")
	}
}

func TestCheckCommitTypeWithConfig_ExtraTypes(t *testing.T) {
	cfg := committer.Config{
		Length: committer.LengthConfig{Min: 10, Max: 60},
		Types:  committer.TypesConfig{Extra: []string{"hotfix", "mytype"}},
	}

	tests := []struct {
		name   string
		msg    committer.Message
		wantOk bool
	}{
		{"built-in feat", "feat: something", true},
		{"extra hotfix", "hotfix: patch critical bug", true},
		{"extra mytype", "mytype: do stuff here now", true},
		{"unknown", "unknown: blah", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, ok := committer.CheckCommitTypeWithConfig(tc.msg, cfg)
			if ok != tc.wantOk {
				t.Errorf("CheckCommitTypeWithConfig(%q) ok=%v, want %v", tc.msg, ok, tc.wantOk)
			}
		})
	}
}
