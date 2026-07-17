package messagevalue

import (
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strconv"
	"strings"

	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
	"github.com/kaptinlin/messageformat-go/internal/intlbridge"
	"github.com/kaptinlin/messageformat-go/pkg/bidi"
)

// Static errors to avoid dynamic error creation
var (
	ErrInvalidNumber        = errors.New("invalid number value")
	ErrInvalidNumberOptions = errors.New("invalid number format options")
	ErrNumberNotSelectable  = errors.New("number value does not support selection")
)

// NumberValue implements MessageValue for numbers.
// Formatting and plural selection are both delegated to go-intl (numberformat
// and pluralrules), matching the TypeScript reference's reliance on
// Intl.NumberFormat and Intl.PluralRules.
type NumberValue struct {
	value       any // int64, float64, or other numeric types
	locale      string
	dir         bidi.Direction
	source      string
	options     map[string]any
	selectable  bool
	formatter   *numberformat.NumberFormat
	formatValue numberformat.Value
	pluralRules *pluralrules.PluralRules
}

// NewNumberValue creates a validated number value.
// TypeScript original code:
// return getMessageNumber(ctx, value, options, true);
func NewNumberValue(value any, locale, source string, options map[string]any) (*NumberValue, error) {
	return newNumberValue(value, locale, source, bidi.DirAuto, options, true)
}

// NewNumberValueWithDir creates a validated number value with explicit direction.
// TypeScript original code:
// return getMessageNumber({ ...ctx, dir }, value, options, true);
func NewNumberValueWithDir(value any, locale, source string, dir bidi.Direction, options map[string]any) (*NumberValue, error) {
	return newNumberValue(value, locale, source, dir, options, true)
}

// NewNumberValueWithSelection creates a validated number value with specified selection capability.
// TypeScript original code:
// return getMessageNumber(ctx, value, options, selectable);
func NewNumberValueWithSelection(value any, locale, source string, dir bidi.Direction, options map[string]any, selectable bool) (*NumberValue, error) {
	return newNumberValue(value, locale, source, dir, options, selectable)
}

// newNumberValue compiles the one formatting plan shared by all projections.
// TypeScript original code:
// const formatter = new Intl.NumberFormat(locales, options);
func newNumberValue(value any, locale, source string, dir bidi.Direction, options map[string]any, selectable bool) (*NumberValue, error) {
	formatValue, ok := numberFormatValue(value)
	if !ok {
		return nil, fmt.Errorf("%w: %T", ErrInvalidNumber, value)
	}
	formatter, err := numberformat.New(intlbridge.ParseLocale(locale), intlbridge.NumberOptions(options))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidNumberOptions, err)
	}
	resolved := formatter.ResolvedOptions()
	resolvedLocale := resolved.Locale.String()
	selectionRules, err := newPluralRules(resolvedLocale, resolved, options, selectable)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidNumberOptions, err)
	}

	return &NumberValue{
		value:       value,
		locale:      resolvedLocale,
		dir:         dir,
		source:      source,
		options:     cloneOptions(options),
		selectable:  selectable,
		formatter:   formatter,
		formatValue: formatValue,
		pluralRules: selectionRules,
	}, nil
}

// newPluralRules compiles selection with the formatter's resolved locale and digit options.
// TypeScript original code:
// cat ??= new Intl.PluralRules(locales, pluralOpt).select(Number(numVal));
func newPluralRules(
	locale string,
	resolved numberformat.ResolvedOptions,
	options map[string]any,
	selectable bool,
) (*pluralrules.PluralRules, error) {
	if !selectable || options["select"] == "exact" {
		return nil, nil
	}

	ruleType := string(pluralrules.Cardinal)
	if options["select"] == "ordinal" {
		ruleType = string(pluralrules.Ordinal)
	}
	roundingMode := string(resolved.RoundingMode)
	roundingPriority := string(resolved.RoundingPriority)
	trailingZeroDisplay := string(resolved.TrailingZeroDisplay)
	notation := string(resolved.Notation)
	pluralOptions := pluralrules.Options{
		Type:                     &ruleType,
		MinimumIntegerDigits:     intPointer(resolved.MinimumIntegerDigits),
		MinimumFractionDigits:    resolved.MinimumFractionDigits,
		MaximumFractionDigits:    resolved.MaximumFractionDigits,
		MinimumSignificantDigits: resolved.MinimumSignificantDigits,
		MaximumSignificantDigits: resolved.MaximumSignificantDigits,
		RoundingIncrement:        intPointer(resolved.RoundingIncrement),
		RoundingMode:             &roundingMode,
		RoundingPriority:         &roundingPriority,
		TrailingZeroDisplay:      &trailingZeroDisplay,
		Notation:                 &notation,
	}
	if resolved.CompactDisplay != nil {
		compactDisplay := string(*resolved.CompactDisplay)
		pluralOptions.CompactDisplay = &compactDisplay
	}
	return pluralrules.New(intlbridge.ParseLocale(locale), pluralOptions)
}

// intPointer bridges a resolved Go scalar into a typed dependency option.
// TypeScript original code:
// // JavaScript options do not require pointer conversion.
func intPointer(value int) *int {
	return &value
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
	return cloneOptions(nv.options)
}

// CanSelect reports whether this number value supports pattern selection.
// TypeScript original code:
// // MessageNumber values expose selectKey only when selectable.
func (nv *NumberValue) CanSelect() bool {
	return nv.selectable
}

func (nv *NumberValue) ToString() (string, error) {
	return nv.formatter.Format(nv.formatValue), nil
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
		return bigFloatNumberFormatValue(x)
	case big.Float:
		return bigFloatNumberFormatValue(&x)
	}
	return numberformat.Value{}, false
}

func bigFloatNumberFormatValue(v *big.Float) (numberformat.Value, bool) {
	if v == nil {
		return numberformat.Value{}, false
	}
	if f, accuracy := v.Float64(); accuracy == big.Exact {
		return numberformat.Float(f), true
	}
	value, err := numberformat.Decimal(v.Text('g', -1))
	if err != nil {
		return numberformat.Value{}, false
	}
	return value, true
}

func stringPtr(v string) *string {
	return &v
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
	intlParts := nv.formatter.FormatToParts(nv.formatValue)
	formatted := nv.formatter.Format(nv.formatValue)
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

	// 4. Apply the plural plan compiled with the formatter's resolved options.
	if nv.pluralRules == nil {
		return []string{}, nil
	}
	category := nv.pluralRules.Select(pluralrules.Float(numVal))
	pluralCategory := mapPluralForm(category)
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
	return slices.Clone(np.parts)
}
