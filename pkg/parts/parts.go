// Package parts provides formatted parts types for MessageFormat 2.0
// TypeScript original code: formatted-parts.ts module
package parts

import (
	"github.com/kaptinlin/messageformat-go/pkg/bidi"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// MessagePart represents a part of a formatted message
// TypeScript original code:
// export type MessagePart<P extends string> =
//
//	| MessageBiDiIsolationPart
//	| MessageFallbackPart
//	| MessageMarkupPart
//	| MessageNumberPart
//	| MessageStringPart
//	| MessageTextPart
//	| MessageUnknownPart
//	| MessageExpressionPart<...>
type MessagePart interface {
	Type() string
	Value() any
	Source() string
	Locale() string
	Dir() bidi.Direction
}

// MessageTextPart represents a literal text part
// TypeScript original code:
//
//	export interface MessageTextPart {
//	  type: 'text';
//	  value: string;
//	}
type MessageTextPart interface {
	Type() string
	Value() any
	Source() string
	Locale() string
	Dir() bidi.Direction
}

// MessageBidiIsolationPart represents a bidi isolation character
// TypeScript original code:
//
//	export interface MessageBiDiIsolationPart {
//	  type: 'bidiIsolation';
//	  /** LRI | RLI | FSI | PDI */
//	  value: '\u2066' | '\u2067' | '\u2068' | '\u2069';
//	}
type MessageBidiIsolationPart interface {
	Type() string
	Value() any
	Source() string
	Locale() string
	Dir() bidi.Direction
}

// MessageMarkupPart represents a markup element
// TypeScript original code:
//
//	export interface MessageMarkupPart {
//	  type: 'markup';
//	  kind: 'open' | 'standalone' | 'close';
//	  name: string;
//	  id?: string;
//	  options?: { [key: string]: unknown };
//	}
type MessageMarkupPart interface {
	Type() string
	Value() any
	Source() string
	Locale() string
	Dir() bidi.Direction
}

// MessageFallbackPart represents a fallback value for errors
// TypeScript original code:
//
//	export interface MessageFallbackPart {
//	  type: 'fallback';
//	  source: string;
//	}
type MessageFallbackPart interface {
	Type() string
	Value() any
	Source() string
	Locale() string
	Dir() bidi.Direction
}

// MessageStringPart represents a string part
type MessageStringPart interface {
	Type() string
	Value() any
	Source() string
	Locale() string
	Dir() bidi.Direction
}

// MessageNumberPart represents a number part
type MessageNumberPart interface {
	Type() string
	Value() any
	Source() string
	Locale() string
	Dir() bidi.Direction
}

// NewTextPart creates a new text part
func NewTextPart(value string) messagevalue.MessagePart {
	return messagevalue.NewTextPart(value, value, "")
}

// NewBidiIsolationPart creates a new bidi isolation part
func NewBidiIsolationPart(value string) messagevalue.MessagePart {
	return messagevalue.NewBidiIsolationPart(value)
}

// NewMarkupPart creates a new markup part
func NewMarkupPart(kind, name, source string, options map[string]any) messagevalue.MessagePart {
	return messagevalue.NewMarkupPart(kind, name, source, options)
}

// NewFallbackPart creates a new fallback part
func NewFallbackPart(source, locale string) messagevalue.MessagePart {
	return messagevalue.NewFallbackPart(source, locale)
}
