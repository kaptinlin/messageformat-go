package functions

import (
	"maps"

	"github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// CurrencyFunction implements the :currency function for currency value formatting.
//
// Status: Stable (finalized in LDML 48)
// Specification: https://unicode.org/reports/tr35/tr35-messageFormat.html#currency
//
// The :currency function formats numeric values as currency. It requires either:
// - An operand containing both a numeric value and currency code
// - A numeric operand with a currency option
//
// Example:
//
//	{$amount :currency currency=USD}
//	{$price :currency currency=EUR fractionDigits=2}
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
	options map[string]any,
	operand any,
) messagevalue.MessageValue {
	source := ctx.Source()

	numericOperand, err := readNumericOperand(operand, source)
	if err != nil {
		ctx.OnError(err)
		return messagevalue.NewFallbackValue(source, GetFirstLocale(ctx.Locales()))
	}

	mergedOptions := make(map[string]any)
	maps.Copy(mergedOptions, numericOperand.Options)
	if existingStyle, ok := numericOperand.Options["style"]; ok && existingStyle == "percent" {
		ctx.OnError(errors.NewBadOperandError("Cannot format a percent-formatted number as currency", source))
		return messagevalue.NewFallbackValue(source, GetFirstLocale(ctx.Locales()))
	}
	mergedOptions["localeMatcher"] = ctx.LocaleMatcher()
	mergedOptions["style"] = "currency"

	badOptionError := func(name string, value any) {
		msg := "Value " + toString(value) + " is not valid for :currency option " + name
		ctx.OnError(errors.NewBadOptionError(msg, source))
	}

	for name, optval := range options {
		if optval == nil {
			continue
		}

		switch name {
		case "currency", "currencySign", "roundingMode", "roundingPriority", "trailingZeroDisplay", "useGrouping":
			if strval, err := asString(optval); err == nil {
				mergedOptions[name] = strval
			} else {
				badOptionError(name, optval)
			}

		case "minimumIntegerDigits", "minimumSignificantDigits", "maximumSignificantDigits", "roundingIncrement":
			if intval, err := asPositiveInteger(optval); err == nil {
				mergedOptions[name] = intval
			} else {
				badOptionError(name, optval)
			}

		case "currencyDisplay":
			if strval, err := asString(optval); err == nil {
				if strval == "never" {
					ctx.OnError(errors.NewMessageResolutionError(errors.ErrorTypeUnsupportedOperation, "Currency display \"never\" is not yet supported", source))
				} else {
					mergedOptions[name] = strval
				}
			} else {
				badOptionError(name, optval)
			}

		case "fractionDigits":
			if strval, err := asString(optval); err == nil {
				if strval == "auto" {
					delete(mergedOptions, "minimumFractionDigits")
					delete(mergedOptions, "maximumFractionDigits")
				} else if numval, err := asPositiveInteger(strval); err == nil {
					mergedOptions["minimumFractionDigits"] = numval
					mergedOptions["maximumFractionDigits"] = numval
				} else {
					badOptionError(name, optval)
				}
			} else {
				badOptionError(name, optval)
			}
		}
	}

	if _, hasCurrency := mergedOptions["currency"]; !hasCurrency {
		ctx.OnError(errors.NewBadOperandError("A currency code is required for :currency", source))
		return messagevalue.NewFallbackValue(source, GetFirstLocale(ctx.Locales()))
	}

	return getMessageNumber(ctx, numericOperand.Value, mergedOptions, false)
}

// toString converts a value to string for error messages
func toString(value any) string {
	if str, ok := value.(string); ok {
		return str
	}
	return "unknown"
}
