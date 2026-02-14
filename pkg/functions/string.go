package functions

import (
	"fmt"

	"github.com/kaptinlin/messageformat-go/pkg/bidi"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// StringFunction implements the :string function
// TypeScript original code:
// export function string(
//
//	ctx: Pick<MessageFunctionContext, 'dir' | 'locales' | 'source'>,
//	_options: Record<string, unknown>,
//	operand?: unknown
//
//	): MessageString {
//	  const str = operand === undefined ? '' : String(operand);
//	  const selStr = str.normalize();
//	  return {
//	    type: 'string',
//	    source: ctx.source,
//	    dir: ctx.dir ?? 'auto',
//	    selectKey: keys => (keys.has(selStr) ? selStr : null),
//	    toParts() {
//	      const { dir } = ctx;
//	      const locale = ctx.locales[0];
//	      return dir === 'ltr' || dir === 'rtl'
//	        ? [{ type: 'string', dir, locale, value: str }]
//	        : [{ type: 'string', locale, value: str }];
//	    },
//	    toString: () => str,
//	    valueOf: () => str
//	  };
//	}
func StringFunction(
	ctx MessageFunctionContext,
	options map[string]any,
	operand any,
) messagevalue.MessageValue {
	// Convert input to string
	var stringValue string
	if operand == nil {
		stringValue = ""
	} else {
		// Handle MessageValue operands - get their string representation
		if mv, ok := operand.(messagevalue.MessageValue); ok {
			str, err := mv.ToString()
			if err != nil {
				// If ToString fails, fall back to basic formatting
				stringValue = fmt.Sprintf("%v", operand)
			} else {
				stringValue = str
			}
		} else {
			stringValue = fmt.Sprintf("%v", operand)
		}
	}

	// Get locale from context or options
	locale := getFirstLocale(ctx.Locales())
	if localeOpt, ok := options["locale"]; ok {
		if localeStr, ok := localeOpt.(string); ok {
			locale = localeStr
		}
	}

	// Get direction from context and convert to bidi.Direction
	var dir bidi.Direction
	ctxDir := ctx.Dir()
	switch ctxDir {
	case "ltr":
		dir = bidi.DirLTR
	case "rtl":
		dir = bidi.DirRTL
	default:
		dir = bidi.DirAuto
	}

	return messagevalue.NewStringValueWithDir(stringValue, locale, ctx.Source(), dir)
}
