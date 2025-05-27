package messagevalue

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kaptinlin/messageformat-go/pkg/bidi"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
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
	value     interface{} // int64, float64, or other numeric types
	locale    string
	dir       bidi.Direction
	source    string
	options   map[string]interface{}
	canSelect bool // whether this number value supports selection
}

// NewNumberValue creates a new number value
func NewNumberValue(value interface{}, locale, source string, options map[string]interface{}) *NumberValue {
	if options == nil {
		options = make(map[string]interface{})
	}

	return &NumberValue{
		value:     value,
		locale:    locale,
		dir:       bidi.DirAuto,
		source:    source,
		options:   options,
		canSelect: true, // default to supporting selection
	}
}

// NewNumberValueWithDir creates a new number value with explicit direction
func NewNumberValueWithDir(value interface{}, locale, source string, dir bidi.Direction, options map[string]interface{}) *NumberValue {
	if options == nil {
		options = make(map[string]interface{})
	}

	return &NumberValue{
		value:     value,
		locale:    locale,
		dir:       dir,
		source:    source,
		options:   options,
		canSelect: true, // default to supporting selection
	}
}

// NewNumberValueWithSelection creates a new number value with specified selection capability
// TypeScript original code: new NumberValue(source, value, locale, dir, options) with canSelect parameter
func NewNumberValueWithSelection(value interface{}, locale, source string, dir bidi.Direction, options map[string]interface{}, canSelect bool) *NumberValue {
	if options == nil {
		options = make(map[string]interface{})
	}

	return &NumberValue{
		value:     value,
		locale:    locale,
		dir:       dir,
		source:    source,
		options:   options,
		canSelect: canSelect,
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
	// Apply number formatting options
	return nv.formatNumber()
}

// formatNumber formats the number according to the options
func (nv *NumberValue) formatNumber() (string, error) {
	// Convert value to float64 for formatting
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
		return fmt.Sprintf("%v", v), nil
	}

	// Parse locale
	locale := nv.locale
	if locale == "" {
		locale = "en-US"
	}

	// Parse language tag
	tag, err := language.Parse(locale)
	if err != nil {
		// Fallback to English if locale parsing fails
		tag = language.English
	}

	// Create printer for the locale
	p := message.NewPrinter(tag)

	// Get formatting options
	minFractionDigits := 0
	maxFractionDigits := -1 // -1 means use default
	useGrouping := true     // Default to true for thousand separators

	if val, ok := nv.options["minimumFractionDigits"]; ok {
		if intVal, ok := val.(int); ok {
			minFractionDigits = intVal
		}
	}

	if val, ok := nv.options["maximumFractionDigits"]; ok {
		if intVal, ok := val.(int); ok {
			maxFractionDigits = intVal
		}
	}

	if val, ok := nv.options["useGrouping"]; ok {
		if strVal, ok := val.(string); ok {
			useGrouping = strVal != "never" && strVal != "false"
		} else if boolVal, ok := val.(bool); ok {
			useGrouping = boolVal
		}
	}

	// If maxFractionDigits is not set, use a reasonable default
	if maxFractionDigits == -1 {
		if minFractionDigits > 0 {
			maxFractionDigits = minFractionDigits
		} else {
			// For integers, use 0; for floats, use up to 3
			if num == float64(int64(num)) && minFractionDigits == 0 {
				maxFractionDigits = 0
			} else {
				maxFractionDigits = 3
			}
		}
	}

	// Ensure maxFractionDigits is at least minFractionDigits
	if maxFractionDigits < minFractionDigits {
		maxFractionDigits = minFractionDigits
	}

	// Check if this is currency formatting
	if style, ok := nv.options["style"]; ok && style == "currency" {
		return nv.formatCurrency(num, tag, minFractionDigits, maxFractionDigits)
	}

	// Check if this is unit formatting
	if style, ok := nv.options["style"]; ok && style == "unit" {
		return nv.formatUnit(num, tag, minFractionDigits, maxFractionDigits)
	}

	// Create number formatter and format the number
	var formatted string

	if useGrouping {
		// Use decimal formatting with grouping (default)
		var result strings.Builder
		if _, err := p.Fprintf(&result, "%v", number.Decimal(num)); err != nil {
			// Fallback to simple formatting if there's an error
			return strconv.FormatFloat(num, 'f', maxFractionDigits, 64), nil
		}
		formatted = result.String()
	} else {
		// Format without grouping - use simple Go formatting for now
		// TODO: Implement proper no-grouping formatting with golang.org/x/text
		formatted = strconv.FormatFloat(num, 'f', maxFractionDigits, 64)
	}

	// Handle fraction digits manually since golang.org/x/text doesn't fully support all options
	if minFractionDigits > 0 || maxFractionDigits >= 0 {
		formatted = nv.adjustFractionDigits(formatted, minFractionDigits, maxFractionDigits)
	}

	// Handle signDisplay option
	if val, ok := nv.options["signDisplay"]; ok {
		if signDisplay, ok := val.(string); ok {
			formatted = nv.applySignDisplay(formatted, num, signDisplay)
		}
	}

	return formatted, nil
}

// formatCurrency formats the number as currency
func (nv *NumberValue) formatCurrency(num float64, tag language.Tag, minFractionDigits, maxFractionDigits int) (string, error) {
	// Get currency code
	currency, ok := nv.options["currency"]
	if !ok {
		return fmt.Sprintf("%v", num), nil // Fallback if no currency
	}

	currencyCode, ok := currency.(string)
	if !ok {
		return fmt.Sprintf("%v", num), nil // Fallback if currency is not string
	}

	// Format the number with appropriate fraction digits for currency
	// Most currencies use 2 decimal places by default
	if maxFractionDigits == -1 {
		maxFractionDigits = 2
	}
	if minFractionDigits == 0 && maxFractionDigits >= 2 {
		minFractionDigits = 2
	}

	// Format the number part
	formatted := strconv.FormatFloat(num, 'f', maxFractionDigits, 64)

	// Adjust fraction digits
	formatted = nv.adjustFractionDigits(formatted, minFractionDigits, maxFractionDigits)

	// Handle currency sign for accounting format
	if currencySign, ok := nv.options["currencySign"]; ok && currencySign == "accounting" {
		if num < 0 {
			// Remove negative sign and wrap in parentheses
			if strings.HasPrefix(formatted, "-") {
				formatted = "(" + formatted[1:] + ")"
			}
		}
	}

	// Add currency symbol
	symbol := nv.getCurrencySymbol(currencyCode, tag)

	// For most locales, currency symbol comes before the number
	// This is a simplified implementation - a full implementation would use
	// locale-specific currency formatting rules
	if strings.Contains(formatted, "(") {
		// For accounting format with parentheses, put symbol inside
		formatted = strings.Replace(formatted, "(", "("+symbol, 1)
	} else {
		formatted = symbol + formatted
	}

	return formatted, nil
}

// getCurrencySymbol returns the currency symbol for the given currency code
func (nv *NumberValue) getCurrencySymbol(currencyCode string, tag language.Tag) string {
	// Check currencyDisplay option
	if display, ok := nv.options["currencyDisplay"]; ok {
		switch display {
		case "code":
			return currencyCode + " "
		case "name":
			// Return currency name - simplified implementation
			switch currencyCode {
			case "USD":
				return "US dollars "
			case "EUR":
				return "euros "
			default:
				return currencyCode + " "
			}
		case "narrowSymbol":
			// Use narrow symbols - simplified implementation
			switch currencyCode {
			case "USD":
				return "$"
			case "EUR":
				return "€"
			default:
				return currencyCode
			}
		case "symbol":
			fallthrough
		default:
			// Use standard symbols
			switch currencyCode {
			case "USD":
				return "$"
			case "EUR":
				return "€"
			case "GBP":
				return "£"
			case "JPY":
				return "¥"
			default:
				return currencyCode
			}
		}
	}

	// Default to symbol
	switch currencyCode {
	case "USD":
		return "$"
	case "EUR":
		return "€"
	case "GBP":
		return "£"
	case "JPY":
		return "¥"
	default:
		return currencyCode
	}
}

// formatUnit formats the number as unit
func (nv *NumberValue) formatUnit(num float64, tag language.Tag, minFractionDigits, maxFractionDigits int) (string, error) {
	// Get unit identifier
	unit, ok := nv.options["unit"]
	if !ok {
		return fmt.Sprintf("%v", num), nil // Fallback if no unit
	}

	unitCode, ok := unit.(string)
	if !ok {
		return fmt.Sprintf("%v", num), nil // Fallback if unit is not string
	}

	// Format the number with appropriate fraction digits
	if maxFractionDigits == -1 {
		maxFractionDigits = 2 // Default for units
	}

	// Format the number part
	formatted := strconv.FormatFloat(num, 'f', maxFractionDigits, 64)

	// Adjust fraction digits
	formatted = nv.adjustFractionDigits(formatted, minFractionDigits, maxFractionDigits)

	// Handle signDisplay option
	if val, ok := nv.options["signDisplay"]; ok {
		if signDisplay, ok := val.(string); ok {
			formatted = nv.applySignDisplay(formatted, num, signDisplay)
		}
	}

	// Add unit symbol
	symbol := nv.getUnitSymbol(unitCode, tag)

	// For most units, symbol comes after the number with a space
	formatted = formatted + " " + symbol

	return formatted, nil
}

// getUnitSymbol returns the unit symbol for the given unit code
func (nv *NumberValue) getUnitSymbol(unitCode string, tag language.Tag) string {
	// Check unitDisplay option
	if display, ok := nv.options["unitDisplay"]; ok {
		switch display {
		case "long":
			// Return full unit name
			switch unitCode {
			case "meter":
				return "meters"
			case "kilometer":
				return "kilometers"
			case "gram":
				return "grams"
			case "kilogram":
				return "kilograms"
			case "second":
				return "seconds"
			case "minute":
				return "minutes"
			case "hour":
				return "hours"
			default:
				return unitCode
			}
		case "narrow":
			// Use narrow symbols
			switch unitCode {
			case "meter":
				return "m"
			case "kilometer":
				return "km"
			case "gram":
				return "g"
			case "kilogram":
				return "kg"
			case "second":
				return "s"
			case "minute":
				return "min"
			case "hour":
				return "h"
			default:
				return unitCode
			}
		case "short":
			fallthrough
		default:
			// Use short symbols (default)
			switch unitCode {
			case "meter":
				return "m"
			case "kilometer":
				return "km"
			case "gram":
				return "g"
			case "kilogram":
				return "kg"
			case "second":
				return "s"
			case "minute":
				return "min"
			case "hour":
				return "h"
			default:
				return unitCode
			}
		}
	}

	// Default to short symbols
	switch unitCode {
	case "meter":
		return "m"
	case "kilometer":
		return "km"
	case "gram":
		return "g"
	case "kilogram":
		return "kg"
	case "second":
		return "s"
	case "minute":
		return "min"
	case "hour":
		return "h"
	default:
		return unitCode
	}
}

// adjustFractionDigits adjusts the number of fraction digits in a formatted number string
func (nv *NumberValue) adjustFractionDigits(formatted string, minFractionDigits, maxFractionDigits int) string {
	// Find the decimal separator (could be . or , depending on locale)
	decimalIndex := -1
	for i, ch := range formatted {
		if ch == '.' || ch == ',' {
			// Check if this is actually a decimal separator, not a thousands separator
			// Decimal separator should be followed by digits and be near the end
			remaining := formatted[i+1:]
			if len(remaining) <= 6 && isAllDigits(remaining) { // Reasonable assumption for decimal part
				decimalIndex = i
				break
			}
		}
	}

	if decimalIndex == -1 {
		// No decimal point found
		if minFractionDigits > 0 {
			// Add decimal point and required digits
			formatted += "."
			for i := 0; i < minFractionDigits; i++ {
				formatted += "0"
			}
		}
		return formatted
	}

	// Count existing fraction digits
	fractionPart := formatted[decimalIndex+1:]
	existingFractionDigits := len(fractionPart)

	if existingFractionDigits < minFractionDigits {
		// Add more zeros
		for i := existingFractionDigits; i < minFractionDigits; i++ {
			formatted += "0"
		}
	} else if existingFractionDigits > maxFractionDigits && maxFractionDigits >= 0 {
		// Truncate excess digits
		formatted = formatted[:decimalIndex+1+maxFractionDigits]
	}

	return formatted
}

// isAllDigits checks if a string contains only digits
func isAllDigits(s string) bool {
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

// applySignDisplay applies the signDisplay option to the formatted number
// TypeScript original code: signDisplay option handling in Intl.NumberFormat
func (nv *NumberValue) applySignDisplay(formatted string, num float64, signDisplay string) string {
	switch signDisplay {
	case "always":
		// Always display sign, even for positive numbers
		if num > 0 && !strings.HasPrefix(formatted, "+") && !strings.HasPrefix(formatted, "-") {
			return "+" + formatted
		}
		return formatted
	case "exceptZero":
		// Display sign for positive and negative numbers, but not zero
		if num == 0 {
			// Remove any existing sign for zero
			if strings.HasPrefix(formatted, "+") || strings.HasPrefix(formatted, "-") {
				return formatted[1:]
			}
			return formatted
		}
		// For non-zero numbers, apply "always" logic
		if num > 0 && !strings.HasPrefix(formatted, "+") && !strings.HasPrefix(formatted, "-") {
			return "+" + formatted
		}
		return formatted
	case "negative":
		// Only display sign for negative numbers, excluding negative zero
		if num == 0 && strings.HasPrefix(formatted, "-") {
			// Remove sign from negative zero
			return formatted[1:]
		}
		return formatted
	case "never":
		// Never display sign
		if strings.HasPrefix(formatted, "+") || strings.HasPrefix(formatted, "-") {
			return formatted[1:]
		}
		return formatted
	case "auto":
		fallthrough
	default:
		// Default behavior: sign for negative numbers only (including negative zero)
		return formatted
	}
}

func (nv *NumberValue) ToParts() ([]MessagePart, error) {
	// Format the number to get the string representation
	formattedValue, err := nv.formatNumber()
	if err != nil {
		return nil, err
	}

	// Create detailed number parts based on the formatted value
	// For now, we'll create a simple integer part for the whole number
	// TODO: Implement more detailed parsing of formatted number parts
	var parts []MessagePart

	// Determine the type of the number part
	partType := "integer"
	if v, ok := nv.value.(float64); ok && v != float64(int64(v)) {
		partType = "decimal"
	} else if v, ok := nv.value.(float32); ok && v != float32(int32(v)) {
		partType = "decimal"
	}

	// Create an integer/decimal part
	parts = append(parts, &NumberSubPart{
		partType: partType,
		value:    formattedValue,
		source:   nv.source,
		locale:   nv.locale,
		dir:      nv.dir,
	})

	return []MessagePart{
		&NumberPart{
			value:  formattedValue,
			source: nv.source,
			locale: nv.locale,
			dir:    nv.dir,
			parts:  parts,
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
	// Check if this NumberValue supports selection
	if !nv.canSelect {
		return nil, fmt.Errorf("number value does not support selection")
	}

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

// NumberSubPart represents a sub-part of a number (like integer, decimal, etc.)
type NumberSubPart struct {
	partType string
	value    interface{}
	source   string
	locale   string
	dir      bidi.Direction
}

func (nsp *NumberSubPart) Type() string {
	return nsp.partType
}

func (nsp *NumberSubPart) Value() interface{} {
	return nsp.value
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
// TypeScript original code: number part implementation
type NumberPart struct {
	value  interface{}
	source string
	locale string
	dir    bidi.Direction
	parts  []MessagePart
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

func (np *NumberPart) Parts() []MessagePart {
	return np.parts
}
