package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *MessageError
		expected string
	}{
		{
			name: "syntax error with position",
			err: &MessageError{
				Type:    ErrorTypeSyntax,
				Message: "unexpected token",
				Source:  "Hello {$name",
				Start:   6,
				End:     12,
			},
			expected: "syntax error in 'Hello {$name' at 6-12: unexpected token",
		},
		{
			name: "resolution error without position",
			err: &MessageError{
				Type:    ErrorTypeUnresolvedVar,
				Message: "variable not found",
			},
			expected: "unresolved-variable error: variable not found",
		},
		{
			name: "parse error with source",
			err: &MessageError{
				Type:    ErrorTypeParseError,
				Message: "invalid syntax",
				Source:  "{{invalid}}",
				Start:   2,
				End:     9,
			},
			expected: "parse-error error in '{{invalid}}' at 2-9: invalid syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMessageError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &MessageError{
		Type:    ErrorTypeResolution,
		Message: "resolution failed",
		Cause:   cause,
	}

	unwrapped := err.Unwrap()
	assert.Equal(t, cause, unwrapped)
}

func TestNewSyntaxError(t *testing.T) {
	err := NewSyntaxError("missing closing brace", 10, 15)

	require.NotNil(t, err)
	assert.Equal(t, ErrorTypeSyntax, err.Type)
	assert.Equal(t, "missing closing brace", err.Message)
	assert.Equal(t, 10, err.Start)
	assert.Equal(t, 15, err.End)
	assert.Empty(t, err.Source)
	assert.Nil(t, err.Cause)
}

func TestNewParseError(t *testing.T) {
	err := NewParseError("invalid token", "{{test}}", 2, 6)

	require.NotNil(t, err)
	assert.Equal(t, ErrorTypeParseError, err.Type)
	assert.Equal(t, "invalid token", err.Message)
	assert.Equal(t, "{{test}}", err.Source)
	assert.Equal(t, 2, err.Start)
	assert.Equal(t, 6, err.End)
	assert.Nil(t, err.Cause)
}

func TestNewResolutionError(t *testing.T) {
	err := NewResolutionError(ErrorTypeUnresolvedVar, "variable 'name' not found", "$name")

	require.NotNil(t, err)
	assert.Equal(t, ErrorTypeUnresolvedVar, err.Type)
	assert.Equal(t, "variable 'name' not found", err.Message)
	assert.Equal(t, "$name", err.Source)
	assert.Equal(t, 0, err.Start)
	assert.Equal(t, 0, err.End)
	assert.Nil(t, err.Cause)
}

func TestNewSelectionError(t *testing.T) {
	err := NewSelectionError(ErrorTypeBadSelector, "invalid selector")

	require.NotNil(t, err)
	assert.Equal(t, ErrorTypeBadSelector, err.Type)
	assert.Equal(t, "invalid selector", err.Message)
	assert.Empty(t, err.Source)
	assert.Equal(t, 0, err.Start)
	assert.Equal(t, 0, err.End)
	assert.Nil(t, err.Cause)
}

func TestNewFormattingError(t *testing.T) {
	cause := errors.New("formatting failed")
	err := NewFormattingError(ErrorTypeNotFormattable, "cannot format value", cause)

	require.NotNil(t, err)
	assert.Equal(t, ErrorTypeNotFormattable, err.Type)
	assert.Equal(t, "cannot format value", err.Message)
	assert.Empty(t, err.Source)
	assert.Equal(t, 0, err.Start)
	assert.Equal(t, 0, err.End)
	assert.Equal(t, cause, err.Cause)
}

func TestErrorTypes(t *testing.T) {
	// Test that all error type constants are properly defined
	errorTypes := []ErrorType{
		ErrorTypeSyntax,
		ErrorTypeParseError,
		ErrorTypeMissingSyntax,
		ErrorTypeDuplicateDecl,
		ErrorTypeMissingFallback,
		ErrorTypeResolution,
		ErrorTypeUnresolvedVar,
		ErrorTypeBadOperand,
		ErrorTypeBadOption,
		ErrorTypeBadFunctionResult,
		ErrorTypeSelection,
		ErrorTypeBadSelector,
		ErrorTypeNoMatch,
		ErrorTypeNotFormattable,
		ErrorTypeUnknownFunction,
	}

	for _, errType := range errorTypes {
		assert.NotEmpty(t, string(errType), "Error type should not be empty")
	}
}
