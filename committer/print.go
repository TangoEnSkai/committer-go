package committer

import "fmt"

// Print prints error message with the commit message
func Print(m Message, errMsg string) {
	fmt.Printf("%s\n\ncommit message:\n\n%s", errMsg, m)
}
