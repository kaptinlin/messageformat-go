package messagevalue

import (
	"fmt"

	"github.com/kaptinlin/messageformat-go/pkg/bidi"
)

// UnknownValue implements MessageValue for unknown function results
// TypeScript original code:
//
//	export interface MessageUnknownValue extends MessageValue<'unknown'> {
//	  readonly type: 'unknown';
//	  readonly source: string;
//	  readonly dir: 'auto';
//	  toParts(): [MessageUnknownPart];
//	  toString(): string;
//	  valueOf(): unknown;
//	}
//
// export const unknown = (
//
//	source: string,
//	input: unknown
//
//	): MessageUnknownValue => ({
//	  type: 'unknown',
//	  source,
//	  dir: 'auto',
//	  toParts: () => [{ type: 'unknown', value: input }],
//	  toString: () => String(input),
//	  valueOf: () => input
//	});
type UnknownValue struct {
	source string
	value  interface{}
	locale string
}

// NewUnknownValue creates a new unknown value
func NewUnknownValue(source string, value interface{}, locale string) *UnknownValue {
	return &UnknownValue{
		source: source,
		value:  value,
		locale: locale,
	}
}

func (uv *UnknownValue) Type() string {
	return "unknown"
}

func (uv *UnknownValue) Source() string {
	return uv.source
}

func (uv *UnknownValue) Dir() bidi.Direction {
	return bidi.DirAuto // TypeScript: readonly dir: 'auto'
}

func (uv *UnknownValue) Locale() string {
	return uv.locale
}

func (uv *UnknownValue) Options() map[string]interface{} {
	return nil
}

func (uv *UnknownValue) ToString() (string, error) {
	return fmt.Sprintf("%v", uv.value), nil // TypeScript: toString: () => String(input)
}

func (uv *UnknownValue) ToParts() ([]MessagePart, error) {
	return []MessagePart{
		&UnknownPart{
			source: uv.source,
			value:  uv.value,
			locale: uv.locale,
		},
	}, nil
}

func (uv *UnknownValue) ValueOf() (interface{}, error) {
	return uv.value, nil // TypeScript: valueOf: () => input
}

func (uv *UnknownValue) SelectKeys(keys []string) ([]string, error) {
	// Unknown values don't participate in selection
	return []string{}, nil
}

// UnknownPart implements MessagePart for unknown function results
// TypeScript original code:
//
//	export interface MessageUnknownPart extends MessageExpressionPart<'unknown'> {
//	  type: 'unknown';
//	  value: unknown;
//	}
type UnknownPart struct {
	source string
	value  interface{}
	locale string
}

// NewUnknownPart creates a new unknown part
func NewUnknownPart(source string, value interface{}, locale string) *UnknownPart {
	return &UnknownPart{
		source: source,
		value:  value,
		locale: locale,
	}
}

func (up *UnknownPart) Type() string {
	return "unknown"
}

func (up *UnknownPart) Value() interface{} {
	return up.value
}

func (up *UnknownPart) Source() string {
	return up.source
}

func (up *UnknownPart) Locale() string {
	return up.locale
}

func (up *UnknownPart) Dir() bidi.Direction {
	return bidi.DirAuto
}
