package committer

import (
	"fmt"
	"strings"
)

// Suggestion holds a suggested corrected commit message.
type Suggestion struct {
	Message string
	Reason  string
}

// Suggest analyses the commit message m against cfg and returns zero or more
// corrected message suggestions.
func Suggest(m Message, cfg Config) []Suggestion {
	var suggestions []Suggestion

	header, err := NewFirstLine(m)
	if err != nil {
		// Can't parse – suggest a canonical example.
		suggestions = append(suggestions, Suggestion{
			Message: "feat: describe your change here",
			Reason:  "use the canonical format: <type>: <description>",
		})
		return suggestions
	}

	// Collect which checks failed.
	_, lengthOK := CheckLengthWithConfig(header, cfg)
	_, typeOK := CheckCommitTypeWithConfig(header, cfg)
	_, patternOK := PatternMatch(header)

	if lengthOK && typeOK && patternOK {
		// Nothing to suggest.
		return suggestions
	}

	headerStr := string(header)

	// Extract the raw type token (everything before ':', '(', or '!').
	rawType := extractType(headerStr)

	// Extract description (after the first ': ').
	desc := extractDescription(headerStr)

	// --- Invalid type ---
	if !typeOK {
		allTypes := append(validTypes, cfg.Types.Extra...)
		best := findTypeSuggestion(rawType, allTypes)

		if best == "" {
			// No close match – offer the two most common types.
			for _, t := range []string{"feat", "fix"} {
				suggestions = append(suggestions, Suggestion{
					Message: fmt.Sprintf("%s: %s", t, descOrPlaceholder(desc)),
					Reason:  fmt.Sprintf("replace unknown type %q with %q", rawType, t),
				})
			}
		} else {
			suggestions = append(suggestions, Suggestion{
				Message: fmt.Sprintf("%s: %s", best, descOrPlaceholder(desc)),
				Reason:  fmt.Sprintf("replace %q with the closest valid type %q", rawType, best),
			})
		}
		return suggestions
	}

	// --- Length problems (type was valid) ---
	effectiveType := rawType
	if effectiveType == "" {
		effectiveType = "feat"
	}

	if !lengthOK {
		l := len(headerStr)
		if l < cfg.Length.Min {
			suggestions = append(suggestions, Suggestion{
				Message: fmt.Sprintf("%s: %s (add more context)", effectiveType, descOrPlaceholder(desc)),
				Reason:  fmt.Sprintf("message is too short (%d chars); add more context to reach at least %d chars", l, cfg.Length.Min),
			})
		} else if l > cfg.Length.Max {
			truncated := truncateAt(headerStr, cfg.Length.Max)
			suggestions = append(suggestions, Suggestion{
				Message: truncated,
				Reason:  fmt.Sprintf("message is too long (%d chars); truncated to %d chars", l, cfg.Length.Max),
			})
		}
	}

	// --- Format wrong (type + length were OK) ---
	if !patternOK {
		suggestions = append(suggestions, Suggestion{
			Message: fmt.Sprintf("%s: %s", effectiveType, descOrPlaceholder(desc)),
			Reason:  "message does not match the conventional commit format",
		})
	}

	return suggestions
}

// FormatSuggestions returns the suggestions block as a printable string.
func FormatSuggestions(suggestions []Suggestion) string {
	if len(suggestions) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n\U0001f4a1 Suggestions:\n")
	for _, s := range suggestions {
		b.WriteString(fmt.Sprintf("    %s\n", s.Message))
	}
	return b.String()
}

// extractType returns the type token from the commit header.
func extractType(header string) string {
	for i, c := range header {
		if c == ':' || c == '(' || c == '!' {
			return header[:i]
		}
	}
	return header
}

// extractDescription returns the description part after ": ".
func extractDescription(header string) string {
	idx := strings.Index(header, ": ")
	if idx < 0 {
		return ""
	}
	return strings.TrimSpace(header[idx+2:])
}

func descOrPlaceholder(desc string) string {
	if desc == "" {
		return "describe your change here"
	}
	return desc
}

// truncateAt shortens the string to maxLen chars, appending "..." if needed.
// It tries to preserve the type prefix.
func truncateAt(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
