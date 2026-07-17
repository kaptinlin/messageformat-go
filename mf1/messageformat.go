// Package v1 provides ICU MessageFormat implementation for Go.
//
// This package implements the ICU MessageFormat specification, providing
// internationalization support for formatting messages with variables,
// pluralization rules, and conditional text selection.
//
// Example usage:
//
//	mf, err := v1.New("en", nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	msg, err := mf.Compile("Hello {name}, you have {count, plural, one {# item} other {# items}}!")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	result, err := msg.Format(map[string]any{
//		"name":  "Alice",
//		"count": 5,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(result) // Output: Hello Alice, you have 5 items!
//
// Key Features:
//   - ICU MessageFormat specification compliance
//   - CLDR-backed plural and Intl formatting through go-intl
//   - High-performance compilation and execution
//   - Thread-safe message formatters
//   - Go-native typed constructors and compiled values
//   - Support for number formatting, date/time formatting, and currency
//   - Nested message templates and complex conditionals
//
// Behavioral reference:
// .reference/messageformat/mf1/packages/core/src/messageformat.ts

package v1

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

type compiledEvaluator func(map[string]any) ([]any, error)

// CompiledMessage is an immutable parsed MF1 message with typed terminal projections.
// TypeScript original code:
// const message = mf.compile(source);
type CompiledMessage struct {
	evaluate compiledEvaluator
}

// Format renders the message as text. A nil values map is treated as empty.
// TypeScript original code:
// message(values);
func (message *CompiledMessage) Format(values map[string]any) (string, error) {
	valuesResult, err := message.evaluate(values)
	if err != nil {
		return "", err
	}
	var result strings.Builder
	for _, value := range valuesResult {
		fmt.Fprint(&result, value)
	}
	return result.String(), nil
}

// FormatValues renders the message as its ordered terminal values. A nil map is treated as empty.
// TypeScript original code:
// new MessageFormat(locale, { returnType: 'values' }).compile(source)(values);
func (message *CompiledMessage) FormatValues(values map[string]any) ([]any, error) {
	return message.evaluate(values)
}

// Formatter formats one custom argument with the effective locale and style.
// TypeScript original code:
// customFormatters[name](value, locale, arg)
type Formatter func(value any, locale, style string) (string, error)

// MessageFormatOptions represents options for the MessageFormat constructor
// Uses zero-value semantics to simplify API usage
type MessageFormatOptions struct {
	// Add Unicode control characters to all input parts to preserve the
	// integrity of the output when mixing LTR and RTL text
	// Default: false (zero value)
	BiDiSupport bool `json:"biDiSupport,omitempty"`

	// The currency to use when formatting {V, number, currency}
	// Default: "USD" (empty string uses default)
	Currency string `json:"currency,omitempty"`

	// The time zone to use when formatting {V, date}
	// Default: "" (empty string uses system timezone)
	TimeZone string `json:"timeZone,omitempty"`

	// Map of custom formatting functions to include
	// Default: nil (zero value)
	CustomFormatters map[string]Formatter `json:"customFormatters,omitempty"`

	// Used to identify and map keys to locale identifiers
	// Return empty string for null/undefined (following TypeScript pattern)
	// Default: nil (zero value)
	LocaleCodeFromKey func(key string) string `json:"-"`

	// Require all message arguments to be set with a defined value
	// Default: false (zero value)
	RequireAllArguments bool `json:"requireAllArguments,omitempty"`

	// Follow the ICU MessageFormat spec more closely
	// Default: false (zero value)
	Strict bool `json:"strict,omitempty"`

	// Enable strict checks for plural keys according to Unicode CLDR
	// Default: PluralKeyModeDefault (which means strict=true)
	StrictPluralKeys PluralKeyMode `json:"strictPluralKeys,omitempty"`
}

// MessageFormatOptionsWithDefaults represents options with default values applied
type MessageFormatOptionsWithDefaults struct {
	BiDiSupport         bool                    `json:"biDiSupport"`
	Currency            string                  `json:"currency"`
	TimeZone            string                  `json:"timeZone"`
	CustomFormatters    map[string]Formatter    `json:"customFormatters"`
	LocaleCodeFromKey   func(key string) string `json:"-"`
	RequireAllArguments bool                    `json:"requireAllArguments"`
	Strict              bool                    `json:"strict"`
	StrictPluralKeys    bool                    `json:"strictPluralKeys"`
}

// ResolvedMessageFormatOptions represents resolved options returned by resolvedOptions
type ResolvedMessageFormatOptions struct {
	MessageFormatOptionsWithDefaults
	Locale  string         `json:"locale"`
	Plurals []PluralObject `json:"plurals"`
}

// Note: PluralFunction and PluralObject are defined in plurals.go

// MessageFormat represents the core MessageFormat-to-JavaScript compiler
type MessageFormat struct {
	options MessageFormatOptionsWithDefaults
	plurals []PluralObject
}

const defaultLocale = "en"

var (
	escapeRegexp           = regexp.MustCompile(`[{}]`)
	escapeOctothorpeRegexp = regexp.MustCompile(`[#{}]`)
)

// Escape escapes characters that may be considered as MessageFormat markup
// This surrounds the characters {, } and optionally # with 'quotes'.
// This will allow those characters to not be considered as MessageFormat control characters.
// TypeScript original code:
//
//	static escape(str: string, octothorpe?: boolean) {
//	  const esc = octothorpe ? /[#{}]/g : /[{}]/g;
//	  return String(str).replace(esc, "'$&'");
//	}
func Escape(str string, octothorpe bool) string {
	re := escapeRegexp
	if octothorpe {
		re = escapeOctothorpeRegexp
	}

	return re.ReplaceAllString(str, "'$0'")
}

// SupportedLocalesOf returns the locales with built-in plural category support.
// TypeScript original code:
// static supportedLocalesOf(locales: string | string[]) { return la.filter(hasPlural); }
func SupportedLocalesOf(locales []string) ([]string, error) {
	var result []string
	for _, localeTag := range locales {
		locale, err := parseStrictLocale(localeTag)
		if err != nil {
			return nil, WrapInvalidLocale(localeTag)
		}
		if hasPluralLocale(locale) {
			result = append(result, localeTag)
		}
	}
	return result, nil
}

// hasPlural checks if a locale has plural support by consulting go-intl's
// CLDR-backed pluralrules package. Unlike intlbridge.ParseLocale, this uses
// strict parsing: malformed tags ("x", "xx") and valid unsupported tags
// ("eo") return false instead of silently aliasing to English.
func hasPlural(loc string) bool {
	if len(loc) < 2 {
		return false
	}
	parsed, err := parseStrictLocale(loc)
	if err != nil {
		return false
	}
	return hasPluralLocale(parsed)
}

// getPlural resolves the plural object for a validated locale.
// TypeScript original code:
// const pl = getPlural(locale);
func getPlural(locale string) *PluralObject {
	plural, err := GetPlural(locale)
	if err != nil {
		return nil
	}
	return &plural
}

// New creates a new MessageFormat compiler for a locale.
// TypeScript original code:
// constructor(locale, options)
func New(locale string, options *MessageFormatOptions) (*MessageFormat, error) {
	if _, err := parseStrictLocale(locale); err != nil {
		return nil, WrapInvalidLocale(locale)
	}

	mf := &MessageFormat{}

	// Apply options with zero-value semantics and defaults
	var opts MessageFormatOptions
	if options != nil {
		opts = *options
	}
	if err := validateFormatters(opts.CustomFormatters); err != nil {
		return nil, err
	}

	// Set default options with user overrides
	mf.options = MessageFormatOptionsWithDefaults{
		BiDiSupport:         opts.BiDiSupport, // false is meaningful default
		Currency:            "USD",            // Default currency
		TimeZone:            opts.TimeZone,    // Empty string means system timezone
		CustomFormatters:    maps.Clone(opts.CustomFormatters),
		LocaleCodeFromKey:   opts.LocaleCodeFromKey,
		RequireAllArguments: opts.RequireAllArguments, // false is meaningful default
		Strict:              opts.Strict,              // false is meaningful default
		StrictPluralKeys:    true,                     // Default to true
	}

	// Apply non-zero user values
	if opts.Currency != "" {
		mf.options.Currency = opts.Currency
	}
	// Handle StrictPluralKeys special case
	switch opts.StrictPluralKeys {
	case PluralKeyModeDefault:
		mf.options.StrictPluralKeys = true // Default behavior
	case PluralKeyModeStrict:
		mf.options.StrictPluralKeys = true
	case PluralKeyModeRelaxed:
		mf.options.StrictPluralKeys = false
	}

	if plural := getPlural(locale); plural != nil {
		mf.plurals = []PluralObject{*plural}
	} else {
		return nil, WrapInvalidLocale(locale)
	}

	return mf, nil
}

// validateFormatters validates custom formatter registrations at construction.
// TypeScript original code:
// customFormatters?: { [key: string]: CustomFormatter };
func validateFormatters(formatters map[string]Formatter) error {
	for name, formatter := range formatters {
		if !validFormatterName(name) {
			return fmt.Errorf("formatter name %q: %w", name, ErrInvalidFormatter)
		}
		if formatter == nil {
			return fmt.Errorf("formatter %q has nil handler: %w", name, ErrInvalidFormatter)
		}
	}
	return nil
}

// validFormatterName reports whether the lexer can resolve a custom formatter name.
// TypeScript original code:
// match: /,\s*[^\p{Pat_Syn}\p{Pat_WS}]+\s*/u;
func validFormatterName(name string) bool {
	if name == "" || strings.ContainsAny(name, "{},") || strings.ContainsFunc(name, unicode.IsSpace) {
		return false
	}
	switch strings.ToLower(name) {
	case "number", "date", "time", "plural", "select", "selectordinal":
		return false
	default:
		return true
	}
}

// NewWithPlural creates a MessageFormat using caller-supplied plural behavior.
// TypeScript original code:
// new MessageFormat(pluralFunction, options)
func NewWithPlural(profile PluralProfile, options *MessageFormatOptions) (*MessageFormat, error) {
	if err := validatePluralProfile(profile); err != nil {
		return nil, err
	}

	mf, err := New(defaultLocale, options)
	if err != nil {
		return nil, err
	}
	custom := newCustomPlural(profile)
	mf.plurals = []PluralObject{custom}
	return mf, nil
}

// ResolvedOptions returns a new object with properties reflecting the default locale,
// plurals, and other options computed during initialization.
func (mf *MessageFormat) ResolvedOptions() ResolvedMessageFormatOptions {
	var locale string
	if len(mf.plurals) > 0 {
		locale = mf.plurals[0].Locale
	} else {
		locale = defaultLocale
	}
	options := mf.options
	options.CustomFormatters = maps.Clone(mf.options.CustomFormatters)
	plurals := slices.Clone(mf.plurals)
	for i := range plurals {
		plurals[i].Cardinals = slices.Clone(plurals[i].Cardinals)
		plurals[i].Ordinals = slices.Clone(plurals[i].Ordinals)
	}

	return ResolvedMessageFormatOptions{
		MessageFormatOptionsWithDefaults: options,
		Locale:                           locale,
		Plurals:                          plurals,
	}
}

// Compile parses source into an immutable message with text and values projections.
// TypeScript original code:
// const message = mf.compile(source);
func (mf *MessageFormat) Compile(message string) (*CompiledMessage, error) {
	var plural *PluralObject
	if len(mf.plurals) > 0 {
		plural = &mf.plurals[0]
	}

	tokens, err := Parse(message, &ParseOptions{
		Strict:           mf.options.Strict,
		StrictPluralKeys: &mf.options.StrictPluralKeys,
		Cardinal:         plural.Cardinals,
		Ordinal:          plural.Ordinals,
	})
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	return &CompiledMessage{
		evaluate: func(values map[string]any) ([]any, error) {
			result, err := mf.executeTokens(tokens, values, plural, nil)
			if err != nil {
				return nil, fmt.Errorf("execution error: %w", err)
			}
			return result, nil
		},
	}, nil
}

// executeTokens evaluates tokens into ordered terminal values.
// TypeScript original code:
// const values = tokens.map(token => this.token(token, pluralToken));
func (mf *MessageFormat) executeTokens(
	tokens []Token,
	values map[string]any,
	plural *PluralObject,
	pluralContext *Select,
) ([]any, error) {
	result := make([]any, 0, len(tokens))
	for _, token := range tokens {
		switch token := token.(type) {
		case *Content:
			result = append(result, token.Value)

		case *PlainArg:
			value, exists := values[token.Arg]
			if !exists {
				if mf.options.RequireAllArguments {
					return nil, WrapMissingArgument(token.Arg)
				}
				value = ""
			}
			result = append(result, value)

		case *FunctionArg:
			value, exists := values[token.Arg]
			if !exists && mf.options.RequireAllArguments {
				return nil, WrapMissingArgument(token.Arg)
			}
			formatted, err := mf.formatValue(value, token.Key, token.Param, values, plural)
			if err != nil {
				return nil, err
			}
			result = append(result, formatted)

		case *Select:
			if _, exists := values[token.Arg]; !exists && mf.options.RequireAllArguments {
				return nil, WrapMissingArgument(token.Arg)
			}
			selectedCase, err := mf.selectCase(token, values, plural)
			if err != nil {
				return nil, err
			}

			nestedPluralContext := pluralContext
			if token.Type == "plural" || token.Type == "selectordinal" {
				nestedPluralContext = token
			}
			nested, err := mf.executeTokens(
				selectedCase.Tokens,
				values,
				plural,
				nestedPluralContext,
			)
			if err != nil {
				return nil, err
			}
			result = append(result, nested...)

		case *Octothorpe:
			if pluralContext == nil {
				result = append(result, "#")
				continue
			}
			value, exists := values[pluralContext.Arg]
			if !exists {
				result = append(result, "#")
				continue
			}

			offset := 0
			if pluralContext.PluralOffset != nil {
				offset = *pluralContext.PluralOffset
			}
			locale := defaultLocale
			if plural != nil && plural.Locale != "" {
				locale = plural.Locale
			}
			formatted, err := mf.numberFormatter(locale, value, offset)
			if err != nil {
				result = append(result, "#")
				continue
			}
			result = append(result, formatted)
		}
	}
	return result, nil
}

func (mf *MessageFormat) formatValue(value any, key string, param []Token, _ map[string]any, plural *PluralObject) (string, error) {
	locale := defaultLocale
	if plural != nil && plural.Locale != "" {
		locale = plural.Locale
	}
	style := literalFormatterStyle(param)

	switch strings.ToLower(key) {
	case "number":
		return formatNumber(value, locale, style, mf.options.Currency)
	case "date":
		return formatDate(value, locale, style, mf.options.TimeZone)
	case "time":
		return formatTime(value, locale, style, mf.options.TimeZone)
	default:
		if formatter, ok := mf.options.CustomFormatters[key]; ok {
			return formatter(value, locale, style)
		}
	}

	return fmt.Sprintf("%v", value), nil
}

// literalFormatterStyle returns the literal style encoded in a function argument.
// TypeScript original code:
// const arg = param.map(tok => this.token(tok, pluralToken)).join(”).trim();
func literalFormatterStyle(tokens []Token) string {
	var style strings.Builder
	for _, token := range tokens {
		if content, ok := token.(*Content); ok {
			style.WriteString(content.Value)
		}
	}
	return strings.TrimSpace(style.String())
}

// selectCase selects the appropriate case from a select statement
func (mf *MessageFormat) selectCase(sel *Select, paramMap map[string]any, plural *PluralObject) (*SelectCase, error) {
	value, exists := paramMap[sel.Arg]
	if !exists {
		// Find "other" case
		for _, c := range sel.Cases {
			if c.Key == "other" {
				return &c, nil
			}
		}
		return nil, ErrNoOtherCase
	}

	switch sel.Type {
	case "select":
		// String matching
		valueStr := fmt.Sprintf("%v", value)
		for _, c := range sel.Cases {
			if c.Key == valueStr {
				return &c, nil
			}
		}
		// Fall back to "other"
		for _, c := range sel.Cases {
			if c.Key == "other" {
				return &c, nil
			}
		}

	case "plural", "selectordinal":
		// Numeric matching with plural rules
		var numValue float64
		switch v := value.(type) {
		case int:
			numValue = float64(v)
		case float64:
			numValue = v
		case string:
			// Try to parse as number
			_, _ = fmt.Sscanf(v, "%f", &numValue) // Explicitly ignore parsing errors
		default:
			numValue = 0
		}

		// Check exact matches first (=n)
		exactKey := formatExactKey(numValue)
		for _, c := range sel.Cases {
			if c.Key == exactKey {
				return &c, nil
			}
		}

		// Exact keys use the original value; only plural category selection uses the offset.
		if sel.PluralOffset != nil {
			numValue -= float64(*sel.PluralOffset)
		}

		// Use plural function to determine category
		if plural != nil && plural.Func != nil {
			category, err := plural.Func(numValue, sel.Type == "selectordinal")
			if err == nil {
				for _, c := range sel.Cases {
					if c.Key == string(category) {
						return &c, nil
					}
				}
			}
		}

		// Fall back to "other"
		for _, c := range sel.Cases {
			if c.Key == "other" {
				return &c, nil
			}
		}
	}

	return nil, WrapNoMatchingCase(sel.Arg, sel.Type)
}

// numberFormatter provides locale-aware number formatting
// TypeScript original code:
//
//	export function number(lc: string, value: number, offset: number) {
//	  return _nf(lc).format(value - offset);
//	}
func (mf *MessageFormat) numberFormatter(locale string, value any, offset int) (string, error) {
	// Convert value to number
	var num float64
	switch v := value.(type) {
	case int:
		num = float64(v)
	case int64:
		num = float64(v)
	case float64:
		num = v
	case float32:
		num = float64(v)
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return "", WrapInvalidNumberStr(v)
		}
		num = parsed
	default:
		return "", WrapInvalidType(fmt.Sprintf("%T", value))
	}

	// Apply offset
	result := num - float64(offset)

	return formatNumber(result, locale, "", mf.options.Currency)
}
