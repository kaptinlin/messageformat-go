package messagevalue

import (
	"fmt"
	"strconv"

	"github.com/kaptinlin/messageformat-go/pkg/bidi"
)

// NumberValue implements MessageValue for numbers
// TypeScript original code:
//
//	export class NumberValue implements MessageValue<'number'> {
//	  readonly type = 'number';
//	  constructor(
//	    public readonly source: string,
//	    public readonly value: number,
//	    public readonly locale?: string,
//	    public readonly dir?: Direction,
//	    public readonly options?: Intl.NumberFormatOptions
//	  ) {}
//	  valueOf() { return this.value; }
//	  toString() { return String(this.value); }
//	  toParts() { return [{ type: 'number', value: this.value, source: this.source }]; }
//	  selectKeys(keys: string[]) { /* plural selection logic */ }
//	}
type NumberValue struct {
	value   interface{} // int64, float64, or other numeric types
	locale  string
	dir     bidi.Direction
	source  string
	options map[string]interface{}
}

// NewNumberValue creates a new number value
func NewNumberValue(value interface{}, locale, source string, options map[string]interface{}) *NumberValue {
	if options == nil {
		options = make(map[string]interface{})
	}

	return &NumberValue{
		value:   value,
		locale:  locale,
		dir:     bidi.DirectionAuto,
		source:  source,
		options: options,
	}
}

// NewNumberValueWithDir creates a new number value with explicit direction
func NewNumberValueWithDir(value interface{}, locale, source string, dir bidi.Direction, options map[string]interface{}) *NumberValue {
	if options == nil {
		options = make(map[string]interface{})
	}

	return &NumberValue{
		value:   value,
		locale:  locale,
		dir:     dir,
		source:  source,
		options: options,
	}
}

func (nv *NumberValue) Type() string {
	return "number"
}

func (nv *NumberValue) Source() string {
	return nv.source
}

func (nv *NumberValue) Dir() bidi.Direction {
	return nv.dir
}

func (nv *NumberValue) Locale() string {
	return nv.locale
}

func (nv *NumberValue) Options() map[string]interface{} {
	return nv.options
}

func (nv *NumberValue) ToString() (string, error) {
	switch v := nv.value.(type) {
	case int:
		return strconv.Itoa(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64), nil
	case float32:
		return strconv.FormatFloat(float64(v), 'g', -1, 32), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

func (nv *NumberValue) ToParts() ([]MessagePart, error) {
	return []MessagePart{
		&NumberPart{
			value:  nv.value,
			source: nv.source,
			locale: nv.locale,
			dir:    nv.dir,
		},
	}, nil
}

func (nv *NumberValue) ValueOf() (interface{}, error) {
	return nv.value, nil
}

// SelectKeys implements plural selection logic
// TypeScript original code: plural selection implementation
func (nv *NumberValue) SelectKeys(keys []string) ([]string, error) {
	// Convert value to float64 for comparison
	var num float64
	switch v := nv.value.(type) {
	case int:
		num = float64(v)
	case int64:
		num = float64(v)
	case float64:
		num = v
	case float32:
		num = float64(v)
	default:
		return []string{}, nil
	}

	// Simplified plural selection logic
	// In a full implementation, this would use CLDR plural rules
	for _, key := range keys {
		switch key {
		case "zero":
			if num == 0 {
				return []string{key}, nil
			}
		case "one":
			if num == 1 {
				return []string{key}, nil
			}
		case "two":
			if num == 2 {
				return []string{key}, nil
			}
		case "few":
			if num >= 3 && num <= 6 {
				return []string{key}, nil
			}
		case "many":
			if num > 6 {
				return []string{key}, nil
			}
		case "other":
			// "other" is the fallback case
			return []string{key}, nil
		}
	}

	return []string{}, nil
}

// NumberPart implements MessagePart for number parts
// TypeScript original code: number part implementation
type NumberPart struct {
	value  interface{}
	source string
	locale string
	dir    bidi.Direction
}

func (np *NumberPart) Type() string {
	return "number"
}

func (np *NumberPart) Value() interface{} {
	return np.value
}

func (np *NumberPart) Source() string {
	return np.source
}

func (np *NumberPart) Locale() string {
	return np.locale
}

func (np *NumberPart) Dir() bidi.Direction {
	return np.dir
}
