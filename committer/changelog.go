package committer

import (
	"fmt"
	"os/exec"
	"strings"
)

// ChangelogEntry represents a single parsed commit for the changelog.
type ChangelogEntry struct {
	Hash        string
	Type        string
	Scope       string
	Breaking    bool
	Description string
	Raw         string
}

// ChangelogSection groups entries by type.
type ChangelogSection struct {
	Title   string
	Entries []ChangelogEntry
}

// sectionOrder defines display order and headings for changelog sections.
var sectionOrder = []struct {
	Types []string
	Title string
}{
	{[]string{"feat"}, "Features"},
	{[]string{"fix"}, "Bug Fixes"},
	{[]string{"perf"}, "Performance"},
	{[]string{"refactor"}, "Refactoring"},
	{[]string{"docs"}, "Documentation"},
	{[]string{"build", "ci"}, "Build & CI"},
	{[]string{"chore", "deps"}, "Chores"},
	{[]string{"test"}, "Tests"},
	{[]string{"style"}, "Style"},
	{[]string{"revert"}, "Reverts"},
}

// ParseChangelogEntry parses a "HASH SUBJECT" line into a ChangelogEntry.
// Returns (entry, true) if the subject is a valid Conventional Commit.
func ParseChangelogEntry(line string) (ChangelogEntry, bool) {
	parts := strings.SplitN(line, " ", 2)
	if len(parts) != 2 {
		return ChangelogEntry{}, false
	}
	hash, subject := parts[0], parts[1]

	msg := Message(subject)
	if _, ok := ValidateMessage(msg, DefaultConfig()); !ok {
		return ChangelogEntry{Raw: subject}, false
	}

	entry := ChangelogEntry{
		Hash: hash,
		Raw:  subject,
	}

	// Parse type, scope, breaking from subject
	rest := subject
	typeEnd := strings.IndexAny(rest, ":(!")
	if typeEnd < 0 {
		return ChangelogEntry{}, false
	}
	entry.Type = rest[:typeEnd]
	rest = rest[typeEnd:]

	if strings.HasPrefix(rest, "(") {
		scopeEnd := strings.Index(rest, ")")
		if scopeEnd > 0 {
			entry.Scope = rest[1:scopeEnd]
			rest = rest[scopeEnd+1:]
		}
	}

	if strings.HasPrefix(rest, "!") {
		entry.Breaking = true
		rest = rest[1:]
	}

	if strings.HasPrefix(rest, ": ") {
		entry.Description = rest[2:]
	}

	return entry, true
}

// BuildChangelog groups entries into sections for markdown output.
func BuildChangelog(entries []ChangelogEntry, includeBreaking bool) []ChangelogSection {
	var breaking []ChangelogEntry
	byType := make(map[string][]ChangelogEntry)
	for _, e := range entries {
		if e.Breaking && includeBreaking {
			breaking = append(breaking, e)
		}
		byType[e.Type] = append(byType[e.Type], e)
	}

	var sections []ChangelogSection

	if len(breaking) > 0 {
		sections = append(sections, ChangelogSection{Title: "⚠ Breaking Changes", Entries: breaking})
	}

	for _, group := range sectionOrder {
		var grouped []ChangelogEntry
		for _, t := range group.Types {
			grouped = append(grouped, byType[t]...)
		}
		if len(grouped) > 0 {
			sections = append(sections, ChangelogSection{Title: group.Title, Entries: grouped})
		}
	}

	return sections
}

// FormatChangelog renders sections as Markdown.
func FormatChangelog(sections []ChangelogSection, version string) string {
	var sb strings.Builder
	if version != "" {
		sb.WriteString(fmt.Sprintf("## %s\n\n", version))
	} else {
		sb.WriteString("## [Unreleased]\n\n")
	}

	for _, sec := range sections {
		sb.WriteString(fmt.Sprintf("### %s\n", sec.Title))
		for _, e := range sec.Entries {
			var line string
			if e.Scope != "" {
				line = fmt.Sprintf("- %s(%s): %s", e.Type, e.Scope, e.Description)
			} else {
				line = fmt.Sprintf("- %s: %s", e.Type, e.Description)
			}
			if e.Hash != "" {
				line += fmt.Sprintf(" (%s)", e.Hash[:func(a, b int) int { if a < b { return a }; return b }(7, len(e.Hash))])
			}
			sb.WriteString(line + "\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// GitLog runs git log between from..to and returns parsed entries.
// Pass empty string for from to use the last tag; "HEAD" for to is default.
func GitLog(from, to string, strict bool) ([]ChangelogEntry, []string, error) {
	if to == "" {
		to = "HEAD"
	}

	if from == "" {
		tagOut, err := exec.Command("git", "describe", "--tags", "--abbrev=0").Output()
		if err == nil {
			from = strings.TrimSpace(string(tagOut))
		}
	}

	gitArgs := []string{"log", "--format=%H %s"}
	if from != "" {
		gitArgs = append(gitArgs, fmt.Sprintf("%s..%s", from, to))
	} else {
		gitArgs = append(gitArgs, to)
	}

	out, err := exec.Command("git", gitArgs...).Output()
	if err != nil {
		return nil, nil, fmt.Errorf("git log failed: %w", err)
	}

	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	var entries []ChangelogEntry
	var skipped []string

	for _, line := range lines {
		if line == "" {
			continue
		}
		entry, ok := ParseChangelogEntry(line)
		if !ok {
			if strict {
				return nil, nil, fmt.Errorf("invalid commit: %q", line)
			}
			skipped = append(skipped, line)
			continue
		}
		entries = append(entries, entry)
	}

	return entries, skipped, nil
}
