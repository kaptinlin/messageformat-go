// Package functions provides MessageFormat 2.0 function implementations
// TypeScript original code: functions/ module
package functions

import (
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// MessageFunction represents a function that can be called within a message
// TypeScript original code:
// export type MessageFunction = (
//
//	ctx: MessageFunctionContext,
//	options: Record<string, unknown>,
//	operand?: unknown
//
// ) => MessageValue;
type MessageFunction func(
	ctx MessageFunctionContext,
	options map[string]interface{},
	operand interface{},
) messagevalue.MessageValue

// MessageFunctionContext provides context for function execution
// TypeScript original code:
//
//	export class MessageFunctionContext {
//	  readonly dir: 'ltr' | 'rtl' | 'auto' | undefined;
//	  readonly id: string | undefined;
//	  readonly source: string;
//	  get literalOptionKeys(): Set<string>;
//	  get localeMatcher(): string;
//	  get locales(): string[];
//	  get onError(): (error: Error) => void;
//	}
type MessageFunctionContext struct {
	// Text direction override (optional)
	dir string

	// Unique identifier for the expression (optional)
	id string

	// Source string for error reporting
	source string

	// Available locales
	locales []string

	// Locale matcher strategy
	localeMatcher string

	// Error handler
	onError func(error)

	// Set of literal option keys
	literalOptionKeys map[string]bool
}

// NewMessageFunctionContext creates a new function context
func NewMessageFunctionContext(
	locales []string,
	source string,
	localeMatcher string,
	onError func(error),
	literalOptionKeys map[string]bool,
	dir string,
	id string,
) MessageFunctionContext {
	if literalOptionKeys == nil {
		literalOptionKeys = make(map[string]bool)
	}

	return MessageFunctionContext{
		dir:               dir,
		id:                id,
		source:            source,
		locales:           locales,
		localeMatcher:     localeMatcher,
		onError:           onError,
		literalOptionKeys: literalOptionKeys,
	}
}

// Dir returns the text direction override
func (ctx MessageFunctionContext) Dir() string {
	return ctx.dir
}

// ID returns the unique identifier
func (ctx MessageFunctionContext) ID() string {
	return ctx.id
}

// Source returns the source string
func (ctx MessageFunctionContext) Source() string {
	return ctx.source
}

// Locales returns the available locales
func (ctx MessageFunctionContext) Locales() []string {
	return ctx.locales
}

// LocaleMatcher returns the locale matcher strategy
func (ctx MessageFunctionContext) LocaleMatcher() string {
	return ctx.localeMatcher
}

// OnError calls the error handler
func (ctx MessageFunctionContext) OnError(err error) {
	if ctx.onError != nil {
		ctx.onError(err)
	}
}

// LiteralOptionKeys returns the set of literal option keys
func (ctx MessageFunctionContext) LiteralOptionKeys() map[string]bool {
	return ctx.literalOptionKeys
}
