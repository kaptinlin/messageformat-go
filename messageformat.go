// Package messageformat provides the main MessageFormat 2.0 API
package messageformat

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/kaptinlin/messageformat-go/internal/cst"
	"github.com/kaptinlin/messageformat-go/internal/resolve"
	"github.com/kaptinlin/messageformat-go/internal/selector"
	"github.com/kaptinlin/messageformat-go/pkg/bidi"
	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	"github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/logger"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// BidiIsolation represents the bidi isolation strategy
type BidiIsolation string

const (
	BidiDefault BidiIsolation = "default"
	BidiNone    BidiIsolation = "none"
)

// Direction represents text direction
// Use the Direction type from bidi package as the authoritative definition
type Direction = bidi.Direction

// Re-export constants from bidi package for API compatibility
const (
	DirLTR  = bidi.DirLTR
	DirRTL  = bidi.DirRTL
	DirAuto = bidi.DirAuto
)

// LocaleMatcher represents locale matching strategy
type LocaleMatcher string

const (
	LocaleBestFit LocaleMatcher = "best fit"
	LocaleLookup  LocaleMatcher = "lookup"
)

// MessageFormatOptions represents options for creating a MessageFormat
// TypeScript original code:
// export interface MessageFormatOptions<
//
//	T extends string = never,
//	P extends string = T
//
//	> {
//	  bidiIsolation?: 'default' | 'none';
//	  dir?: 'ltr' | 'rtl' | 'auto';
//	  localeMatcher?: 'best fit' | 'lookup';
//	  functions?: Record<string, MessageFunction<T, P>>;
//	}
type MessageFormatOptions struct {
	// The bidi isolation strategy for message formatting.
	// "default" isolates all expression placeholders except when both message and placeholder are LTR.
	// "none" applies no isolation at all.
	BidiIsolation BidiIsolation `json:"bidiIsolation,omitempty"`

	// Explicitly set the message's base direction.
	// If not set, the direction is detected from the primary locale.
	Dir Direction `json:"dir,omitempty"`

	// Locale matching algorithm for multiple locales.
	LocaleMatcher LocaleMatcher `json:"localeMatcher,omitempty"`

	// Custom functions to make available during message resolution.
	// Extends the default functions.
	Functions map[string]functions.MessageFunction `json:"functions,omitempty"`

	// Logger for this MessageFormat instance. If nil, uses global logger.
	Logger *slog.Logger `json:"-"`
}

// NewMessageFormatOptions creates a new MessageFormatOptions with defaults
func NewMessageFormatOptions(opts *MessageFormatOptions) *MessageFormatOptions {
	if opts == nil {
		opts = &MessageFormatOptions{}
	}
	if opts.BidiIsolation == "" {
		// Default to BidiNone for simpler output (KISS principle)
		// TypeScript defaults to 'default', but Go implementation prioritizes simplicity
		opts.BidiIsolation = BidiNone
	}
	if opts.Dir == "" {
		opts.Dir = DirAuto
	}
	if opts.LocaleMatcher == "" {
		opts.LocaleMatcher = LocaleBestFit
	}
	return opts
}

// MessageFormat represents a compiled MessageFormat 2.0 message
// TypeScript original code:
//
//	export class MessageFormat<T extends string = never, P extends string = T> {
//	  readonly #bidiIsolation: boolean;
//	  readonly #dir: 'ltr' | 'rtl' | 'auto';
//	  readonly #localeMatcher: 'best fit' | 'lookup';
//	  readonly #locales: Intl.Locale[];
//	  readonly #message: Message;
//	  readonly #functions: Record<string, MessageFunction<T | DefaultFunctionTypes, P | DefaultFunctionTypes>>;
//	  constructor(locales, source, options) { ... }
//	  format(msgParams, onError) { ... }
//	  formatToParts(msgParams, onError) { ... }
//	}
type MessageFormat struct {
	message       datamodel.Message
	locales       []string
	functions     map[string]functions.MessageFunction
	bidiIsolation bool         // true for "default", false for "none"
	dir           string       // "ltr" | "rtl" | "auto"
	localeMatcher string       // "best fit" | "lookup"
	logger        *slog.Logger // instance-specific logger
}

// New creates a new MessageFormat from locales, source, and options
// Supports both traditional options struct and functional options pattern
// TypeScript original code:
// constructor(
//
//	locales: string | string[] | undefined,
//	source: string | Message,
//	options?: MessageFormatOptions<T, P>
//
//	) {
//	  this.#bidiIsolation = options?.bidiIsolation !== 'none';
//	  this.#localeMatcher = options?.localeMatcher ?? 'best fit';
//	  this.#locales = Array.isArray(locales) ? locales.map(lc => new Intl.Locale(lc)) : locales ? [new Intl.Locale(locales)] : [];
//	  this.#dir = options?.dir ?? getLocaleDir(this.#locales[0]);
//	  this.#message = typeof source === 'string' ? parseMessage(source) : source;
//	  validate(this.#message);
//	  this.#functions = options?.functions ? Object.assign(Object.create(null), DefaultFunctions, options.functions) : DefaultFunctions;
//	}
func New(
	locales interface{}, // string | []string | nil
	source interface{}, // string | datamodel.Message
	options ...interface{}, // *MessageFormatOptions or ...Option
) (*MessageFormat, error) {
	// Parse locales parameter - matches TypeScript: string | string[] | undefined
	var localeList []string
	switch l := locales.(type) {
	case string:
		if l != "" {
			localeList = []string{l}
		} else {
			localeList = []string{}
		}
	case []string:
		if l == nil {
			localeList = []string{}
		} else {
			localeList = l
		}
	case nil:
		localeList = []string{}
	default:
		return nil, errors.NewCustomSyntaxError("locales must be string, []string, or nil")
	}

	// Parse source parameter - matches TypeScript: string | Message
	var message datamodel.Message
	var err error

	switch s := source.(type) {
	case string:
		// Parse the string using CST parser and convert to datamodel
		cstMessage := cst.ParseCST(s, false)

		// Check for CST parsing errors
		if len(cstMessage.Errors()) > 0 {
			// Return the first error
			firstError := cstMessage.Errors()[0]
			end := firstError.End
			return nil, errors.NewMessageSyntaxError(errors.ErrorTypeParseError, firstError.Start, &end, nil)
		}

		// Convert CST to datamodel
		message, err = datamodel.FromCST(cstMessage)
		if err != nil {
			return nil, err
		}
	case datamodel.Message:
		message = s
	case nil:
		return nil, errors.NewCustomSyntaxError("source cannot be nil")
	default:
		return nil, errors.NewCustomSyntaxError("source must be string or datamodel.Message")
	}

	// Validate the message
	_, err = datamodel.ValidateMessage(message, nil)
	if err != nil {
		return nil, err
	}

	// Handle options - support both traditional struct and functional options
	var opts *MessageFormatOptions
	switch len(options) {
	case 0:
		opts = &MessageFormatOptions{}
	case 1:
		// Check if it's nil (traditional way of passing no options)
		if options[0] == nil {
			opts = &MessageFormatOptions{}
		} else if structOpts, ok := options[0].(*MessageFormatOptions); ok {
			// Traditional options struct - check if the pointer itself is nil
			if structOpts == nil {
				opts = &MessageFormatOptions{}
			} else {
				opts = structOpts
			}
		} else if optFunc, ok := options[0].(Option); ok {
			// Single functional option
			opts = applyOptions(optFunc)
		} else {
			end := 1
			return nil, errors.NewMessageSyntaxError(errors.ErrorTypeParseError, 0, &end, nil)
		}
	default:
		// Multiple functional options
		var funcOpts []Option
		for _, opt := range options {
			if optFunc, ok := opt.(Option); ok {
				funcOpts = append(funcOpts, optFunc)
			} else {
				end := 1
				return nil, errors.NewMessageSyntaxError(errors.ErrorTypeParseError, 0, &end, nil)
			}
		}
		opts = applyOptions(funcOpts...)
	}

	// Apply defaults to options
	opts = NewMessageFormatOptions(opts)

	// Resolve bidiIsolation option (default is "default" which means true)
	bidiIsolation := opts.BidiIsolation != BidiNone

	// Resolve dir option
	dir := string(opts.Dir)
	if dir == "" || dir == string(DirAuto) {
		if len(localeList) > 0 {
			// Determine direction from first locale
			dir = string(bidi.GetLocaleDirection(localeList[0]))
		} else {
			dir = "auto"
		}
	}

	// Resolve localeMatcher option
	localeMatcher := string(opts.LocaleMatcher)

	// Set up functions
	functionMap := make(map[string]functions.MessageFunction)

	// Add default functions
	addDefaultFunctions(functionMap)

	// Add custom functions (override defaults if provided)
	if opts.Functions != nil {
		for name, fn := range opts.Functions {
			functionMap[name] = fn
		}
	}

	// Set up logger - use instance logger if provided, otherwise use global logger
	var instanceLogger *slog.Logger
	if opts.Logger != nil {
		instanceLogger = opts.Logger
	} else {
		instanceLogger = logger.GetLogger()
	}

	mf := &MessageFormat{
		message:       message,
		locales:       localeList,
		functions:     functionMap,
		bidiIsolation: bidiIsolation,
		dir:           dir,
		localeMatcher: localeMatcher,
		logger:        instanceLogger,
	}

	return mf, nil
}

// MustNew creates a new MessageFormat and panics if there's an error
// This is a convenience function for cases where you're certain the input is valid
func MustNew(
	locales interface{}, // string | []string | nil
	source interface{}, // string | datamodel.Message
	options ...interface{}, // *MessageFormatOptions or ...Option
) *MessageFormat {
	mf, err := New(locales, source, options...)
	if err != nil {
		panic(err)
	}
	return mf
}

// Format formats the message with the given values and optional error handler
// Supports both traditional onError callback and functional options pattern
func (mf *MessageFormat) Format(
	values map[string]interface{},
	options ...interface{}, // func(error) or ...FormatOption
) (string, error) {
	parts, err := mf.FormatToParts(values, options...)
	if err != nil {
		mf.logger.Error("failed to format message", "error", err)
		return "", err
	}

	// Concatenate all parts into a string
	var result strings.Builder

	for _, part := range parts {
		switch p := part.(type) {
		case *messagevalue.TextPart:
			result.WriteString(p.Value().(string))
		case *messagevalue.BidiIsolationPart:
			result.WriteString(p.Value().(string))
		case *messagevalue.FallbackPart:
			result.WriteString(p.Value().(string))
		case *messagevalue.MarkupPart:
			// Markup elements format as empty string - matches TypeScript behavior
			// TypeScript: formatMarkup(ctx, elem); // Handle errors, but discard results
			// Do nothing - markup doesn't contribute to string output
		default:
			// For other parts, try to get string representation
			if str, ok := p.Value().(string); ok {
				result.WriteString(str)
			} else {
				result.WriteString(fmt.Sprintf("%v", p.Value()))
			}
		}
	}

	return result.String(), nil
}

// FormatToParts formats the message and returns detailed parts
// Supports both traditional onError callback and functional options pattern
func (mf *MessageFormat) FormatToParts(
	values map[string]interface{},
	options ...interface{}, // func(error) or ...FormatOption
) ([]messagevalue.MessagePart, error) {
	// Parse options - support both traditional callback and functional options
	var onError func(error)

	switch len(options) {
	case 0:
		// Default error handler that emits warnings (matches TypeScript behavior)
		// TypeScript: process.emitWarning(error) or console.warn(error)
		onError = func(err error) {
			// Default: emit warning using logger (matches TypeScript behavior)
			mf.logger.Warn("MessageFormat error", "error", err)
		}
	case 1:
		// Check if it's nil (traditional way of passing no error handler)
		if options[0] == nil {
			onError = func(err error) {
				mf.logger.Warn("MessageFormat error", "error", err)
			}
		} else if errorFunc, ok := options[0].(func(error)); ok {
			// Traditional error callback
			onError = errorFunc
		} else if formatOpt, ok := options[0].(FormatOption); ok {
			// Single functional option
			formatOpts := applyFormatOptions(formatOpt)
			if formatOpts.OnError != nil {
				onError = formatOpts.OnError
			} else {
				onError = func(err error) {
					mf.logger.Warn("MessageFormat error", "error", err)
				}
			}
		} else {
			end := 1
			return nil, errors.NewMessageSyntaxError(errors.ErrorTypeParseError, 0, &end, nil)
		}
	default:
		// Multiple functional options
		var funcOpts []FormatOption
		for _, opt := range options {
			if formatOpt, ok := opt.(FormatOption); ok {
				funcOpts = append(funcOpts, formatOpt)
			} else {
				end := 1
				return nil, errors.NewMessageSyntaxError(errors.ErrorTypeParseError, 0, &end, nil)
			}
		}
		formatOpts := applyFormatOptions(funcOpts...)
		if formatOpts.OnError != nil {
			onError = formatOpts.OnError
		} else {
			onError = func(err error) {
				mf.logger.Warn("MessageFormat error", "error", err)
			}
		}
	}

	// Create resolution context with provided values
	ctx := mf.createContext(values, onError)

	// Select the pattern to format based on message type
	pattern := selector.SelectPattern(ctx, mf.message)

	// Format the selected pattern
	return mf.formatPattern(ctx, pattern)
}

// createContext creates a resolution context with the given values and error handler
// TypeScript original code:
// #createContext(
//
//	msgParams?: Record<string, unknown>,
//	onError: Context['onError'] = (error: Error) => {
//	  try {
//	    process.emitWarning(error);
//	  } catch {
//	    console.warn(error);
//	  }
//	}
//
//	) {
//	  const scope = { ...msgParams };
//	  for (const decl of this.#message.declarations) {
//	    scope[decl.name] = new UnresolvedExpression(decl.value, decl.type === 'input' ? (msgParams ?? {}) : undefined);
//	  }
//	  const ctx: Context = { onError, localeMatcher: this.#localeMatcher, locales: this.#locales, localVars: new WeakSet(), functions: this.#functions, scope };
//	  return ctx;
//	}
func (mf *MessageFormat) createContext(
	values map[string]interface{},
	onError func(error),
) *resolve.Context {
	// Start with base context scope
	scope := make(map[string]interface{})

	// Add provided values first
	for k, v := range values {
		scope[k] = v
	}

	// Add message declarations
	if err := mf.addDeclarationsToScope(scope, values); err != nil {
		onError(err)
	}

	return resolve.NewContext(mf.locales, mf.functions, scope, onError)
}

// addDeclarationsToScope adds message declarations to the scope
func (mf *MessageFormat) addDeclarationsToScope(
	scope map[string]interface{},
	msgParams map[string]interface{},
) error {
	declarations := mf.message.Declarations()

	for _, decl := range declarations {
		switch d := decl.(type) {
		case *datamodel.InputDeclaration:
			// For input declarations, create an unresolved expression
			// that will be resolved with the provided msgParams
			expr := d.Value()
			if varRefExpr, ok := expr.(*datamodel.VariableRefExpression); ok {
				// Convert VariableRefExpression to Expression for resolve package
				generalExpr := datamodel.NewExpression(varRefExpr.Arg(), varRefExpr.FunctionRef(), varRefExpr.Attributes())
				scope[d.Name()] = resolve.NewUnresolvedExpression(generalExpr, msgParams)
			}
		case *datamodel.LocalDeclaration:
			// For local declarations, create an unresolved expression
			// that will be resolved without external parameters
			expr := d.Value()
			if localExpr, ok := expr.(*datamodel.Expression); ok {
				scope[d.Name()] = resolve.NewUnresolvedExpression(localExpr, nil)
			}
		}
	}

	return nil
}

// formatPattern formats a pattern into message parts with bidi isolation
// TypeScript original code: pattern formatting logic
func (mf *MessageFormat) formatPattern(
	ctx *resolve.Context,
	pattern datamodel.Pattern,
) ([]messagevalue.MessagePart, error) {
	var parts []messagevalue.MessagePart

	for _, element := range pattern.Elements() {
		switch elem := element.(type) {
		case *datamodel.TextElement:
			// Text element
			parts = append(parts, messagevalue.NewTextPart(elem.Value(), elem.Value(), ""))

		case *datamodel.Expression:
			// Expression placeholder
			mv := resolve.ResolveExpression(ctx, elem)

			// Check if resolution failed
			if mv == nil {
				// Add fallback part for failed resolution
				parts = append(parts, messagevalue.NewFallbackPart("", getFirstLocale(ctx.Locales)))
				continue
			}

			// Apply bidi isolation if needed (matches TypeScript logic)
			if mf.shouldApplyBidiIsolation(mv) {
				// Add opening isolation
				isolationStart := mf.getBidiIsolationStart(mv.Dir())
				parts = append(parts, messagevalue.NewBidiIsolationPart(isolationStart))
			}

			// Convert MessageValue to parts
			valueParts, err := mv.ToParts()
			if err != nil {
				// Error during formatting - emit error and use fallback
				ctx.OnError(err)
				// Add fallback part
				parts = append(parts, messagevalue.NewFallbackPart(mv.Source(), getFirstLocale(ctx.Locales)))
			} else {
				// Add the actual parts
				parts = append(parts, valueParts...)
			}

			// Apply closing bidi isolation if we opened it
			if mf.shouldApplyBidiIsolation(mv) {
				// Add closing isolation
				parts = append(parts, messagevalue.NewBidiIsolationPart("\u2069")) // PDI
			}

		case *datamodel.Markup:
			// Markup element - format using resolve package
			markupPart := resolve.FormatMarkup(ctx, elem)
			parts = append(parts, markupPart)
		}
	}

	return parts, nil
}

// shouldApplyBidiIsolation determines if bidi isolation should be applied
// TypeScript original code:
// if (
//
//	this.#bidiIsolation &&
//	(this.#dir !== 'ltr' || mv.dir !== 'ltr' || mv[BIDI_ISOLATE])
//
// )
func (mf *MessageFormat) shouldApplyBidiIsolation(value messagevalue.MessageValue) bool {
	if !mf.bidiIsolation {
		return false
	}

	// Apply isolation if:
	// 1. Message direction is not LTR, OR
	// 2. Value direction is not LTR, OR
	// 3. Value has BIDI_ISOLATE flag set
	valueDir := value.Dir()

	// Check if value needs isolation based on TypeScript logic
	needsIsolation := mf.dir != "ltr" || valueDir != bidi.DirLTR

	// Check for BIDI_ISOLATE flag - matches TypeScript: mv[BIDI_ISOLATE]
	if hasIsolateFlag, ok := value.(interface{ HasBidiIsolate() bool }); ok {
		needsIsolation = needsIsolation || hasIsolateFlag.HasBidiIsolate()
	}

	return needsIsolation
}

// getBidiIsolationStart returns the appropriate bidi isolation start character
func (mf *MessageFormat) getBidiIsolationStart(valueDir bidi.Direction) string {
	switch valueDir {
	case bidi.DirLTR:
		return string(bidi.LRI)
	case bidi.DirRTL:
		return string(bidi.RLI)
	case bidi.DirAuto:
		return string(bidi.FSI)
	default:
		return string(bidi.FSI)
	}
}

// Dir returns the message's base direction
func (mf *MessageFormat) Dir() string {
	return mf.dir
}

// BidiIsolation returns whether bidi isolation is enabled
func (mf *MessageFormat) BidiIsolation() bool {
	return mf.bidiIsolation
}

// addDefaultFunctions adds default and draft functions to the function map
// TypeScript original code:
// this.#functions = options?.functions ? Object.assign(Object.create(null), DefaultFunctions, options.functions) : DefaultFunctions;
func addDefaultFunctions(functionMap map[string]functions.MessageFunction) {
	// Add default functions first
	defaults := functions.DefaultFunctions
	for name, fn := range defaults {
		functionMap[name] = fn
	}

	// Add draft functions (beta functions like math, currency, etc.)
	drafts := functions.DraftFunctions
	for name, fn := range drafts {
		functionMap[name] = fn
	}
}

// getFirstLocale returns the first locale from a list, or "en" as fallback
func getFirstLocale(locales []string) string {
	if len(locales) > 0 {
		return locales[0]
	}
	return "en"
}

// ResolvedMessageFormatOptions represents the resolved options for a MessageFormat instance
// Based on TC39 Intl.MessageFormat proposal
// https://github.com/tc39/proposal-intl-messageformat#constructor-options-and-resolvedoptions
type ResolvedMessageFormatOptions struct {
	BidiIsolation BidiIsolation                        `json:"bidiIsolation"`
	Dir           Direction                            `json:"dir"`
	Functions     map[string]functions.MessageFunction `json:"functions"`
	LocaleMatcher LocaleMatcher                        `json:"localeMatcher"`
}

// ResolvedOptions returns the resolved options for this MessageFormat instance
// This method is required by the TC39 Intl.MessageFormat proposal
// https://github.com/tc39/proposal-intl-messageformat#constructor-options-and-resolvedoptions
func (mf *MessageFormat) ResolvedOptions() ResolvedMessageFormatOptions {
	// Convert internal bidiIsolation boolean to BidiIsolation enum
	var bidiIsolation BidiIsolation
	if mf.bidiIsolation {
		bidiIsolation = BidiDefault
	} else {
		bidiIsolation = BidiNone
	}

	// Convert internal dir string to Direction type
	var dir Direction
	switch mf.dir {
	case "ltr":
		dir = DirLTR
	case "rtl":
		dir = DirRTL
	case "auto":
		dir = DirAuto
	default:
		dir = DirAuto
	}

	// Convert internal localeMatcher string to LocaleMatcher type
	var localeMatcher LocaleMatcher
	switch mf.localeMatcher {
	case "best fit":
		localeMatcher = LocaleBestFit
	case "lookup":
		localeMatcher = LocaleLookup
	default:
		localeMatcher = LocaleBestFit
	}

	// Create a copy of the functions map to avoid external modification
	functionsCopy := make(map[string]functions.MessageFunction)
	for k, v := range mf.functions {
		functionsCopy[k] = v
	}

	return ResolvedMessageFormatOptions{
		BidiIsolation: bidiIsolation,
		Dir:           dir,
		Functions:     functionsCopy,
		LocaleMatcher: localeMatcher,
	}
}
