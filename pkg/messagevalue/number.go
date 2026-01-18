package messagevalue

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Rhymond/go-money"
	"github.com/kaptinlin/messageformat-go/pkg/bidi"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

// Static errors to avoid dynamic error creation
var (
	ErrNumberNotSelectable = errors.New("number value does not support selection")
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
		// For currency, pass the original values and let formatCurrency handle defaults
		origMinFractionDigits := 0
		origMaxFractionDigits := -1
		if val, ok := nv.options["minimumFractionDigits"]; ok {
			if intVal, ok := val.(int); ok {
				origMinFractionDigits = intVal
			}
		}
		if val, ok := nv.options["maximumFractionDigits"]; ok {
			if intVal, ok := val.(int); ok {
				origMaxFractionDigits = intVal
			}
		}
		return nv.formatCurrency(num, tag, origMinFractionDigits, origMaxFractionDigits)
	}

	// Check if this is percentage formatting
	if style, ok := nv.options["style"]; ok && style == "percent" {
		// For percentage, pass the original values and let formatPercent handle defaults
		origMinFractionDigits := 0
		origMaxFractionDigits := -1
		if val, ok := nv.options["minimumFractionDigits"]; ok {
			if intVal, ok := val.(int); ok {
				origMinFractionDigits = intVal
			}
		}
		if val, ok := nv.options["maximumFractionDigits"]; ok {
			if intVal, ok := val.(int); ok {
				origMaxFractionDigits = intVal
			}
		}
		return nv.formatPercent(num, tag, origMinFractionDigits, origMaxFractionDigits)
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
			return strconv.FormatFloat(num, 'f', maxFractionDigits, 64), err
		}
		formatted = result.String()
	} else {
		// Format without grouping - strconv.FormatFloat is appropriate here
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

// formatCurrency formats the number as currency using go-money
func (nv *NumberValue) formatCurrency(num float64, tag language.Tag, minFractionDigits, maxFractionDigits int) (string, error) {
	// Get currency code
	currency, ok := nv.options["currency"]
	if !ok {
		return fmt.Sprintf("%v", num), nil
	}

	currencyCode, ok := currency.(string)
	if !ok {
		return fmt.Sprintf("%v", num), nil
	}

	// Create money object from float using go-money
	moneyObj := money.NewFromFloat(num, strings.ToUpper(currencyCode))
	if moneyObj == nil {
		// Fallback if currency not supported
		return fmt.Sprintf("%v %s", num, currencyCode), nil
	}

	// Handle currency sign for accounting format
	isAccounting := false
	if currencySign, ok := nv.options["currencySign"]; ok && currencySign == "accounting" {
		isAccounting = true
	}

	// Get currency display option
	currencyDisplay := "symbol" // default
	if display, ok := nv.options["currencyDisplay"]; ok {
		if displayStr, ok := display.(string); ok {
			currencyDisplay = displayStr
		}
	}

	// Format using go-money's capabilities
	var formatted string

	switch currencyDisplay {
	case "code":
		// Use currency code instead of symbol
		if isAccounting && moneyObj.IsNegative() {
			formatted = fmt.Sprintf("(%s %s)", moneyObj.Currency().Code, moneyObj.Absolute().Display())
		} else {
			formatted = fmt.Sprintf("%s %s", moneyObj.Currency().Code, moneyObj.Display())
		}
		// Remove the symbol from go-money's display and replace with code
		formatted = nv.replaceCurrencySymbolWithCode(formatted, moneyObj.Currency())

	case "name":
		// Use currency name
		currencyName := nv.getCurrencyName(moneyObj.Currency().Code)
		if isAccounting && moneyObj.IsNegative() {
			amountStr := moneyObj.Absolute().Display()
			amountStr = nv.removeCurrencySymbol(amountStr, moneyObj.Currency())
			formatted = fmt.Sprintf("(%s %s)", amountStr, currencyName)
		} else {
			amountStr := moneyObj.Display()
			amountStr = nv.removeCurrencySymbol(amountStr, moneyObj.Currency())
			formatted = fmt.Sprintf("%s %s", amountStr, currencyName)
		}

	case "narrowSymbol", "symbol":
		fallthrough
	default:
		// Use go-money's default formatting with symbol
		if isAccounting && moneyObj.IsNegative() {
			formatted = fmt.Sprintf("(%s)", moneyObj.Absolute().Display())
		} else {
			formatted = moneyObj.Display()
		}
	}

	return formatted, nil
}

// getCurrencyName returns the currency name for common currencies
func (nv *NumberValue) getCurrencyName(currencyCode string) string {
	switch currencyCode {
	case "USD":
		return "US dollars"
	case "EUR":
		return "euros"
	case "GBP":
		return "British pounds"
	case "JPY":
		return "Japanese yen"
	case "CNY":
		return "Chinese yuan"
	case "CAD":
		return "Canadian dollars"
	case "AUD":
		return "Australian dollars"
	case "CHF":
		return "Swiss francs"
	case "SEK":
		return "Swedish kronor"
	case "NOK":
		return "Norwegian kroner"
	case "DKK":
		return "Danish kroner"
	case "PLN":
		return "Polish zloty"
	case "CZK":
		return "Czech koruna"
	case "HUF":
		return "Hungarian forint"
	case "RUB":
		return "Russian rubles"
	case "INR":
		return "Indian rupees"
	case "KRW":
		return "South Korean won"
	case "SGD":
		return "Singapore dollars"
	case "HKD":
		return "Hong Kong dollars"
	case "NZD":
		return "New Zealand dollars"
	case "MXN":
		return "Mexican pesos"
	case "BRL":
		return "Brazilian reais"
	case "ZAR":
		return "South African rand"
	case "TRY":
		return "Turkish lira"
	case "ILS":
		return "Israeli shekels"
	case "THB":
		return "Thai baht"
	case "MYR":
		return "Malaysian ringgit"
	case "PHP":
		return "Philippine pesos"
	case "IDR":
		return "Indonesian rupiah"
	case "VND":
		return "Vietnamese dong"
	default:
		// Fallback to code if name not available
		return currencyCode
	}
}

// replaceCurrencySymbolWithCode replaces currency symbol with code in formatted string
func (nv *NumberValue) replaceCurrencySymbolWithCode(formatted string, currency *money.Currency) string {
	// Replace the symbol with code
	symbol := currency.Grapheme
	if symbol != "" && strings.Contains(formatted, symbol) {
		return strings.Replace(formatted, symbol, currency.Code, 1)
	}
	return formatted
}

// removeCurrencySymbol removes currency symbol from formatted string
func (nv *NumberValue) removeCurrencySymbol(formatted string, currency *money.Currency) string {
	// Remove the symbol
	symbol := currency.Grapheme
	if symbol != "" {
		formatted = strings.Replace(formatted, symbol, "", 1)
		// Clean up any extra spaces
		formatted = strings.TrimSpace(formatted)
	}
	return formatted
}

// formatPercent formats the number as percentage
// TypeScript original code: similar to formatCurrency but for percentage
func (nv *NumberValue) formatPercent(num float64, tag language.Tag, minFractionDigits, maxFractionDigits int) (string, error) {
	// Convert to percentage (multiply by 100)
	percentNum := num * 100

	// Format the number with appropriate fraction digits for percentage
	// Default behavior: if the percentage has meaningful decimal places, show them
	if maxFractionDigits == -1 {
		// Check if the percentage has meaningful decimal places
		if percentNum == float64(int64(percentNum)) {
			maxFractionDigits = 0 // Integer percentage
		} else {
			maxFractionDigits = 1 // Show one decimal place for non-integer percentages
		}
	}
	// Format the number part
	formatted := strconv.FormatFloat(percentNum, 'f', maxFractionDigits, 64)

	// Adjust fraction digits
	formatted = nv.adjustFractionDigits(formatted, minFractionDigits, maxFractionDigits)

	// Handle signDisplay option
	if val, ok := nv.options["signDisplay"]; ok {
		if signDisplay, ok := val.(string); ok {
			formatted = nv.applySignDisplay(formatted, percentNum, signDisplay)
		}
	}

	// Add percentage symbol
	formatted += "%"

	return formatted, nil
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
	// For currency formatting, we need to be very careful about distinguishing
	// between thousands separators and decimal separators

	// Strategy: Look for the rightmost separator that has 1-2 digits after it
	// (typical for decimal places). Separators with exactly 3 digits after them
	// are likely thousands separators.

	decimalIndex := -1

	// Search from right to left
	for i := len(formatted) - 1; i >= 0; i-- {
		ch := rune(formatted[i])
		if ch == '.' || ch == ',' {
			remaining := formatted[i+1:]

			// Check if this looks like a decimal separator:
			// 1. Should have digits after it
			// 2. Should be 1-2 digits for currency (not exactly 3, which suggests thousands)
			// 3. Should be the rightmost such separator
			if len(remaining) > 0 && isAllDigits(remaining) {
				// Check if this is the rightmost separator
				hasOtherSeparator := false
				for j := i + 1; j < len(formatted); j++ {
					if formatted[j] == '.' || formatted[j] == ',' {
						hasOtherSeparator = true
						break
					}
				}

				if !hasOtherSeparator {
					// For currency formatting, decimal separators typically have 1-2 digits
					// If it has exactly 3 digits, it's likely a thousands separator
					if len(remaining) == 3 {
						// This is likely a thousands separator, skip it
						continue
					}

					// This looks like a decimal separator
					decimalIndex = i
					break
				}
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

	// Check the style to determine how to parse parts
	style := "decimal"
	if styleVal, ok := nv.options["style"]; ok {
		if styleStr, ok := styleVal.(string); ok {
			style = styleStr
		}
	}

	switch style {
	case "currency":
		return nv.parseCurrencyParts(formattedValue)
	case "percent":
		return nv.parsePercentParts(formattedValue)
	case "unit":
		return nv.parseUnitParts(formattedValue)
	default:
		return nv.parseDecimalParts(formattedValue)
	}
}

// parseCurrencyParts parses currency formatted string into detailed parts
func (nv *NumberValue) parseCurrencyParts(formatted string) ([]MessagePart, error) {
	var parts []MessagePart

	// Get currency code for symbol detection
	currencyCode := "USD"
	if currency, ok := nv.options["currency"]; ok {
		if currencyStr, ok := currency.(string); ok {
			currencyCode = currencyStr
		}
	}

	// Get currency symbol
	var currencySymbol string
	if currencyObj := money.GetCurrency(strings.ToUpper(currencyCode)); currencyObj != nil {
		currencySymbol = currencyObj.Grapheme
	} else {
		currencySymbol = currencyCode // fallback
	}

	// Parse the formatted string
	remaining := formatted

	// Handle accounting format (parentheses for negative)
	isNegative := false
	if strings.HasPrefix(remaining, "(") && strings.HasSuffix(remaining, ")") {
		isNegative = true
		remaining = remaining[1 : len(remaining)-1]
		parts = append(parts, &NumberSubPart{
			partType: "literal",
			value:    "(",
			source:   nv.source,
			locale:   nv.locale,
			dir:      nv.dir,
		})
	}

	// Find currency symbol position
	currencyIndex := strings.Index(remaining, currencySymbol)
	if currencyIndex == -1 {
		// Fallback: look for common currency symbols
		for _, symbol := range []string{"$", "€", "£", "¥", "₹"} {
			if idx := strings.Index(remaining, symbol); idx != -1 {
				currencyIndex = idx
				currencySymbol = symbol
				break
			}
		}
	}

	if currencyIndex != -1 {
		// Currency symbol at the beginning
		if currencyIndex == 0 {
			parts = append(parts, &NumberSubPart{
				partType: "currency",
				value:    currencySymbol,
				source:   nv.source,
				locale:   nv.locale,
				dir:      nv.dir,
			})
			// Parse the numeric part after the currency symbol
			numericPart := remaining[len(currencySymbol):]
			numericParts := nv.parseNumericParts(numericPart)
			parts = append(parts, numericParts...)
		} else {
			// Currency symbol at the end
			// Parse the numeric part before the currency symbol
			numericPart := remaining[:currencyIndex]
			numericParts := nv.parseNumericParts(numericPart)
			parts = append(parts, numericParts...)

			parts = append(parts, &NumberSubPart{
				partType: "currency",
				value:    currencySymbol,
				source:   nv.source,
				locale:   nv.locale,
				dir:      nv.dir,
			})
		}
	} else {
		// No currency symbol found, parse as numeric
		numericParts := nv.parseNumericParts(remaining)
		parts = append(parts, numericParts...)
	}

	// Close parenthesis for accounting format
	if isNegative {
		parts = append(parts, &NumberSubPart{
			partType: "literal",
			value:    ")",
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
			parts:  parts,
		},
	}, nil
}

// parsePercentParts parses percentage formatted string into detailed parts
func (nv *NumberValue) parsePercentParts(formatted string) ([]MessagePart, error) {
	var parts []MessagePart

	// Remove the % symbol and parse the numeric part
	remaining := formatted
	if strings.HasSuffix(remaining, "%") {
		remaining = remaining[:len(remaining)-1]

		// Parse numeric parts
		numericParts := nv.parseNumericParts(remaining)
		parts = append(parts, numericParts...)

		// Add percent symbol
		parts = append(parts, &NumberSubPart{
			partType: "percentSign",
			value:    "%",
			source:   nv.source,
			locale:   nv.locale,
			dir:      nv.dir,
		})
	} else {
		// No percent symbol, parse as numeric
		numericParts := nv.parseNumericParts(remaining)
		parts = append(parts, numericParts...)
	}

	return []MessagePart{
		&NumberPart{
			value:  formatted,
			source: nv.source,
			locale: nv.locale,
			dir:    nv.dir,
			parts:  parts,
		},
	}, nil
}

// parseUnitParts parses unit formatted string into detailed parts
func (nv *NumberValue) parseUnitParts(formatted string) ([]MessagePart, error) {
	var parts []MessagePart

	// Find the space that separates number from unit
	spaceIndex := strings.LastIndex(formatted, " ")
	if spaceIndex != -1 {
		// Parse numeric part
		numericPart := formatted[:spaceIndex]
		unitPart := formatted[spaceIndex+1:]

		numericParts := nv.parseNumericParts(numericPart)
		parts = append(parts, numericParts...)

		// Add literal space
		parts = append(parts, &NumberSubPart{
			partType: "literal",
			value:    " ",
			source:   nv.source,
			locale:   nv.locale,
			dir:      nv.dir,
		})

		// Add unit
		parts = append(parts, &NumberSubPart{
			partType: "unit",
			value:    unitPart,
			source:   nv.source,
			locale:   nv.locale,
			dir:      nv.dir,
		})
	} else {
		// No space found, parse as numeric
		numericParts := nv.parseNumericParts(formatted)
		parts = append(parts, numericParts...)
	}

	return []MessagePart{
		&NumberPart{
			value:  formatted,
			source: nv.source,
			locale: nv.locale,
			dir:    nv.dir,
			parts:  parts,
		},
	}, nil
}

// parseDecimalParts parses decimal formatted string into detailed parts
func (nv *NumberValue) parseDecimalParts(formatted string) ([]MessagePart, error) {
	var parts []MessagePart

	numericParts := nv.parseNumericParts(formatted)
	parts = append(parts, numericParts...)

	return []MessagePart{
		&NumberPart{
			value:  formatted,
			source: nv.source,
			locale: nv.locale,
			dir:    nv.dir,
			parts:  parts,
		},
	}, nil
}

// parseNumericParts parses a numeric string into integer, decimal, fraction parts
func (nv *NumberValue) parseNumericParts(numeric string) []MessagePart {
	var parts []MessagePart

	// Handle sign
	remaining := numeric
	if strings.HasPrefix(remaining, "+") || strings.HasPrefix(remaining, "-") {
		signType := "plusSign"
		if remaining[0] == '-' {
			signType = "minusSign"
		}
		parts = append(parts, &NumberSubPart{
			partType: signType,
			value:    remaining[:1],
			source:   nv.source,
			locale:   nv.locale,
			dir:      nv.dir,
		})
		remaining = remaining[1:]
	}

	// Find decimal separator (rightmost . or , that looks like decimal)
	decimalIndex := -1

	// For currency formatting, we need to be more careful about decimal detection
	// Look for the rightmost separator that has 1-2 digits after it (typical for currency)
	for i := len(remaining) - 1; i >= 0; i-- {
		ch := remaining[i]
		if ch == '.' || ch == ',' {
			afterDecimal := remaining[i+1:]
			// Check if this looks like a decimal separator:
			// 1. Should have digits after it
			// 2. Should be 1-3 digits (typical for decimal places)
			// 3. Should be the rightmost such separator
			if len(afterDecimal) > 0 && len(afterDecimal) <= 3 && isAllDigits(afterDecimal) {
				// Check if there are any other separators after this one
				hasOtherSeparator := false
				for j := i + 1; j < len(remaining); j++ {
					if remaining[j] == '.' || remaining[j] == ',' {
						hasOtherSeparator = true
						break
					}
				}

				if !hasOtherSeparator {
					// This is likely the decimal separator
					decimalIndex = i
					break
				}
			}
		}
	}

	if decimalIndex != -1 {
		// Split into integer and fraction parts
		integerPart := remaining[:decimalIndex]
		decimalSeparator := remaining[decimalIndex : decimalIndex+1]
		fractionPart := remaining[decimalIndex+1:]

		// Parse integer part (may contain group separators)
		if integerPart != "" {
			integerParts := nv.parseIntegerWithGroups(integerPart)
			parts = append(parts, integerParts...)
		}

		// Add decimal separator
		parts = append(parts, &NumberSubPart{
			partType: "decimal",
			value:    decimalSeparator,
			source:   nv.source,
			locale:   nv.locale,
			dir:      nv.dir,
		})

		// Add fraction part
		if fractionPart != "" {
			parts = append(parts, &NumberSubPart{
				partType: "fraction",
				value:    fractionPart,
				source:   nv.source,
				locale:   nv.locale,
				dir:      nv.dir,
			})
		}
	} else {
		// No decimal separator, parse as integer
		integerParts := nv.parseIntegerWithGroups(remaining)
		parts = append(parts, integerParts...)
	}

	return parts
}

// parseIntegerWithGroups parses integer part with potential group separators
func (nv *NumberValue) parseIntegerWithGroups(integer string) []MessagePart {
	var parts []MessagePart

	// The integer string already contains formatted group separators from formatNumber()
	// We return it as a single integer part for ToParts() output
	parts = append(parts, &NumberSubPart{
		partType: "integer",
		value:    integer,
		source:   nv.source,
		locale:   nv.locale,
		dir:      nv.dir,
	})

	return parts
}

func (nv *NumberValue) ValueOf() (interface{}, error) {
	return nv.value, nil
}

// SelectKeys performs selection for the number value
// TypeScript reference: getMessageNumber.selectKey (number.ts:116-134)
func (nv *NumberValue) SelectKeys(keys []string) ([]string, error) {
	if !nv.canSelect {
		return nil, ErrNumberNotSelectable
	}

	// Convert value to float64 for consistent comparison
	var num float64
	switch v := nv.value.(type) {
	case int:
		num = float64(v)
	case int32:
		num = float64(v)
	case int64:
		num = float64(v)
	case float32:
		num = float64(v)
	case float64:
		num = v
	default:
		return []string{}, nil
	}

	// TypeScript: if (options.style === 'percent') { numVal *= 100; }
	// For percent style, multiply by 100 for selection (number.ts:119-122)
	numVal := num
	if style, hasStyle := nv.options["style"]; hasStyle {
		if styleStr, ok := style.(string); ok && styleStr == "percent" {
			numVal = num * 100
		}
	}

	// TC39: "exact numeric match will be preferred over plural category"
	// 1. Check for exact numeric match with =N syntax (e.g., =0, =1, =42)
	for _, key := range keys {
		if strings.HasPrefix(key, "=") {
			if keyNum, err := strconv.ParseFloat(key[1:], 64); err == nil && keyNum == numVal {
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

	// 3. Check if select option is set to 'exact' only (TypeScript: if (options.select === 'exact') return null)
	if selectOpt, hasSelect := nv.options["select"]; hasSelect {
		if selectStr, ok := selectOpt.(string); ok && selectStr == "exact" {
			return []string{}, nil
		}
	}

	// 4. Apply plural rules (TypeScript: new Intl.PluralRules(...).select(Number(numVal)))
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

// getPluralCategory determines the plural category for a number
// TypeScript reference: new Intl.PluralRules(locales, pluralOpt).select(Number(numVal))
// Note: The percent multiplication (*100) is now handled by the caller (SelectKeys)
func getPluralCategory(num float64, options map[string]interface{}, locale string) string {
	// Check select option type (cardinal, ordinal, or default to cardinal)
	selectType := "cardinal"
	if selectOpt, hasSelect := options["select"]; hasSelect {
		if selectStr, ok := selectOpt.(string); ok && (selectStr == "ordinal" || selectStr == "cardinal") {
			selectType = selectStr
		}
	}

	if selectType == "ordinal" {
		// Ordinal rules for English: 1st, 2nd, 3rd, 4th, etc.
		switch int(num) % 100 {
		case 11, 12, 13:
			return "other"
		default:
			switch int(num) % 10 {
			case 1:
				return "one"
			case 2:
				return "two"
			case 3:
				return "few"
			default:
				return "other"
			}
		}
	} else {
		// Cardinal rules for English: simplified implementation
		if num == 1 {
			return "one"
		}
		return "other"
	}
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
