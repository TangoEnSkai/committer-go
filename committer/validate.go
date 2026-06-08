package committer

// ValidateMessage runs all three checks (length, type, pattern) on the first
// line of the commit message using the provided Config.
// It returns the first error string encountered, or ("", true) on success.
func ValidateMessage(m Message, cfg Config) (errMsg string, ok bool) {
	header, err := NewFirstLine(m)
	if err != nil {
		return err.Error(), false
	}

	if errMsg, ok = CheckLengthWithConfig(header, cfg); !ok {
		return errMsg, false
	}

	if errMsg, ok = CheckCommitTypeWithConfig(header, cfg); !ok {
		return errMsg, false
	}

	if errMsg, ok = PatternMatch(header); !ok {
		return errMsg, false
	}

	return "", true
}
