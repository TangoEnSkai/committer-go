//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/TangoEnSkai/committer-go/committer"
)

func main() {
	js.Global().Set("validateCommit", js.FuncOf(validateCommit))
	// Keep the program alive.
	select {}
}

// validateCommit is the JS-callable function exposed as window.validateCommit.
// It accepts a single string argument (the commit message) and returns a JSON
// string matching the ValidationResult schema.
func validateCommit(this js.Value, args []js.Value) interface{} {
	if len(args) == 0 {
		return marshalError("no argument provided")
	}
	msg := committer.Message(args[0].String())
	cfg := committer.DefaultConfig()
	result := committer.ValidateMessageStructured(msg, cfg)
	b, err := json.Marshal(result)
	if err != nil {
		return marshalError(err.Error())
	}
	return string(b)
}

func marshalError(msg string) string {
	b, _ := json.Marshal(map[string]interface{}{
		"valid":  false,
		"input":  "",
		"errors": []map[string]string{{"code": "INTERNAL", "message": msg}},
	})
	return string(b)
}
