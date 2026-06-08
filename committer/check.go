package committer

import "fmt"

const (
	maxLength = 60 // default maximum length of a commit message
	minLength = 10 // default minimum length of a commit message
)

// CheckLength checks the length of commit messages against the provided config.
// For the commit that has proper length, it will NOOP.
// For longer/shorter commit messages, it returns the error message.
func CheckLength(m Message) (errStr string, ok bool) {
	return CheckLengthWithConfig(m, DefaultConfig())
}

// CheckLengthWithConfig checks the length using the given Config.
func CheckLengthWithConfig(m Message, cfg Config) (errStr string, ok bool) {
	l := len(m)

	if l < cfg.Length.Min {
		return fmt.Sprintf("\tthe commit `%s` is too short: got(%d), need >= (%d)", m, l, cfg.Length.Min), false
	}

	if l > cfg.Length.Max {
		return fmt.Sprintf(
			"\tthe commit `%s` is too long: got(%d), need <= (%d)", m, l, cfg.Length.Max,
		), false
	}

	return "", true
}
