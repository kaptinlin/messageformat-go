// Package messageformat provides functional options for MessageFormat configuration
package messageformat

import (
	"log/slog"

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
func WithBidiIsolation(strategy BidiIsolation) Option {
	return func(opts *MessageFormatOptions) {
		opts.BidiIsolation = strategy
	}
}

// WithBidiIsolationString sets the bidi isolation strategy from string (for backward compatibility)
func WithBidiIsolationString(strategy string) Option {
	return func(opts *MessageFormatOptions) {
		switch strategy {
		case "default":
			opts.BidiIsolation = BidiDefault
		case "none":
			opts.BidiIsolation = BidiNone
		default:
			opts.BidiIsolation = BidiDefault
		}
	}
}

// WithDir sets the message's base direction
// TypeScript original code:
// dir?: 'ltr' | 'rtl' | 'auto';
func WithDir(direction Direction) Option {
	return func(opts *MessageFormatOptions) {
		opts.Dir = direction
	}
}

// WithDirString sets the message's base direction from string (for backward compatibility)
func WithDirString(direction string) Option {
	return func(opts *MessageFormatOptions) {
		switch direction {
		case "ltr":
			opts.Dir = DirLTR
		case "rtl":
			opts.Dir = DirRTL
		case "auto":
			opts.Dir = DirAuto
		default:
			opts.Dir = DirAuto
		}
	}
}

// WithLocaleMatcher sets the locale matching algorithm
// TypeScript original code:
// localeMatcher?: 'best fit' | 'lookup';
func WithLocaleMatcher(matcher LocaleMatcher) Option {
	return func(opts *MessageFormatOptions) {
		opts.LocaleMatcher = matcher
	}
}

// WithLocaleMatcherString sets the locale matching algorithm from string (for backward compatibility)
func WithLocaleMatcherString(matcher string) Option {
	return func(opts *MessageFormatOptions) {
		switch matcher {
		case "best fit":
			opts.LocaleMatcher = LocaleBestFit
		case "lookup":
			opts.LocaleMatcher = LocaleLookup
		default:
			opts.LocaleMatcher = LocaleBestFit
		}
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

// WithLogger sets a custom logger for this MessageFormat instance
func WithLogger(logger *slog.Logger) Option {
	return func(opts *MessageFormatOptions) {
		opts.Logger = logger
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
