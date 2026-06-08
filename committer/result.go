package committer

// ErrorCode is a machine-readable code identifying which validation rule failed.
type ErrorCode string

const (
	// ErrInvalidLength is returned when the commit header is too short or too long.
	ErrInvalidLength ErrorCode = "INVALID_LENGTH"
	// ErrInvalidType is returned when the commit type is not in the allowed list.
	ErrInvalidType ErrorCode = "INVALID_TYPE"
	// ErrInvalidFormat is returned when the commit message does not match the regex pattern.
	ErrInvalidFormat ErrorCode = "INVALID_FORMAT"
	// ErrReadError is returned when the commit message file cannot be read.
	ErrReadError ErrorCode = "READ_ERROR"
)

// ValidationError holds one structured validation failure.
type ValidationError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

// ValidationResult is the structured output of a full validation run.
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Input  string            `json:"input"`
	Errors []ValidationError `json:"errors"`
}

// ValidateMessageStructured runs all three checks and returns a ValidationResult.
// Unlike ValidateMessage it does not stop at the first failure — it collects all errors.
func ValidateMessageStructured(m Message, cfg Config) ValidationResult {
	result := ValidationResult{
		Input:  string(m),
		Errors: []ValidationError{},
	}

	header, err := NewFirstLine(m)
	if err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrReadError,
			Message: err.Error(),
		})
		return result
	}

	if errMsg, ok := CheckLengthWithConfig(header, cfg); !ok {
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrInvalidLength,
			Message: errMsg,
		})
	}

	if errMsg, ok := CheckCommitTypeWithConfig(header, cfg); !ok {
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrInvalidType,
			Message: errMsg,
		})
	}

	if errMsg, ok := PatternMatch(header); !ok {
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrInvalidFormat,
			Message: errMsg,
		})
	}

	result.Valid = len(result.Errors) == 0
	return result
}
