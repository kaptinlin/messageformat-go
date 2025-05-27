package tests

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kaptinlin/messageformat-go/pkg/bidi"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// TestFunctions provides test-only functions for the MessageFormat test suite
var TestFunctions = map[string]functions.MessageFunction{
	"test:function": testFunction,
	"test:select":   testSelectFunction,
	"test:format":   testFormatFunction,
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
		return nil, fmt.Errorf("not selectable")
	}
	if tv.badOption {
		return nil, fmt.Errorf("bad option")
	}
	if tv.failsSelect {
		return nil, fmt.Errorf("selection failed")
	}

	if tv.input == 1 {
		if tv.decimalPlaces == 1 {
			for _, key := range keys {
				if key == "1.0" {
					return []string{"1.0"}, nil
				}
			}
		}
		for _, key := range keys {
			if key == "1" {
				return []string{"1"}, nil
			}
		}
	}

	return []string{}, nil // No match
}

// ToString formats the test value as a string
func (tv *testValue) ToString() (string, error) {
	if !tv.canFormat {
		return "", fmt.Errorf("not formattable")
	}
	if tv.failsFormat {
		return "", fmt.Errorf("formatting failed")
	}

	// Format the number directly
	var result strings.Builder

	// Handle negative sign
	if tv.input < 0 {
		result.WriteString("-")
	}

	// Get absolute value and format integer part
	abs := tv.input
	if abs < 0 {
		abs = -abs
	}

	intPart := int64(abs)
	result.WriteString(strconv.FormatInt(intPart, 10))

	// Add decimal part if needed
	if tv.decimalPlaces == 1 {
		result.WriteString(".")
		fracPart := int64((abs - float64(intPart)) * 10)
		result.WriteString(strconv.FormatInt(fracPart, 10))
	}

	return result.String(), nil
}

// ToParts formats the test value as message parts
func (tv *testValue) ToParts() ([]messagevalue.MessagePart, error) {
	if !tv.canFormat {
		return nil, fmt.Errorf("not formattable")
	}
	if tv.failsFormat {
		return nil, fmt.Errorf("formatting failed")
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

	// Handle operand that might be another test value
	if testVal, ok := operand.(*testValue); ok {
		tv.input = testVal.input
		tv.decimalPlaces = testVal.decimalPlaces
		tv.failsFormat = testVal.failsFormat
		tv.failsSelect = testVal.failsSelect
		tv.badOption = testVal.badOption
	} else {
		// Check if operand is a fallback value
		if fallbackVal, ok := operand.(messagevalue.MessageValue); ok && fallbackVal.Type() == "fallback" {
			// For fallback operands, return fallback immediately
			return messagevalue.NewFallbackValue(ctx.Source(), locale)
		}

		// Try to parse numeric input, but don't fail if it's not numeric
		input, err := parseNumericInput(operand)
		if err != nil {
			// For non-numeric operands, use a default value but continue processing
			tv.input = 0
		} else {
			tv.input = input
		}
	}

	// Process options
	if decimalPlaces, exists := options["decimalPlaces"]; exists {
		if dp, err := asPositiveInteger(decimalPlaces); err == nil {
			if dp == 0 || dp == 1 {
				tv.decimalPlaces = dp
			} else {
				// Invalid decimalPlaces value - this should cause a bad-option error
				tv.badOption = true
			}
		} else {
			// Invalid decimalPlaces format - this should cause a bad-option error
			tv.badOption = true
		}
	}

	if fails, exists := options["fails"]; exists {
		if failsStr, err := asString(fails); err == nil {
			switch failsStr {
			case "never":
				// Default behavior
			case "select":
				tv.failsSelect = true
			case "format":
				tv.failsFormat = true
			case "always":
				tv.failsSelect = true
				tv.failsFormat = true
			default:
				return messagevalue.NewFallbackValue(ctx.Source(), locale)
			}
		} else {
			return messagevalue.NewFallbackValue(ctx.Source(), locale)
		}
	}

	return tv
}

// parseNumericInput parses various input types into a numeric value
func parseNumericInput(input interface{}) (float64, error) {
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
	case string:
		// Try to parse as JSON number
		if val, err := strconv.ParseFloat(v, 64); err == nil {
			return val, nil
		}
		return 0, fmt.Errorf("invalid numeric string: %s", v)
	default:
		return 0, fmt.Errorf("input is not numeric: %T", input)
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
	return 0, fmt.Errorf("not a positive integer")
}

func asString(value interface{}) (string, error) {
	if str, ok := value.(string); ok {
		return str, nil
	}
	return "", fmt.Errorf("not a string")
}
