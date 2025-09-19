package tests

import (
	"errors"
	"strconv"
	"strings"

	"github.com/kaptinlin/messageformat-go/pkg/bidi"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// Static error variables to avoid dynamic error creation
var (
	ErrInvalidNumeric       = errors.New("invalid numeric input")
	ErrNotPositiveInt       = errors.New("not a positive integer")
	ErrNotString            = errors.New("not a string")
	ErrInvalidDecimalPlaces = errors.New("invalid option decimalPlaces")
	ErrInvalidFailsOption   = errors.New("invalid option fails")
	ErrNotSelectable        = errors.New("not selectable")
	ErrBadOption            = errors.New("bad option")
	ErrSelectionFailed      = errors.New("selection failed")
	ErrNotFormattable       = errors.New("not formattable")
	ErrFormattingFailed     = errors.New("formatting failed")
)

// TestFunctions returns a map of test functions for MessageFormat testing
func TestFunctions() map[string]functions.MessageFunction {
	return map[string]functions.MessageFunction{
		"test":        testFunction,
		"test:select": testSelectFunction,
		"test:format": testFormatFunction,
		"placeholder": placeholderFunction,
	}
}

// testValue represents a test function result with specific capabilities
type testValue struct {
	source        string
	input         float64
	canFormat     bool
	canSelect     bool
	decimalPlaces int
	failsFormat   bool
	failsSelect   bool
	badOption     bool
	locale        string
	dir           bidi.Direction
}

// Type returns the type of this message value
func (tv *testValue) Type() string {
	return "test"
}

// Source returns the source expression that created this value
func (tv *testValue) Source() string {
	return tv.source
}

// Dir returns the text direction
func (tv *testValue) Dir() bidi.Direction {
	return tv.dir
}

// Locale returns the locale
func (tv *testValue) Locale() string {
	return tv.locale
}

// Options returns formatting options
func (tv *testValue) Options() map[string]interface{} {
	return nil
}

// ValueOf returns the underlying value
func (tv *testValue) ValueOf() (interface{}, error) {
	return tv.input, nil
}

// SelectKeys performs selection for the test value
func (tv *testValue) SelectKeys(keys []string) ([]string, error) {
	if !tv.canSelect {
		return nil, ErrNotSelectable
	}
	if tv.badOption {
		// When there's a bad option, return an error that will cause fallback selection
		// This is critical for proper error handling - bad options should fail selection
		return nil, ErrBadOption
	}
	if tv.failsSelect {
		return nil, ErrSelectionFailed
	}

	// Follow TypeScript logic: if value === 1
	if tv.input == 1 {
		// Check for "1.0" first if decimalPlaces === 1
		if tv.decimalPlaces == 1 {
			for _, key := range keys {
				if key == "1.0" {
					return []string{"1.0"}, nil
				}
			}
		}
		// Then check for "1"
		for _, key := range keys {
			if key == "1" {
				return []string{"1"}, nil
			}
		}
	}

	// Return null (empty slice) if no match found, like TypeScript
	return []string{}, nil
}

// ToString formats the test value as a string
func (tv *testValue) ToString() (string, error) {
	if !tv.canFormat {
		return "", ErrNotFormattable
	}
	if tv.failsFormat {
		return "", ErrFormattingFailed
	}

	// If there's a bad option, return special value
	// This happens when formatting is attempted despite a bad option
	if tv.badOption {
		return "bad-option-value", nil
	}

	// Follow TypeScript testParts logic
	var result strings.Builder

	// Handle negative numbers
	if tv.input < 0 {
		result.WriteString("-")
	}

	// Get absolute value and format integer part
	abs := tv.input
	if abs < 0 {
		abs = -abs
	}

	// Integer part (Math.floor equivalent)
	intPart := int64(abs)
	result.WriteString(strconv.FormatInt(intPart, 10))

	// Add decimal part if decimalPlaces === 1
	if tv.decimalPlaces == 1 {
		result.WriteString(".")
		// Fractional part: Math.floor((abs - Math.floor(abs)) * 10)
		fracPart := int64((abs - float64(intPart)) * 10)
		result.WriteString(strconv.FormatInt(fracPart, 10))
	}

	return result.String(), nil
}

// ToParts formats the test value as message parts
func (tv *testValue) ToParts() ([]messagevalue.MessagePart, error) {
	if !tv.canFormat {
		return nil, ErrNotFormattable
	}
	if tv.failsFormat {
		return nil, ErrFormattingFailed
	}

	// Create a text part with the formatted value
	str, err := tv.ToString()
	if err != nil {
		return nil, err
	}

	return []messagevalue.MessagePart{
		messagevalue.NewTextPart(str, tv.source, "und"),
	}, nil
}

// testFunction implements the :test:function behavior
func testFunction(ctx functions.MessageFunctionContext, options map[string]interface{}, operand interface{}) messagevalue.MessageValue {
	return createTestValue(ctx, options, operand, true, true)
}

// testSelectFunction implements the :test:select behavior
func testSelectFunction(ctx functions.MessageFunctionContext, options map[string]interface{}, operand interface{}) messagevalue.MessageValue {
	return createTestValue(ctx, options, operand, false, true)
}

// testFormatFunction implements the :test:format behavior
func testFormatFunction(ctx functions.MessageFunctionContext, options map[string]interface{}, operand interface{}) messagevalue.MessageValue {
	return createTestValue(ctx, options, operand, true, false)
}

// createTestValue creates a test value with the specified capabilities
func createTestValue(ctx functions.MessageFunctionContext, options map[string]interface{}, operand interface{}, canFormat, canSelect bool) messagevalue.MessageValue {
	// Get locale from context
	locale := "en"
	if locales := ctx.Locales(); len(locales) > 0 {
		locale = locales[0]
	}

	tv := &testValue{
		source:        ctx.Source(),
		canFormat:     canFormat,
		canSelect:     canSelect,
		decimalPlaces: 0,
		failsFormat:   false,
		failsSelect:   false,
		badOption:     false,
		locale:        locale,
		dir:           bidi.DirAuto,
	}

	// Handle operand that might be another test value (like TypeScript valueOf)
	// We need to check both direct *testValue and MessageValue interface wrapping
	var inheritedTestValue *testValue

	// First try to get test value from MessageValue interface (most common case)
	if msgVal, ok := operand.(messagevalue.MessageValue); ok {
		// Check if it's a test value wrapped in MessageValue interface
		if msgVal.Type() == "test" {
			if testVal, ok := msgVal.(*testValue); ok {
				inheritedTestValue = testVal
				tv.input = testVal.input
			}
		} else if msgVal.Type() == "fallback" {
			// For fallback operands, return fallback immediately
			return messagevalue.NewFallbackValue(ctx.Source(), locale)
		}
	} else if testVal, ok := operand.(*testValue); ok {
		// Direct test value (less common)
		inheritedTestValue = testVal
		tv.input = testVal.input
	} else {
		// Check if operand is a fallback value
		if fallbackVal, ok := operand.(messagevalue.MessageValue); ok && fallbackVal.Type() == "fallback" {
			// For fallback operands, return fallback immediately
			return messagevalue.NewFallbackValue(ctx.Source(), locale)
		}

		// Try to parse numeric input - be more strict like TypeScript
		input, err := parseNumericInputStrict(operand)
		if err != nil {
			// For non-numeric input, use 0 as default but continue processing
			// This allows the fails option to be processed and take effect
			tv.input = 0
		} else {
			tv.input = input
		}
	}

	// If we have an inherited test value, inherit its properties
	if inheritedTestValue != nil {
		// Always inherit the badOption state - this is critical for error propagation
		tv.badOption = inheritedTestValue.badOption
		tv.failsFormat = inheritedTestValue.failsFormat
		tv.failsSelect = inheritedTestValue.failsSelect

		// Only inherit decimalPlaces if not explicitly set in current options
		if _, hasDecimalPlaces := options["decimalPlaces"]; !hasDecimalPlaces {
			tv.decimalPlaces = inheritedTestValue.decimalPlaces
		}
	}

	// Process decimalPlaces option with strict validation like TypeScript
	if decimalPlaces, exists := options["decimalPlaces"]; exists {
		if dp, err := asPositiveInteger(decimalPlaces); err == nil {
			if dp == 0 || dp == 1 {
				tv.decimalPlaces = dp
				// Important: If we're setting a valid decimalPlaces, it overrides inherited badOption
				// This allows fixing a bad option by providing a valid one
				if inheritedTestValue != nil && inheritedTestValue.badOption {
					// We still have the badOption from inheritance but apply the new valid option
					tv.decimalPlaces = dp
				}
			} else {
				// Invalid decimalPlaces value - TypeScript throws bad-option error
				tv.badOption = true
				// In TypeScript, this would call onError with MessageResolutionError
				ctx.OnError(ErrInvalidDecimalPlaces)
			}
		} else {
			// Invalid decimalPlaces format
			tv.badOption = true
			ctx.OnError(ErrInvalidDecimalPlaces)
		}
	}

	// If we have a badOption state (either inherited or from current options), propagate the error
	if tv.badOption {
		ctx.OnError(ErrBadOption)
	}

	// Process fails option with exact TypeScript logic
	if fails, exists := options["fails"]; exists {
		if failsStr, err := asString(fails); err == nil {
			switch failsStr {
			case "never":
				// Default behavior - no changes
			case "select":
				tv.failsSelect = true
			case "format":
				tv.failsFormat = true
			case "always":
				tv.failsSelect = true
				tv.failsFormat = true
			default:
				// Invalid fails value - TypeScript calls onError
				ctx.OnError(ErrInvalidFailsOption)
			}
		} else {
			// Invalid fails format
			ctx.OnError(ErrInvalidFailsOption)
		}
	}

	return tv
}

// parseNumericInputStrict parses input more strictly like TypeScript version
func parseNumericInputStrict(input interface{}) (float64, error) {
	// Handle valueOf() method like TypeScript
	if obj, ok := input.(interface{ ValueOf() (interface{}, error) }); ok {
		if val, err := obj.ValueOf(); err == nil {
			input = val
		}
	}

	// Try JSON parsing for strings like TypeScript
	if str, ok := input.(string); ok {
		// Try to parse as number directly
		if val, err := strconv.ParseFloat(str, 64); err == nil {
			return val, nil
		}
		// If not a valid number string, return error
		return 0, ErrInvalidNumeric
	}

	// Handle numeric types
	switch v := input.(type) {
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	default:
		// TypeScript throws error for non-numeric input
		return 0, ErrInvalidNumeric
	}
}

// Helper functions (simplified versions of functions package utilities)
func asPositiveInteger(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		if v >= 0 {
			return v, nil
		}
	case float64:
		if v >= 0 && v == float64(int(v)) {
			return int(v), nil
		}
	case string:
		if i, err := strconv.Atoi(v); err == nil && i >= 0 {
			return i, nil
		}
	}
	return 0, ErrNotPositiveInt
}

func asString(value interface{}) (string, error) {
	if str, ok := value.(string); ok {
		return str, nil
	}
	return "", ErrNotString
}

// placeholderFunction implements a simple placeholder function for testing
func placeholderFunction(ctx functions.MessageFunctionContext, options map[string]interface{}, operand interface{}) messagevalue.MessageValue {
	locale := "en"
	if locales := ctx.Locales(); len(locales) > 0 {
		locale = locales[0]
	}

	// Return a simple text value
	return messagevalue.NewStringValue("placeholder", locale, ctx.Source())
}
