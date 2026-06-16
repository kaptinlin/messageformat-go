package tests

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/kaptinlin/messageformat-go/pkg/bidi"
	pkgerrors "github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// Static error variables for non-typed failure modes.
var (
	ErrBadOption            = errors.New("bad option")
	ErrFormattingFailed     = errors.New("formatting failed")
	ErrInvalidDecimalPlaces = errors.New("invalid decimalPlaces option")
	ErrInvalidFailsOption   = errors.New("invalid fails option")
	ErrInvalidNumeric       = errors.New("invalid numeric input")
	ErrNotFormattable       = errors.New("not formattable")
	ErrNotPositiveInt       = errors.New("not a positive integer")
	ErrNotSelectable        = errors.New("not selectable")
	ErrNotString            = errors.New("not a string")
	ErrSelectionFailed      = errors.New("selection failed")
)

func newFunctionError(errorType pkgerrors.ErrorKind, cause error, format string, args ...any) error {
	err := pkgerrors.NewMessageFunctionError(errorType, fmt.Sprintf(format, args...))
	err.SetCause(cause)
	return err
}

// newBadOptionError constructs a spec-typed bad-option MessageFunctionError
// matching the TS reference test-functions, which throws/emits a
// MessageFunctionError('bad-option', ...) for invalid option values.
func newBadOptionError(format string, args ...any) error {
	return newFunctionError(pkgerrors.ErrorTypeBadOption, ErrBadOption, format, args...)
}

func newOptionError(cause error, format string, args ...any) error {
	return newFunctionError(pkgerrors.ErrorTypeBadOption, errors.Join(cause, ErrBadOption), format, args...)
}

// TestFunctions returns a map of test functions for MessageFormat testing.
// The naming mirrors the TS reference's test-functions: a value tagged
// "function" can both format and select, "format" can only format, and
// "select" can only select.
func TestFunctions() map[string]functions.MessageFunction {
	return map[string]functions.MessageFunction{
		"test":          testFunction,
		"test:function": testFunction,
		"test:select":   testSelectFunction,
		"test:format":   testFormatFunction,
		"placeholder":   placeholderFunction,
	}
}

// testValue represents a test function result with specific capabilities.
type testValue struct {
	source        string
	input         float64
	canFormat     bool
	selectable    bool
	decimalPlaces int
	badOption     bool
	failsFormat   bool
	failsSelect   bool
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
func (tv *testValue) Options() map[string]any {
	return nil
}

// ValueOf returns the underlying value
func (tv *testValue) ValueOf() (any, error) {
	return tv.input, nil
}

// SelectKeys performs selection for the test value
func (tv *testValue) SelectKeys(keys []string) ([]string, error) {
	if !tv.selectable {
		return nil, newFunctionError(pkgerrors.ErrorTypeUnsupportedOperation, ErrNotSelectable, "not selectable")
	}
	if tv.badOption {
		return nil, newBadOptionError("bad option")
	}
	if tv.failsSelect {
		return nil, newFunctionError(pkgerrors.ErrorTypeBadOption, ErrSelectionFailed, "Selection failed")
	}

	// Follow TypeScript logic: if value === 1
	if tv.input == 1 {
		// Check for "1.0" first if decimalPlaces === 1
		if tv.decimalPlaces == 1 {
			if slices.Contains(keys, "1.0") {
				return []string{"1.0"}, nil
			}
		}
		// Then check for "1"
		if slices.Contains(keys, "1") {
			return []string{"1"}, nil
		}
	}

	// Return null (empty slice) if no match found, like TypeScript
	return []string{}, nil
}

// ToString formats the test value as a string
func (tv *testValue) ToString() (string, error) {
	if !tv.canFormat {
		return "", newFunctionError(pkgerrors.ErrorTypeNotFormattable, ErrNotFormattable, "not formattable")
	}
	if tv.failsFormat {
		return "", newFunctionError(pkgerrors.ErrorTypeBadOption, ErrFormattingFailed, "Formatting failed")
	}
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
		return nil, newFunctionError(pkgerrors.ErrorTypeNotFormattable, ErrNotFormattable, "not formattable")
	}
	if tv.failsFormat {
		return nil, newFunctionError(pkgerrors.ErrorTypeBadOption, ErrFormattingFailed, "Formatting failed")
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
func testFunction(ctx functions.MessageFunctionContext, options functions.Options, operand any) messagevalue.MessageValue {
	return createTestValue(ctx, options, operand, true, true)
}

// testSelectFunction implements the :test:select behavior
func testSelectFunction(ctx functions.MessageFunctionContext, options functions.Options, operand any) messagevalue.MessageValue {
	return createTestValue(ctx, options, operand, false, true)
}

// testFormatFunction implements the :test:format behavior
func testFormatFunction(ctx functions.MessageFunctionContext, options functions.Options, operand any) messagevalue.MessageValue {
	return createTestValue(ctx, options, operand, true, false)
}

// createTestValue creates a test value with the specified capabilities
func createTestValue(ctx functions.MessageFunctionContext, options functions.Options, operand any, canFormat, selectable bool) messagevalue.MessageValue {
	// Get locale from context
	locale := "en"
	if locales := ctx.Locales(); len(locales) > 0 {
		locale = locales[0]
	}

	tv := &testValue{
		source:     ctx.Source(),
		canFormat:  canFormat,
		selectable: selectable,
		locale:     locale,
		dir:        bidi.DirAuto,
	}

	// Handle operand that might be another test value (mirrors TS valueOf
	// handling). Non-numeric input is a bad-operand error, including the case
	// where the operand resolved to a fallback value upstream — matching the
	// TS reference, which throws MessageFunctionError('bad-operand', ...).
	var inheritedTestValue *testValue

	if msgVal, ok := operand.(messagevalue.MessageValue); ok {
		if msgVal.Type() == "test" {
			if testVal, ok := msgVal.(*testValue); ok {
				inheritedTestValue = testVal
				tv.input = testVal.input
			}
		} else {
			input, err := parseNumericInputStrict(operand)
			if err != nil {
				ctx.OnError(pkgerrors.NewMessageFunctionError(pkgerrors.ErrorTypeBadOperand, "Input is not numeric"))
				return messagevalue.NewFallbackValue(ctx.Source(), locale)
			}
			tv.input = input
		}
	} else if testVal, ok := operand.(*testValue); ok {
		inheritedTestValue = testVal
		tv.input = testVal.input
	} else {
		input, err := parseNumericInputStrict(operand)
		if err != nil {
			ctx.OnError(pkgerrors.NewMessageFunctionError(pkgerrors.ErrorTypeBadOperand, "Input is not numeric"))
			return messagevalue.NewFallbackValue(ctx.Source(), locale)
		}
		tv.input = input
	}

	// Inherit fails* state from the upstream test value (matches the TS
	// reference's Object.assign over opt). decimalPlaces is only inherited if
	// not explicitly set in current options.
	if inheritedTestValue != nil {
		tv.failsFormat = inheritedTestValue.failsFormat
		tv.failsSelect = inheritedTestValue.failsSelect
		tv.badOption = inheritedTestValue.badOption
		if tv.badOption {
			ctx.OnError(newFunctionError(pkgerrors.ErrorTypeBadOperand, ErrBadOption, "bad operand"))
		}
		if _, hasDecimalPlaces := options["decimalPlaces"]; !hasDecimalPlaces {
			tv.decimalPlaces = inheritedTestValue.decimalPlaces
		}
	}

	// Process decimalPlaces option with strict validation like TypeScript.
	// The TS reference throws MessageFunctionError('bad-option', ...) on
	// invalid values; the runtime catches that and replaces the value with a
	// fallback. We mirror that here by emitting the error and returning a
	// fallback so downstream consumers (selectors, other functions) see the
	// fallback and emit their own cascade errors.
	if decimalPlaces, exists := options["decimalPlaces"]; exists {
		if dp, err := asPositiveInteger(decimalPlaces); err == nil && (dp == 0 || dp == 1) {
			tv.decimalPlaces = dp
		} else {
			tv.badOption = true
			ctx.OnError(newOptionError(ErrInvalidDecimalPlaces, "Invalid option decimalPlaces=%v", decimalPlaces))
			return tv
		}
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
				ctx.OnError(newOptionError(ErrInvalidFailsOption, "Invalid option fails=%v", fails))
			}
		} else {
			ctx.OnError(newOptionError(ErrInvalidFailsOption, "Invalid option fails=%v", fails))
		}
	}

	return tv
}

// parseNumericInputStrict parses input more strictly like TypeScript version
func parseNumericInputStrict(input any) (float64, error) {
	// Handle valueOf() method like TypeScript
	if obj, ok := input.(interface{ ValueOf() (any, error) }); ok {
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
func asPositiveInteger(value any) (int, error) {
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

func asString(value any) (string, error) {
	if str, ok := value.(string); ok {
		return str, nil
	}
	return "", ErrNotString
}

// placeholderFunction implements a simple placeholder function for testing
func placeholderFunction(ctx functions.MessageFunctionContext, options functions.Options, operand any) messagevalue.MessageValue {
	locale := "en"
	if locales := ctx.Locales(); len(locales) > 0 {
		locale = locales[0]
	}

	// Return a simple text value
	return messagevalue.NewStringValue("placeholder", locale, ctx.Source())
}
