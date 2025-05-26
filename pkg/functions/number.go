package functions

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"

	"github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// NumericInput represents parsed numeric input
type NumericInput struct {
	Value   interface{}
	Options map[string]interface{}
}

// readNumericOperand parses numeric operand
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
func readNumericOperand(value interface{}, source string) (*NumericInput, error) {
	var options map[string]interface{}

	// Check if value has valueOf method and options
	if obj, ok := value.(map[string]interface{}); ok {
		if valueOf, hasValueOf := obj["valueOf"]; hasValueOf {
			if optionsVal, hasOptions := obj["options"]; hasOptions {
				if optMap, ok := optionsVal.(map[string]interface{}); ok {
					options = optMap
				}
			}
			value = valueOf
		}
	}

	// Parse string values as JSON numbers
	if str, ok := value.(string); ok {
		if parsed, err := parseJSONNumber(str); err == nil {
			value = parsed
		} else {
			return nil, errors.NewResolutionError(
				errors.ErrorTypeBadOperand,
				"Input is not numeric",
				source,
			)
		}
	}

	// Validate numeric type
	switch value.(type) {
	case int, int8, int16, int32, int64:
	case uint, uint8, uint16, uint32, uint64:
	case float32, float64:
	case *big.Int, *big.Float:
	default:
		return nil, errors.NewResolutionError(
			errors.ErrorTypeBadOperand,
			"Input is not numeric",
			source,
		)
	}

	return &NumericInput{
		Value:   value,
		Options: options,
	}, nil
}

// NumberFunction implements the :number function
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
//	  // ... option processing ...
//	  return getMessageNumber(ctx, value, options, true);
//	}
func NumberFunction(
	ctx MessageFunctionContext,
	options map[string]interface{},
	operand interface{},
) messagevalue.MessageValue {
	// Read numeric operand
	numInput, err := readNumericOperand(operand, ctx.Source())
	if err != nil {
		ctx.OnError(err)
		return messagevalue.NewFallbackValue(ctx.Source(), getFirstLocale(ctx.Locales()))
	}

	// Merge options from operand and expression
	mergedOptions := mergeNumberOptions(numInput.Options, options, ctx.LocaleMatcher())

	return getMessageNumber(ctx, numInput.Value, mergedOptions, true)
}

// IntegerFunction implements the :integer function
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
//	  // ... option processing ...
//	  return getMessageNumber(ctx, value, options, true);
//	}
func IntegerFunction(
	ctx MessageFunctionContext,
	options map[string]interface{},
	operand interface{},
) messagevalue.MessageValue {
	// Read numeric operand
	numInput, err := readNumericOperand(operand, ctx.Source())
	if err != nil {
		ctx.OnError(err)
		return messagevalue.NewFallbackValue(ctx.Source(), getFirstLocale(ctx.Locales()))
	}

	// Round to integer
	var value interface{}
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

	// Merge options
	mergedOptions := mergeNumberOptions(numInput.Options, options, ctx.LocaleMatcher())
	mergedOptions["maximumFractionDigits"] = 0

	return getMessageNumber(ctx, value, mergedOptions, true)
}

// getMessageNumber creates a MessageNumber value
func getMessageNumber(
	ctx MessageFunctionContext,
	value interface{},
	options map[string]interface{},
	canSelect bool,
) messagevalue.MessageValue {
	// Validate select option
	if canSelect {
		if selectVal, hasSelect := options["select"]; hasSelect {
			if !ctx.LiteralOptionKeys()["select"] {
				ctx.OnError(errors.NewResolutionError(
					errors.ErrorTypeBadOption,
					"The option select may only be set by a literal value",
					ctx.Source(),
				))
				canSelect = false
			} else {
				// Validate select value
				if selectStr, ok := selectVal.(string); ok {
					if selectStr != "exact" && selectStr != "cardinal" && selectStr != "ordinal" {
						ctx.OnError(errors.NewResolutionError(
							errors.ErrorTypeBadOption,
							fmt.Sprintf("invalid select value: %s", selectStr),
							ctx.Source(),
						))
					}
				}
			}
		}
	}

	// Get first locale
	locale := getFirstLocale(ctx.Locales())

	return messagevalue.NewNumberValue(value, locale, ctx.Source(), options)
}

// mergeNumberOptions merges options from different sources
func mergeNumberOptions(
	operandOptions map[string]interface{},
	exprOptions map[string]interface{},
	localeMatcher string,
) map[string]interface{} {
	merged := map[string]interface{}{
		"localeMatcher": localeMatcher,
		"style":         "decimal",
	}

	// Add operand options first
	if operandOptions != nil {
		for k, v := range operandOptions {
			merged[k] = v
		}
	}

	// Add expression options (override operand options)
	if exprOptions != nil {
		for k, v := range exprOptions {
			if k == "locale" {
				continue // Handle locale separately
			}

			// Validate and convert option values
			switch k {
			case "minimumIntegerDigits", "minimumFractionDigits", "maximumFractionDigits",
				"minimumSignificantDigits", "maximumSignificantDigits", "roundingIncrement":
				if intVal, err := asPositiveInteger(v); err == nil {
					merged[k] = intVal
				}
			case "roundingMode", "roundingPriority", "select", "signDisplay",
				"trailingZeroDisplay", "useGrouping", "style":
				if strVal, err := asString(v); err == nil {
					merged[k] = strVal
				}
			}
		}
	}

	return merged
}

// Helper functions
func parseJSONNumber(s string) (interface{}, error) {
	// Try integer first
	if intVal, err := strconv.ParseInt(s, 10, 64); err == nil {
		return intVal, nil
	}

	// Try float
	if floatVal, err := strconv.ParseFloat(s, 64); err == nil {
		return floatVal, nil
	}

	// Try JSON parsing for edge cases
	var jsonVal interface{}
	if err := json.Unmarshal([]byte(s), &jsonVal); err == nil {
		switch v := jsonVal.(type) {
		case float64, int64:
			return v, nil
		}
	}

	return nil, fmt.Errorf("not a valid number: %s", s)
}

func isFinite(f float64) bool {
	return !math.IsInf(f, 0) && !math.IsNaN(f)
}
