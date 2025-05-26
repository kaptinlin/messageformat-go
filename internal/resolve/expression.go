// Package resolve provides expression resolution functions for MessageFormat 2.0
// TypeScript original code: resolve/resolve-expression.ts module
package resolve

import (
	"fmt"

	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// ResolveExpression resolves an expression to a MessageValue
// TypeScript original code:
// export function resolveExpression<T extends string, P extends string>(
//
//	ctx: Context<T, P>,
//	{ arg, functionRef }: Expression
//
//	): MessageFallback | MessageString | MessageUnknownValue | MessageValue<T, P> {
//	  if (functionRef) {
//	    return resolveFunctionRef(ctx, arg, functionRef);
//	  }
//	  switch (arg?.type) {
//	    case 'literal':
//	      return resolveLiteral(ctx, arg);
//	    case 'variable':
//	      return resolveVariableRef(ctx, arg);
//	    default:
//	      // @ts-expect-error - should never happen
//	      throw new Error(`Unsupported expression: ${arg?.type}`);
//	  }
//	}
func ResolveExpression(ctx *Context, expr *datamodel.Expression) messagevalue.MessageValue {
	if expr == nil {
		// Should not happen in well-formed messages
		return messagevalue.NewFallbackValue("unknown", getFirstLocale(ctx.Locales))
	}

	// Check if expression has a function reference - matches TypeScript: if (functionRef)
	if functionRef := expr.FunctionRef(); functionRef != nil {
		// Convert interface{} to datamodel.Node for operand
		var operand datamodel.Node
		if arg := expr.Arg(); arg != nil {
			if node, ok := arg.(datamodel.Node); ok {
				operand = node
			}
		}
		// matches TypeScript: return resolveFunctionRef(ctx, arg, functionRef);
		return ResolveFunctionRef(ctx, operand, functionRef)
	}

	// Handle operand-only expressions - matches TypeScript: switch (arg?.type)
	arg := expr.Arg()
	if arg == nil {
		// Should not happen in well-formed expressions
		return messagevalue.NewFallbackValue("unknown", getFirstLocale(ctx.Locales))
	}

	switch v := arg.(type) {
	case *datamodel.Literal:
		// matches TypeScript: case 'literal': return resolveLiteral(ctx, arg);
		return ResolveLiteral(ctx, v)
	case *datamodel.VariableRef:
		// matches TypeScript: case 'variable': return resolveVariableRef(ctx, arg);
		return ResolveVariableRef(ctx, v)
	default:
		// matches TypeScript: @ts-expect-error - should never happen
		// matches TypeScript: throw new Error(`Unsupported expression: ${arg?.type}`);
		if node, ok := v.(datamodel.Node); ok {
			panic(fmt.Sprintf("Unsupported expression: %s", node.Type()))
		} else {
			panic(fmt.Sprintf("Unsupported expression: %T", v))
		}
	}
}
