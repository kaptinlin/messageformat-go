// Package resolve provides variable resolution functions for MessageFormat 2.0
// TypeScript original code: resolve/resolve-variable.ts module
package resolve

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	"github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// UnresolvedExpression represents an unresolved expression
// TypeScript original code:
//
//	export class UnresolvedExpression {
//	  expression: Expression;
//	  scope: Record<string, unknown> | undefined;
//	  constructor(expression: Expression, scope?: Record<string, unknown>) {
//	    this.expression = expression;
//	    this.scope = scope;
//	  }
//	}
type UnresolvedExpression struct {
	Expression *datamodel.Expression
	Scope      map[string]any
}

// NewUnresolvedExpression creates a new unresolved expression
// TypeScript original code: UnresolvedExpression constructor
func NewUnresolvedExpression(
	expression *datamodel.Expression,
	scope map[string]any,
) *UnresolvedExpression {
	return &UnresolvedExpression{
		Expression: expression,
		Scope:      scope,
	}
}

// isScope checks if a value is a scope-like object
// TypeScript original code:
// const isScope = (scope: unknown): scope is Record<string, unknown> =>
//
//	scope !== null && (typeof scope === 'object' || typeof scope === 'function');
func isScope(scope any) bool {
	if scope == nil {
		return false
	}
	switch reflect.ValueOf(scope).Kind() {
	case reflect.Map, reflect.Struct, reflect.Pointer, reflect.Func:
		return true
	default:
		return false
	}
}

// getValue looks for the longest matching `.` delimited starting substring of name
// TypeScript original code:
//
//	function getValue(scope: unknown, name: string): unknown {
//	  if (isScope(scope)) {
//	    if (name in scope) return scope[name];
//
//	    const parts = name.split('.');
//	    for (let i = parts.length - 1; i > 0; --i) {
//	      const head = parts.slice(0, i).join('.');
//	      if (head in scope) {
//	        const tail = parts.slice(i).join('.');
//	        return getValue(scope[head], tail);
//	      }
//	    }
//
//	    for (const [key, value] of Object.entries(scope)) {
//	      if (key.normalize() === name) return value;
//	    }
//	  }
//
//	  return undefined;
//	}
func getValue(scope any, name string) any {
	if !isScope(scope) {
		return nil
	}

	// Handle map types
	if m, ok := scope.(map[string]any); ok {
		// Direct lookup - matches TypeScript: if (name in scope) return scope[name];
		if value, exists := m[name]; exists {
			return value
		}

		// Dotted property access - matches TypeScript parts logic
		if strings.Contains(name, ".") {
			parts := strings.Split(name, ".")
			for i := len(parts) - 1; i > 0; i-- {
				head := strings.Join(parts[:i], ".")
				if headValue, exists := m[head]; exists {
					tail := strings.Join(parts[i:], ".")
					return getValue(headValue, tail)
				}
			}
		}

	}

	// Handle map[interface{}]interface{} types
	if m, ok := scope.(map[any]any); ok {
		if value, exists := m[name]; exists {
			return value
		}
	}

	return nil
}

// lookupVariableRef looks up a variable reference and resolves it
// TypeScript original code:
//
//	export function lookupVariableRef(ctx: Context, { name }: VariableRef) {
//	  const value = getValue(ctx.scope, name);
//	  if (value === undefined) {
//	    const source = '$' + name;
//	    const msg = `Variable not available: ${source}`;
//	    ctx.onError(new MessageResolutionError('unresolved-variable', msg, source));
//	  } else if (value instanceof UnresolvedExpression) {
//	    const local = resolveExpression(
//	      value.scope ? { ...ctx, scope: value.scope } : ctx,
//	      value.expression
//	    );
//	    ctx.scope[name] = local;
//	    ctx.localVars.add(local);
//	    return local;
//	  }
//	  return value;
//	}
func lookupVariableRef(ctx *Context, ref *datamodel.VariableRef) any {
	name := ref.Name()
	value := getValue(ctx.Scope, name)

	if value == nil {
		source := "$" + name
		msg := fmt.Sprintf("variable not available: %s", source)
		if ctx.OnError != nil {
			ctx.OnError(errors.NewMessageResolutionError(
				errors.ErrorTypeUnresolvedVariable,
				msg,
				source,
			))
		}
		return nil
	}

	// Handle unresolved expressions - matches TypeScript: value instanceof UnresolvedExpression
	if unresolvedExpr, ok := value.(*UnresolvedExpression); ok {
		// Check for .input declarations first - these are special cases where we should use the original parameter value
		// But only for simple variable references without functions
		if unresolvedExpr.Scope != nil && unresolvedExpr.Expression.Arg() != nil && unresolvedExpr.Expression.FunctionRef() == nil {
			if varRef, ok := unresolvedExpr.Expression.Arg().(*datamodel.VariableRef); ok {
				varRefName := varRef.Name()
				// Check if this is an .input declaration by looking for the original parameter in the scope
				if originalValue, exists := unresolvedExpr.Scope[varRefName]; exists {
					if _, isUnresolved := originalValue.(*UnresolvedExpression); !isUnresolved {
						// This is an .input declaration without function - return the original parameter value directly
						// This avoids infinite recursion in Unicode normalization cases
						return originalValue
					}
				}

				// Also check for Unicode normalization cases - if the variable names normalize to the same value
				// we should look for any non-unresolved value in the scope
				// But only if there's exactly one non-unresolved value (to avoid ambiguity)
				var nonUnresolvedValue any
				var count int
				for _, scopeValue := range unresolvedExpr.Scope {
					if _, isUnresolved := scopeValue.(*UnresolvedExpression); !isUnresolved {
						nonUnresolvedValue = scopeValue
						count++
					}
				}
				if count == 1 {
					// This could be the original parameter value for a Unicode normalization case
					// Return it directly to avoid infinite recursion
					return nonUnresolvedValue
				}
			}
		}

		// Check for circular reference by looking if we're already resolving this variable
		if ctx.ResolvingVars == nil {
			ctx.ResolvingVars = make(map[string]bool)
		}

		if ctx.ResolvingVars[name] {
			// Circular reference detected - return fallback
			source := "$" + name
			if ctx.OnError != nil {
				ctx.OnError(errors.NewMessageResolutionError(
					errors.ErrorTypeUnresolvedVariable,
					fmt.Sprintf("circular reference detected for variable: %s", source),
					source,
				))
			}
			return messagevalue.NewFallbackValue(source, getFirstLocale(ctx.Locales))
		}

		// Mark this variable as being resolved
		ctx.ResolvingVars[name] = true
		defer func() {
			delete(ctx.ResolvingVars, name)
		}()

		// Create new context with expression scope - matches TypeScript: value.scope ? { ...ctx, scope: value.scope } : ctx
		var newCtx *Context
		if unresolvedExpr.Scope != nil {
			newCtx = ctx.CloneWithScope(unresolvedExpr.Scope)
		} else {
			newCtx = ctx
		}

		// Resolve the expression
		local := ResolveExpression(newCtx, unresolvedExpr.Expression)

		// Cache the resolved value - matches TypeScript: ctx.scope[name] = local; ctx.localVars.add(local);
		ctx.Scope[name] = local
		ctx.LocalVars[local] = true

		return local
	}

	return value
}

// ResolveVariableRef resolves a variable reference to a MessageValue
// TypeScript original code:
// export function resolveVariableRef<T extends string, P extends string>(
//
//	ctx: Context<T, P>,
//	ref: VariableRef
//
//	) {
//	  const source = '$' + ref.name;
//	  const value = lookupVariableRef(ctx, ref);
//
//	  let type = typeof value;
//	  if (type === 'object') {
//	    const mv = value as MessageValue<T, P>;
//	    if (mv.type === 'fallback') return fallback(source);
//	    if (ctx.localVars.has(mv)) return mv;
//	    if (value instanceof Number) type = 'number';
//	    else if (value instanceof String) type = 'string';
//	  }
//
//	  switch (type) {
//	    case 'bigint':
//	    case 'number': {
//	      const msgCtx = new MessageFunctionContext(ctx, source);
//	      return ctx.functions.number(msgCtx, {}, value);
//	    }
//	    case 'string': {
//	      const msgCtx = new MessageFunctionContext(ctx, source);
//	      return ctx.functions.string(msgCtx, {}, value);
//	    }
//	  }
//
//	  return value === undefined ? fallback(source) : unknown(source, value);
//	}
func ResolveVariableRef(ctx *Context, ref *datamodel.VariableRef) messagevalue.MessageValue {
	source := "$" + ref.Name()
	value := lookupVariableRef(ctx, ref)

	// Determine type - matches TypeScript: let type = typeof value;
	valueType := getValueType(value)

	// Handle object types - matches TypeScript: if (type === 'object')
	if valueType == "object" {
		// Check if it's already a MessageValue - matches TypeScript: const mv = value as MessageValue<T, P>;
		if mv, ok := value.(messagevalue.MessageValue); ok {
			// Check for fallback type - matches TypeScript: if (mv.type === 'fallback') return fallback(source);
			if mv.Type() == "fallback" {
				return messagevalue.NewFallbackValue(source, getFirstLocale(ctx.Locales))
			}
			// Check if it's a local variable - matches TypeScript: if (ctx.localVars.has(mv)) return mv;
			if ctx.LocalVars[mv] {
				return mv
			}
		}

		// Handle Number and String objects - matches TypeScript: if (value instanceof Number) type = 'number';
		switch value.(type) {
		case *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64, *float32, *float64:
			valueType = "number"
		case *string:
			valueType = "string"
		}
	}

	// Switch on type - matches TypeScript switch statement
	switch valueType {
	case "bigint", "number":
		// matches TypeScript: const msgCtx = new MessageFunctionContext(ctx, source);
		if numberFunc, exists := ctx.Functions["number"]; exists {
			msgCtx := functions.NewMessageFunctionContext(
				ctx.Locales,
				source,
				ctx.LocaleMatcher,
				ctx.OnError,
				nil,
				"",
				"",
			)
			// matches TypeScript: return ctx.functions.number(msgCtx, {}, value);
			return numberFunc(msgCtx, make(map[string]any), value)
		}
	case "string":
		// matches TypeScript: const msgCtx = new MessageFunctionContext(ctx, source);
		if stringFunc, exists := ctx.Functions["string"]; exists {
			msgCtx := functions.NewMessageFunctionContext(
				ctx.Locales,
				source,
				ctx.LocaleMatcher,
				ctx.OnError,
				nil,
				"",
				"",
			)
			// matches TypeScript: return ctx.functions.string(msgCtx, {}, value);
			return stringFunc(msgCtx, make(map[string]any), value)
		}
	}

	// matches TypeScript: return value === undefined ? fallback(source) : unknown(source, value);
	if value == nil {
		return messagevalue.NewFallbackValue(source, getFirstLocale(ctx.Locales))
	}

	// For unknown types, create a string representation (equivalent to TypeScript unknown function)
	// TypeScript unknown function typically converts to string representation
	return messagevalue.NewStringValue(fmt.Sprintf("%v", value), getFirstLocale(ctx.Locales), source)
}

// getValueType determines the type of a value similar to TypeScript typeof
// TypeScript original code: typeof value
func getValueType(value any) string {
	if value == nil {
		return "undefined"
	}

	switch value.(type) {
	case bool:
		return "boolean"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return "number"
	case string:
		return "string"
	case func(...any) any:
		return "function"
	default:
		return "object"
	}
}

// getFirstLocale returns the first locale from a list, or "en" as fallback
// TypeScript original code: locale fallback logic
func getFirstLocale(locales []string) string {
	if len(locales) > 0 {
		return locales[0]
	}
	return "en"
}
