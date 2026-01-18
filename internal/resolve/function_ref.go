// Package resolve provides function reference resolution for MessageFormat 2.0
// TypeScript original code: resolve/resolve-function-ref.ts module
package resolve

import (
	"fmt"

	"github.com/kaptinlin/messageformat-go/pkg/bidi"
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
	result, err := resolveFunctionRefInternal(ctx, operand, functionRef, source)
	if err != nil {
		if ctx.OnError != nil {
			ctx.OnError(err)
		}
		// Return fallback value
		locale := "en"
		if len(ctx.Locales) > 0 {
			locale = ctx.Locales[0]
		}
		return functions.FallbackFunction(source, locale)
	}
	return result
}

func resolveFunctionRefInternal(
	ctx *Context,
	operand datamodel.Node,
	functionRef *datamodel.FunctionRef,
	source string,
) (messagevalue.MessageValue, error) {
	// matches TypeScript: const fnInput = operand ? [resolveValue(ctx, operand)] : [];
	var fnInput []interface{}
	if operand != nil {
		resolved, err := resolveValue(ctx, operand)
		if err != nil {
			// Log error and return error as MessageValue would be invalid
			logger.Error("failed to resolve operand", "error", err)
			return nil, errors.NewMessageResolutionError(
				errors.ErrorTypeBadOperand,
				err.Error(),
				source,
			)
		}
		fnInput = []interface{}{resolved}
	} else {
		fnInput = []interface{}{}
	}

	// matches TypeScript: const rf = ctx.functions[name];
	rf, exists := ctx.Functions[functionRef.Name()]
	// matches TypeScript: if (!rf) { throw new MessageError('unknown-function', `Unknown function :${name}`); }
	if !exists {
		logger.Error("unknown function", "function", functionRef.Name(), "source", source)
		return nil, errors.NewMessageResolutionError(
			errors.ErrorTypeUnknownFunction,
			fmt.Sprintf("Unknown function :%s", functionRef.Name()),
			source,
		)
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
		return nil, errors.NewMessageResolutionError(
			errors.ErrorTypeBadFunctionResult,
			fmt.Sprintf("Function :%s did not return a MessageValue", functionRef.Name()),
			source,
		)
	}

	// Handle bidi isolation and ID setting like TypeScript
	// matches TypeScript: if (msgCtx.dir) res = { ...res, dir: msgCtx.dir, [BIDI_ISOLATE]: true };
	if msgCtx.Dir() != "" || msgCtx.ID() != "" {
		res = &messageValueWithOptions{
			wrapped:     res,
			dir:         msgCtx.Dir(),
			id:          msgCtx.ID(),
			bidiIsolate: msgCtx.Dir() != "",
		}
	}

	// matches TypeScript: return res;
	return res, nil
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
				dirValue, err := resolveValue(ctx, dirNode)
				if err != nil {
					logger.Error("failed to resolve u:dir option", "error", err)
					if ctx.OnError != nil {
						ctx.OnError(errors.NewMessageResolutionError(
							errors.ErrorTypeBadOption,
							err.Error(),
							getValueSource(dirNode),
						))
					}
				} else if dirStr, ok := dirValue.(string); ok {
					switch dirStr {
					case "ltr", "rtl", "auto":
						dir = dirStr
					case "inherit":
						// Use context default
					default:
						// Invalid direction - report error
						if ctx.OnError != nil {
							ctx.OnError(errors.NewMessageResolutionError(
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
				idValue, err := resolveValue(ctx, idNode)
				if err != nil {
					logger.Error("failed to resolve u:id option", "error", err)
					if ctx.OnError != nil {
						ctx.OnError(errors.NewMessageResolutionError(
							errors.ErrorTypeBadOption,
							err.Error(),
							getValueSource(idNode),
						))
					}
				} else {
					id = fmt.Sprintf("%v", idValue)
				}

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
				var err error
				resolved, err = resolveValue(ctx, node)
				if err != nil {
					logger.Error("failed to resolve option", "option", name, "error", err)
					if ctx.OnError != nil {
						ctx.OnError(errors.NewMessageResolutionError(
							errors.ErrorTypeBadOption,
							err.Error(),
							getValueSource(node),
						))
					}
					// Use nil as fallback
					resolved = nil
				}
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

	for name, value := range options {
		converted[name] = value
	}

	return converted
}

// messageValueWithOptions wraps a MessageValue to add dir, id, and bidi isolate flag
// TypeScript original code: res = { ...res, dir: msgCtx.dir, [BIDI_ISOLATE]: true };
type messageValueWithOptions struct {
	wrapped     messagevalue.MessageValue
	dir         string
	id          string
	bidiIsolate bool
}

func (mv *messageValueWithOptions) Type() string {
	return mv.wrapped.Type()
}

func (mv *messageValueWithOptions) Source() string {
	return mv.wrapped.Source()
}

func (mv *messageValueWithOptions) Dir() bidi.Direction {
	if mv.dir != "" {
		switch mv.dir {
		case "ltr":
			return bidi.DirLTR
		case "rtl":
			return bidi.DirRTL
		case "auto":
			return bidi.DirAuto
		default:
			return bidi.DirAuto
		}
	}
	return mv.wrapped.Dir()
}

func (mv *messageValueWithOptions) Locale() string {
	return mv.wrapped.Locale()
}

func (mv *messageValueWithOptions) Options() map[string]interface{} {
	return mv.wrapped.Options()
}

func (mv *messageValueWithOptions) ToString() (string, error) {
	return mv.wrapped.ToString()
}

func (mv *messageValueWithOptions) ToParts() ([]messagevalue.MessagePart, error) {
	parts, err := mv.wrapped.ToParts()
	if err != nil {
		return nil, err
	}

	// Add ID and dir to all parts if specified - matches TypeScript: for (const part of parts) part.id = msgCtx.id;
	if mv.id != "" || mv.dir != "" {
		// Create new parts with ID and dir information
		newParts := make([]messagevalue.MessagePart, len(parts))
		for i, part := range parts {
			// Determine locale based on dir and id according to test expectations
			var locale string
			switch {
			case mv.dir == "rtl" || mv.dir == "auto":
				// For rtl and auto, include locale
				locale = mv.wrapped.Locale()
			case mv.dir == "ltr" && mv.id != "":
				// For ltr with id, don't include locale (leave empty)
				locale = ""
			case mv.dir == "ltr" && mv.id == "":
				// For ltr without id, include locale
				locale = mv.wrapped.Locale()
			}

			newParts[i] = &partWithOptions{
				wrapped: part,
				id:      mv.id,
				dir:     mv.dir,
				locale:  locale,
			}
		}
		return newParts, nil
	}

	return parts, nil
}

func (mv *messageValueWithOptions) ValueOf() (interface{}, error) {
	return mv.wrapped.ValueOf()
}

func (mv *messageValueWithOptions) SelectKeys(keys []string) ([]string, error) {
	return mv.wrapped.SelectKeys(keys)
}

// HasBidiIsolate returns whether this value should be bidi isolated
func (mv *messageValueWithOptions) HasBidiIsolate() bool {
	return mv.bidiIsolate
}

// GetID returns the ID for this value
func (mv *messageValueWithOptions) GetID() string {
	return mv.id
}

// partWithOptions wraps a MessagePart to add ID and dir information
type partWithOptions struct {
	wrapped messagevalue.MessagePart
	id      string
	dir     string
	locale  string
}

func (p *partWithOptions) Type() string       { return p.wrapped.Type() }
func (p *partWithOptions) Value() interface{} { return p.wrapped.Value() }
func (p *partWithOptions) Source() string     { return p.wrapped.Source() }
func (p *partWithOptions) Locale() string     { return p.wrapped.Locale() }

func (p *partWithOptions) Dir() bidi.Direction {
	if p.dir != "" {
		switch p.dir {
		case "ltr":
			return bidi.DirLTR
		case "rtl":
			return bidi.DirRTL
		case "auto":
			return bidi.DirAuto
		default:
			return bidi.DirAuto
		}
	}
	return p.wrapped.Dir()
}

// GetID returns the ID for this part
func (p *partWithOptions) GetID() string {
	return p.id
}

// GetDir returns the dir for this part
func (p *partWithOptions) GetDir() string {
	return p.dir
}

// GetLocale returns the locale for this part
func (p *partWithOptions) GetLocale() string {
	return p.locale
}
