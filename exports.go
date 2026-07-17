// Package messageformat provides the main MessageFormat 2.0 API
package messageformat

import (
	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// Function types - exported for custom function implementations
type (
	MessageFunction        = functions.MessageFunction
	MessageFunctionContext = functions.MessageFunctionContext
)

// DefaultFunctionMap returns a snapshot of built-in functions.
func DefaultFunctionMap() map[string]functions.MessageFunction {
	return functions.DefaultFunctionMap()
}

// DraftFunctionMap returns a snapshot of draft functions (beta).
func DraftFunctionMap() map[string]functions.MessageFunction {
	return functions.DraftFunctionMap()
}

// MessageValue types for parts formatting
type (
	Part         = messagevalue.MessagePart
	LiteralPart  = messagevalue.TextPart
	StringPart   = messagevalue.StringPart
	NumberPart   = messagevalue.NumberPart
	DateTimePart = messagevalue.DateTimePart
	FallbackPart = messagevalue.FallbackPart
	UnknownPart  = messagevalue.UnknownPart
	MarkupPart   = messagevalue.MarkupPart
)
