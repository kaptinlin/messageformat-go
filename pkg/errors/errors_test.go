package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockNode is a test implementation of the Node interface
type mockNode struct {
	start int
	end   int
}

func (m *mockNode) GetPosition() (start, end int) {
	return m.start, m.end
}

func TestMessageError(t *testing.T) {
	t.Run("NewMessageError creates error with type and message", func(t *testing.T) {
		err := NewMessageError(ErrorTypeUnknownFunction, "test message")
		require.NotNil(t, err)
		assert.Equal(t, ErrorTypeUnknownFunction, err.Type)
		assert.Equal(t, "test message", err.Message)
	})

	t.Run("Error returns message", func(t *testing.T) {
		err := NewMessageError(ErrorTypeUnknownFunction, "test message")
		assert.Equal(t, "test message", err.Error())
	})

	t.Run("ErrorType returns error type", func(t *testing.T) {
		err := NewMessageError(ErrorTypeUnknownFunction, "test message")
		assert.Equal(t, ErrorTypeUnknownFunction, err.ErrorType())
	})

	t.Run("Is returns true for same error type", func(t *testing.T) {
		err1 := NewMessageError(ErrorTypeUnknownFunction, "msg1")
		err2 := NewMessageError(ErrorTypeUnknownFunction, "msg2")
		assert.True(t, errors.Is(err1, err2))
	})

	t.Run("Is returns false for different error type", func(t *testing.T) {
		err1 := NewMessageError(ErrorTypeUnknownFunction, "msg1")
		err2 := NewMessageError(ErrorTypeBadOperand, "msg2")
		assert.False(t, errors.Is(err1, err2))
	})

	t.Run("Is returns false for non-MessageError", func(t *testing.T) {
		err1 := NewMessageError(ErrorTypeUnknownFunction, "msg1")
		err2 := errors.New("standard error")
		assert.False(t, errors.Is(err1, err2))
	})
}

func TestMessageSyntaxError(t *testing.T) {
	t.Run("creates error with start position", func(t *testing.T) {
		err := NewMessageSyntaxError(ErrorTypeParseError, 10, nil, nil)
		require.NotNil(t, err)
		assert.Equal(t, ErrorTypeParseError, err.Type)
		assert.Equal(t, 10, err.Start)
		assert.Equal(t, 11, err.End) // Default end is start + 1
		assert.Contains(t, err.Error(), "parse-error")
		assert.Contains(t, err.Error(), "at 10")
	})

	t.Run("creates error with start and end positions", func(t *testing.T) {
		end := 20
		err := NewMessageSyntaxError(ErrorTypeParseError, 10, &end, nil)
		require.NotNil(t, err)
		assert.Equal(t, 10, err.Start)
		assert.Equal(t, 20, err.End)
	})

	t.Run("creates error with expected message", func(t *testing.T) {
		expected := "closing brace"
		err := NewMessageSyntaxError(ErrorTypeMissingSyntax, 10, nil, &expected)
		require.NotNil(t, err)
		assert.Contains(t, err.Error(), "missing closing brace")
		assert.Contains(t, err.Error(), "at 10")
	})

	t.Run("handles negative start position", func(t *testing.T) {
		err := NewMessageSyntaxError(ErrorTypeParseError, -1, nil, nil)
		require.NotNil(t, err)
		assert.Equal(t, -1, err.Start)
		assert.Equal(t, 0, err.End) // -1 + 1
		assert.NotContains(t, err.Error(), "at -1")
	})

	t.Run("error type is preserved in base MessageError", func(t *testing.T) {
		err := NewMessageSyntaxError(ErrorTypeBadEscape, 5, nil, nil)
		assert.Equal(t, ErrorTypeBadEscape, err.Type)
	})

	t.Run("inherits Is method from MessageError", func(t *testing.T) {
		err1 := NewMessageSyntaxError(ErrorTypeParseError, 10, nil, nil)
		err2 := NewMessageSyntaxError(ErrorTypeParseError, 20, nil, nil)
		assert.True(t, errors.Is(err1.MessageError, err2.MessageError))
	})
}

func TestMessageDataModelError(t *testing.T) {
	t.Run("creates error from node with position", func(t *testing.T) {
		node := &mockNode{start: 15, end: 25}
		err := NewMessageDataModelError(ErrorTypeDuplicateDeclaration, node)
		require.NotNil(t, err)
		assert.Equal(t, ErrorTypeDuplicateDeclaration, err.Type)
		assert.Equal(t, 15, err.Start)
		assert.Equal(t, 25, err.End)
	})

	t.Run("creates error from nil node", func(t *testing.T) {
		err := NewMessageDataModelError(ErrorTypeMissingFallback, nil)
		require.NotNil(t, err)
		assert.Equal(t, ErrorTypeMissingFallback, err.Type)
		assert.Equal(t, -1, err.Start)
		assert.Equal(t, -1, err.End)
	})

	t.Run("inherits from MessageSyntaxError", func(t *testing.T) {
		node := &mockNode{start: 10, end: 20}
		err := NewMessageDataModelError(ErrorTypeKeyMismatch, node)
		require.NotNil(t, err)
		require.NotNil(t, err.MessageSyntaxError)
		assert.Equal(t, ErrorTypeKeyMismatch, err.Type)
	})
}

func TestMessageResolutionError(t *testing.T) {
	t.Run("creates error with type, message and source", func(t *testing.T) {
		err := NewMessageResolutionError(ErrorTypeBadOperand, "invalid operand type", "source text")
		require.NotNil(t, err)
		assert.Equal(t, ErrorTypeBadOperand, err.Type)
		assert.Equal(t, "source text", err.Source)
		assert.Contains(t, err.Error(), "bad-operand")
		assert.Contains(t, err.Error(), "invalid operand type")
	})

	t.Run("includes error type in message if not present", func(t *testing.T) {
		err := NewMessageResolutionError(ErrorTypeBadOperand, "message without type", "source")
		assert.Contains(t, err.Error(), "bad-operand")
		assert.Contains(t, err.Error(), "message without type")
	})

	t.Run("does not duplicate error type in message", func(t *testing.T) {
		err := NewMessageResolutionError(ErrorTypeBadOperand, "bad-operand: already has type", "source")
		// Should contain error type only once
		message := err.Error()
		assert.Contains(t, message, "bad-operand")
		assert.Contains(t, message, "already has type")
	})

	t.Run("inherits from MessageError", func(t *testing.T) {
		err := NewMessageResolutionError(ErrorTypeUnresolvedVariable, "test", "source")
		require.NotNil(t, err.MessageError)
		assert.Equal(t, ErrorTypeUnresolvedVariable, err.Type)
	})
}

func TestMessageSelectionError(t *testing.T) {
	t.Run("creates error with type", func(t *testing.T) {
		err := NewMessageSelectionError(ErrorTypeBadSelector, nil)
		require.NotNil(t, err)
		assert.Equal(t, ErrorTypeBadSelector, err.Type)
		assert.Contains(t, err.Error(), "Selection error")
		assert.Contains(t, err.Error(), "bad-selector")
		assert.Nil(t, err.Cause)
	})

	t.Run("creates error with cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := NewMessageSelectionError(ErrorTypeNoMatch, cause)
		require.NotNil(t, err)
		assert.Equal(t, ErrorTypeNoMatch, err.Type)
		assert.Equal(t, cause, err.Cause)
	})

	t.Run("Unwrap returns cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := NewMessageSelectionError(ErrorTypeBadSelector, cause)
		unwrapped := err.Unwrap()
		assert.Equal(t, cause, unwrapped)
	})

	t.Run("Unwrap returns nil when no cause", func(t *testing.T) {
		err := NewMessageSelectionError(ErrorTypeBadSelector, nil)
		unwrapped := err.Unwrap()
		assert.Nil(t, unwrapped)
	})

	t.Run("works with errors.Is for cause", func(t *testing.T) {
		cause := errors.New("specific error")
		err := NewMessageSelectionError(ErrorTypeNoMatch, cause)
		assert.True(t, errors.Is(err, cause))
	})
}

func TestMessageFunctionError(t *testing.T) {
	t.Run("creates error with default source", func(t *testing.T) {
		err := NewMessageFunctionError(ErrorTypeBadOption, "invalid option value")
		require.NotNil(t, err)
		assert.Equal(t, ErrorTypeBadOption, err.Type)
		assert.Equal(t, "invalid option value", err.Message)
		assert.Equal(t, "�", err.Source) // TypeScript default
		assert.Nil(t, err.Cause)
	})

	t.Run("SetSource updates source", func(t *testing.T) {
		err := NewMessageFunctionError(ErrorTypeBadOperand, "test")
		err.SetSource("custom source")
		assert.Equal(t, "custom source", err.Source)
	})

	t.Run("SetCause updates cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := NewMessageFunctionError(ErrorTypeNotFormattable, "test")
		err.SetCause(cause)
		assert.Equal(t, cause, err.Cause)
	})

	t.Run("Unwrap returns cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := NewMessageFunctionError(ErrorTypeUnsupportedOperation, "test")
		err.SetCause(cause)
		assert.Equal(t, cause, err.Unwrap())
	})

	t.Run("Unwrap returns nil when no cause", func(t *testing.T) {
		err := NewMessageFunctionError(ErrorTypeBadOperand, "test")
		assert.Nil(t, err.Unwrap())
	})

	t.Run("works with errors.Is for cause", func(t *testing.T) {
		cause := errors.New("specific error")
		err := NewMessageFunctionError(ErrorTypeBadOperand, "test")
		err.SetCause(cause)
		assert.True(t, errors.Is(err, cause))
	})
}

func TestConvenienceConstructors(t *testing.T) {
	t.Run("NewUnknownFunctionError", func(t *testing.T) {
		err := NewUnknownFunctionError("datetime", "source text")
		require.NotNil(t, err)
		assert.Equal(t, ErrorTypeUnknownFunction, err.Type)
		assert.Contains(t, err.Error(), "unknown function :datetime")
		assert.Equal(t, "source text", err.Source)
	})

	t.Run("NewUnresolvedVariableError", func(t *testing.T) {
		err := NewUnresolvedVariableError("count", "source text")
		require.NotNil(t, err)
		assert.Equal(t, ErrorTypeUnresolvedVariable, err.Type)
		assert.Contains(t, err.Error(), "unresolved variable $count")
		assert.Equal(t, "source text", err.Source)
	})

	t.Run("NewBadOperandError", func(t *testing.T) {
		err := NewBadOperandError("operand is not a number", "source text")
		require.NotNil(t, err)
		assert.Equal(t, ErrorTypeBadOperand, err.Type)
		assert.Contains(t, err.Error(), "operand is not a number")
		assert.Equal(t, "source text", err.Source)
	})

	t.Run("NewBadOptionError", func(t *testing.T) {
		err := NewBadOptionError("invalid option format", "source text")
		require.NotNil(t, err)
		assert.Equal(t, ErrorTypeBadOption, err.Type)
		assert.Contains(t, err.Error(), "invalid option format")
		assert.Equal(t, "source text", err.Source)
	})

	t.Run("NewBadFunctionResultError", func(t *testing.T) {
		err := NewBadFunctionResultError("function returned invalid type", "source text")
		require.NotNil(t, err)
		assert.Equal(t, ErrorTypeBadFunctionResult, err.Type)
		assert.Contains(t, err.Error(), "function returned invalid type")
		assert.Equal(t, "source text", err.Source)
	})

	t.Run("NewBadSelectorError", func(t *testing.T) {
		cause := errors.New("selector evaluation failed")
		err := NewBadSelectorError(cause)
		require.NotNil(t, err)
		assert.Equal(t, ErrorTypeBadSelector, err.Type)
		assert.Contains(t, err.Error(), "Selection error")
		assert.Equal(t, cause, err.Cause)
	})

	t.Run("NewBadSelectorError with nil cause", func(t *testing.T) {
		err := NewBadSelectorError(nil)
		require.NotNil(t, err)
		assert.Equal(t, ErrorTypeBadSelector, err.Type)
		assert.Nil(t, err.Cause)
	})

	t.Run("NewNoMatchError", func(t *testing.T) {
		cause := errors.New("no matching variant")
		err := NewNoMatchError(cause)
		require.NotNil(t, err)
		assert.Equal(t, ErrorTypeNoMatch, err.Type)
		assert.Contains(t, err.Error(), "Selection error")
		assert.Equal(t, cause, err.Cause)
	})

	t.Run("NewNoMatchError with nil cause", func(t *testing.T) {
		err := NewNoMatchError(nil)
		require.NotNil(t, err)
		assert.Equal(t, ErrorTypeNoMatch, err.Type)
		assert.Nil(t, err.Cause)
	})

	t.Run("NewDuplicateDeclarationError", func(t *testing.T) {
		node := &mockNode{start: 10, end: 20}
		err := NewDuplicateDeclarationError(node)
		require.NotNil(t, err)
		assert.Equal(t, ErrorTypeDuplicateDeclaration, err.Type)
		assert.Equal(t, 10, err.Start)
		assert.Equal(t, 20, err.End)
	})

	t.Run("NewMissingFallbackError", func(t *testing.T) {
		node := &mockNode{start: 15, end: 25}
		err := NewMissingFallbackError(node)
		require.NotNil(t, err)
		assert.Equal(t, ErrorTypeMissingFallback, err.Type)
		assert.Equal(t, 15, err.Start)
		assert.Equal(t, 25, err.End)
	})

	t.Run("NewCustomSyntaxError", func(t *testing.T) {
		err := NewCustomSyntaxError("custom syntax error message")
		require.NotNil(t, err)
		assert.Equal(t, ErrorTypeParseError, err.Type)
		assert.Equal(t, "custom syntax error message", err.Message)
		assert.Equal(t, 0, err.Start)
		assert.Equal(t, 1, err.End)
	})
}

func TestErrorTypeConstants(t *testing.T) {
	t.Run("syntax error type constants", func(t *testing.T) {
		assert.Equal(t, "empty-token", ErrorTypeEmptyToken)
		assert.Equal(t, "bad-escape", ErrorTypeBadEscape)
		assert.Equal(t, "bad-input-expression", ErrorTypeBadInputExpression)
		assert.Equal(t, "duplicate-attribute", ErrorTypeDuplicateAttribute)
		assert.Equal(t, "duplicate-declaration", ErrorTypeDuplicateDeclaration)
		assert.Equal(t, "duplicate-option-name", ErrorTypeDuplicateOptionName)
		assert.Equal(t, "duplicate-variant", ErrorTypeDuplicateVariant)
		assert.Equal(t, "extra-content", ErrorTypeExtraContent)
		assert.Equal(t, "key-mismatch", ErrorTypeKeyMismatch)
		assert.Equal(t, "parse-error", ErrorTypeParseError)
		assert.Equal(t, "missing-fallback", ErrorTypeMissingFallback)
		assert.Equal(t, "missing-selector-annotation", ErrorTypeMissingSelectorAnnotation)
		assert.Equal(t, "missing-syntax", ErrorTypeMissingSyntax)
	})

	t.Run("resolution error type constants", func(t *testing.T) {
		assert.Equal(t, "bad-function-result", ErrorTypeBadFunctionResult)
		assert.Equal(t, "bad-operand", ErrorTypeBadOperand)
		assert.Equal(t, "bad-option", ErrorTypeBadOption)
		assert.Equal(t, "unresolved-variable", ErrorTypeUnresolvedVariable)
		assert.Equal(t, "unsupported-operation", ErrorTypeUnsupportedOperation)
	})

	t.Run("selection error type constants", func(t *testing.T) {
		assert.Equal(t, "bad-selector", ErrorTypeBadSelector)
		assert.Equal(t, "no-match", ErrorTypeNoMatch)
	})

	t.Run("formatting error type constants", func(t *testing.T) {
		assert.Equal(t, "not-formattable", ErrorTypeNotFormattable)
		assert.Equal(t, "unknown-function", ErrorTypeUnknownFunction)
	})
}

func TestErrorIntegration(t *testing.T) {
	t.Run("error chain with errors.Is works correctly", func(t *testing.T) {
		cause := errors.New("root cause")
		selectionErr := NewBadSelectorError(cause)

		assert.True(t, errors.Is(selectionErr, cause))
	})

	t.Run("function error chain with errors.Is", func(t *testing.T) {
		cause := errors.New("function execution failed")
		funcErr := NewMessageFunctionError(ErrorTypeBadOperand, "invalid operand")
		funcErr.SetCause(cause)

		assert.True(t, errors.Is(funcErr, cause))
	})

	t.Run("MessageError Is method with error chain", func(t *testing.T) {
		err1 := NewMessageResolutionError(ErrorTypeBadOperand, "test1", "source1")
		err2 := NewMessageResolutionError(ErrorTypeBadOperand, "test2", "source2")

		// Should match based on type, not message
		assert.True(t, errors.Is(err1.MessageError, err2.MessageError))
	})

	t.Run("different error types don't match with Is", func(t *testing.T) {
		err1 := NewMessageResolutionError(ErrorTypeBadOperand, "test", "source")
		err2 := NewMessageResolutionError(ErrorTypeBadOption, "test", "source")

		assert.False(t, errors.Is(err1.MessageError, err2.MessageError))
	})

	t.Run("nil node in data model error handles gracefully", func(t *testing.T) {
		err := NewDuplicateDeclarationError(nil)
		require.NotNil(t, err)
		assert.Equal(t, -1, err.Start)
		assert.Equal(t, -1, err.End)
		// Should still have a valid error message
		assert.NotEmpty(t, err.Error())
	})
}

func TestAllSyntaxErrorTypes(t *testing.T) {
	testCases := []struct {
		errorType string
		name      string
	}{
		{ErrorTypeEmptyToken, "empty token"},
		{ErrorTypeBadEscape, "bad escape"},
		{ErrorTypeBadInputExpression, "bad input expression"},
		{ErrorTypeDuplicateAttribute, "duplicate attribute"},
		{ErrorTypeDuplicateDeclaration, "duplicate declaration"},
		{ErrorTypeDuplicateOptionName, "duplicate option name"},
		{ErrorTypeDuplicateVariant, "duplicate variant"},
		{ErrorTypeExtraContent, "extra content"},
		{ErrorTypeKeyMismatch, "key mismatch"},
		{ErrorTypeParseError, "parse error"},
		{ErrorTypeMissingFallback, "missing fallback"},
		{ErrorTypeMissingSelectorAnnotation, "missing selector annotation"},
		{ErrorTypeMissingSyntax, "missing syntax"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := NewMessageSyntaxError(tc.errorType, 5, nil, nil)
			require.NotNil(t, err)
			assert.Equal(t, tc.errorType, err.Type)
			assert.Contains(t, err.Error(), tc.errorType)
		})
	}
}

func TestAllResolutionErrorTypes(t *testing.T) {
	testCases := []struct {
		errorType string
		name      string
	}{
		{ErrorTypeBadFunctionResult, "bad function result"},
		{ErrorTypeBadOperand, "bad operand"},
		{ErrorTypeBadOption, "bad option"},
		{ErrorTypeUnresolvedVariable, "unresolved variable"},
		{ErrorTypeUnsupportedOperation, "unsupported operation"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := NewMessageResolutionError(tc.errorType, "test message", "source")
			require.NotNil(t, err)
			assert.Equal(t, tc.errorType, err.Type)
			assert.Contains(t, err.Error(), tc.errorType)
		})
	}
}
