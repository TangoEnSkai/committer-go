package committer

import (
	"fmt"
	"regexp"
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

// getPattern return a regular expression of conventional commit message
// The regex pattern is highly inspired by convention of Angular and the Conventional Commits
// FYI, the types of commits in regex sorted in alphabetical order for better readability
func getPattern() string {
	const pattern = `^(BREAKING CHANGE|build|chore|ci|deps|docs|feat|fix|perf|refactor|revert|style|test)(\([a-zA-Z0-9 _\-]+\))?!?: [\w \-.,!':/@()]+$`

	return pattern
}
