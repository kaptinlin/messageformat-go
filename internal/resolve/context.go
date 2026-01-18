// Package resolve provides expression resolution for MessageFormat 2.0
// TypeScript original code: format-context.ts module
package resolve

import (
	"maps"

	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// Context represents the resolution context for message formatting
// TypeScript original code:
//
//	export interface Context<T extends string = string, P extends string = T> {
//	  functions: Record<string, MessageFunction<T, P>>;
//	  onError(error: unknown): void;
//	  localeMatcher: 'best fit' | 'lookup';
//	  locales: Intl.Locale[];
//	  localVars: WeakSet<MessageValue<T, P>>;
//	  scope: Record<string, unknown>;
//	}
type Context struct {
	// Available functions
	Functions map[string]functions.MessageFunction

	// Error handler for resolution errors
	OnError func(error)

	// Locale matcher strategy
	LocaleMatcher string

	// Available locales
	Locales []string

	// Set of local variables (for cycle detection)
	LocalVars map[messagevalue.MessageValue]bool

	// Variable scope
	Scope map[string]interface{}

	// Track variables currently being resolved (for circular reference detection)
	ResolvingVars map[string]bool
}

// NewContext creates a new resolution context
// TypeScript original code: Context constructor logic
func NewContext(
	locales []string,
	funcs map[string]functions.MessageFunction,
	scope map[string]interface{},
	onError func(error),
) *Context {
	if funcs == nil {
		funcs = make(map[string]functions.MessageFunction)
	}
	if scope == nil {
		scope = make(map[string]interface{})
	}

	return &Context{
		Functions:     funcs,
		OnError:       onError,
		LocaleMatcher: "best fit",
		Locales:       locales,
		LocalVars:     make(map[messagevalue.MessageValue]bool),
		Scope:         scope,
		ResolvingVars: make(map[string]bool),
	}
}

// Clone creates a copy of the context
// TypeScript original code: { ...ctx } spread operator equivalent
func (ctx *Context) Clone() *Context {
	return &Context{
		Functions:     ctx.Functions, // Immutable, safe to share
		OnError:       ctx.OnError,
		LocaleMatcher: ctx.LocaleMatcher,
		Locales:       ctx.Locales, // Immutable, safe to share
		LocalVars:     maps.Clone(ctx.LocalVars),
		Scope:         maps.Clone(ctx.Scope),
		ResolvingVars: ctx.ResolvingVars, // Share the resolving vars tracking
	}
}

// CloneWithScope creates a copy of the context with a new scope
// TypeScript original code: { ...ctx, scope: newScope } spread operator equivalent
func (ctx *Context) CloneWithScope(newScope map[string]interface{}) *Context {
	cloned := ctx.Clone()

	// Merge new scope with existing scope
	for k, v := range newScope {
		cloned.Scope[k] = v
	}

	return cloned
}
