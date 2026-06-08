package committer

import (
	"fmt"
	"strings"
)

// Message represents a raw git commit message.
type Message string

// validTypes is the single source of truth for all allowed commit types.
// Add or remove a type here and every validator/printer updates automatically.
var validTypes = []string{
	"BREAKING CHANGE",
	"build",
	"chore",
	"ci",
	"deps",
	"docs",
	"feat",
	"fix",
	"perf",
	"refactor",
	"revert",
	"style",
	"test",
}

// validTypeSet is a fast lookup map derived from validTypes.
var validTypeSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(validTypes))
	for _, t := range validTypes {
		m[t] = struct{}{}
	}
	return m
}()

// CheckCommitType checks that the type prefix of the commit message is valid.
func CheckCommitType(m Message) (errMsg string, ok bool) {
	var b strings.Builder

	for i, n := 0, len(m); i < n; i++ {
		c := m[i]

		// stop at scope, breaking-change bang, or colon
		if c == ':' || c == '(' || c == '!' {
			break
		}

		b.WriteByte(c)
	}

	if _, found := validTypeSet[b.String()]; !found {
		errMsg = fmt.Sprintf(
			"%v has invalid commit type\n\tavailable commit types are:\n\t%v",
			m, printTypes(),
		)

		return errMsg, false
	}

	return "", true
}

// printTypes returns a formatted list of valid commit types for display in errors.
func printTypes() string {
	return "(" + strings.Join(validTypes, " | ") + ")"
}
