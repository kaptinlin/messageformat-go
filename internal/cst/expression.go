// Package cst provides expression parsing for CST
// TypeScript original code: cst/expression.ts module
package cst

import (
	"github.com/kaptinlin/messageformat-go/pkg/errors"
)

// parseExpression parses a placeholder expression
// TypeScript original code:
// export function parseExpression(
//
//	ctx: ParseContext,
//	start: number
//
//	): CST.Expression {
//	  const { source } = ctx;
//	  let pos = start + 1; // '{'
//	  pos = whitespaces(source, pos).end;
//
//	  const arg =
//	    source[pos] === '$'
//	      ? parseVariable(ctx, pos)
//	      : parseLiteral(ctx, pos, false);
//	  if (arg) {
//	    pos = arg.end;
//	    const ws = whitespaces(source, pos);
//	    if (!ws.hasWS && source[pos] !== '}') {
//	      ctx.onError('missing-syntax', pos, ' ');
//	    }
//	    pos = ws.end;
//	  }
//
//	  let functionRef: CST.FunctionRef | CST.Junk | undefined;
//	  let markup: CST.Markup | undefined;
//	  let junkError: MessageSyntaxError | undefined;
//	  switch (source[pos]) {
//	    case ':':
//	      functionRef = parseFunctionRefOrMarkup(ctx, pos, 'function');
//	      pos = functionRef.end;
//	      break;
//	    case '#':
//	    case '/':
//	      if (arg) ctx.onError('extra-content', arg.start, arg.end);
//	      markup = parseFunctionRefOrMarkup(ctx, pos, 'markup');
//	      pos = markup.end;
//	      break;
//	    case '@':
//	    case '}':
//	      if (!arg) ctx.onError('empty-token', start, pos);
//	      break;
//	    default:
//	      if (!arg) {
//	        const end = pos + 1;
//	        functionRef = { type: 'junk', start: pos, end, source: source[pos] };
//	        junkError = new MessageSyntaxError('parse-error', start, end);
//	        ctx.errors.push(junkError);
//	      }
//	  }
//	  // ... rest of function
//	}
//
// isDigit checks if a character is a digit
func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

// isIdentifierStart checks if a character can start an identifier
func isIdentifierStart(ch byte) bool {
	// Use the existing notNameStartRegex to check if character can start a name
	// If it matches notNameStart (-.0-9), then it cannot start an identifier
	return !notNameStartRegex.MatchString(string(ch)) &&
		nameCharsRegex.MatchString(string(ch))
}

// parseVariableRef parses a variable reference without $ prefix
func parseVariableRef(ctx *ParseContext, start int) *VariableRef {
	source := ctx.Source()
	name := ParseNameValue(source, start)
	if name == nil {
		ctx.OnError("empty-token", start, start+1)
		return NewVariableRef(start, start, NewSyntax(start, start, ""), "")
	}

	// For unquoted identifiers, we don't have an explicit $ prefix
	// So we create a VariableRef with an empty open syntax
	open := NewSyntax(start, start, "")
	return NewVariableRef(start, name.End, open, name.Value)
}

func parseExpression(ctx *ParseContext, start int) *Expression {
	source := ctx.Source()
	pos := start + 1 // '{'
	pos = Whitespaces(source, pos).End

	var arg Node
	if pos < len(source) {
		ch := source[pos]
		if ch == '$' {
			// Explicit variable reference: {$name}
			variable := ParseVariable(ctx, pos)
			if variable != nil {
				arg = variable
			}
		} else if ch == '|' {
			// Quoted literal: {|text|} (MessageFormat 2.0 spec)
			literal := ParseLiteral(ctx, pos, false)
			if literal != nil {
				arg = literal
			}
		} else if isDigit(ch) || ch == '-' || ch == '+' {
			// Numeric literal: {123} {-456} {+789}
			literal := ParseLiteral(ctx, pos, false)
			if literal != nil {
				arg = literal
			}
		} else if isIdentifierStart(ch) {
			// Unquoted identifier = unquoted literal: {name} {count}
			// According to MessageFormat 2.0 spec, these should be literals, not variable references
			literal := ParseLiteral(ctx, pos, false)
			if literal != nil {
				arg = literal
			}
		} else {
			// Fall back to literal parsing for other cases
			literal := ParseLiteral(ctx, pos, false)
			if literal != nil {
				arg = literal
			}
		}
	}

	if arg != nil {
		pos = arg.End()
		ws := Whitespaces(source, pos)
		if !ws.HasWS && pos < len(source) && source[pos] != '}' {
			ctx.OnError("missing-syntax", pos, " ")
		}
		pos = ws.End
	}

	var functionRef Node
	var markup *Markup
	var junkError *errors.MessageSyntaxError

	if pos < len(source) {
		switch source[pos] {
		case ':':
			functionRef = parseFunctionRefOrMarkup(ctx, pos, "function")
			pos = functionRef.End()
		case '#', '/':
			if arg != nil {
				ctx.OnError("extra-content", arg.Start(), arg.End())
			}
			markupNode := parseFunctionRefOrMarkup(ctx, pos, "markup")
			if m, ok := markupNode.(*Markup); ok {
				markup = m
			}
			pos = markupNode.End()
		case '@', '}':
			if arg == nil {
				ctx.OnError("empty-token", start, pos)
			}
		default:
			if arg == nil {
				end := pos + 1
				functionRef = NewJunk(pos, end, string(source[pos]))
				junkError = errors.NewMessageSyntaxError(errors.ErrorTypeParseError, start, &end, nil)
				ctx.errors = append(ctx.errors, junkError)
			}
		}
	}

	var attributes []Attribute
	reqWS := functionRef != nil || markup != nil
	ws := Whitespaces(source, pos)

	for pos < len(source) && source[ws.End] == '@' {
		if reqWS && !ws.HasWS {
			ctx.OnError("missing-syntax", pos, " ")
		}
		pos = ws.End
		attr := parseAttribute(ctx, pos)
		attributes = append(attributes, *attr)
		pos = attr.End()
		reqWS = true
		ws = Whitespaces(source, pos)
	}
	pos = ws.End

	open := NewSyntax(start, start+1, "{")
	var close *Syntax

	if pos >= len(source) {
		ctx.OnError("missing-syntax", pos, "}")
	} else {
		if source[pos] != '}' {
			errStart := pos
			for pos < len(source) && source[pos] != '}' {
				pos++
			}
			if junk, ok := functionRef.(*Junk); ok {
				junk.end = pos
				junk.source = source[junk.start:pos]
				if junkError != nil {
					junkError.End = pos
				}
			} else {
				ctx.OnError("extra-content", errStart, pos)
			}
		}
		if pos < len(source) && source[pos] == '}' {
			closeSyntax := NewSyntax(pos, pos+1, "}")
			close = &closeSyntax
			pos++
		}
	}

	var braces []Syntax
	if close != nil {
		braces = []Syntax{open, *close}
	} else {
		braces = []Syntax{open}
	}

	end := pos
	if markup != nil {
		return NewExpression(start, end, braces, nil, nil, markup, attributes)
	} else {
		return NewExpression(start, end, braces, arg, functionRef, nil, attributes)
	}
}

// parseFunctionRefOrMarkup parses a function reference or markup
func parseFunctionRefOrMarkup(ctx *ParseContext, start int, nodeType string) Node {
	source := ctx.Source()
	id := parseIdentifier(ctx, start+1)
	pos := id.End
	var options []Option
	var close *Syntax

	// Track option names to detect duplicates
	optionNames := make(map[string]bool)

	for pos < len(source) {
		ws := Whitespaces(source, pos)
		next := byte(0)
		if ws.End < len(source) {
			next = source[ws.End]
		}

		if next == '@' || next == '}' {
			break
		}

		if next == '/' && source[start] == '#' {
			pos = ws.End + 1
			closeSyntax := NewSyntax(pos-1, pos, "/")
			close = &closeSyntax
			ws = Whitespaces(source, pos)
			if ws.HasWS {
				ctx.OnError("extra-content", pos, ws.End)
			}
			break
		}

		if !ws.HasWS {
			ctx.OnError("missing-syntax", pos, " ")
		}
		pos = ws.End

		opt := parseOption(ctx, pos)
		if opt.End() == pos {
			break // error
		}

		// Check for duplicate option names
		optionName := getOptionName(opt.Name())
		if optionNames[optionName] {
			ctx.OnError("duplicate-option-name", opt.Start(), opt.End())
		} else {
			optionNames[optionName] = true
		}

		options = append(options, *opt)
		pos = opt.End()
	}

	if nodeType == "function" {
		open := NewSyntax(start, start+1, ":")
		return NewFunctionRef(start, pos, open, id.Parts, options)
	} else {
		open := NewSyntax(start, start+1, string(source[start]))
		return NewMarkup(start, pos, open, id.Parts, options, close)
	}
}

// getOptionName extracts the full option name from identifier parts
func getOptionName(identifier Identifier) string {
	var name string
	for _, part := range identifier {
		name += part.Value()
	}
	return name
}

// IdentifierResult represents the result of parsing an identifier
type IdentifierResult struct {
	Parts Identifier
	End   int
}

// parseIdentifier parses an identifier (name or namespace:name)
// TypeScript original code:
// function parseIdentifier(
//
//	ctx: ParseContext,
//	start: number
//
//	): { parts: CST.Identifier; end: number } {
//	  const { source } = ctx;
//	  const name0 = parseNameValue(source, start);
//	  if (!name0) {
//	    ctx.onError('empty-token', start, start + 1);
//	    return { parts: [{ start, end: start, value: '' }], end: start };
//	  }
//	  let pos = name0.end;
//	  const id0 = { start, end: pos, value: name0.value };
//	  if (source[pos] !== ':') return { parts: [id0], end: pos };
//
//	  const sep = { start: pos, end: pos + 1, value: ':' as const };
//	  pos += 1;
//
//	  const name1 = parseNameValue(source, pos);
//	  if (name1) {
//	    const id1 = { start: pos, end: name1.end, value: name1.value };
//	    return { parts: [id0, sep, id1], end: name1.end };
//	  } else {
//	    ctx.onError('empty-token', pos, pos + 1);
//	    return { parts: [id0, sep], end: pos };
//	  }
//	}
func parseIdentifier(ctx *ParseContext, start int) *IdentifierResult {
	source := ctx.Source()
	name0 := ParseNameValue(source, start)

	if name0 == nil {
		ctx.OnError("empty-token", start, start+1)
		return &IdentifierResult{
			Parts: Identifier{NewSyntax(start, start, "")},
			End:   start,
		}
	}

	pos := name0.End
	id0 := NewSyntax(start, pos, name0.Value)

	if pos >= len(source) || source[pos] != ':' {
		return &IdentifierResult{
			Parts: Identifier{id0},
			End:   pos,
		}
	}

	sep := NewSyntax(pos, pos+1, ":")
	pos++

	name1 := ParseNameValue(source, pos)
	if name1 != nil {
		id1 := NewSyntax(pos, name1.End, name1.Value)
		return &IdentifierResult{
			Parts: Identifier{id0, sep, id1},
			End:   name1.End,
		}
	} else {
		ctx.OnError("empty-token", pos, pos+1)
		return &IdentifierResult{
			Parts: Identifier{id0, sep},
			End:   pos,
		}
	}
}

// parseOption parses a function or markup option
// TypeScript original code:
//
//	function parseOption(ctx: ParseContext, start: number): CST.Option {
//	  const id = parseIdentifier(ctx, start);
//	  let pos = whitespaces(ctx.source, id.end).end;
//	  let equals: CST.Syntax<'='> | undefined;
//	  if (ctx.source[pos] === '=') {
//	    equals = { start: pos, end: pos + 1, value: '=' };
//	    pos += 1;
//	  } else {
//	    ctx.onError('missing-syntax', pos, '=');
//	  }
//	  pos = whitespaces(ctx.source, pos).end;
//	  const value =
//	    ctx.source[pos] === '$'
//	      ? parseVariable(ctx, pos)
//	      : parseLiteral(ctx, pos, true);
//	  return { start, end: value.end, name: id.parts, equals, value };
//	}
func parseOption(ctx *ParseContext, start int) *Option {
	id := parseIdentifier(ctx, start)
	pos := Whitespaces(ctx.Source(), id.End).End

	var equals *Syntax
	if pos < len(ctx.Source()) && ctx.Source()[pos] == '=' {
		equalsSyntax := NewSyntax(pos, pos+1, "=")
		equals = &equalsSyntax
		pos++
	} else {
		ctx.OnError("missing-syntax", pos, "=")
	}

	pos = Whitespaces(ctx.Source(), pos).End

	var value Node
	if pos < len(ctx.Source()) {
		if ctx.Source()[pos] == '$' {
			value = ParseVariable(ctx, pos)
		} else {
			value = ParseLiteral(ctx, pos, true)
		}
	}

	// Ensure value is never nil to avoid nil pointer dereference
	if value == nil {
		// Create a dummy literal to avoid nil - ensure proper initialization
		value = NewLiteral(pos, pos, false, nil, "", nil)
	}

	return NewOption(start, value.End(), id.Parts, equals, value)
}

// parseAttribute parses an expression attribute
// TypeScript original code:
//
//	function parseAttribute(ctx: ParseContext, start: number): CST.Attribute {
//	  const { source } = ctx;
//	  const id = parseIdentifier(ctx, start + 1);
//	  let pos = id.end;
//	  const ws = whitespaces(source, pos);
//	  let equals: CST.Syntax<'='> | undefined;
//	  let value: CST.Literal | undefined;
//	  if (source[ws.end] === '=') {
//	    pos = ws.end + 1;
//	    equals = { start: pos - 1, end: pos, value: '=' };
//	    pos = whitespaces(source, pos).end;
//	    value = parseLiteral(ctx, pos, true);
//	    pos = value.end;
//	  }
//	  return {
//	    start,
//	    end: pos,
//	    open: { start, end: start + 1, value: '@' },
//	    name: id.parts,
//	    equals,
//	    value
//	  };
//	}
func parseAttribute(ctx *ParseContext, start int) *Attribute {
	source := ctx.Source()
	id := parseIdentifier(ctx, start+1)
	pos := id.End
	ws := Whitespaces(source, pos)

	var equals *Syntax
	var value *Literal

	if ws.End < len(source) && source[ws.End] == '=' {
		pos = ws.End + 1
		equalsSyntax := NewSyntax(pos-1, pos, "=")
		equals = &equalsSyntax
		pos = Whitespaces(source, pos).End
		value = ParseLiteral(ctx, pos, true)
		if value != nil {
			pos = value.End()
		}
	} else {
		// Fix: Don't include trailing whitespace in attribute end position
		// This matches the TypeScript original behavior
		pos = id.End
	}

	open := NewSyntax(start, start+1, "@")
	return NewAttribute(start, pos, open, id.Parts, equals, value)
}
