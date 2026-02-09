// Package cst provides value parsing for CST
// TypeScript original code: cst/values.ts module
package cst

import (
	"strconv"
	"strings"
)

// ParseText parses literal text content
// TypeScript original code:
//
//	export function parseText(ctx: ParseContext, start: number): CST.Text {
//	  let value = '';
//	  let pos = start;
//	  let i = start;
//	  loop: for (; i < ctx.source.length; ++i) {
//	    switch (ctx.source[i]) {
//	      case '\\': {
//	        const esc = parseEscape(ctx, i);
//	        if (esc) {
//	          value += ctx.source.substring(pos, i) + esc.value;
//	          i += esc.length;
//	          pos = i + 1;
//	        }
//	        break;
//	      }
//	      case '{':
//	      case '}':
//	        break loop;
//	      case '\n':
//	        if (ctx.resource) {
//	          const nl = i;
//	          let next = ctx.source[i + 1];
//	          while (next === ' ' || next === '\t') {
//	            i += 1;
//	            next = ctx.source[i + 1];
//	          }
//	          if (i > nl) {
//	            value += ctx.source.substring(pos, nl + 1);
//	            pos = i + 1;
//	          }
//	        }
//	        break;
//	    }
//	  }
//	  value += ctx.source.substring(pos, i);
//	  return { type: 'text', start, end: i, value };
//	}
func ParseText(ctx *ParseContext, start int) *Text {
	var value strings.Builder
	pos := start
	i := start
	source := ctx.Source()

	for i < len(source) {
		ch := source[i]
		switch ch {
		case '\\':
			esc := parseEscape(ctx, i)
			if esc != nil {
				value.WriteString(source[pos:i])
				value.WriteString(esc.Value)
				i += esc.Length
				pos = i + 1
			}
		case '{':
			// Single { means end of text ({{ is handled by parsePattern for quoted patterns)
			goto loop_end
		case '}':
			// Single } means end of text (}} is handled by parsePattern for quoted patterns)
			goto loop_end
		case '\n':
			if ctx.Resource() {
				nl := i
				next := byte(0)
				if i+1 < len(source) {
					next = source[i+1]
				}
				for next == ' ' || next == '\t' {
					i++
					if i+1 < len(source) {
						next = source[i+1]
					} else {
						break
					}
				}
				if i > nl {
					value.WriteString(source[pos : nl+1])
					pos = i + 1
				}
			}
		}
		i++
	}

loop_end:
	value.WriteString(source[pos:i])
	return NewText(start, i, value.String())
}

// ParseLiteral parses a literal value (quoted or unquoted)
// TypeScript original code:
// export function parseLiteral(
//
//	ctx: ParseContext,
//	start: number,
//	required: boolean
//
//	): CST.Literal | undefined {
//	  if (ctx.source[start] === '|') return parseQuotedLiteral(ctx, start);
//	  const value = parseUnquotedLiteralValue(ctx.source, start);
//	  if (!value) {
//	    if (required) ctx.onError('empty-token', start, start);
//	    else return undefined;
//	  }
//	  const end = start + value.length;
//	  return { type: 'literal', quoted: false, start, end, value };
//	}
func ParseLiteral(ctx *ParseContext, start int, required bool) *Literal {
	source := ctx.Source()

	if start < len(source) && source[start] == '|' {
		return parseQuotedLiteral(ctx, start)
	}

	value := ParseUnquotedLiteralValue(source, start)
	if value == "" {
		if required {
			ctx.OnError("empty-token", start, start)
			// Return an empty literal instead of nil when required
			return NewLiteral(start, start, false, nil, "", nil)
		}
		return nil
	}

	end := start + len(value)
	return NewLiteral(start, end, false, nil, value, nil)
}

// parseQuotedLiteral parses a quoted literal value
func parseQuotedLiteral(ctx *ParseContext, start int) *Literal {
	var value strings.Builder
	pos := start + 1
	source := ctx.Source()

	open := NewSyntax(start, pos, "|")

	for i := pos; i < len(source); i++ {
		ch := source[i]
		switch ch {
		case '\\':
			esc := parseEscape(ctx, i)
			if esc != nil {
				value.WriteString(source[pos:i])
				value.WriteString(esc.Value)
				i += esc.Length
				pos = i + 1
			}
		case '|':
			value.WriteString(source[pos:i])
			close := NewSyntax(i, i+1, "|")
			return NewLiteral(start, i+1, true, &open, value.String(), &close)
		case '\n':
			if ctx.Resource() {
				nl := i
				next := byte(0)
				if i+1 < len(source) {
					next = source[i+1]
				}
				for next == ' ' || next == '\t' {
					i++
					if i+1 < len(source) {
						next = source[i+1]
					} else {
						break
					}
				}
				if i > nl {
					value.WriteString(source[pos : nl+1])
					pos = i + 1
				}
			}
		}
	}

	value.WriteString(source[pos:])
	ctx.OnError("missing-syntax", len(source), "|")
	return NewLiteral(start, len(source), true, &open, value.String(), nil)
}

// ParseVariable parses a variable reference
// TypeScript original code:
// export function parseVariable(
//
//	ctx: ParseContext,
//	start: number
//
//	): CST.VariableRef {
//	  const pos = start + 1;
//	  const open = { start, end: pos, value: '$' as const };
//	  const name = parseNameValue(ctx.source, pos);
//	  if (!name) {
//	    ctx.onError('empty-token', pos, pos + 1);
//	    return { type: 'variable', start, end: pos, open, name: '' };
//	  }
//	  return { type: 'variable', start, end: name.end, open, name: name.value };
//	}
func ParseVariable(ctx *ParseContext, start int) *VariableRef {
	pos := start + 1
	open := NewSyntax(start, pos, "$")

	name := ParseNameValue(ctx.Source(), pos)
	if name == nil {
		ctx.OnError("empty-token", pos, pos+1)
		return NewVariableRef(start, pos, open, "")
	}

	return NewVariableRef(start, name.End, open, name.Value)
}

// EscapeResult represents the result of parsing an escape sequence
type EscapeResult struct {
	Value  string
	Length int
}

// parseEscape parses an escape sequence
// TypeScript original code:
// function parseEscape(
//
//	ctx: ParseContext,
//	start: number
//
//	): { value: string; length: number } | null {
//	  const raw = ctx.source[start + 1];
//	  if ('\\{|}'.includes(raw)) return { value: raw, length: 1 };
//	  if (ctx.resource) {
//	    let hexLen = 0;
//	    switch (raw) {
//	      case '\t':
//	      case ' ':
//	        return { value: raw, length: 1 };
//	      case 'n':
//	        return { value: '\n', length: 1 };
//	      case 'r':
//	        return { value: '\r', length: 1 };
//	      case 't':
//	        return { value: '\t', length: 1 };
//	      case 'u':
//	        hexLen = 4;
//	        break;
//	      case 'U':
//	        hexLen = 6;
//	        break;
//	      case 'x':
//	        hexLen = 2;
//	        break;
//	    }
//	    if (hexLen > 0) {
//	      const h0 = start + 2;
//	      const raw = ctx.source.substring(h0, h0 + hexLen);
//	      if (raw.length === hexLen && /^[0-9A-Fa-f]+$/.test(raw)) {
//	        return {
//	          value: String.fromCharCode(parseInt(raw, 16)),
//	          length: 1 + hexLen
//	        };
//	      }
//	    }
//	  }
//	  ctx.onError('bad-escape', start, start + 2);
//	  return null;
//	}
func parseEscape(ctx *ParseContext, start int) *EscapeResult {
	source := ctx.Source()

	if start+1 >= len(source) {
		ctx.OnError("bad-escape", start, start+2)
		return nil
	}

	raw := source[start+1]

	// Basic escape sequences
	switch raw {
	case '\\', '{', '|', '}':
		return &EscapeResult{Value: string(raw), Length: 1}
	}

	// Resource-specific escape sequences
	if ctx.Resource() {
		switch raw {
		case '\t', ' ':
			return &EscapeResult{Value: string(raw), Length: 1}
		case 'n':
			return &EscapeResult{Value: "\n", Length: 1}
		case 'r':
			return &EscapeResult{Value: "\r", Length: 1}
		case 't':
			return &EscapeResult{Value: "\t", Length: 1}
		case 'u':
			return parseHexEscape(ctx, start, 4)
		case 'U':
			return parseHexEscape(ctx, start, 6)
		case 'x':
			return parseHexEscape(ctx, start, 2)
		}
	}

	ctx.OnError("bad-escape", start, start+2)
	return nil
}

// parseHexEscape parses a hexadecimal escape sequence
func parseHexEscape(ctx *ParseContext, start int, hexLen int) *EscapeResult {
	source := ctx.Source()
	h0 := start + 2

	if h0+hexLen > len(source) {
		ctx.OnError("bad-escape", start, start+2)
		return nil
	}

	raw := source[h0 : h0+hexLen]

	// Check if all characters are hex digits
	for _, ch := range raw {
		if (ch < '0' || ch > '9') && (ch < 'A' || ch > 'F') && (ch < 'a' || ch > 'f') {
			ctx.OnError("bad-escape", start, start+2)
			return nil
		}
	}

	// Parse hex value
	value, err := strconv.ParseInt(raw, 16, 32)
	if err != nil {
		ctx.OnError("bad-escape", start, start+2)
		return nil
	}

	return &EscapeResult{
		Value:  string(rune(value)),
		Length: 1 + hexLen,
	}
}

// ParseSimpleText parses literal text content with escape sequence support
// This is used for simple messages that are not quoted patterns
func ParseSimpleText(ctx *ParseContext, start int) *Text {
	var value strings.Builder
	pos := start
	i := start
	source := ctx.Source()

	for i < len(source) {
		ch := source[i]
		switch ch {
		case '\\':
			esc := parseEscape(ctx, i)
			if esc != nil {
				value.WriteString(source[pos:i])
				value.WriteString(esc.Value)
				i += esc.Length
				pos = i + 1
			}
		case '{':
			// Check for {{ escape sequence
			if i+1 < len(source) && source[i+1] == '{' {
				// This is an escaped {, add the text before it and the literal {
				value.WriteString(source[pos:i])
				value.WriteString("{")
				i += 2 // Skip both {{
				pos = i
				continue
			}
			// Single { means end of text
			goto loop_end
		case '}':
			// Check for }} escape sequence
			if i+1 < len(source) && source[i+1] == '}' {
				// This is an escaped }, add the text before it and the literal }
				value.WriteString(source[pos:i])
				value.WriteString("}")
				i += 2 // Skip both }}
				pos = i
				continue
			}
			// Single } means end of text
			goto loop_end
		case '\n':
			if ctx.Resource() {
				nl := i
				next := byte(0)
				if i+1 < len(source) {
					next = source[i+1]
				}
				for next == ' ' || next == '\t' {
					i++
					if i+1 < len(source) {
						next = source[i+1]
					} else {
						break
					}
				}
				if i > nl {
					value.WriteString(source[pos : nl+1])
					pos = i + 1
				}
			}
		}
		i++
	}

loop_end:
	value.WriteString(source[pos:i])
	return NewText(start, i, value.String())
}
