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
// TypeScript original code:
// selectKey: canSelect
//
//	? keys => {
//	    const str = String(value);
//	    if (keys.has(str)) return str;
//	    if (options.select === 'exact') return null;
//	    const pluralOpt = options.select
//	      ? { ...options, select: undefined, type: options.select }
//	      : options;
//	    // Intl.PluralRules needs a number, not bigint
//	    cat ??= new Intl.PluralRules(locales, pluralOpt).select(
//	      Number(value)
//	    );
//	    return keys.has(cat) ? cat : null;
//	  }
//	: undefined,
func (nv *NumberValue) SelectKeys(keys []string) ([]string, error) {
	// Convert value to string for exact matching
	var valueStr string
	switch v := nv.value.(type) {
	case int:
		valueStr = strconv.Itoa(v)
	case int64:
		valueStr = strconv.FormatInt(v, 10)
	case float64:
		valueStr = strconv.FormatFloat(v, 'g', -1, 64)
	case float32:
		valueStr = strconv.FormatFloat(float64(v), 'g', -1, 32)
	default:
		valueStr = fmt.Sprintf("%v", v)
	}

	// First, check for exact match (matches TypeScript: if (keys.has(str)) return str;)
	for _, key := range keys {
		if key == valueStr {
			return []string{key}, nil
		}
	}

	// Check if select option is set to 'exact' only
	if selectOpt, hasSelect := nv.options["select"]; hasSelect {
		if selectStr, ok := selectOpt.(string); ok && selectStr == "exact" {
			return []string{}, nil
		}
	}

	// Convert value to float64 for plural rule comparison
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

	// Apply plural rules based on locale and select option
	// This is a simplified implementation - a full implementation would use CLDR plural rules
	// For English (default), the rules are:
	// - "one" for 1
	// - "other" for everything else

	// Check select option type (cardinal, ordinal, or default to cardinal)
	selectType := "cardinal"
	if selectOpt, hasSelect := nv.options["select"]; hasSelect {
		if selectStr, ok := selectOpt.(string); ok && (selectStr == "ordinal" || selectStr == "cardinal") {
			selectType = selectStr
		}
	}

	var pluralCategory string
	if selectType == "ordinal" {
		// Ordinal rules for English: 1st, 2nd, 3rd, 4th, etc.
		// This is simplified - real implementation would use CLDR
		switch int(num) % 100 {
		case 11, 12, 13:
			pluralCategory = "other"
		default:
			switch int(num) % 10 {
			case 1:
				pluralCategory = "one"
			case 2:
				pluralCategory = "two"
			case 3:
				pluralCategory = "few"
			default:
				pluralCategory = "other"
			}
		}
	} else {
		// Cardinal rules for English: simplified implementation
		if num == 1 {
			pluralCategory = "one"
		} else {
			pluralCategory = "other"
		}
	}

	// Check if the calculated plural category is in the available keys
	for _, key := range keys {
		if key == pluralCategory {
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
