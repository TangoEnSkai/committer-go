package committer

import (
	"fmt"
	"regexp"
	"strings"
)

// PatternMatch checks whether the given msg follows proper style or out
// the pattern is defined under getPattern and
// built in regexp.MatchString check the format of commit message
// if it fails, it let users know which rules they have to follow
func PatternMatch(m Message) (errMsg string, ok bool) {
	p := getPattern()

	mStr := string(m)
	// style check
	matched, err := regexp.MatchString(p, mStr)
	if err != nil {
		return fmt.Errorf("error whilst checking commit message %v: %w", m, err).Error(), false
	}

	// since this function is final check, print out the commit message check result
	if !matched {
		return fmt.Sprintf("invalid commit message. \n\tmust follow this rule: %v\n\t\t", p), false
	}

	return "", matched
}

// getPattern returns a regular expression of conventional commit message.
// The type alternation is built from validTypes so there is one place to
// add or remove a type.
// The regex pattern follows the Angular / Conventional Commits convention.
func getPattern() string {
	alternation := strings.Join(validTypes, "|")
	return `^(` + alternation + `)(\([a-zA-Z0-9 _\-]+\))?!?: [\w \-.,!':/@()]+$`
}
