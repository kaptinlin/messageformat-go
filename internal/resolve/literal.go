// Package resolve provides literal resolution functions for MessageFormat 2.0
// TypeScript original code: resolve/resolve-literal.ts module
package resolve

import (
	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// ResolveLiteral resolves a literal expression to a MessageValue
// TypeScript original code:
//
//	export function resolveLiteral(ctx: Context, lit: Literal) {
//	  const msgCtx = new MessageFunctionContext(ctx, `|${lit.value}|`);
//	  return string(msgCtx, {}, lit.value);
//	}
func ResolveLiteral(ctx *Context, literal *datamodel.Literal) messagevalue.MessageValue {
	// matches TypeScript: const msgCtx = new MessageFunctionContext(ctx, `|${lit.value}|`);
	source := getValueSource(literal) // This creates |value| format
	msgCtx := functions.NewMessageFunctionContext(
		ctx.Locales,
		source,
		ctx.LocaleMatcher,
		ctx.OnError,
		nil,
		"",
		"",
	)

	// Use string function to handle literal values - matches TypeScript: return string(msgCtx, {}, lit.value);
	stringFunc, exists := ctx.Functions["string"]
	if !exists {
		// Fallback if string function not available
		return messagevalue.NewStringValue(
			literal.Value(),
			getFirstLocale(ctx.Locales),
			source,
		)
	}

	// matches TypeScript: return string(msgCtx, {}, lit.value);
	return stringFunc(msgCtx, make(map[string]interface{}), literal.Value())
}
