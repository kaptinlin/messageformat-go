package messagevalue

import (
	"github.com/kaptinlin/messageformat-go/pkg/bidi"
	"golang.org/x/text/unicode/norm"
)

// StringValue implements MessageValue for strings
// TypeScript original code:
//
//	export class StringValue implements MessageValue<'string'> {
//	  readonly type = 'string';
//	  constructor(
//	    public readonly source: string,
//	    public readonly value: string,
//	    public readonly locale?: string,
//	    public readonly dir?: Direction
//	  ) {}
//	  valueOf() { return this.value; }
//	  toString() { return this.value; }
//	  toParts() { return [{ type: 'string', value: this.value, source: this.source }]; }
//	  selectKeys(keys: string[]) { return keys.includes(this.value) ? [this.value] : []; }
//	}
type StringValue struct {
	value  string
	locale string
	dir    bidi.Direction
	source string
}

// NewStringValue creates a new string value
func NewStringValue(value, locale, source string) *StringValue {
	return &StringValue{
		value:  value,
		locale: locale,
		dir:    bidi.DirAuto,
		source: source,
	}
}

// NewStringValueWithDir creates a new string value with explicit direction
func NewStringValueWithDir(value, locale, source string, dir bidi.Direction) *StringValue {
	return &StringValue{
		value:  value,
		locale: locale,
		dir:    dir,
		source: source,
	}
}

func (sv *StringValue) Type() string {
	return "string"
}

func (sv *StringValue) Source() string {
	return sv.source
}

func (sv *StringValue) Dir() bidi.Direction {
	return sv.dir
}

func (sv *StringValue) Locale() string {
	return sv.locale
}

func (sv *StringValue) Options() map[string]interface{} {
	return nil
}

func (sv *StringValue) ToString() (string, error) {
	return sv.value, nil
}

func (sv *StringValue) ToParts() ([]MessagePart, error) {
	return []MessagePart{
		&StringPart{
			value:  sv.value,
			source: sv.source,
			locale: sv.locale,
			dir:    sv.dir,
		},
	}, nil
}

func (sv *StringValue) ValueOf() (interface{}, error) {
	return sv.value, nil
}

func (sv *StringValue) SelectKeys(keys []string) ([]string, error) {
	// Apply Unicode NFC normalization for comparison (matches TypeScript: str.normalize())
	normalizedValue := norm.NFC.String(sv.value)

	for _, key := range keys {
		normalizedKey := norm.NFC.String(key)
		if normalizedKey == normalizedValue {
			return []string{key}, nil
		}
	}
	return []string{}, nil
}

// StringPart implements MessagePart for string parts
type StringPart struct {
	value  string
	source string
	locale string
	dir    bidi.Direction
}

func (sp *StringPart) Type() string {
	return "string"
}

func (sp *StringPart) Value() interface{} {
	return sp.value
}

func (sp *StringPart) Source() string {
	return sp.source
}

func (sp *StringPart) Locale() string {
	return sp.locale
}

func (sp *StringPart) Dir() bidi.Direction {
	return sp.dir
}
