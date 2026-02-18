package functions

import (
	"fmt"
	"math/big"

	pkgErrors "github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// OffsetFunction accepts a numeric value as input and adds or subtracts an integer value from it
// TypeScript original code:
// export function offset(
//
//	ctx: MessageFunctionContext,
//	exprOpt: Record<string | symbol, unknown>,
//	operand?: unknown
//
//	): MessageNumber {
//	  let { value, options } = readNumericOperand(operand);
//	  let add: number;
//	  try {
//	    add = 'add' in exprOpt ? asPositiveInteger(exprOpt.add) : -1;
//	  } catch {
//	    throw new MessageFunctionError(
//	      'bad-option',
//	      `Value ${exprOpt.add} is not valid for :offset option add`
//	    );
//	  }
//	  let sub: number;
//	  try {
//	    sub = 'subtract' in exprOpt ? asPositiveInteger(exprOpt.subtract) : -1;
//	  } catch {
//	    throw new MessageFunctionError(
//	      'bad-option',
//	      `Value ${exprOpt.subtract} is not valid for :offset option subtract`
//	    );
//	  }
//	  if (add < 0 === sub < 0) {
//	    const msg =
//	      'Exactly one of "add" or "subtract" is required as an :offset option';
//	    throw new MessageFunctionError('bad-option', msg);
//	  }
//	  const delta = add < 0 ? -sub : add;
//	  if (typeof value === 'number') value += delta;
//	  else value += BigInt(delta);
//	  return number(ctx, {}, { valueOf: () => value, options });
//	}
func OffsetFunction(
	ctx MessageFunctionContext,
	options map[string]any,
	operand any,
) messagevalue.MessageValue {
	// Read numeric operand - matches TypeScript: let { value, options } = readNumericOperand(operand);
	numInput, err := readNumericOperand(operand, ctx.Source())
	if err != nil {
		ctx.OnError(err)
		return messagevalue.NewFallbackValue(ctx.Source(), GetFirstLocale(ctx.Locales()))
	}

	value := numInput.Value

	// Parse add option - matches TypeScript: add = 'add' in exprOpt ? asPositiveInteger(exprOpt.add) : -1;
	add := -1
	if addVal, hasAdd := options["add"]; hasAdd {
		if addInt, err := asPositiveInteger(addVal); err == nil {
			add = addInt
		} else {
			msg := fmt.Sprintf("Value %v is not valid for :offset option add", addVal)
			ctx.OnError(pkgErrors.NewBadOptionError(msg, ctx.Source()))
			return messagevalue.NewFallbackValue(ctx.Source(), GetFirstLocale(ctx.Locales()))
		}
	}

	// Parse subtract option - matches TypeScript: sub = 'subtract' in exprOpt ? asPositiveInteger(exprOpt.subtract) : -1;
	sub := -1
	if subVal, hasSubtract := options["subtract"]; hasSubtract {
		if subInt, err := asPositiveInteger(subVal); err == nil {
			sub = subInt
		} else {
			msg := fmt.Sprintf("Value %v is not valid for :offset option subtract", subVal)
			ctx.OnError(pkgErrors.NewBadOptionError(msg, ctx.Source()))
			return messagevalue.NewFallbackValue(ctx.Source(), GetFirstLocale(ctx.Locales()))
		}
	}

	// Validate that exactly one of add or subtract is provided - matches TypeScript: if (add < 0 === sub < 0)
	if (add < 0) == (sub < 0) {
		msg := "Exactly one of \"add\" or \"subtract\" is required as an :offset option"
		ctx.OnError(pkgErrors.NewBadOptionError(msg, ctx.Source()))
		return messagevalue.NewFallbackValue(ctx.Source(), GetFirstLocale(ctx.Locales()))
	}

	// Calculate delta - matches TypeScript: const delta = add < 0 ? -sub : add;
	delta := add
	if add < 0 {
		delta = -sub
	}

	// Apply offset to value - matches TypeScript: if (typeof value === 'number') value += delta; else value += BigInt(delta);
	switch v := value.(type) {
	case int:
		value = v + delta
	case int64:
		value = v + int64(delta)
	case float64:
		value = v + float64(delta)
	case float32:
		value = v + float32(delta)
	case *big.Int:
		value = new(big.Int).Add(v, big.NewInt(int64(delta)))
	case *big.Float:
		deltaFloat := big.NewFloat(float64(delta))
		value = new(big.Float).Add(v, deltaFloat)
	default:
		// For other numeric types, try to convert to float64 and add delta
		if floatVal, ok := convertToFloat64(value); ok {
			value = floatVal + float64(delta)
		} else {
			msg := fmt.Sprintf("Cannot apply offset to value of type %T", value)
			ctx.OnError(pkgErrors.NewBadOperandError(msg, ctx.Source()))
			return messagevalue.NewFallbackValue(ctx.Source(), GetFirstLocale(ctx.Locales()))
		}
	}

	// Return number function result - matches TypeScript: return number(ctx, {}, { valueOf: () => value, options });
	// Create a value object that can be processed by NumberFunction
	valueObj := map[string]any{
		"valueOf": value,
		"options": numInput.Options,
	}

	return NumberFunction(ctx, map[string]any{}, valueObj)
}
