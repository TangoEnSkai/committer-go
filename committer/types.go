package committer

import (
	"fmt"
	"strings"
)

type (
	Message    string
	commitType uint8
)

// const that we are using
// from unknown to the end, it defines the type of commits
// if you want to add another type, please add in alphabetical order for better readability
const (
	breakingChange commitType = iota // breaking change for our service
	build                            // change that build system or external dependencies
	chore                            // something else
	ci                               // change to our CI configuration files and scripts
	deps                             // dependency updates
	docs                             // documentation only changes
	feat                             // a new feature
	fix                              // a bug fix
	perf                             // a code change that improves performance
	refactor                         // a code change that neither fixes a bug nor add a feature
	revert                           // reverts a previous commit
	style                            // changes that do not affect the meaning of the code (e.g. whitespace)
	test                             // adding missing tests or correcting existing tests
)

// this String method returns string for each commit type
// for example, if the commit type is breakingChange, it returns `BREAKING CHANGE`
func (t commitType) String() string {
	switch t {
	case breakingChange:
		return "BREAKING CHANGE"
	case build:
		return "build"
	case chore:
		return "chore"
	case ci:
		return "ci"
	case deps:
		return "deps"
	case docs:
		return "docs"
	case feat:
		return "feat"
	case fix:
		return "fix"
	case perf:
		return "perf"
	case refactor:
		return "refactor"
	case revert:
		return "revert"
	case style:
		return "style"
	case test:
		return "test"
	default:
		panic("unexpected error")
	}
}

// CheckCommitType check a type of commit out from the commit message
func CheckCommitType(m Message) (errMsg string, ok bool) {
	var b strings.Builder

	for i, n := 0, len(m); i < n; i++ {
		c := m[i]

		// this is attempt to get the type before actual scope or commit message
		if c == ':' || c == '(' || c == '!' {
			break
		}

		b.WriteByte(c)
	}

	// be aware that this switch status use implicit break unlike C
	// therefore, if the commit message matches the type
	// it will jump out this switch statement
	switch b.String() {
	case breakingChange.String():
	case build.String():
	case chore.String():
	case ci.String():
	case deps.String():
	case docs.String():
	case feat.String():
	case fix.String():
	case perf.String():
	case refactor.String():
	case revert.String():
	case style.String():
	case test.String():
	default:
		errMsg = fmt.Sprintf(
			"%v has invalid commit type\n\tavailable commit types are:\n\t%v",
			m, printTypes(),
		)

		return errMsg, false
	}

	return "", true
}

// getTypes returns available types
// this is only used for letting user know when they failed to commit
// because of type mismatch issue
func printTypes() string {
	const types = `(BREAKING CHANGE | build | chore | ci | deps | docs | feat | fix | perf | refactor | revert | style | test)`

	return types
}
