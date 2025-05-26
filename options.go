// Package messageformat provides functional options for MessageFormat configuration
package messageformat

import (
	"github.com/kaptinlin/messageformat-go/pkg/functions"
)

// Option represents a functional option for MessageFormat constructor
// TypeScript original code: MessageFormatOptions interface
type Option func(*MessageFormatOptions)

// FormatOption represents a functional option for Format methods
type FormatOption func(*FormatOptions)

// FormatOptions represents options for Format and FormatToParts methods
type FormatOptions struct {
	OnError func(error)
}

// WithBidiIsolation sets the bidi isolation strategy
// TypeScript original code:
// bidiIsolation?: 'default' | 'none';
func WithBidiIsolation(strategy string) Option {
	return func(opts *MessageFormatOptions) {
		opts.BidiIsolation = &strategy
	}
}

// WithDir sets the message's base direction
// TypeScript original code:
// dir?: 'ltr' | 'rtl' | 'auto';
func WithDir(direction string) Option {
	return func(opts *MessageFormatOptions) {
		opts.Dir = &direction
	}
}

// WithLocaleMatcher sets the locale matching algorithm
// TypeScript original code:
// localeMatcher?: 'best fit' | 'lookup';
func WithLocaleMatcher(matcher string) Option {
	return func(opts *MessageFormatOptions) {
		opts.LocaleMatcher = &matcher
	}
}

// WithFunction adds a single custom function
// TypeScript original code:
// functions?: Record<string, MessageFunction<T, P>>;
func WithFunction(name string, fn functions.MessageFunction) Option {
	return func(opts *MessageFormatOptions) {
		if opts.Functions == nil {
			opts.Functions = make(map[string]functions.MessageFunction)
		}
		opts.Functions[name] = fn
	}
}

// WithFunctions adds multiple custom functions
// TypeScript original code:
// functions?: Record<string, MessageFunction<T, P>>;
func WithFunctions(funcs map[string]functions.MessageFunction) Option {
	return func(opts *MessageFormatOptions) {
		if opts.Functions == nil {
			opts.Functions = make(map[string]functions.MessageFunction)
		}
		for name, fn := range funcs {
			opts.Functions[name] = fn
		}
	}
}

// WithErrorHandler sets an error handler for Format methods
// TypeScript original code:
// format(msgParams?: Record<string, unknown>, onError?: (error: Error) => void): string
func WithErrorHandler(handler func(error)) FormatOption {
	return func(opts *FormatOptions) {
		opts.OnError = handler
	}
}

// applyOptions applies functional options to MessageFormatOptions
func applyOptions(options ...Option) *MessageFormatOptions {
	opts := &MessageFormatOptions{}
	for _, option := range options {
		option(opts)
	}
	return opts
}

// applyFormatOptions applies functional options to FormatOptions
func applyFormatOptions(options ...FormatOption) *FormatOptions {
	opts := &FormatOptions{}
	for _, option := range options {
		option(opts)
	}
	return opts
}
