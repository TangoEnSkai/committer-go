package committer

import (
	"fmt"
	"os"
	"strings"
)

// Read takes the commit message from the commit path
// it returns error when we cannot read the file from the given path
// when the file successfully read, it returns the message in string
func Read(path string) (Message, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("error whilst reading file %s: %w", path, err)
	}

	return Message(string(bytes)), nil
}

// NewFirstLine extracts a commit's header from entire commit messages
// entire commit message means header, body, and footer
// in this program, we want to check only the style of header
func NewFirstLine(m Message) (Message, error) {
	if len(m) == 0 {
		return "", fmt.Errorf("commit message is empty")
	}

	lines := strings.Split(string(m), "\n")
	return Message(lines[0]), nil
}
