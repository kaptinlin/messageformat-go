package functions

import (
	"fmt"
	"maps"

	"github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// MathFunction implements the :math function (DRAFT)
// math accepts a numeric value as input and adds or subtracts an integer value from it
//
// TypeScript original code:
// export function math(
//
//	ctx: MessageFunctionContext,
//	exprOpt: Record<string | symbol, unknown>,
//	operand?: unknown
//
//	): MessageNumber {
//	  const { source } = ctx;
//	  let { value, options } = readNumericOperand(operand, source);
//
//	  let add: number;
//	  let sub: number;
//	  try {
//	    add = 'add' in exprOpt ? asPositiveInteger(exprOpt.add) : -1;
//	    sub = 'subtract' in exprOpt ? asPositiveInteger(exprOpt.subtract) : -1;
//	  } catch (error) {
//	    throw new MessageResolutionError('bad-option', String(error), source);
//	  }
//	  if (add < 0 === sub < 0) {
//	    const msg =
//	      'Exactly one of "add" or "subtract" is required as a :math option';
//	    throw new MessageResolutionError('bad-option', msg, source);
//	  }
//	  const delta = add < 0 ? -sub : add;
//	  if (typeof value === 'number') value += delta;
//	  else value += BigInt(delta);
//
//	  return number(ctx, {}, { valueOf: () => value, options });
//	}
func MathFunction(
	ctx MessageFunctionContext,
	options map[string]any,
	operand any,
) messagevalue.MessageValue {
	source := ctx.Source()

	// Read numeric operand
	numericOperand, err := readNumericOperand(operand, source)
	if err != nil {
		ctx.OnError(err)
		return messagevalue.NewFallbackValue(source, GetFirstLocale(ctx.Locales()))
	}

	value := numericOperand.Value
	operandOptions := numericOperand.Options

	// Parse add and subtract options
	add, subtract := -1, -1

	// Try to get add option
	if addVal, ok := options["add"]; ok {
		if addInt, err := asPositiveInteger(addVal); err == nil {
			add = addInt
		} else {
			ctx.OnError(errors.NewBadOptionError(fmt.Sprintf("Invalid add option: %v", err), source))
			return messagevalue.NewFallbackValue(source, GetFirstLocale(ctx.Locales()))
		}
	}

	// Try to get subtract option
	if subVal, ok := options["subtract"]; ok {
		if subInt, err := asPositiveInteger(subVal); err == nil {
			subtract = subInt
		} else {
			ctx.OnError(errors.NewBadOptionError(fmt.Sprintf("Invalid subtract option: %v", err), source))
			return messagevalue.NewFallbackValue(source, GetFirstLocale(ctx.Locales()))
		}
	}

	// Exactly one of "add" or "subtract" is required
	if (add < 0) == (subtract < 0) {
		msg := "Exactly one of \"add\" or \"subtract\" is required as a :math option"
		ctx.OnError(errors.NewBadOptionError(msg, source))
		return messagevalue.NewFallbackValue(source, GetFirstLocale(ctx.Locales()))
	}

	// Calculate delta
	var delta int
	if add >= 0 {
		delta = add
	} else {
		delta = -subtract
	}

	// Apply delta to value
	var newValue any
	switch v := value.(type) {
	case int:
		newValue = v + delta
	case int64:
		newValue = v + int64(delta)
	case float64:
		newValue = v + float64(delta)
	case float32:
		newValue = float64(v) + float64(delta)
	default:
		// Try to convert to float64 and add
		if floatVal, ok := convertToFloat64(v); ok {
			newValue = floatVal + float64(delta)
		} else {
			ctx.OnError(errors.NewBadOperandError("Cannot perform math operation on non-numeric value", source))
			return messagevalue.NewFallbackValue(source, GetFirstLocale(ctx.Locales()))
		}
	}

	// Delegate to number function with the new value and merged options
	// Merge the original operand options with any new options
	mergedOptions := make(map[string]any)
	maps.Copy(mergedOptions, operandOptions)

	// Delegate to number function with the new value
	return NumberFunction(ctx, mergedOptions, newValue)
}

// convertToFloat64 attempts to convert various numeric types to float64
func convertToFloat64(value any) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case float64:
		return v, true
	case float32:
		return float64(v), true
	default:
		return 0, false
	}
}
