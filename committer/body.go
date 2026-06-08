package committer

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	footerPattern        = regexp.MustCompile(`^[A-Za-z-]+: .+`)
	breakingChangeFooter = regexp.MustCompile(`^BREAKING CHANGE: .+`)
)

// ErrBodySeparator is returned when line 2 is not blank in a multi-line message.
const ErrBodySeparatorCode ErrorCode = "INVALID_BODY_SEPARATOR"

// ErrBodyLineLength is returned when a body line exceeds the configured max length.
const ErrBodyLineLengthCode ErrorCode = "BODY_LINE_TOO_LONG"

// ValidateFullMessage validates the full commit message including body and footer.
// It runs header validation plus body/footer rules.
// Returns a slice of ValidationError and a bool indicating overall validity.
func ValidateFullMessage(m Message, cfg Config) ([]ValidationError, bool) {
	var errs []ValidationError

	lines := strings.Split(string(m), "\n")
	header := Message(lines[0])

	// Run existing header validations.
	if errMsg, ok := CheckLengthWithConfig(header, cfg); !ok {
		errs = append(errs, ValidationError{Code: ErrInvalidLength, Message: errMsg})
	}
	if errMsg, ok := CheckCommitTypeWithConfig(header, cfg); !ok {
		errs = append(errs, ValidationError{Code: ErrInvalidType, Message: errMsg})
	}
	if errMsg, ok := PatternMatch(header); !ok {
		errs = append(errs, ValidationError{Code: ErrInvalidFormat, Message: errMsg})
	}

	if len(lines) > 1 {
		// Rule 1: line 2 (index 1) must be blank.
		if lines[1] != "" {
			errs = append(errs, ValidationError{
				Code:    ErrBodySeparatorCode,
				Message: "line 2 must be blank to separate header from body",
			})
		}

		// Rule 2: body line length (warn, not hard fail — still a ValidationError but code is distinct).
		if cfg.Body.MaxLineLength > 0 {
			for i, line := range lines[2:] {
				if len(line) > cfg.Body.MaxLineLength {
					errs = append(errs, ValidationError{
						Code: ErrBodyLineLengthCode,
						Message: fmt.Sprintf(
							"body line %d exceeds %d characters (got %d): %q",
							i+3, cfg.Body.MaxLineLength, len(line), line,
						),
					})
				}
			}
		}
	}

	// Rule 4: if header has '!', suggest BREAKING CHANGE footer.
	headerStr := string(header)
	hasBreakingBang := strings.Contains(headerStr, "!")
	if hasBreakingBang {
		hasBreakingFooter := false
		for _, line := range lines[1:] {
			if breakingChangeFooter.MatchString(line) {
				hasBreakingFooter = true
				break
			}
		}
		if !hasBreakingFooter {
			errs = append(errs, ValidationError{
				Code:    ErrInvalidFormat,
				Message: "header contains '!' (breaking change) — consider adding a 'BREAKING CHANGE: <description>' footer",
			})
		}
	}

	return errs, len(errs) == 0
}
