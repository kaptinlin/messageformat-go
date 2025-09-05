package functions

import (
	"fmt"

	pkgErrors "github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// PercentFunction implements the :percent function for percent value formatting
// TypeScript original code:
// export function percent(
//
//	ctx: MessageFunctionContext,
//	exprOpt: Record<string | symbol, unknown>,
//	operand?: unknown
//
//	): MessageNumber {
//	  const input = readNumericOperand(operand);
//	  const options: MessageNumberOptions = Object.assign({}, input.options, {
//	    localeMatcher: ctx.localeMatcher,
//	    style: 'percent'
//	  } as const);
//
//	  for (const [name, optval] of Object.entries(exprOpt)) {
//	    if (optval === undefined) continue;
//	    try {
//	      switch (name) {
//	        case 'roundingMode':
//	        case 'roundingPriority':
//	        case 'signDisplay':
//	        case 'trailingZeroDisplay':
//	        case 'useGrouping':
//	          // @ts-expect-error Let Intl.NumberFormat construction fail
//	          options[name] = asString(optval);
//	          break;
//	        case 'minimumFractionDigits':
//	        case 'maximumFractionDigits':
//	        case 'minimumSignificantDigits':
//	        case 'maximumSignificantDigits':
//	          options[name] = asPositiveInteger(optval);
//	          break;
//	      }
//	    } catch {
//	      ctx.onError(
//	        'bad-option',
//	        `Value ${optval} is not valid for :percent option ${name}`
//	      );
//	    }
//	  }
//
//	  return getMessageNumber(ctx, input.value, options, true);
//	}
func PercentFunction(
	ctx MessageFunctionContext,
	options map[string]interface{},
	operand interface{},
) messagevalue.MessageValue {
	// Read numeric operand - matches TypeScript: const input = readNumericOperand(operand);
	numInput, err := readNumericOperand(operand, ctx.Source())
	if err != nil {
		ctx.OnError(err)
		return messagevalue.NewFallbackValue(ctx.Source(), getFirstLocale(ctx.Locales()))
	}

	// Start with operand options and set defaults - matches TypeScript Object.assign
	mergedOptions := mergeNumberOptions(numInput.Options, nil, ctx.LocaleMatcher())
	mergedOptions["style"] = "percent" // Set percent style

	// Process expression options - matches TypeScript for loop
	for name, optval := range options {
		if optval == nil {
			continue // matches TypeScript: if (optval === undefined) continue;
		}

		// Process options with validation - matches TypeScript try/catch blocks
		switch name {
		case "roundingMode", "roundingPriority", "signDisplay", "trailingZeroDisplay", "useGrouping":
			if strVal, err := asString(optval); err == nil {
				mergedOptions[name] = strVal
			} else {
				msg := fmt.Sprintf("Value %v is not valid for :percent option %s", optval, name)
				ctx.OnError(pkgErrors.NewBadOptionError(msg, ctx.Source()))
			}
		case "minimumFractionDigits", "maximumFractionDigits", "minimumSignificantDigits", "maximumSignificantDigits":
			if intVal, err := asPositiveInteger(optval); err == nil {
				mergedOptions[name] = intVal
			} else {
				msg := fmt.Sprintf("Value %v is not valid for :percent option %s", optval, name)
				ctx.OnError(pkgErrors.NewBadOptionError(msg, ctx.Source()))
			}
		default:
			// Unknown option - silently ignore to match TypeScript behavior
		}
	}

	return getMessageNumber(ctx, numInput.Value, mergedOptions, true)
}
