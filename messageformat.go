// Package messageformat provides the main MessageFormat 2.0 API.
// Construction failures are returned as error values.
package messageformat

import (
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"

	"github.com/kaptinlin/messageformat-go/internal/cst"
	"github.com/kaptinlin/messageformat-go/internal/resolve"
	"github.com/kaptinlin/messageformat-go/internal/selector"
	"github.com/kaptinlin/messageformat-go/pkg/bidi"
	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
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

	bidiIsolationSet bool
}

// NewOptions creates a new MessageFormatOptions with defaults
func NewOptions(opts *MessageFormatOptions) *MessageFormatOptions {
	if opts == nil {
		opts = &MessageFormatOptions{}
	}
	if opts.BidiIsolation == "" {
		// Default to BidiNone for backward compatibility
		// However, enable BidiDefault for RTL locales to match TypeScript reference implementation
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

// Parse creates a MessageFormat by parsing source text and applying options.
func Parse(locales []string, source string, options ...Option) (*MessageFormat, error) {
	cstMessage := cst.ParseCST(source, false)
	if len(cstMessage.Errors()) > 0 {
		return nil, cstMessage.Errors()[0]
	}

	message, err := datamodel.FromCST(cstMessage)
	if err != nil {
		return nil, err
	}

	return Compile(locales, message, options...)
}

// Compile creates a MessageFormat from a prebuilt data model and applies options.
func Compile(locales []string, message datamodel.Message, options ...Option) (*MessageFormat, error) {
	localeList := slices.Clone(locales)

	_, err := datamodel.ValidateMessage(message, nil)
	if err != nil {
		return nil, err
	}

	opts := NewOptions(NewMessageFormatOptions(options...))
	if !opts.bidiIsolationSet && opts.BidiIsolation == BidiNone && len(localeList) > 0 && bidi.GetLocaleDirection(localeList[0]) == bidi.DirRTL {
		opts.BidiIsolation = BidiDefault
	}

	bidiIsolation := opts.BidiIsolation != BidiNone

	dir := string(opts.Dir)
	if dir == "" || dir == string(DirAuto) {
		if len(localeList) > 0 {
			dir = string(bidi.GetLocaleDirection(localeList[0]))
		} else {
			dir = "auto"
		}
	}

	localeMatcher := string(opts.LocaleMatcher)

	functionMap := make(map[string]functions.MessageFunction)
	addDefaultFunctions(functionMap)
	if opts.Functions != nil {
		maps.Copy(functionMap, opts.Functions)
	}

	instanceLogger := logger.GetLogger()
	if opts.Logger != nil {
		instanceLogger = opts.Logger
	}

	return &MessageFormat{
		message:       message,
		locales:       localeList,
		functions:     functionMap,
		bidiIsolation: bidiIsolation,
		dir:           dir,
		localeMatcher: localeMatcher,
		logger:        instanceLogger,
	}, nil
}

// Format formats the message with the given values and format options.
func (mf *MessageFormat) Format(
	values map[string]any,
	options ...FormatOption,
) (string, error) {
	parts, err := mf.FormatToParts(values, options...)
	if err != nil {
		mf.logger.Error("failed to format message", "error", err)
		return "", err
	}

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
			continue
		default:
			if str, ok := p.Value().(string); ok {
				result.WriteString(str)
			} else {
				fmt.Fprintf(&result, "%v", p.Value())
			}
		}
	}

	return result.String(), nil
}

// FormatToParts formats the message and returns detailed parts.
func (mf *MessageFormat) FormatToParts(
	values map[string]any,
	options ...FormatOption,
) ([]messagevalue.MessagePart, error) {
	formatOptions := NewFormatOptions(options...)
	onError := func(err error) {
		mf.logger.Warn("MessageFormat error", "error", err)
	}
	if formatOptions.OnError != nil {
		onError = formatOptions.OnError
	}

	ctx := mf.createContext(values, onError)
	pattern := selector.SelectPattern(ctx, mf.message)
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
	values map[string]any,
	onError func(error),
) *resolve.Context {
	scope := make(map[string]any)
	maps.Copy(scope, values)

	for _, decl := range mf.message.Declarations() {
		switch d := decl.(type) {
		case *datamodel.InputDeclaration:
			expr := d.Value()
			if varRefExpr, ok := expr.(*datamodel.VariableRefExpression); ok {
				generalExpr := datamodel.NewExpression(varRefExpr.Arg(), varRefExpr.FunctionRef(), varRefExpr.Attributes())
				scope[d.Name()] = resolve.NewUnresolvedExpression(generalExpr, values)
			}
		case *datamodel.LocalDeclaration:
			expr := d.Value()
			if localExpr, ok := expr.(*datamodel.Expression); ok {
				combinedScope := make(map[string]any)
				maps.Copy(combinedScope, values)
				// Prefer unresolved declarations from scope over raw message parameters.
				maps.Copy(combinedScope, scope)
				scope[d.Name()] = resolve.NewUnresolvedExpression(localExpr, combinedScope)
			}
		}
	}

	return resolve.NewContext(mf.locales, mf.functions, scope, onError)
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
			parts = append(parts, messagevalue.NewTextPart(elem.Value(), elem.Value(), ""))

		case *datamodel.Expression:
			mv := resolve.ResolveExpression(ctx, elem)

			if mv == nil {
				parts = append(parts, messagevalue.NewFallbackPart("", functions.GetFirstLocale(ctx.Locales)))
				continue
			}

			if mf.shouldApplyBidiIsolation(mv) {
				isolationStart := mf.getBidiIsolationStart(mv.Dir())
				parts = append(parts, messagevalue.NewBidiIsolationPart(isolationStart))
			}

			valueParts, err := mv.ToParts()
			if err != nil {
				ctx.OnError(err)
				valueParts = []messagevalue.MessagePart{
					messagevalue.NewFallbackPart(mv.Source(), functions.GetFirstLocale(ctx.Locales)),
				}
			}
			parts = append(parts, valueParts...)

			if mf.shouldApplyBidiIsolation(mv) {
				parts = append(parts, messagevalue.NewBidiIsolationPart("\u2069")) // PDI
			}

		case *datamodel.Markup:
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

	// Apply isolation if message or value direction is not LTR
	if mf.dir != "ltr" || value.Dir() != bidi.DirLTR {
		return true
	}

	// Check for BIDI_ISOLATE flag - matches TypeScript: mv[BIDI_ISOLATE]
	if hasIsolateFlag, ok := value.(interface{ HasBidiIsolate() bool }); ok {
		return hasIsolateFlag.HasBidiIsolate()
	}

	return false
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

// addDefaultFunctions adds default and draft functions to the function map.
// TypeScript original code:
// this.#functions = options?.functions ? Object.assign(Object.create(null), DefaultFunctions, options.functions) : DefaultFunctions;
func addDefaultFunctions(functionMap map[string]functions.MessageFunction) {
	maps.Copy(functionMap, functions.DefaultFunctions)
	maps.Copy(functionMap, functions.DraftFunctions)
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
	bidiIsolation := BidiNone
	if mf.bidiIsolation {
		bidiIsolation = BidiDefault
	}

	dir := bidi.ParseDirection(mf.dir)

	var localeMatcher LocaleMatcher
	switch mf.localeMatcher {
	case "best fit":
		localeMatcher = LocaleBestFit
	case "lookup":
		localeMatcher = LocaleLookup
	default:
		localeMatcher = LocaleBestFit
	}

	functionsCopy := maps.Clone(mf.functions)

	return ResolvedMessageFormatOptions{
		BidiIsolation: bidiIsolation,
		Dir:           dir,
		Functions:     functionsCopy,
		LocaleMatcher: localeMatcher,
	}
}
