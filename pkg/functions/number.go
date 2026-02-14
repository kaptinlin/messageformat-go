package functions

import (
	"errors"
	"fmt"
	"maps"
	"math"
	"math/big"

	"github.com/go-json-experiment/json"

	"github.com/kaptinlin/messageformat-go/pkg/bidi"
	pkgErrors "github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// Static errors to avoid dynamic error creation
var (
	ErrNotValidJSONNumber = errors.New("not a valid JSON number")
)

// NumericInput represents parsed numeric input with value and options
// TypeScript original code:
// { value: number | bigint; options: unknown }
type NumericInput struct {
	Value   any
	Options map[string]any
}

// readNumericOperand parses numeric operand and extracts value and options
// TypeScript original code:
// export function readNumericOperand(
//
//	value: unknown,
//	source: string
//
//	): { value: number | bigint; options: unknown } {
//	  let options: unknown = undefined;
//	  if (typeof value === 'object') {
//	    const valueOf = value?.valueOf;
//	    if (typeof valueOf === 'function') {
//	      options = (value as { options: unknown }).options;
//	      value = valueOf.call(value);
//	    }
//	  }
//	  if (typeof value === 'string') {
//	    try {
//	      value = JSON.parse(value);
//	    } catch {
//	      // handled below
//	    }
//	  }
//	  if (typeof value !== 'bigint' && typeof value !== 'number') {
//	    const msg = 'Input is not numeric';
//	    throw new MessageResolutionError('bad-operand', msg, source);
//	  }
//	  return { value, options };
//	}
func readNumericOperand(value any, source string) (*NumericInput, error) {
	// Check for nil operand first - should return bad-operand error
	if value == nil {
		return nil, pkgErrors.NewMessageResolutionError(
			pkgErrors.ErrorTypeBadOperand,
			"Input is not numeric",
			source,
		)
	}

	var options map[string]any

	// Check if it's a MessageValue with valueOf method
	if mv, ok := value.(messagevalue.MessageValue); ok {
		// Special case: if it's a FallbackValue, it's a bad operand
		if mv.Type() == "fallback" {
			return nil, pkgErrors.NewMessageResolutionError(
				pkgErrors.ErrorTypeBadOperand,
				"Input is not numeric",
				source,
			)
		}
		// Get the underlying value from MessageValue
		rawValue, err := mv.ValueOf()
		if err != nil {
			return nil, pkgErrors.NewMessageResolutionError(
				pkgErrors.ErrorTypeBadOperand,
				"Input is not numeric",
				source,
			)
		}
		value = rawValue

		// If it's a NumberValue, extract its options
		if nv, ok := mv.(*messagevalue.NumberValue); ok {
			if nvOptions := nv.Options(); nvOptions != nil {
				options = nvOptions
			}
		}
	} else if obj, ok := value.(map[string]any); ok {
		// Check if value has valueOf method and options
		if valueOf, hasValueOf := obj["valueOf"]; hasValueOf {
			if optionsVal, hasOptions := obj["options"]; hasOptions {
				if optMap, ok := optionsVal.(map[string]any); ok {
					options = optMap
				}
			}
			value = valueOf
		}
	}

	// Parse string values as JSON numbers - matches TypeScript logic
	// TypeScript: if (typeof value === 'string') { try { value = JSON.parse(value); } catch { } }
	if str, ok := value.(string); ok {
		if parsed, err := parseJSONNumber(str); err == nil {
			value = parsed
		} else {
			return nil, pkgErrors.NewMessageResolutionError(
				pkgErrors.ErrorTypeBadOperand,
				"Input is not numeric",
				source,
			)
		}
	}

	// Validate numeric type - matches TypeScript logic
	// TypeScript: if (typeof value !== 'bigint' && typeof value !== 'number') { throw ... }
	switch value.(type) {
	case int, int8, int16, int32, int64:
	case uint, uint8, uint16, uint32, uint64:
	case float32, float64:
	case *big.Int, *big.Float:
	default:
		return nil, pkgErrors.NewMessageResolutionError(
			pkgErrors.ErrorTypeBadOperand,
			"Input is not numeric",
			source,
		)
	}

	return &NumericInput{
		Value:   value,
		Options: options,
	}, nil
}

// NumberFunction implements the :number function for numeric value formatting and selection.
//
// Status: Stable (REQUIRED in LDML 48)
// Specification: https://www.unicode.org/reports/tr35/tr35-76/tr35-messageFormat.html#the-number-function
//
// The :number function is a selector and formatter for numeric values.
//
// Operand Requirements:
// - Must be a number, BigInt, or string representing a JSON number
// - Can be an object with valueOf() method and optional options map
//
// Supported Options:
// - select: 'exact' | 'plural' | 'ordinal' (controls selection behavior)
// - minimumIntegerDigits, minimumFractionDigits, maximumFractionDigits (digit size options)
// - minimumSignificantDigits, maximumSignificantDigits (digit size options)
// - roundingMode, roundingPriority, roundingIncrement (rounding control)
// - signDisplay, useGrouping, trailingZeroDisplay (formatting control)
//
// Selection Behavior:
// - Exact numeric match is preferred over plural category
// - Supports plural rules (zero, one, two, few, many, other)
// - The select option must be set by a literal value, not a variable
//
// TypeScript Reference: .reference/messageformat/mf2/messageformat/src/functions/number.ts
//
// TypeScript original code:
// export function number(
//
//	ctx: MessageFunctionContext,
//	exprOpt: Record<string, unknown>,
//	operand?: unknown
//
//	): MessageNumber {
//	  const input = readNumericOperand(operand, ctx.source);
//	  const value = input.value;
//	  const options: MessageNumberOptions = Object.assign({}, input.options, {
//	    localeMatcher: ctx.localeMatcher,
//	    style: 'decimal'
//	  } as const);
//	  for (const [name, optval] of Object.entries(exprOpt)) {
//	    if (optval === undefined) continue;
//	    try {
//	      switch (name) {
//	        case 'minimumIntegerDigits':
//	        case 'minimumFractionDigits':
//	        case 'maximumFractionDigits':
//	        case 'minimumSignificantDigits':
//	        case 'maximumSignificantDigits':
//	        case 'roundingIncrement':
//	          options[name] = asPositiveInteger(optval);
//	          break;
//	        case 'roundingMode':
//	        case 'roundingPriority':
//	        case 'select':
//	        case 'signDisplay':
//	        case 'trailingZeroDisplay':
//	        case 'useGrouping':
//	          options[name] = asString(optval);
//	      }
//	    } catch {
//	      const msg = `Value ${optval} is not valid for :number option ${name}`;
//	      ctx.onError(new MessageResolutionError('bad-option', msg, ctx.source));
//	    }
//	  }
//	  return getMessageNumber(ctx, value, options, true);
//	}
func NumberFunction(
	ctx MessageFunctionContext,
	options map[string]any,
	operand any,
) messagevalue.MessageValue {
	// Read numeric operand - matches TypeScript: const input = readNumericOperand(operand, ctx.source);
	numInput, err := readNumericOperand(operand, ctx.Source())
	if err != nil {
		ctx.OnError(err)
		return messagevalue.NewFallbackValue(ctx.Source(), getFirstLocale(ctx.Locales()))
	}

	// Start with operand options and set defaults - matches TypeScript Object.assign
	mergedOptions := mergeNumberOptions(numInput.Options, nil, ctx.LocaleMatcher())
	// Don't force style=decimal here - let user options override

	// Process expression options - matches TypeScript for loop
	for name, optval := range options {
		if optval == nil {
			continue // matches TypeScript: if (optval === undefined) continue;
		}

		// Process options with validation - matches TypeScript try/catch blocks
		switch name {
		case "minimumIntegerDigits", "minimumFractionDigits", "maximumFractionDigits",
			"minimumSignificantDigits", "maximumSignificantDigits", "roundingIncrement":
			if intVal, err := asPositiveInteger(optval); err == nil {
				mergedOptions[name] = intVal
			} else {
				msg := fmt.Sprintf("Value %v is not valid for :number option %s", optval, name)
				ctx.OnError(pkgErrors.NewBadOptionError(msg, ctx.Source()))
			}
		case "roundingMode", "roundingPriority", "select", "signDisplay",
			"trailingZeroDisplay", "useGrouping", "style", "currency", "currencyDisplay", "currencySign":
			if strVal, err := asString(optval); err == nil {
				mergedOptions[name] = strVal
			} else {
				msg := fmt.Sprintf("Value %v is not valid for :number option %s", optval, name)
				ctx.OnError(pkgErrors.NewBadOptionError(msg, ctx.Source()))
			}
		default:
			// Unknown option - silently ignore to match TypeScript behavior
		}
	}

	// Set default style if not specified
	if _, hasStyle := mergedOptions["style"]; !hasStyle {
		mergedOptions["style"] = "decimal"
	}

	return getMessageNumber(ctx, numInput.Value, mergedOptions, true)
}

// IntegerFunction implements the :integer function for integer number formatting
// TypeScript original code:
// export function integer(
//
//	ctx: MessageFunctionContext,
//	exprOpt: Record<string, unknown>,
//	operand?: unknown
//
//	) {
//	  const input = readNumericOperand(operand, ctx.source);
//	  const value = Number.isFinite(input.value)
//	    ? Math.round(input.value as number)
//	    : input.value;
//	  const options: MessageNumberOptions = Object.assign({}, input.options, {
//	    localeMatcher: ctx.localeMatcher,
//	    maximumFractionDigits: 0
//	  } as const);
//	  for (const [name, optval] of Object.entries(exprOpt)) {
//	    if (optval === undefined) continue;
//	    try {
//	      switch (name) {
//	        case 'minimumIntegerDigits':
//	        case 'maximumSignificantDigits':
//	          options[name] = asPositiveInteger(optval);
//	          break;
//	        case 'select':
//	        case 'signDisplay':
//	        case 'useGrouping':
//	          options[name] = asString(optval);
//	      }
//	    } catch {
//	      const msg = `Value ${optval} is not valid for :integer option ${name}`;
//	      ctx.onError(new MessageResolutionError('bad-option', msg, ctx.source));
//	    }
//	  }
//	  return getMessageNumber(ctx, value, options, true);
//	}
func IntegerFunction(
	ctx MessageFunctionContext,
	options map[string]any,
	operand any,
) messagevalue.MessageValue {
	// Read numeric operand - matches TypeScript: const input = readNumericOperand(operand, ctx.source);
	numInput, err := readNumericOperand(operand, ctx.Source())
	if err != nil {
		ctx.OnError(err)
		return messagevalue.NewFallbackValue(ctx.Source(), getFirstLocale(ctx.Locales()))
	}

	// Round to integer - matches TypeScript: Number.isFinite(input.value) ? Math.round(input.value as number) : input.value;
	var value any
	switch v := numInput.Value.(type) {
	case float64:
		if isFinite(v) {
			value = int64(math.Round(v))
		} else {
			value = v
		}
	case float32:
		if isFinite(float64(v)) {
			value = int64(math.Round(float64(v)))
		} else {
			value = v
		}
	case *big.Float:
		if v.IsInf() {
			value = v
		} else {
			intVal, _ := v.Int64()
			value = intVal
		}
	default:
		value = numInput.Value
	}

	// Start with operand options and set defaults - matches TypeScript Object.assign
	mergedOptions := mergeNumberOptions(numInput.Options, nil, ctx.LocaleMatcher())
	mergedOptions["maximumFractionDigits"] = 0

	// Process expression options - matches TypeScript for loop
	for name, optval := range options {
		if optval == nil {
			continue // matches TypeScript: if (optval === undefined) continue;
		}

		// Process options with validation - matches TypeScript try/catch blocks
		switch name {
		case "minimumIntegerDigits", "maximumSignificantDigits":
			if intVal, err := asPositiveInteger(optval); err == nil {
				mergedOptions[name] = intVal
			} else {
				msg := fmt.Sprintf("Value %v is not valid for :integer option %s", optval, name)
				ctx.OnError(pkgErrors.NewBadOptionError(msg, ctx.Source()))
			}
		case "select", "signDisplay", "useGrouping":
			if strVal, err := asString(optval); err == nil {
				mergedOptions[name] = strVal
			} else {
				msg := fmt.Sprintf("Value %v is not valid for :integer option %s", optval, name)
				ctx.OnError(pkgErrors.NewBadOptionError(msg, ctx.Source()))
			}
		default:
			// Unknown option - silently ignore to match TypeScript behavior
		}
	}

	return getMessageNumber(ctx, value, mergedOptions, true)
}

// getMessageNumber creates a MessageNumber value with formatting options
// TypeScript original code:
// export function getMessageNumber(
//
//	ctx: MessageFunctionContext,
//	value: number | bigint,
//	options: MessageNumberOptions,
//	canSelect: boolean
//
// ): MessageNumber
func getMessageNumber(
	ctx MessageFunctionContext,
	value any,
	options map[string]any,
	canSelect bool,
) messagevalue.MessageValue {
	// Validate select option - matches TypeScript select validation logic
	if canSelect {
		if selectVal, hasSelect := options["select"]; hasSelect {
			// Check if select option is set by literal value
			if !ctx.LiteralOptionKeys()["select"] {
				ctx.OnError(pkgErrors.NewBadOptionError("The option select may only be set by a literal value", ctx.Source()))
				canSelect = false
			} else {
				// Validate select value - matches TypeScript select value validation
				if selectStr, ok := selectVal.(string); ok {
					if selectStr != "exact" && selectStr != "cardinal" && selectStr != "ordinal" {
						msg := fmt.Sprintf("invalid select value: %s", selectStr)
						ctx.OnError(pkgErrors.NewBadOptionError(msg, ctx.Source()))
					}
				} else {
					ctx.OnError(pkgErrors.NewBadOptionError("select option must be a string", ctx.Source()))
				}
			}
		}
	}

	// Get first locale - matches TypeScript locale handling
	locale := getFirstLocale(ctx.Locales())

	// Determine direction - matches TypeScript: let { dir, locales } = ctx;
	dir := ctx.Dir()
	if dir == "" {
		// If dir is not set in context, determine from locale
		// matches TypeScript: dir = getLocaleDir(locale);
		dir = string(bidi.GetLocaleDirection(locale))
	}

	// Convert string direction to bidi.Direction
	var bidiDir bidi.Direction
	switch dir {
	case "ltr":
		bidiDir = bidi.DirLTR
	case "rtl":
		bidiDir = bidi.DirRTL
	default:
		bidiDir = bidi.DirAuto
	}

	return messagevalue.NewNumberValueWithSelection(value, locale, ctx.Source(), bidiDir, options, canSelect)
}

// mergeNumberOptions merges options from operand and expression sources
// Combines operand options with expression options, with expression options taking precedence
// Matches TypeScript Object.assign({}, input.options, { localeMatcher: ctx.localeMatcher, ... })
func mergeNumberOptions(
	operandOptions map[string]any,
	exprOptions map[string]any,
	localeMatcher string,
) map[string]any {
	// Start with empty map - matches TypeScript Object.assign({}, ...)
	merged := make(map[string]any)

	// Add operand options first - matches TypeScript Object.assign({}, input.options, ...)
	maps.Copy(merged, operandOptions)

	// Add default options - matches TypeScript Object.assign(..., { localeMatcher: ctx.localeMatcher })
	merged["localeMatcher"] = localeMatcher

	// Add expression options (override operand options) - matches TypeScript Object.assign(..., exprOpt)
	maps.Copy(merged, exprOptions)

	return merged
}

// parseJSONNumber parses a string as a JSON number (integer or float)
// Matches TypeScript JSON.parse() behavior for numeric values
// This function strictly follows JSON number format rules to match TypeScript behavior
func parseJSONNumber(s string) (any, error) {
	// Use JSON.Unmarshal to strictly validate JSON number format
	// This will reject invalid JSON numbers like "00", "042", "1.", ".1", "+1", etc.
	var jsonVal any
	if err := json.Unmarshal([]byte(s), &jsonVal); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNotValidJSONNumber, s)
	}

	// JSON.Unmarshal only returns float64 for numbers
	if floatVal, ok := jsonVal.(float64); ok {
		// Check if it's actually an integer value
		if floatVal == float64(int64(floatVal)) && floatVal >= float64(math.MinInt64) && floatVal <= float64(math.MaxInt64) {
			return int64(floatVal), nil
		}
		return floatVal, nil
	}

	return nil, fmt.Errorf("%w: %s", ErrNotValidJSONNumber, s)
}

// isFinite checks if a float64 value is finite (not infinite or NaN)
// Matches TypeScript Number.isFinite() behavior
func isFinite(f float64) bool {
	return !math.IsInf(f, 0) && !math.IsNaN(f)
}
