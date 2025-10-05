// Package cst provides CST parsing for MessageFormat 2.0
// TypeScript original code: cst/parse-cst.ts module
package cst

import (
	"strings"

	"github.com/kaptinlin/messageformat-go/pkg/errors"
)

// ParseContext provides context for parsing
// TypeScript original code:
//
//	export class ParseContext {
//	  readonly errors: MessageSyntaxError[] = [];
//	  readonly resource: boolean;
//	  readonly source: string;
//
//	  constructor(source: string, opt?: { resource?: boolean }) {
//	    this.resource = opt?.resource ?? false;
//	    this.source = source;
//	  }
//	}
type ParseContext struct {
	errors   []*errors.MessageSyntaxError
	resource bool
	source   string
}

// NewParseContext creates a new parse context
// TypeScript original code: ParseContext constructor
func NewParseContext(source string, resource bool) *ParseContext {
	return &ParseContext{
		errors:   make([]*errors.MessageSyntaxError, 0),
		resource: resource,
		source:   source,
	}
}

// Errors returns the collected errors
// TypeScript original code: readonly errors: MessageSyntaxError[]
func (ctx *ParseContext) Errors() []*errors.MessageSyntaxError {
	return ctx.errors
}

// Resource returns whether this is a resource context
// TypeScript original code: readonly resource: boolean
func (ctx *ParseContext) Resource() bool {
	return ctx.resource
}

// Source returns the source string
// TypeScript original code: readonly source: string
func (ctx *ParseContext) Source() string {
	return ctx.source
}

// OnError adds an error to the context
// TypeScript original code:
// onError(
//
//	type: Exclude<
//	  typeof MessageSyntaxError.prototype.type,
//	  'missing-syntax' | typeof MessageDataModelError.prototype.type
//	>,
//	start: number,
//	end: number
//
// ): void;
// onError(type: 'missing-syntax', start: number, char: string): void;
func (ctx *ParseContext) OnError(errorType string, start int, endOrChar interface{}) {
	// Convert string error types to constants
	var errorTypeConstant string
	switch errorType {
	case "missing-syntax":
		errorTypeConstant = errors.ErrorTypeMissingSyntax
	case "extra-content":
		errorTypeConstant = errors.ErrorTypeExtraContent
	case "empty-token":
		errorTypeConstant = errors.ErrorTypeEmptyToken
	case "bad-escape":
		errorTypeConstant = errors.ErrorTypeBadEscape
	case "bad-input-expression":
		errorTypeConstant = errors.ErrorTypeBadInputExpression
	case "duplicate-option-name":
		errorTypeConstant = errors.ErrorTypeDuplicateOptionName
	case "parse-error":
		errorTypeConstant = errors.ErrorTypeParseError
	default:
		// Fallback to parse-error for unknown types
		errorTypeConstant = errors.ErrorTypeParseError
	}

	var err *errors.MessageSyntaxError
	var endPos int
	var expected *string

	// Handle different parameter types
	if errorType == "missing-syntax" && endOrChar != nil {
		if expectedStr, ok := endOrChar.(string); ok {
			// matches TypeScript: err = new MessageSyntaxError(type, start, start + exp.length, exp);
			endPos = start + len(expectedStr)
			expected = &expectedStr
		} else {
			endPos = start + 1
		}
	} else if end, ok := endOrChar.(int); ok {
		// matches TypeScript: err = new MessageSyntaxError(type, start, Number(end));
		endPos = end
	} else {
		endPos = start + 1
	}

	err = errors.NewMessageSyntaxError(errorTypeConstant, start, &endPos, expected)

	// matches TypeScript: this.errors.push(err);
	ctx.errors = append(ctx.errors, err)
}

// ParseCST parses a message source into a CST
// TypeScript original code:
// export function parseCST(
//
//	source: string,
//	opt?: { resource?: boolean }
//
//	): CST.Message {
//	  const ctx = new ParseContext(source, opt);
//
//	  const pos = whitespaces(source, 0).end;
//	  if (source.startsWith('.', pos)) {
//	    const { declarations, end } = parseDeclarations(ctx, pos);
//	    return source.startsWith('.match', end)
//	      ? parseSelectMessage(ctx, end, declarations)
//	      : parsePatternMessage(ctx, end, declarations, true);
//	  } else {
//	    return source.startsWith('{{', pos)
//	      ? parsePatternMessage(ctx, pos, [], true)
//	      : parsePatternMessage(ctx, 0, [], false);
//	  }
//	}
func ParseCST(source string, resource bool) Message {
	// matches TypeScript: const ctx = new ParseContext(source, opt);
	ctx := NewParseContext(source, resource)

	// matches TypeScript: const pos = whitespaces(source, 0).end;
	pos := Whitespaces(source, 0).End

	// matches TypeScript: if (source.startsWith('.', pos))
	if pos < len(source) && source[pos] == '.' {
		// matches TypeScript: const { declarations, end } = parseDeclarations(ctx, pos);
		declarations, end := parseDeclarations(ctx, pos)

		// matches TypeScript: return source.startsWith('.match', end) ? parseSelectMessage(...) : parsePatternMessage(...);
		if strings.HasPrefix(source[end:], ".match") {
			return parseSelectMessage(ctx, end, declarations)
		} else {
			return parsePatternMessage(ctx, end, declarations, true)
		}
	} else {
		// matches TypeScript: return source.startsWith('{{', pos) ? parsePatternMessage(...) : parsePatternMessage(...);
		if strings.HasPrefix(source[pos:], "{{") {
			return parsePatternMessage(ctx, 0, []Declaration{}, true)
		} else {
			return parsePatternMessage(ctx, 0, []Declaration{}, false)
		}
	}
}

// parsePatternMessage parses a simple or complex message
func parsePatternMessage(
	ctx *ParseContext,
	start int,
	declarations []Declaration,
	complex bool,
) Message {
	pattern := parsePattern(ctx, start, complex)
	pos := Whitespaces(ctx.source, pattern.End()).End

	if pos < len(ctx.source) {
		ctx.OnError("extra-content", pos, len(ctx.source))
	}

	if complex {
		return NewComplexMessage(declarations, *pattern, ctx.errors)
	} else {
		return NewSimpleMessage(*pattern, ctx.errors)
	}
}

// parseSelectMessage parses a select message
func parseSelectMessage(
	ctx *ParseContext,
	start int,
	declarations []Declaration,
) *SelectMessage {
	pos := start + 6 // ".match"
	match := NewSyntax(start, pos, ".match")

	ws := Whitespaces(ctx.source, pos)
	if !ws.HasWS {
		ctx.OnError("missing-syntax", pos, " ")
	}
	pos = ws.End

	var selectors []VariableRef
	for pos < len(ctx.source) {
		ch := ctx.source[pos]
		switch ch {
		case '{':
			// Selectors in .match must NOT be in braces according to MessageFormat 2.0
			// This is a syntax error
			expr := parseExpression(ctx, pos)
			ctx.OnError("bad-selector", expr.Start(), expr.End())
			pos = expr.End()
		case '$':
			// Correct: explicit $ prefix variables
			sel := ParseVariable(ctx, pos)
			selectors = append(selectors, *sel)
			pos = sel.End()
		default:
			// No more selectors
			goto selectorsEnd
		}

		ws = Whitespaces(ctx.source, pos)
		if !ws.HasWS {
			// Check if we're at the end or at a variant key - need whitespace between selectors
			// Also need whitespace before '*' variant key
			if pos < len(ctx.source) {
				// Always require whitespace after selector, even before '*'
				ctx.OnError("missing-syntax", pos, " ")
			}
		}
		pos = ws.End
	}

selectorsEnd:
	if len(selectors) == 0 {
		ctx.OnError("empty-token", pos, pos+1)
	}

	var variants []Variant
	for pos < len(ctx.source) {
		variant := parseVariant(ctx, pos)
		if variant.End() > pos {
			variants = append(variants, *variant)
			pos = variant.End()
		} else {
			pos++
		}
		pos = Whitespaces(ctx.source, pos).End
	}

	if pos < len(ctx.source) {
		ctx.OnError("extra-content", pos, len(ctx.source))
	}

	return NewSelectMessage(declarations, match, selectors, variants, ctx.errors)
}

// parseVariant parses a select variant
func parseVariant(ctx *ParseContext, start int) *Variant {
	pos := start
	var keys []Key

	for pos < len(ctx.source) {
		ws := Whitespaces(ctx.source, pos)
		pos = ws.End

		if pos >= len(ctx.source) {
			break
		}

		ch := ctx.source[pos]
		if ch == '{' {
			break
		}

		if pos > start && !ws.HasWS {
			ctx.OnError("missing-syntax", pos, " ")
		}

		var key Key
		if ch == '*' {
			key = NewCatchallKey(pos, pos+1)
			pos++
		} else {
			literal := ParseLiteral(ctx, pos, true)
			if literal != nil {
				// Normalize the literal value (Unicode normalization)
				literal.value = strings.ToValidUTF8(literal.value, "")
				key = literal
				pos = literal.End()
			} else {
				// If literal is nil, we can't proceed
				break
			}
		}

		if key == nil || key.End() == key.Start() {
			break // error; reported in pattern.errors
		}
		keys = append(keys, key)
	}

	value := parsePattern(ctx, pos, true)
	return NewVariant(start, value.End(), keys, *value)
}

// parsePattern parses a message pattern
func parsePattern(ctx *ParseContext, start int, quoted bool) *Pattern {
	pos := start
	var braces []Syntax

	if quoted {
		// Skip optional whitespace and bidi characters before {{
		pos = Whitespaces(ctx.source, pos).End
		if strings.HasPrefix(ctx.source[pos:], "{{") {
			braces = append(braces, NewSyntax(pos, pos+2, "{{"))
			pos += 2
		} else {
			ctx.OnError("missing-syntax", start, "{{")
			return NewPattern(start, start, []Node{}, nil)
		}
	}

	var body []Node
	for pos < len(ctx.source) {
		ch := ctx.source[pos]
		switch ch {
		case '{':
			expr := parseExpression(ctx, pos)
			body = append(body, expr)
			pos = expr.End()
		case '}':
			goto loop_end
		default:
			var text *Text
			if quoted {
				// In quoted patterns, use regular ParseText (no escape sequences)
				text = ParseText(ctx, pos)
			} else {
				// In simple patterns, use ParseSimpleText (with escape sequences)
				text = ParseSimpleText(ctx, pos)
			}
			body = append(body, text)
			pos = text.End()
		}
	}
loop_end:

	if quoted {
		// Skip optional whitespace and bidi characters before }}
		pos = Whitespaces(ctx.source, pos).End
		if strings.HasPrefix(ctx.source[pos:], "}}") {
			braces = append(braces, NewSyntax(pos, pos+2, "}}"))
			pos += 2
		} else {
			ctx.OnError("missing-syntax", pos, "}}")
		}
	}

	return NewPattern(start, pos, body, braces)
}

// parseDeclarations parses message declarations
func parseDeclarations(ctx *ParseContext, start int) ([]Declaration, int) {
	// Pre-allocate with small initial capacity to reduce allocations
	declarations := make([]Declaration, 0, 4)
	pos := start
	source := ctx.source

	for pos < len(source) && source[pos] == '.' {
		if strings.HasPrefix(source[pos:], ".match") {
			break
		}

		var decl Declaration
		switch {
		case strings.HasPrefix(source[pos:], ".input"):
			decl = parseInputDeclaration(ctx, pos)
		case strings.HasPrefix(source[pos:], ".local"):
			decl = parseLocalDeclaration(ctx, pos)
		default:
			decl = parseDeclarationJunk(ctx, pos)
		}

		declarations = append(declarations, decl)
		pos = Whitespaces(source, decl.End()).End
	}

	return declarations, pos
}

// parseInputDeclaration parses an input declaration
func parseInputDeclaration(ctx *ParseContext, start int) *InputDeclaration {
	pos := start + 6 // ".input"
	keyword := NewSyntax(start, pos, ".input")
	pos = Whitespaces(ctx.source, pos).End

	value := parseDeclarationValue(ctx, pos)
	if expr, ok := value.(*Expression); ok {
		if expr.markup != nil || (expr.arg != nil && expr.arg.Type() != "variable") {
			ctx.OnError("bad-input-expression", value.Start(), value.End())
		}
	}

	return NewInputDeclaration(start, value.End(), keyword, value)
}

// parseLocalDeclaration parses a local declaration
func parseLocalDeclaration(ctx *ParseContext, start int) *LocalDeclaration {
	source := ctx.source
	pos := start + 6 // ".local"
	keyword := NewSyntax(start, pos, ".local")

	ws := Whitespaces(source, pos)
	pos = ws.End

	if !ws.HasWS {
		ctx.OnError("missing-syntax", pos, " ")
	}

	var target Node
	if pos < len(source) && source[pos] == '$' {
		target = ParseVariable(ctx, pos)
		pos = target.End()
	} else {
		junkStart := pos
		junkEnd := pos
		for junkEnd < len(source) {
			ch := source[junkEnd]
			if ch == '\t' || ch == '\n' || ch == '\r' || ch == ' ' || ch == '=' || ch == '{' || ch == '}' {
				break
			}
			junkEnd++
		}
		target = NewJunk(junkStart, junkEnd, source[junkStart:junkEnd])
		ctx.OnError("missing-syntax", junkStart, "$")
		pos = junkEnd
	}

	pos = Whitespaces(source, pos).End
	var equals *Syntax
	if pos < len(source) && source[pos] == '=' {
		equalsSyntax := NewSyntax(pos, pos+1, "=")
		equals = &equalsSyntax
		pos++
	} else {
		ctx.OnError("missing-syntax", pos, "=")
	}

	pos = Whitespaces(source, pos).End
	value := parseDeclarationValue(ctx, pos)

	return NewLocalDeclaration(start, value.End(), keyword, target, equals, value)
}

// parseDeclarationValue parses a declaration value (expression or junk)
func parseDeclarationValue(ctx *ParseContext, start int) Node {
	if start < len(ctx.source) && ctx.source[start] == '{' {
		return parseExpression(ctx, start)
	} else {
		return parseDeclarationJunk(ctx, start)
	}
}

// parseDeclarationJunk parses junk content in declarations
func parseDeclarationJunk(ctx *ParseContext, start int) *Junk {
	source := ctx.source
	end := len(source)

	// Look for next declaration or pattern start
	for i := start + 1; i < len(source)-1; i++ {
		if (source[i] == '.' && i+1 < len(source) &&
			(source[i+1] >= 'a' && source[i+1] <= 'z')) ||
			(source[i] == '{' && i+1 < len(source) && source[i+1] == '{') {
			end = i
			break
		}
	}

	// Trim trailing whitespace
	for end > start && (source[end-1] == ' ' || source[end-1] == '\t' ||
		source[end-1] == '\n' || source[end-1] == '\r') {
		end--
	}

	ctx.OnError("missing-syntax", start, "{")
	return NewJunk(start, end, source[start:end])
}
