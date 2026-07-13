// Package messageformat provides functional options for MessageFormat configuration
package messageformat

import (
	"errors"
	"fmt"
	"log/slog"
	"maps"

	"github.com/kaptinlin/messageformat-go/pkg/functions"
)

// ErrInvalidOption indicates that a constructor option is outside its supported vocabulary.
var ErrInvalidOption = errors.New("invalid MessageFormat option")

// Option represents a functional option for MessageFormat constructor
// TypeScript original code: MessageFormatOptions interface
type Option func(*MessageFormatOptions)

// FormatOption represents a functional option for Format methods
type FormatOption func(*FormatOptions)

// FormatOptions represents options for Format and FormatToParts methods
type FormatOptions struct {
	OnError func(error)
}

// NewMessageFormatOptions applies functional options to a fresh MessageFormatOptions value.
func NewMessageFormatOptions(options ...Option) *MessageFormatOptions {
	opts := &MessageFormatOptions{}
	for _, option := range options {
		if option == nil {
			continue
		}
		option(opts)
	}
	return opts
}

// validateMessageFormatOptions validates the closed constructor vocabularies.
// TypeScript original code:
//
//	interface MessageFormatOptions {
//	  bidiIsolation?: 'default' | 'none';
//	  dir?: 'ltr' | 'rtl' | 'auto';
//	  localeMatcher?: 'best fit' | 'lookup';
//	}
func validateMessageFormatOptions(options *MessageFormatOptions) error {
	switch options.BidiIsolation {
	case BidiDefault, BidiNone:
	default:
		return fmt.Errorf("%w: bidiIsolation %q", ErrInvalidOption, options.BidiIsolation)
	}
	switch options.Dir {
	case DirLTR, DirRTL, DirAuto:
	default:
		return fmt.Errorf("%w: dir %q", ErrInvalidOption, options.Dir)
	}
	switch options.LocaleMatcher {
	case LocaleBestFit, LocaleLookup:
	default:
		return fmt.Errorf("%w: localeMatcher %q", ErrInvalidOption, options.LocaleMatcher)
	}

	return nil
}

// NewFormatOptions applies functional options to a fresh FormatOptions value.
func NewFormatOptions(options ...FormatOption) *FormatOptions {
	opts := &FormatOptions{}
	for _, option := range options {
		if option == nil {
			continue
		}
		option(opts)
	}
	return opts
}

// Options converts a configuration struct into a constructor option.
func Options(options MessageFormatOptions) Option {
	return func(opts *MessageFormatOptions) {
		*opts = options
	}
}

// WithBidiIsolation sets the bidi isolation strategy
// TypeScript original code:
// bidiIsolation?: 'default' | 'none';
func WithBidiIsolation(strategy BidiIsolation) Option {
	return func(opts *MessageFormatOptions) {
		opts.BidiIsolation = strategy
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

// WithLocaleMatcher sets the locale matching algorithm
// TypeScript original code:
// localeMatcher?: 'best fit' | 'lookup';
func WithLocaleMatcher(matcher LocaleMatcher) Option {
	return func(opts *MessageFormatOptions) {
		opts.LocaleMatcher = matcher
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
		maps.Copy(opts.Functions, funcs)
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
