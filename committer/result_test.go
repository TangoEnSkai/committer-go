package committer_test

import (
	"encoding/json"
	"testing"

	"github.com/TangoEnSkai/committer-go/committer"
)

func TestValidateMessageStructured_Valid(t *testing.T) {
	cfg := committer.DefaultConfig()
	result := committer.ValidateMessageStructured("feat: add awesome feature", cfg)

	if !result.Valid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got: %v", result.Errors)
	}
	if result.Input != "feat: add awesome feature" {
		t.Errorf("unexpected Input: %q", result.Input)
	}
}

func TestValidateMessageStructured_TooShort(t *testing.T) {
	cfg := committer.DefaultConfig()
	result := committer.ValidateMessageStructured("hi", cfg)

	if result.Valid {
		t.Error("expected invalid")
	}
	if len(result.Errors) == 0 {
		t.Fatal("expected at least one error")
	}
	if result.Errors[0].Code != committer.ErrInvalidLength {
		t.Errorf("expected INVALID_LENGTH, got %q", result.Errors[0].Code)
	}
}

func TestValidateMessageStructured_InvalidType(t *testing.T) {
	cfg := committer.DefaultConfig()
	result := committer.ValidateMessageStructured("unknown: some commit message here", cfg)

	if result.Valid {
		t.Error("expected invalid")
	}
	found := false
	for _, e := range result.Errors {
		if e.Code == committer.ErrInvalidType {
			found = true
		}
	}
	if !found {
		t.Errorf("expected INVALID_TYPE error, got: %v", result.Errors)
	}
}

func TestValidateMessageStructured_JSONRoundtrip(t *testing.T) {
	cfg := committer.DefaultConfig()
	result := committer.ValidateMessageStructured("bad", cfg)

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	var decoded committer.ValidationResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	if decoded.Valid != result.Valid {
		t.Errorf("Valid: got %v, want %v", decoded.Valid, result.Valid)
	}
	if len(decoded.Errors) != len(result.Errors) {
		t.Errorf("Errors length: got %d, want %d", len(decoded.Errors), len(result.Errors))
	}
}

func TestValidateMessageStructured_ErrorCodes(t *testing.T) {
	cfg := committer.DefaultConfig()

	tests := []struct {
		name     string
		msg      committer.Message
		wantCode committer.ErrorCode
	}{
		{"too short", "hi", committer.ErrInvalidLength},
		{"invalid type", "unknown: some valid message here", committer.ErrInvalidType},
		{"invalid format", "feat:no space after colon", committer.ErrInvalidFormat},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := committer.ValidateMessageStructured(tc.msg, cfg)
			if result.Valid {
				t.Error("expected invalid")
			}
			found := false
			for _, e := range result.Errors {
				if e.Code == tc.wantCode {
					found = true
				}
			}
			if !found {
				t.Errorf("expected error code %q, got: %v", tc.wantCode, result.Errors)
			}
		})
	}
}
