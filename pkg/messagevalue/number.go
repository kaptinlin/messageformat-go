package messagevalue

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
	"github.com/kaptinlin/messageformat-go/internal/intlbridge"
	"github.com/kaptinlin/messageformat-go/pkg/bidi"
)

// Static errors to avoid dynamic error creation
var (
	ErrNumberNotSelectable = errors.New("number value does not support selection")
)

// NumberValue implements MessageValue for numbers.
// Formatting and plural selection are both delegated to go-intl (numberformat
// and pluralrules), matching the TypeScript reference's reliance on
// Intl.NumberFormat and Intl.PluralRules.
type NumberValue struct {
	value      any // int64, float64, or other numeric types
	locale     string
	dir        bidi.Direction
	source     string
	options    map[string]any
	selectable bool
}

// NewNumberValue creates a new number value
func NewNumberValue(value any, locale, source string, options map[string]any) *NumberValue {
	return &NumberValue{
		value:      value,
		locale:     locale,
		dir:        bidi.DirAuto,
		source:     source,
		options:    cloneOptions(options),
		selectable: true,
	}
}

// NewNumberValueWithDir creates a new number value with explicit direction
func NewNumberValueWithDir(value any, locale, source string, dir bidi.Direction, options map[string]any) *NumberValue {
	return &NumberValue{
		value:      value,
		locale:     locale,
		dir:        dir,
		source:     source,
		options:    cloneOptions(options),
		selectable: true,
	}
}

// NewNumberValueWithSelection creates a new number value with specified selection capability
func NewNumberValueWithSelection(value any, locale, source string, dir bidi.Direction, options map[string]any, selectable bool) *NumberValue {
	return &NumberValue{
		value:      value,
		locale:     locale,
		dir:        dir,
		source:     source,
		options:    cloneOptions(options),
		selectable: selectable,
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

func (nv *NumberValue) Options() map[string]any {
	return nv.options
}

func (nv *NumberValue) ToString() (string, error) {
	formatter, value, ok, err := nv.newFormatter()
	if err != nil {
		return fmt.Sprintf("%v", nv.value), err
	}
	if !ok {
		return fmt.Sprintf("%v", nv.value), nil
	}
	return formatter.Format(value), nil
}

// newFormatter resolves the value into go-intl's numeric input and constructs
// a NumberFormat using the bridge-translated options. The bool indicates whether a formatter
// could be built (false for non-numeric values, which the caller falls back to
// fmt.Sprintf for).
//
// If go-intl rejects the options (e.g., currency style with no currency code,
// unit style with no unit identifier), the style-specific fields are dropped
// and the formatter is rebuilt with the remaining options. ECMA-402 throws on
// these inputs, but MF2's spec calls for graceful fallback instead of failing
// the whole message.
func (nv *NumberValue) newFormatter() (*numberformat.NumberFormat, numberformat.Value, bool, error) {
	value, ok := numberFormatValue(nv.value)
	if !ok {
		return nil, numberformat.Value{}, false, nil
	}
	loc := intlbridge.ParseLocale(nv.locale)
	opts := intlbridge.NumberOptions(nv.options)
	f, err := numberformat.New(loc, opts)
	if err == nil {
		return f, value, true, nil
	}
	opts.Style = ""
	opts.Currency = ""
	opts.Unit = ""
	if f2, err2 := numberformat.New(loc, opts); err2 == nil {
		return f2, value, true, nil
	}
	return nil, value, false, err
}

func numberFormatValue(v any) (numberformat.Value, bool) {
	switch x := v.(type) {
	case int:
		return numberformat.Int(int64(x)), true
	case int8:
		return numberformat.Int(int64(x)), true
	case int16:
		return numberformat.Int(int64(x)), true
	case int32:
		return numberformat.Int(int64(x)), true
	case int64:
		return numberformat.Int(x), true
	case uint:
		return numberformat.Uint(uint64(x)), true
	case uint8:
		return numberformat.Uint(uint64(x)), true
	case uint16:
		return numberformat.Uint(uint64(x)), true
	case uint32:
		return numberformat.Uint(uint64(x)), true
	case uint64:
		return numberformat.Uint(x), true
	case float32:
		return numberformat.Float(float64(x)), true
	case float64:
		return numberformat.Float(x), true
	case *big.Int:
		return numberformat.BigInt(x), true
	case big.Int:
		return numberformat.BigInt(&x), true
	case *big.Float:
		return numberformat.BigFloat(x), true
	case big.Float:
		return numberformat.BigFloat(&x), true
	}
	return numberformat.Value{}, false
}

func numberAsFloat(v any) (float64, bool) {
	switch x := v.(type) {
	case int:
		return float64(x), true
	case int8:
		return float64(x), true
	case int16:
		return float64(x), true
	case int32:
		return float64(x), true
	case int64:
		return float64(x), true
	case uint:
		return float64(x), true
	case uint8:
		return float64(x), true
	case uint16:
		return float64(x), true
	case uint32:
		return float64(x), true
	case uint64:
		return float64(x), true
	case float32:
		return float64(x), true
	case float64:
		return x, true
	}
	return 0, false
}

func (nv *NumberValue) ToParts() ([]MessagePart, error) {
	formatter, value, ok, err := nv.newFormatter()
	if err != nil {
		return nil, err
	}
	if !ok {
		return []MessagePart{
			&NumberPart{
				value:  fmt.Sprintf("%v", nv.value),
				source: nv.source,
				locale: nv.locale,
				dir:    nv.dir,
			},
		}, nil
	}
	intlParts := formatter.FormatToParts(value)
	formatted := formatter.Format(value)
	sub := make([]MessagePart, 0, len(intlParts))
	for _, p := range intlParts {
		sub = append(sub, &NumberSubPart{
			partType: string(p.Type),
			value:    p.Value,
			source:   nv.source,
			locale:   nv.locale,
			dir:      nv.dir,
		})
	}
	return []MessagePart{
		&NumberPart{
			value:  formatted,
			source: nv.source,
			locale: nv.locale,
			dir:    nv.dir,
			parts:  sub,
		},
	}, nil
}

func (nv *NumberValue) ValueOf() (any, error) {
	return nv.value, nil
}

func (nv *NumberValue) Number() any {
	return nv.value
}

// SelectKeys performs selection for the number value
// TypeScript reference: getMessageNumber.selectKey (number.ts:116-134)
func (nv *NumberValue) SelectKeys(keys []string) ([]string, error) {
	if !nv.selectable {
		return nil, ErrNumberNotSelectable
	}

	num, ok := numberAsFloat(nv.value)
	if !ok {
		return []string{}, nil
	}

	// TypeScript: if (options.style === 'percent') { numVal *= 100; }
	numVal := num
	if style, hasStyle := nv.options["style"]; hasStyle {
		if styleStr, ok := style.(string); ok && styleStr == "percent" {
			numVal = num * 100
		}
	}

	// TC39: "exact numeric match will be preferred over plural category"
	// 1. Check for exact numeric match with =N syntax (e.g., =0, =1, =42)
	for _, key := range keys {
		if suffix, ok := strings.CutPrefix(key, "="); ok {
			if keyNum, err := strconv.ParseFloat(suffix, 64); err == nil && keyNum == numVal {
				return []string{key}, nil
			}
		}
	}

	// 2. Check for exact string match (TypeScript: if (keys.has(str)) return str)
	valueStr := formatNumberForSelection(numVal)
	for _, key := range keys {
		if key == valueStr {
			return []string{key}, nil
		}
	}

	// 3. Check if select option is set to 'exact' only
	if selectOpt, hasSelect := nv.options["select"]; hasSelect {
		if selectStr, ok := selectOpt.(string); ok && selectStr == "exact" {
			return []string{}, nil
		}
	}

	// 4. Apply plural rules
	pluralCategory := getPluralCategory(numVal, nv.options, nv.locale)
	for _, key := range keys {
		if key == pluralCategory {
			return []string{key}, nil
		}
	}

	return []string{}, nil
}

// formatNumberForSelection formats a number for selection matching
func formatNumberForSelection(num float64) string {
	if num == float64(int64(num)) {
		return strconv.FormatInt(int64(num), 10)
	}
	return strconv.FormatFloat(num, 'g', -1, 64)
}

// getPluralCategory determines the plural category for a number using
// go-intl's pluralrules (CLDR-backed, ECMA-402 compliant).
//
// TypeScript reference: new Intl.PluralRules(locales, pluralOpt).select(Number(numVal))
// Note: The percent multiplication (*100) is handled by the caller (SelectKeys).
func getPluralCategory(num float64, options map[string]any, loc string) string {
	ruleType := pluralrules.Cardinal
	if selectOpt, ok := options["select"].(string); ok && selectOpt == "ordinal" {
		ruleType = pluralrules.Ordinal
	}
	// select=cardinal and select=exact both keep Cardinal; exact selection has
	// already been handled by SelectKeys before plural rules are reached.

	parsed := intlbridge.ParseLocale(loc)

	rules, err := pluralrules.New(parsed, pluralrules.Options{Type: ruleType})
	if err != nil || rules == nil {
		return "other"
	}

	cat, err := rules.Select(pluralrules.Float(num))
	if err != nil {
		return "other"
	}
	return mapPluralForm(cat)
}

func mapPluralForm(c pluralrules.Category) string {
	switch c {
	case pluralrules.Zero:
		return "zero"
	case pluralrules.One:
		return "one"
	case pluralrules.Two:
		return "two"
	case pluralrules.Few:
		return "few"
	case pluralrules.Many:
		return "many"
	default:
		return "other"
	}
}

// NumberSubPart represents a sub-part of a number (like integer, decimal, etc.)
type NumberSubPart struct {
	partType string
	value    any
	source   string
	locale   string
	dir      bidi.Direction
}

func (nsp *NumberSubPart) Type() string {
	return nsp.partType
}

func (nsp *NumberSubPart) Value() any {
	return nsp.value
}

func (nsp *NumberSubPart) Text() string {
	return fmt.Sprintf("%v", nsp.value)
}

func (nsp *NumberSubPart) Source() string {
	return nsp.source
}

func (nsp *NumberSubPart) Locale() string {
	return nsp.locale
}

func (nsp *NumberSubPart) Dir() bidi.Direction {
	return nsp.dir
}

// NumberPart implements MessagePart for number parts
type NumberPart struct {
	value  any
	source string
	locale string
	dir    bidi.Direction
	parts  []MessagePart
}

func (np *NumberPart) Type() string {
	return "number"
}

func (np *NumberPart) Value() any {
	return np.value
}

func (np *NumberPart) Text() string {
	return fmt.Sprintf("%v", np.value)
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

func (np *NumberPart) Parts() []MessagePart {
	return np.parts
}
