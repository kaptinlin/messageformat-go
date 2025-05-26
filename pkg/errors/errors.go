// Package errors provides error types for MessageFormat 2.0 implementation
// TypeScript original code: errors.ts module
package errors

import "fmt"

// ErrorType represents the type of MessageFormat error
// TypeScript original code:
// export type ErrorType =
//
//	| 'syntax'
//	| 'parse-error'
//	| 'missing-syntax'
//	| 'duplicate-declaration'
//	| 'missing-fallback'
//	| 'resolution'
//	| 'unresolved-variable'
//	| 'bad-operand'
//	| 'bad-option'
//	| 'bad-function-result'
//	| 'selection'
//	| 'bad-selector'
//	| 'no-match'
//	| 'not-formattable'
//	| 'unknown-function';
type ErrorType string

const (
	// Syntax errors
	// TypeScript original code: syntax error types
	ErrorTypeSyntax          ErrorType = "syntax"
	ErrorTypeParseError      ErrorType = "parse-error"
	ErrorTypeMissingSyntax   ErrorType = "missing-syntax"
	ErrorTypeDuplicateDecl   ErrorType = "duplicate-declaration"
	ErrorTypeMissingFallback ErrorType = "missing-fallback"

	// Resolution errors
	// TypeScript original code: resolution error types
	ErrorTypeResolution        ErrorType = "resolution"
	ErrorTypeUnresolvedVar     ErrorType = "unresolved-variable"
	ErrorTypeBadOperand        ErrorType = "bad-operand"
	ErrorTypeBadOption         ErrorType = "bad-option"
	ErrorTypeBadFunctionResult ErrorType = "bad-function-result"

	// Selection errors
	// TypeScript original code: selection error types
	ErrorTypeSelection   ErrorType = "selection"
	ErrorTypeBadSelector ErrorType = "bad-selector"
	ErrorTypeNoMatch     ErrorType = "no-match"

	// Formatting errors
	// TypeScript original code: formatting error types
	ErrorTypeNotFormattable  ErrorType = "not-formattable"
	ErrorTypeUnknownFunction ErrorType = "unknown-function"
)

// MessageError represents all MessageFormat errors
// TypeScript original code:
//
//	export class MessageError extends Error {
//	  readonly type: ErrorType;
//	  readonly start?: number;
//	  readonly end?: number;
//	  readonly source?: string;
//	  readonly cause?: Error;
//	  constructor(
//	    type: ErrorType,
//	    message: string,
//	    options?: {
//	      start?: number;
//	      end?: number;
//	      source?: string;
//	      cause?: Error;
//	    }
//	  ) {
//	    super(message);
//	    this.type = type;
//	    this.start = options?.start;
//	    this.end = options?.end;
//	    this.source = options?.source;
//	    this.cause = options?.cause;
//	  }
//	}
type MessageError struct {
	Type    ErrorType // Error type classification
	Message string    // Error message description
	Source  string    // Source text where error occurred
	Start   int       // Start position in source (optional)
	End     int       // End position in source (optional)
	Cause   error     // Underlying cause error (optional)
}

// Error implements the error interface
func (e *MessageError) Error() string {
	if e.Source != "" {
		return fmt.Sprintf("%s error in '%s' at %d-%d: %s", e.Type, e.Source, e.Start, e.End, e.Message)
	}
	return fmt.Sprintf("%s error: %s", e.Type, e.Message)
}

// Unwrap returns the underlying cause error for error wrapping
func (e *MessageError) Unwrap() error {
	return e.Cause
}

// NewSyntaxError creates a new syntax error
func NewSyntaxError(message string, start, end int) *MessageError {
	return &MessageError{
		Type:    ErrorTypeSyntax,
		Message: message,
		Start:   start,
		End:     end,
	}
}

// NewParseError creates a new parse error
func NewParseError(message, source string, start, end int) *MessageError {
	return &MessageError{
		Type:    ErrorTypeParseError,
		Message: message,
		Source:  source,
		Start:   start,
		End:     end,
	}
}

// NewResolutionError creates a new resolution error
func NewResolutionError(errType ErrorType, message, source string) *MessageError {
	return &MessageError{
		Type:    errType,
		Message: message,
		Source:  source,
	}
}

// NewSelectionError creates a new selection error
func NewSelectionError(errType ErrorType, message string) *MessageError {
	return &MessageError{
		Type:    errType,
		Message: message,
	}
}

// NewFormattingError creates a new formatting error
func NewFormattingError(errType ErrorType, message string, cause error) *MessageError {
	return &MessageError{
		Type:    errType,
		Message: message,
		Cause:   cause,
	}
}
