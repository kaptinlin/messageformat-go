// Package resolve provides function reference resolution for MessageFormat 2.0
// TypeScript original code: resolve/resolve-function-ref.ts module
package resolve

import (
	"fmt"

	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	"github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/logger"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// ResolveFunctionRef resolves a function reference expression
// TypeScript original code:
// export function resolveFunctionRef<T extends string, P extends string>(
//
//	ctx: Context<T, P>,
//	operand: Literal | VariableRef | undefined,
//	{ name, options }: FunctionRef
//
//	): MessageValue<T, P> | MessageFallback {
//	  const source = getValueSource(operand) ?? `:${name}`;
//	  try {
//	    const fnInput = operand ? [resolveValue(ctx, operand)] : [];
//	    const rf = ctx.functions[name];
//	    if (!rf) {
//	      throw new MessageError('unknown-function', `Unknown function :${name}`);
//	    }
//	    const msgCtx = new MessageFunctionContext(ctx, source, options);
//	    const opt = resolveOptions(ctx, options);
//	    let res = rf(msgCtx, opt, ...fnInput);
//	    if (
//	      res === null ||
//	      (typeof res !== 'object' && typeof res !== 'function') ||
//	      typeof res.type !== 'string' ||
//	      typeof res.source !== 'string'
//	    ) {
//	      throw new MessageError(
//	        'bad-function-result',
//	        `Function :${name} did not return a MessageValue`
//	      );
//	    }
//	    if (msgCtx.dir) res = { ...res, dir: msgCtx.dir, [BIDI_ISOLATE]: true };
//	    if (msgCtx.id && typeof res.toParts === 'function') {
//	      return {
//	        ...res,
//	        toParts() {
//	          const parts = res.toParts!();
//	          for (const part of parts) part.id = msgCtx.id;
//	          return parts;
//	        }
//	      };
//	    }
//	    return res;
//	  } catch (error) {
//	    ctx.onError(error);
//	    return fallback(source);
//	  }
//	}
func ResolveFunctionRef(
	ctx *Context,
	operand datamodel.Node,
	functionRef *datamodel.FunctionRef,
) messagevalue.MessageValue {
	// matches TypeScript: const source = getValueSource(operand) ?? `:${name}`;
	source := getValueSource(operand)
	if source == "" {
		source = ":" + functionRef.Name()
	}

	// matches TypeScript: try { ... } catch (error) { ctx.onError(error); return fallback(source); }
	defer func() {
		if r := recover(); r != nil {
			if ctx.OnError != nil {
				if err, ok := r.(error); ok {
					ctx.OnError(err)
				} else {
					ctx.OnError(fmt.Errorf("%v", r))
				}
			}
		}
	}()

	// matches TypeScript: const fnInput = operand ? [resolveValue(ctx, operand)] : [];
	var fnInput []interface{}
	if operand != nil {
		fnInput = []interface{}{resolveValue(ctx, operand)}
	} else {
		fnInput = []interface{}{}
	}

	// matches TypeScript: const rf = ctx.functions[name];
	rf, exists := ctx.Functions[functionRef.Name()]
	// matches TypeScript: if (!rf) { throw new MessageError('unknown-function', `Unknown function :${name}`); }
	if !exists {
		logger.Error("unknown function", "function", functionRef.Name(), "source", source)
		panic(errors.NewResolutionError(
			errors.ErrorTypeUnknownFunction,
			fmt.Sprintf("Unknown function :%s", functionRef.Name()),
			source,
		))
	}

	// matches TypeScript: const msgCtx = new MessageFunctionContext(ctx, source, options);
	msgCtx := createMessageFunctionContext(ctx, source, convertOptionsToMap(functionRef.Options()))

	// matches TypeScript: const opt = resolveOptions(ctx, options);
	opt := resolveOptions(ctx, convertOptionsToMap(functionRef.Options()))

	// matches TypeScript: let res = rf(msgCtx, opt, ...fnInput);
	var res messagevalue.MessageValue
	if len(fnInput) > 0 {
		res = rf(msgCtx, opt, fnInput[0])
	} else {
		res = rf(msgCtx, opt, nil)
	}

	// matches TypeScript: if (res === null || ...) { throw new MessageError('bad-function-result', ...); }
	if res == nil {
		logger.Error("function returned nil result", "function", functionRef.Name(), "source", source)
		panic(errors.NewResolutionError(
			errors.ErrorTypeBadFunctionResult,
			fmt.Sprintf("Function :%s did not return a MessageValue", functionRef.Name()),
			source,
		))
	}

	// TODO: Handle bidi isolation and ID setting like TypeScript
	// matches TypeScript: if (msgCtx.dir) res = { ...res, dir: msgCtx.dir, [BIDI_ISOLATE]: true };
	// matches TypeScript: if (msgCtx.id && typeof res.toParts === 'function') { ... }

	// matches TypeScript: return res;
	return res
}

// createMessageFunctionContext creates a MessageFunctionContext with options
func createMessageFunctionContext(
	ctx *Context,
	source string,
	options map[string]interface{},
) functions.MessageFunctionContext {
	var dir string
	var id string
	literalKeys := make(map[string]bool)

	if options != nil {
		// Process universal options
		if dirOpt, exists := options["u:dir"]; exists {
			// Convert interface{} to datamodel.Node if possible
			if dirNode, ok := dirOpt.(datamodel.Node); ok {
				dirValue := resolveValue(ctx, dirNode)
				if dirStr, ok := dirValue.(string); ok {
					switch dirStr {
					case "ltr", "rtl", "auto":
						dir = dirStr
					case "inherit":
						// Use context default
					default:
						// Invalid direction - report error
						if ctx.OnError != nil {
							ctx.OnError(errors.NewResolutionError(
								errors.ErrorTypeBadOption,
								"Unsupported value for u:dir option",
								getValueSource(dirNode),
							))
						}
					}
				}

				// Mark as literal if it's a literal value
				if _, isLiteral := dirNode.(*datamodel.Literal); isLiteral {
					literalKeys["u:dir"] = true
				}
			} else if dirStr, ok := dirOpt.(string); ok {
				// Handle direct string values
				switch dirStr {
				case "ltr", "rtl", "auto":
					dir = dirStr
				}
			}
		}

		if idOpt, exists := options["u:id"]; exists {
			if idNode, ok := idOpt.(datamodel.Node); ok {
				idValue := resolveValue(ctx, idNode)
				id = fmt.Sprintf("%v", idValue)

				// Mark as literal if it's a literal value
				if _, isLiteral := idNode.(*datamodel.Literal); isLiteral {
					literalKeys["u:id"] = true
				}
			} else {
				id = fmt.Sprintf("%v", idOpt)
			}
		}

		// Mark all literal options
		for key, value := range options {
			if node, ok := value.(datamodel.Node); ok {
				if _, isLiteral := node.(*datamodel.Literal); isLiteral {
					literalKeys[key] = true
				}
			}
		}
	}

	return functions.NewMessageFunctionContext(
		ctx.Locales,
		source,
		ctx.LocaleMatcher,
		ctx.OnError,
		literalKeys,
		dir,
		id,
	)
}

// resolveOptions resolves function options
func resolveOptions(ctx *Context, options map[string]interface{}) map[string]interface{} {
	opt := make(map[string]interface{})

	if options == nil {
		return opt
	}

	for name, value := range options {
		// Skip universal options (they're handled by MessageFunctionContext)
		if !isUniversalOption(name) {
			var resolved interface{}

			// Try to resolve as datamodel.Node first
			if node, ok := value.(datamodel.Node); ok {
				resolved = resolveValue(ctx, node)
			} else {
				// Use value directly if not a Node
				resolved = value
			}

			// If resolved value is a MessageValue with valueOf, use that
			if mv, ok := resolved.(messagevalue.MessageValue); ok {
				if valueOf, err := mv.ValueOf(); err == nil && valueOf != nil {
					opt[name] = valueOf
				} else {
					opt[name] = resolved
				}
			} else {
				opt[name] = resolved
			}
		}
	}

	return opt
}

// isUniversalOption checks if an option is a universal option
func isUniversalOption(name string) bool {
	return len(name) > 2 && name[:2] == "u:"
}

// convertOptionsToMap converts FunctionRef options to a map[string]interface{}
func convertOptionsToMap(options datamodel.Options) map[string]interface{} {
	converted := make(map[string]interface{})

	if options != nil {
		for name, value := range options {
			converted[name] = value
		}
	}

	return converted
}
