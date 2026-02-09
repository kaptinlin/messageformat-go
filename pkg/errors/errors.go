// Package errors provides error types for MessageFormat 2.0 implementation
// Following TypeScript errors.ts module with Go best practices
package errors

import (
	"fmt"
	"strings"
)

// MessageError represents the base error class used by MessageFormat
// TypeScript original code:
//
//	export class MessageError extends Error {
//	  type:
//	    | 'not-formattable'
//	    | 'unknown-function'
//	    | typeof MessageResolutionError.prototype.type
//	    | typeof MessageSelectionError.prototype.type
//	    | typeof MessageSyntaxError.prototype.type;
//	  constructor(type: typeof MessageError.prototype.type, message: string) {
//	    super(message);
//	    this.type = type;
//	  }
//	}
type MessageError struct {
	Type    string // Error type classification
	Message string // Error message description
}

// Error implements the error interface
func (e *MessageError) Error() string {
	return e.Message
}

// ErrorType returns the error type classification.
func (e *MessageError) ErrorType() string {
	return e.Type
}

// Is implements error comparison for errors.Is()
func (e *MessageError) Is(target error) bool {
	if t, ok := target.(*MessageError); ok {
		return e.Type == t.Type
	}
	return false
}

// NewMessageError creates a new base message error
func NewMessageError(errorType, message string) *MessageError {
	return &MessageError{
		Type:    errorType,
		Message: message,
	}
}

// MessageSyntaxError represents errors in the message syntax
// TypeScript original code:
//
//	export class MessageSyntaxError extends MessageError {
//	  declare type:
//	    | 'empty-token'
//	    | 'bad-escape'
//	    | 'bad-input-expression'
//	    | 'duplicate-attribute'
//	    | 'duplicate-declaration'
//	    | 'duplicate-option-name'
//	    | 'duplicate-variant'
//	    | 'extra-content'
//	    | 'key-mismatch'
//	    | 'parse-error'
//	    | 'missing-fallback'
//	    | 'missing-selector-annotation'
//	    | 'missing-syntax';
//	  start: number;
//	  end: number;
//	  constructor(
//	    type: typeof MessageSyntaxError.prototype.type,
//	    start: number,
//	    end?: number,
//	    expected?: string
//	  ) {
//	    let message = expected ? `Missing ${expected}` : type;
//	    if (start >= 0) message += ` at ${start}`;
//	    super(type, message);
//	    this.start = start;
//	    this.end = end ?? start + 1;
//	  }
//	}
type MessageSyntaxError struct {
	*MessageError
	Start int // Start position in source
	End   int // End position in source
}

// NewMessageSyntaxError creates a new syntax error
// TypeScript original code: MessageSyntaxError constructor
func NewMessageSyntaxError(errorType string, start int, end *int, expected *string) *MessageSyntaxError {
	var message string
	if expected != nil {
		message = fmt.Sprintf("missing %s", *expected)
	} else {
		message = errorType
	}

	if start >= 0 {
		message += fmt.Sprintf(" at %d", start)
	}

	endPos := start + 1
	if end != nil {
		endPos = *end
	}

	return &MessageSyntaxError{
		MessageError: NewMessageError(errorType, message),
		Start:        start,
		End:          endPos,
	}
}

// Node represents a minimal interface for data model nodes to avoid import cycles
type Node interface {
	// GetPosition returns the start and end positions if available
	GetPosition() (start, end int)
}

// MessageDataModelError represents errors in the message data model
// TypeScript original code:
//
//	export class MessageDataModelError extends MessageSyntaxError {
//	  declare type:
//	    | 'duplicate-declaration'
//	    | 'duplicate-variant'
//	    | 'key-mismatch'
//	    | 'missing-fallback'
//	    | 'missing-selector-annotation';
//	  constructor(type: typeof MessageDataModelError.prototype.type, node: Node) {
//	    const { start, end } = node[cstKey] ?? { start: -1, end: -1 };
//	    super(type, start, end);
//	  }
//	}
type MessageDataModelError struct {
	*MessageSyntaxError
}

// NewMessageDataModelError creates a new data model error
// TypeScript original code: MessageDataModelError constructor
func NewMessageDataModelError(errorType string, node Node) *MessageDataModelError {
	// Get CST position information from node if available
	start := -1
	end := -1

	if node != nil {
		start, end = node.GetPosition()
	}

	return &MessageDataModelError{
		MessageSyntaxError: NewMessageSyntaxError(errorType, start, &end, nil),
	}
}

// MessageResolutionError represents message runtime resolution errors
// TypeScript original code:
//
//	export class MessageResolutionError extends MessageError {
//	  declare type:
//	    | 'bad-function-result'
//	    | 'bad-operand'
//	    | 'bad-option'
//	    | 'unresolved-variable'
//	    | 'unsupported-operation';
//	  source: string;
//	  constructor(
//	    type: typeof MessageResolutionError.prototype.type,
//	    message: string,
//	    source: string
//	  ) {
//	    super(type, message);
//	    this.source = source;
//	  }
//	}
type MessageResolutionError struct {
	*MessageError
	Source string // Source text where error occurred
}

// NewMessageResolutionError creates a new resolution error
// TypeScript original code: MessageResolutionError constructor
func NewMessageResolutionError(errorType, message, source string) *MessageResolutionError {
	// Include error type in message for compatibility with tests
	if !strings.Contains(message, errorType) {
		message = fmt.Sprintf("%s: %s", errorType, message)
	}

	return &MessageResolutionError{
		MessageError: NewMessageError(errorType, message),
		Source:       source,
	}
}

// MessageSelectionError represents errors in message selection
// TypeScript original code:
//
//	export class MessageSelectionError extends MessageError {
//	  declare type: 'bad-selector' | 'no-match';
//	  cause?: unknown;
//	  constructor(
//	    type: typeof MessageSelectionError.prototype.type,
//	    cause?: unknown
//	  ) {
//	    super(type, `Selection error: ${type}`);
//	    if (cause !== undefined) this.cause = cause;
//	  }
//	}
type MessageSelectionError struct {
	*MessageError
	Cause error // Underlying cause error (optional)
}

// NewMessageSelectionError creates a new selection error
// TypeScript original code: MessageSelectionError constructor
func NewMessageSelectionError(errorType string, cause error) *MessageSelectionError {
	message := fmt.Sprintf("Selection error: %s", errorType)

	return &MessageSelectionError{
		MessageError: NewMessageError(errorType, message),
		Cause:        cause,
	}
}

// Unwrap returns the underlying cause error for error wrapping
func (e *MessageSelectionError) Unwrap() error {
	return e.Cause
}

// MessageFunctionError represents message function errors
// TypeScript original code:
//
//	export class MessageFunctionError extends MessageError {
//	  declare type:
//	    | 'bad-operand'
//	    | 'bad-option'
//	    | 'bad-variant-key'
//	    | 'function-error'
//	    | 'not-formattable'
//	    | 'unsupported-operation';
//	  source: string;
//	  cause?: unknown;
//	  constructor(
//	    type: typeof MessageFunctionError.prototype.type,
//	    message: string
//	  ) {
//	    super(type, message);
//	    this.source = '�';
//	  }
//	}
type MessageFunctionError struct {
	*MessageError
	Source string // Source text where error occurred, defaults to '�'
	Cause  error  // Optional underlying cause error
}

// NewMessageFunctionError creates a new function error
// TypeScript original code: MessageFunctionError constructor
func NewMessageFunctionError(errorType, message string) *MessageFunctionError {
	return &MessageFunctionError{
		MessageError: NewMessageError(errorType, message),
		Source:       "�", // TypeScript default value
	}
}

// SetSource sets the source for a function error
func (e *MessageFunctionError) SetSource(source string) {
	e.Source = source
}

// SetCause sets the cause for a function error
func (e *MessageFunctionError) SetCause(cause error) {
	e.Cause = cause
}

// Unwrap returns the underlying cause error for error wrapping
func (e *MessageFunctionError) Unwrap() error {
	return e.Cause
}

// Error type constants matching TypeScript definitions

// Syntax error types
const (
	ErrorTypeEmptyToken                = "empty-token"
	ErrorTypeBadEscape                 = "bad-escape"
	ErrorTypeBadInputExpression        = "bad-input-expression"
	ErrorTypeDuplicateAttribute        = "duplicate-attribute"
	ErrorTypeDuplicateDeclaration      = "duplicate-declaration"
	ErrorTypeDuplicateOptionName       = "duplicate-option-name"
	ErrorTypeDuplicateVariant          = "duplicate-variant"
	ErrorTypeExtraContent              = "extra-content"
	ErrorTypeKeyMismatch               = "key-mismatch"
	ErrorTypeParseError                = "parse-error"
	ErrorTypeMissingFallback           = "missing-fallback"
	ErrorTypeMissingSelectorAnnotation = "missing-selector-annotation"
	ErrorTypeMissingSyntax             = "missing-syntax"
)

// Resolution error types
const (
	ErrorTypeBadFunctionResult    = "bad-function-result"
	ErrorTypeBadOperand           = "bad-operand"
	ErrorTypeBadOption            = "bad-option"
	ErrorTypeUnresolvedVariable   = "unresolved-variable"
	ErrorTypeUnsupportedOperation = "unsupported-operation"
)

// Selection error types
const (
	ErrorTypeBadSelector = "bad-selector"
	ErrorTypeNoMatch     = "no-match"
)

// Formatting error types
const (
	ErrorTypeNotFormattable  = "not-formattable"
	ErrorTypeUnknownFunction = "unknown-function"
)

// Convenience constructors for common error types

// NewUnknownFunctionError creates an unknown function error
func NewUnknownFunctionError(functionName, source string) *MessageResolutionError {
	return NewMessageResolutionError(
		ErrorTypeUnknownFunction,
		fmt.Sprintf("unknown function :%s", functionName),
		source,
	)
}

// NewUnresolvedVariableError creates an unresolved variable error
func NewUnresolvedVariableError(variableName, source string) *MessageResolutionError {
	return NewMessageResolutionError(
		ErrorTypeUnresolvedVariable,
		fmt.Sprintf("unresolved variable $%s", variableName),
		source,
	)
}

// NewBadOperandError creates a bad operand error
func NewBadOperandError(message, source string) *MessageResolutionError {
	return NewMessageResolutionError(ErrorTypeBadOperand, message, source)
}

// NewBadOptionError creates a bad option error
func NewBadOptionError(message, source string) *MessageResolutionError {
	return NewMessageResolutionError(ErrorTypeBadOption, message, source)
}

// NewBadFunctionResultError creates a bad function result error
func NewBadFunctionResultError(message, source string) *MessageResolutionError {
	return NewMessageResolutionError(ErrorTypeBadFunctionResult, message, source)
}

// NewBadSelectorError creates a bad selector error
func NewBadSelectorError(cause error) *MessageSelectionError {
	return NewMessageSelectionError(ErrorTypeBadSelector, cause)
}

// NewNoMatchError creates a no match error
func NewNoMatchError(cause error) *MessageSelectionError {
	return NewMessageSelectionError(ErrorTypeNoMatch, cause)
}

// NewDuplicateDeclarationError creates a duplicate declaration error
func NewDuplicateDeclarationError(node Node) *MessageDataModelError {
	return NewMessageDataModelError(ErrorTypeDuplicateDeclaration, node)
}

// NewMissingFallbackError creates a missing fallback error
func NewMissingFallbackError(node Node) *MessageDataModelError {
	return NewMessageDataModelError(ErrorTypeMissingFallback, node)
}

// NewCustomSyntaxError creates a syntax error with a custom message
func NewCustomSyntaxError(message string) *MessageSyntaxError {
	return &MessageSyntaxError{
		MessageError: NewMessageError(ErrorTypeParseError, message),
		Start:        0,
		End:          1,
	}
}
