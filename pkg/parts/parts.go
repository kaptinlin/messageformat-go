// Package parts provides compatibility aliases for formatted MessageFormat 2.0 parts.
// Prefer github.com/kaptinlin/messageformat-go/pkg/messagevalue for new code.
package parts

import "github.com/kaptinlin/messageformat-go/pkg/messagevalue"

type MessagePart = messagevalue.MessagePart
type MessageTextPart = messagevalue.TextPart
type MessageBidiIsolationPart = messagevalue.BidiIsolationPart
type MessageMarkupPart = messagevalue.MarkupPart
type MessageFallbackPart = messagevalue.FallbackPart
type MessageStringPart = messagevalue.StringPart
type MessageNumberPart = messagevalue.NumberPart
type MessageDateTimePart = messagevalue.DateTimePart
type MessageUnknownPart = messagevalue.UnknownPart

// NewTextPart creates a new text part.
func NewTextPart(value string) *messagevalue.TextPart {
	return messagevalue.NewTextPart(value, value, "")
}

// NewBidiIsolationPart creates a new bidi isolation part.
func NewBidiIsolationPart(value string) *messagevalue.BidiIsolationPart {
	return messagevalue.NewBidiIsolationPart(value)
}

// NewMarkupPart creates a new markup part.
func NewMarkupPart(kind, name, source string, options map[string]any) *messagevalue.MarkupPart {
	return messagevalue.NewMarkupPart(kind, name, source, "", options)
}

// NewFallbackPart creates a new fallback part.
func NewFallbackPart(source, locale string) *messagevalue.FallbackPart {
	return messagevalue.NewFallbackPart(source, locale)
}
