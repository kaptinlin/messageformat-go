// Package resolve provides markup formatting functions for MessageFormat 2.0
// TypeScript original code: resolve/format-markup.ts module
package resolve

import (
	"fmt"

	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	"github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/kaptinlin/messageformat-go/pkg/logger"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// FormatMarkup formats a markup element to a MessagePart
// TypeScript original code:
// export function formatMarkup(
//
//	ctx: Context,
//	{ kind, name, options }: Markup
//
//	): MessageMarkupPart {
//	  const part: MessageMarkupPart = { type: 'markup', kind, name };
//	  if (options?.size) {
//	    part.options = {};
//	    for (const [name, value] of options) {
//	      if (name === 'u:dir') {
//	        const msg = `The option ${name} is not valid for markup`;
//	        const optSource = getValueSource(value);
//	        ctx.onError(new MessageResolutionError('bad-option', msg, optSource));
//	      } else {
//	        let rv = resolveValue(ctx, value);
//	        if (typeof rv === 'object' && typeof rv?.valueOf === 'function') {
//	          rv = rv.valueOf();
//	        }
//	        if (name === 'u:id') part.id = String(rv);
//	        else part.options[name] = rv;
//	      }
//	    }
//	  }
//	  return part;
//	}
func FormatMarkup(ctx *Context, markup *datamodel.Markup) messagevalue.MessagePart {
	part := messagevalue.NewMarkupPart(
		markup.Kind(),
		markup.Name(),
		"", // source will be set if needed
		make(map[string]any),
	)

	options := markup.Options()
	if len(options) > 0 {
		partOptions := make(map[string]any)

		for name, value := range options {
			if name == "u:dir" {
				msg := fmt.Sprintf("option %s is not valid for markup", name)
				var optSource string
				if node, ok := value.(datamodel.Node); ok {
					optSource = getValueSource(node)
				} else {
					optSource = fmt.Sprintf("%v", value)
				}
				if ctx.OnError != nil {
					ctx.OnError(errors.NewMessageResolutionError(
						errors.ErrorTypeBadOption,
						msg,
						optSource,
					))
				}
				continue
			}

			var rv any
			if node, ok := value.(datamodel.Node); ok {
				var err error
				rv, err = resolveValue(ctx, node)
				if err != nil {
					logger.Error("failed to resolve value in markup", "error", err)
					if ctx.OnError != nil {
						ctx.OnError(errors.NewMessageResolutionError(
							errors.ErrorTypeUnsupportedOperation,
							err.Error(),
							getValueSource(node),
						))
					}
					rv = nil
				}
			} else {
				rv = value
			}

			if mv, ok := rv.(messagevalue.MessageValue); ok {
				if valueOf, err := mv.ValueOf(); err == nil && valueOf != nil {
					rv = valueOf
				}
			}

			partOptions[name] = rv
		}

		part = messagevalue.NewMarkupPart(
			markup.Kind(),
			markup.Name(),
			"", // source
			partOptions,
		)
	}

	return part
}
