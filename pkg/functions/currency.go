package functions

import (
	"github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// CurrencyFunction implements the :currency function (DRAFT)
// currency accepts as input numerical values as well as
// objects wrapping a numerical value that also include a currency property.
//
// TypeScript original code:
// export function currency(
//
//	ctx: MessageFunctionContext,
//	exprOpt: Record<string | symbol, unknown>,
//	operand?: unknown
//
//	): MessageNumber {
//	  const { source } = ctx;
//	  const input = readNumericOperand(operand, source);
//	  const options: MessageNumberOptions = Object.assign({}, input.options, {
//	    localeMatcher: ctx.localeMatcher,
//	    style: 'currency'
//	  } as const);
//
//	  for (const [name, optval] of Object.entries(exprOpt)) {
//	    if (optval === undefined) continue;
//	    try {
//	      switch (name) {
//	        case 'currency':
//	        case 'currencySign':
//	        case 'roundingMode':
//	        case 'roundingPriority':
//	        case 'trailingZeroDisplay':
//	        case 'useGrouping':
//	          options[name] = asString(optval);
//	          break;
//	        case 'minimumIntegerDigits':
//	        case 'minimumSignificantDigits':
//	        case 'maximumSignificantDigits':
//	        case 'roundingIncrement':
//	          options[name] = asPositiveInteger(optval);
//	          break;
//	        case 'currencyDisplay': {
//	          const strval = asString(optval);
//	          if (strval === 'never') {
//	            ctx.onError(
//	              new MessageResolutionError(
//	                'unsupported-operation',
//	                'Currency display "never" is not yet supported',
//	                source
//	              )
//	            );
//	          } else {
//	            options[name] = strval;
//	          }
//	          break;
//	        }
//	        case 'fractionDigits': {
//	          const strval = asString(optval);
//	          if (strval === 'auto') {
//	            options.minimumFractionDigits = undefined;
//	            options.maximumFractionDigits = undefined;
//	          } else {
//	            const numval = asPositiveInteger(strval);
//	            options.minimumFractionDigits = numval;
//	            options.maximumFractionDigits = numval;
//	          }
//	          break;
//	        }
//	      }
//	    } catch (error) {
//	      if (error instanceof MessageError) {
//	        ctx.onError(error);
//	      } else {
//	        const msg = `Value ${optval} is not valid for :currency option ${name}`;
//	        ctx.onError(new MessageResolutionError('bad-option', msg, source));
//	      }
//	    }
//	  }
//
//	  if (!options.currency) {
//	    const msg = 'A currency code is required for :currency';
//	    throw new MessageResolutionError('bad-operand', msg, source);
//	  }
//
//	  return getMessageNumber(ctx, input.value, options, false);
//	}
func CurrencyFunction(
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

	// Start with operand options and set currency style
	mergedOptions := make(map[string]interface{})
	if numericOperand.Options != nil {
		for k, v := range numericOperand.Options {
			mergedOptions[k] = v
		}
	}
	mergedOptions["localeMatcher"] = ctx.LocaleMatcher()
	mergedOptions["style"] = "currency"

	// Process expression options
	for name, optval := range options {
		if optval == nil {
			continue
		}

		switch name {
		case "currency", "currencySign", "roundingMode", "roundingPriority", "trailingZeroDisplay", "useGrouping":
			if strval, err := asString(optval); err == nil {
				mergedOptions[name] = strval
			} else {
				msg := "Value " + toString(optval) + " is not valid for :currency option " + name
				ctx.OnError(errors.NewBadOptionError(msg, source))
			}

		case "minimumIntegerDigits", "minimumSignificantDigits", "maximumSignificantDigits", "roundingIncrement":
			if intval, err := asPositiveInteger(optval); err == nil {
				mergedOptions[name] = intval
			} else {
				msg := "Value " + toString(optval) + " is not valid for :currency option " + name
				ctx.OnError(errors.NewBadOptionError(msg, source))
			}

		case "currencyDisplay":
			if strval, err := asString(optval); err == nil {
				if strval == "never" {
					ctx.OnError(errors.NewMessageResolutionError(errors.ErrorTypeUnsupportedOperation, "Currency display \"never\" is not yet supported", source))
				} else {
					mergedOptions[name] = strval
				}
			} else {
				msg := "Value " + toString(optval) + " is not valid for :currency option " + name
				ctx.OnError(errors.NewBadOptionError(msg, source))
			}

		case "fractionDigits":
			if strval, err := asString(optval); err == nil {
				if strval == "auto" {
					// Remove fraction digit constraints for auto
					delete(mergedOptions, "minimumFractionDigits")
					delete(mergedOptions, "maximumFractionDigits")
				} else {
					if numval, err := asPositiveInteger(strval); err == nil {
						mergedOptions["minimumFractionDigits"] = numval
						mergedOptions["maximumFractionDigits"] = numval
					} else {
						msg := "Value " + toString(optval) + " is not valid for :currency option " + name
						ctx.OnError(errors.NewBadOptionError(msg, source))
					}
				}
			} else {
				msg := "Value " + toString(optval) + " is not valid for :currency option " + name
				ctx.OnError(errors.NewBadOptionError(msg, source))
			}
		}
	}

	// Check that currency is provided
	if _, hasCurrency := mergedOptions["currency"]; !hasCurrency {
		ctx.OnError(errors.NewBadOperandError("A currency code is required for :currency", source))
		return messagevalue.NewFallbackValue(source, getFirstLocale(ctx.Locales()))
	}

	return getMessageNumber(ctx, numericOperand.Value, mergedOptions, false)
}

// toString converts a value to string for error messages
func toString(value interface{}) string {
	if str, ok := value.(string); ok {
		return str
	}
	return "unknown"
}
