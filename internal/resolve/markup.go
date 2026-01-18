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
	// matches TypeScript: const part: MessageMarkupPart = { type: 'markup', kind, name };
	part := messagevalue.NewMarkupPart(
		markup.Kind(),
		markup.Name(),
		"", // source will be set if needed
		make(map[string]interface{}),
	)

	options := markup.Options()
	// matches TypeScript: if (options?.size)
	if len(options) > 0 {
		partOptions := make(map[string]interface{})

		// matches TypeScript: for (const [name, value] of options)
		for name, value := range options {
			// matches TypeScript: if (name === 'u:dir')
			if name == "u:dir" {
				// matches TypeScript: const msg = `The option ${name} is not valid for markup`;
				msg := fmt.Sprintf("The option %s is not valid for markup", name)
				// matches TypeScript: const optSource = getValueSource(value);
				var optSource string
				if node, ok := value.(datamodel.Node); ok {
					optSource = getValueSource(node)
				} else {
					optSource = fmt.Sprintf("%v", value)
				}
				// matches TypeScript: ctx.onError(new MessageResolutionError('bad-option', msg, optSource));
				if ctx.OnError != nil {
					ctx.OnError(errors.NewMessageResolutionError(
						errors.ErrorTypeBadOption,
						msg,
						optSource,
					))
				}
			} else {
				var rv interface{}

				// matches TypeScript: let rv = resolveValue(ctx, value);
				if node, ok := value.(datamodel.Node); ok {
					var err error
					rv, err = resolveValue(ctx, node)
					if err != nil {
						// Log error and use nil as fallback value
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

				// matches TypeScript: if (typeof rv === 'object' && typeof rv?.valueOf === 'function') { rv = rv.valueOf(); }
				if mv, ok := rv.(messagevalue.MessageValue); ok {
					if valueOf, err := mv.ValueOf(); err == nil && valueOf != nil {
						rv = valueOf
					}
				}

				// matches TypeScript: if (name === 'u:id') part.id = String(rv); else part.options[name] = rv;
				partOptions[name] = rv
			}
		}

		// Create new markup part with resolved options
		part = messagevalue.NewMarkupPart(
			markup.Kind(),
			markup.Name(),
			"", // source
			partOptions,
		)
	}

	// matches TypeScript: return part;
	return part
}
