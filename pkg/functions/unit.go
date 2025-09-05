package functions

import (
	"fmt"

	pkgErrors "github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// UnitFunction implements the :unit function (DRAFT)
// unit accepts as input numerical values as well as
// objects wrapping a numerical value that also include a unit property.
//
// TypeScript original code:
// export function unit(
//
//	ctx: MessageFunctionContext,
//	exprOpt: Record<string | symbol, unknown>,
//	operand?: unknown
//
//	): MessageNumber {
//	  const input = readNumericOperand(operand, ctx.source);
//	  const options: MessageNumberOptions = Object.assign({}, input.options, {
//	    localeMatcher: ctx.localeMatcher,
//	    style: 'unit'
//	  } as const);
//
//	  for (const [name, optval] of Object.entries(exprOpt)) {
//	    if (optval === undefined) continue;
//	    try {
//	      switch (name) {
//	        case 'signDisplay':
//	        case 'roundingMode':
//	        case 'roundingPriority':
//	        case 'trailingZeroDisplay':
//	        case 'unit':
//	        case 'unitDisplay':
//	        case 'useGrouping':
//	          options[name] = asString(optval);
//	          break;
//	        case 'minimumIntegerDigits':
//	        case 'minimumFractionDigits':
//	        case 'maximumFractionDigits':
//	        case 'minimumSignificantDigits':
//	        case 'maximumSignificantDigits':
//	        case 'roundingIncrement':
//	          options[name] = asPositiveInteger(optval);
//	          break;
//	      }
//	    } catch (error) {
//	      if (error instanceof MessageError) {
//	        ctx.onError(error);
//	      } else {
//	        const msg = `Value ${optval} is not valid for :currency option ${name}`;
//	        ctx.onError(new MessageResolutionError('bad-option', msg, ctx.source));
//	      }
//	    }
//	  }
//
//	  if (!options.unit) {
//	    const msg = 'A unit identifier is required for :unit';
//	    throw new MessageResolutionError('bad-operand', msg, ctx.source);
//	  }
//
//	  return getMessageNumber(ctx, input.value, options, false);
//	}
func UnitFunction(
	ctx MessageFunctionContext,
	options map[string]interface{},
	operand interface{},
) messagevalue.MessageValue {
	source := ctx.Source()

	// Read numeric operand
	numericOperand, err := readNumericOperand(operand, source)
	if err != nil {
		ctx.OnError(err)
		return messagevalue.NewFallbackValue(source, getFirstLocale(ctx.Locales()))
	}

	// Start with operand options and set unit style
	mergedOptions := mergeNumberOptions(numericOperand.Options, nil, ctx.LocaleMatcher())
	mergedOptions["style"] = "unit"

	// Process expression options
	for name, optval := range options {
		if optval == nil {
			continue
		}

		switch name {
		case "signDisplay", "roundingMode", "roundingPriority", "trailingZeroDisplay", "unit", "unitDisplay", "useGrouping":
			if strval, err := asString(optval); err == nil {
				mergedOptions[name] = strval
			} else {
				msg := fmt.Sprintf("Value %v is not valid for :unit option %s", optval, name)
				ctx.OnError(pkgErrors.NewBadOptionError(msg, source))
			}

		case "minimumIntegerDigits", "minimumFractionDigits", "maximumFractionDigits", "minimumSignificantDigits", "maximumSignificantDigits", "roundingIncrement":
			if intval, err := asPositiveInteger(optval); err == nil {
				mergedOptions[name] = intval
			} else {
				msg := fmt.Sprintf("Value %v is not valid for :unit option %s", optval, name)
				ctx.OnError(pkgErrors.NewBadOptionError(msg, source))
			}

		default:
			// Unknown option - silently ignore to match TypeScript behavior
		}
	}

	// Check that unit is provided
	if _, hasUnit := mergedOptions["unit"]; !hasUnit {
		msg := "A unit identifier is required for :unit"
		ctx.OnError(pkgErrors.NewBadOperandError(msg, source))
		return messagevalue.NewFallbackValue(source, getFirstLocale(ctx.Locales()))
	}

	return getMessageNumber(ctx, numericOperand.Value, mergedOptions, false)
}
