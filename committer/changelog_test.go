package committer

import (
	"strings"
	"testing"
)

func TestParseChangelogEntry(t *testing.T) {
	tests := []struct {
		line         string
		wantOk       bool
		wantType     string
		wantDesc     string
		wantScope    string
		wantBreaking bool
	}{
		{"abc1234 feat: add login", true, "feat", "add login", "", false},
		{"abc1234 fix(auth): fix token bug", true, "fix", "fix token bug", "auth", false},
		{"abc1234 feat(api)!: remove endpoint", true, "feat", "remove endpoint", "api", true},
		{"abc1234 invalid message", false, "", "", "", false},
		{"", false, "", "", "", false},
		{"abc1234 docs: update README for clarity", true, "docs", "update README for clarity", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			entry, ok := ParseChangelogEntry(tt.line)
			if ok != tt.wantOk {
				t.Errorf("ParseChangelogEntry(%q) ok = %v, want %v", tt.line, ok, tt.wantOk)
				return
			}
			if !ok {
				return
			}
			if entry.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", entry.Type, tt.wantType)
			}
			if entry.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", entry.Description, tt.wantDesc)
			}
			if entry.Scope != tt.wantScope {
				t.Errorf("Scope = %q, want %q", entry.Scope, tt.wantScope)
			}
			if entry.Breaking != tt.wantBreaking {
				t.Errorf("Breaking = %v, want %v", entry.Breaking, tt.wantBreaking)
			}
		})
	}
}

func TestBuildChangelog(t *testing.T) {
	entries := []ChangelogEntry{
		{Type: "feat", Description: "add login", Hash: "abc1234"},
		{Type: "fix", Description: "fix auth bug", Hash: "def5678"},
		{Type: "feat", Breaking: true, Description: "remove old API", Hash: "ghi9012"},
	}

	sections := BuildChangelog(entries, true)

	if len(sections) == 0 {
		t.Fatal("expected sections")
	}
	if sections[0].Title != "⚠ Breaking Changes" {
		t.Errorf("first section = %q, want breaking changes", sections[0].Title)
	}
}

func TestFormatChangelog(t *testing.T) {
	sections := []ChangelogSection{
		{
			Title: "Features",
			Entries: []ChangelogEntry{
				{Type: "feat", Description: "add login", Hash: "abc1234567"},
			},
		},
	}

	md := FormatChangelog(sections, "v1.2.0")
	if !strings.Contains(md, "## v1.2.0") {
		t.Error("expected version header")
	}
	if !strings.Contains(md, "### Features") {
		t.Error("expected Features section")
	}
	if !strings.Contains(md, "add login") {
		t.Error("expected entry description")
	}
	if !strings.Contains(md, "abc1234") {
		t.Error("expected short hash")
	}
}

func TestFormatChangelog_Unreleased(t *testing.T) {
	md := FormatChangelog(nil, "")
	if !strings.Contains(md, "[Unreleased]") {
		t.Error("expected Unreleased header when no version given")
	}
}
