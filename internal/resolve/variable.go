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
	Scope      map[string]interface{}
}

// NewUnresolvedExpression creates a new unresolved expression
// TypeScript original code: UnresolvedExpression constructor
func NewUnresolvedExpression(
	expression *datamodel.Expression,
	scope map[string]interface{},
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
func isScope(scope interface{}) bool {
	if scope == nil {
		return false
	}

	v := reflect.ValueOf(scope)
	switch v.Kind() {
	case reflect.Map, reflect.Struct, reflect.Ptr, reflect.Func:
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
func getValue(scope interface{}, name string) interface{} {
	if !isScope(scope) {
		return nil
	}

	// Handle map types
	if m, ok := scope.(map[string]interface{}); ok {
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

		// Normalized key lookup - matches TypeScript: key.normalize() === name
		// Note: Go doesn't have string.normalize(), using ToLower as approximation
		for key, value := range m {
			if strings.ToLower(key) == strings.ToLower(name) {
				return value
			}
		}
	}

	// Handle map[interface{}]interface{} types
	if m, ok := scope.(map[interface{}]interface{}); ok {
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
func lookupVariableRef(ctx *Context, ref *datamodel.VariableRef) interface{} {
	name := ref.Name()
	value := getValue(ctx.Scope, name)

	if value == nil {
		source := "$" + name
		msg := fmt.Sprintf("Variable not available: %s", source)
		if ctx.OnError != nil {
			ctx.OnError(errors.NewResolutionError(
				errors.ErrorTypeUnresolvedVar,
				msg,
				source,
			))
		}
		return nil
	}

	// Handle unresolved expressions - matches TypeScript: value instanceof UnresolvedExpression
	if unresolvedExpr, ok := value.(*UnresolvedExpression); ok {
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
			return numberFunc(msgCtx, make(map[string]interface{}), value)
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
			return stringFunc(msgCtx, make(map[string]interface{}), value)
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
func getValueType(value interface{}) string {
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
	case func(...interface{}) interface{}:
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
