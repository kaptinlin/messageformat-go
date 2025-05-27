// Package messagevalue provides message value interfaces and implementations for MessageFormat 2.0
// TypeScript original code: message-value.ts module
package messagevalue

import (
	"github.com/kaptinlin/messageformat-go/pkg/bidi"
)

// MessageValue represents the base interface for all resolved message values
// TypeScript original code:
//
//	export interface MessageValue<T extends string = string, P extends string = string> {
//	  type: T;
//	  source: string;
//	  dir?: 'ltr' | 'rtl' | 'auto';
//	  valueOf(): unknown;
//	  toString(): string;
//	  toParts?(): MessagePart<P>[];
//	  selectKeys?(keys: string[]): string[];
//	}
type MessageValue interface {
	Type() string                    // Type identifier for the value
	Source() string                  // Source text that produced this value
	Dir() bidi.Direction             // Text direction
	Locale() string                  // Locale for formatting
	Options() map[string]interface{} // Formatting options

	// Core formatting methods
	ToString() (string, error)       // Convert to string representation
	ToParts() ([]MessagePart, error) // Convert to formatted parts
	ValueOf() (interface{}, error)   // Extract underlying value

	// Selection method for plural/select functions
	SelectKeys(keys []string) ([]string, error) // Select matching keys
}

// MessagePart represents a formatted part of a message
// TypeScript original code:
//
//	export interface MessagePart<T extends string = string> {
//	  type: T;
//	  value: unknown;
//	  source?: string;
//	  locale?: string;
//	  dir?: 'ltr' | 'rtl' | 'auto';
//	}
type MessagePart interface {
	Type() string        // Part type identifier
	Value() interface{}  // Part value
	Source() string      // Source text (optional)
	Locale() string      // Locale (optional)
	Dir() bidi.Direction // Text direction (optional)
}

// TextPart represents literal text parts
// TypeScript original code: text part implementation
type TextPart struct {
	value  string
	source string
	locale string
	dir    bidi.Direction
}

// NewTextPart creates a new text part
func NewTextPart(value, source, locale string) *TextPart {
	return &TextPart{
		value:  value,
		source: source,
		locale: locale,
		dir:    bidi.DirAuto,
	}
}

func (tp *TextPart) Type() string        { return "text" }
func (tp *TextPart) Value() interface{}  { return tp.value }
func (tp *TextPart) Source() string      { return tp.source }
func (tp *TextPart) Locale() string      { return tp.locale }
func (tp *TextPart) Dir() bidi.Direction { return tp.dir }

// BidiIsolationPart represents bidirectional isolation characters
type BidiIsolationPart struct {
	value string // LRI, RLI, FSI, or PDI
}

// NewBidiIsolationPart creates a new bidi isolation part
func NewBidiIsolationPart(value string) *BidiIsolationPart {
	return &BidiIsolationPart{value: value}
}

func (bip *BidiIsolationPart) Type() string        { return "bidiIsolation" }
func (bip *BidiIsolationPart) Value() interface{}  { return bip.value }
func (bip *BidiIsolationPart) Source() string      { return "" }
func (bip *BidiIsolationPart) Locale() string      { return "" }
func (bip *BidiIsolationPart) Dir() bidi.Direction { return bidi.DirAuto }

// MarkupPart represents markup elements
type MarkupPart struct {
	kind    string // "open", "close", "standalone"
	name    string
	source  string
	options map[string]interface{}
}

// NewMarkupPart creates a new markup part
func NewMarkupPart(kind, name, source string, options map[string]interface{}) *MarkupPart {
	if options == nil {
		options = make(map[string]interface{})
	}
	return &MarkupPart{
		kind:    kind,
		name:    name,
		source:  source,
		options: options,
	}
}

func (mp *MarkupPart) Type() string                    { return "markup" }
func (mp *MarkupPart) Value() interface{}              { return mp.name }
func (mp *MarkupPart) Source() string                  { return mp.source }
func (mp *MarkupPart) Locale() string                  { return "" }
func (mp *MarkupPart) Dir() bidi.Direction             { return bidi.DirAuto }
func (mp *MarkupPart) Kind() string                    { return mp.kind }
func (mp *MarkupPart) Name() string                    { return mp.name }
func (mp *MarkupPart) Options() map[string]interface{} { return mp.options }

// FallbackPart represents fallback values for errors
type FallbackPart struct {
	source string
	locale string
	dir    bidi.Direction
}

// NewFallbackPart creates a new fallback part
func NewFallbackPart(source, locale string) *FallbackPart {
	return &FallbackPart{
		source: source,
		locale: locale,
		dir:    bidi.DirAuto,
	}
}

func (fp *FallbackPart) Type() string        { return "fallback" }
func (fp *FallbackPart) Value() interface{}  { return "{" + fp.source + "}" }
func (fp *FallbackPart) Source() string      { return fp.source }
func (fp *FallbackPart) Locale() string      { return fp.locale }
func (fp *FallbackPart) Dir() bidi.Direction { return fp.dir }
