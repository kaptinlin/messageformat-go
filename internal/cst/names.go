// Package cst provides name parsing utilities for CST
// TypeScript original code: cst/names.ts module
package cst

import (
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// isBidiControl reports whether r is a bidirectional control character.
func isBidiControl(r rune) bool {
	switch r {
	case 0x061C, 0x200E, 0x200F:
		return true
	}
	return r >= 0x2066 && r <= 0x2069
}

// NameValue represents a parsed name value with its end position
type NameValue struct {
	Value string
	End   int
}

// ParseNameValue parses a name value from the source at the given position
// TypeScript original code:
// export function parseNameValue(
//
//	source: string,
//	start: number
//
//	): { value: string; end: number } | null {
//	  let pos = start;
//	  const startBidi = source.slice(pos).match(bidiChars);
//	  if (startBidi) pos += startBidi[0].length;
//
//	  const match = source.slice(pos).match(nameChars);
//	  if (!match) return null;
//	  const name = match[0];
//	  if (notNameStart.test(name)) return null;
//	  pos += name.length;
//
//	  const endBidi = source.slice(pos).match(bidiChars);
//	  if (endBidi) pos += endBidi[0].length;
//
//	  return { value: name.normalize(), end: pos };
//	}
func ParseNameValue(source string, start int) *NameValue {
	pos := start

	if pos >= len(source) {
		return nil
	}

	pos = skipBidiControls(source, pos)
	if pos >= len(source) {
		return nil
	}

	name, end := parseName(source, pos)
	if name == "" {
		return nil
	}
	first, _ := utf8.DecodeRuneInString(name)
	if !IsNameStartChar(first) {
		return nil
	}

	end = skipBidiControls(source, end)

	return &NameValue{
		Value: norm.NFC.String(name),
		End:   end,
	}
}

func skipBidiControls(source string, start int) int {
	pos := start
	for pos < len(source) {
		r, size := utf8.DecodeRuneInString(source[pos:])
		if !isBidiControl(r) {
			break
		}
		pos += size
	}
	return pos
}

func parseName(source string, start int) (string, int) {
	pos := start
	var result strings.Builder
	for pos < len(source) {
		r, size := utf8.DecodeRuneInString(source[pos:])
		if !isNameChar(r) {
			break
		}
		result.WriteRune(r)
		pos += size
	}
	return result.String(), pos
}

// IsValidUnquotedLiteral checks if a string is a valid unquoted literal
// TypeScript original code:
//
//	export function isValidUnquotedLiteral(str: string): boolean {
//	  const match = str.match(nameChars);
//	  return !!match && match[0].length === str.length;
//	}
func IsValidUnquotedLiteral(str string) bool {
	if str == "" {
		return false
	}

	// Check each character to handle both BMP and non-BMP characters
	for _, r := range str {
		if !isNameChar(r) {
			return false
		}
	}

	return true
}

// ParseUnquotedLiteralValue parses an unquoted literal value
// TypeScript original code:
// export const parseUnquotedLiteralValue = (
//
//	source: string,
//	start: number
//
// ): string => source.slice(start).match(nameChars)?.[0] ?? "";
func ParseUnquotedLiteralValue(source string, start int) string {
	if start >= len(source) {
		return ""
	}

	value, _ := parseName(source, start)
	return value
}

// isNameChar checks if a rune is a valid name character
func isNameChar(r rune) bool {
	// Basic ASCII alphanumeric and allowed symbols
	if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
		return true
	}
	if r == '-' || r == '.' || r == '+' || r == '_' {
		return true
	}

	// Unicode ranges (according to ABNF spec)
	// %xA1-61B but omit BidiControl: %x61C
	if r >= 0xa1 && r <= 0x61b {
		return true
	}
	// Skip %x61C (Arabic Letter Mark - bidirectional control character)
	if r >= 0x61d && r <= 0x167f {
		return true
	}
	if r >= 0x1681 && r <= 0x1fff {
		return true
	}
	if r >= 0x200b && r <= 0x200d {
		return true
	}
	// Skip BidiControl: %x200E-200F (LRM, RLM)
	if r >= 0x2010 && r <= 0x2027 {
		return true
	}
	// Skip Whitespace: %x2028-2029 %x202F, BidiControl: %x202A-202E
	if r >= 0x2030 && r <= 0x205e {
		return true
	}
	// Skip Whitespace: %x205F
	if r >= 0x2060 && r <= 0x2065 {
		return true
	}
	// Skip BidiControl: %x2066-2069 (LRI, RLI, FSI, PDI)
	if r >= 0x206a && r <= 0x2fff {
		return true
	}
	if r >= 0x3001 && r <= 0xd7ff {
		return true
	}
	if r >= 0xe000 && r <= 0xfdcf {
		return true
	}
	if r >= 0xfdf0 && r <= 0xfffd {
		return true
	}
	if r >= 0x10000 && r <= 0x1fffd {
		return true
	}
	if r >= 0x20000 && r <= 0x2fffd {
		return true
	}
	if r >= 0x30000 && r <= 0x3fffd {
		return true
	}
	if r >= 0x40000 && r <= 0x4fffd {
		return true
	}
	if r >= 0x50000 && r <= 0x5fffd {
		return true
	}
	if r >= 0x60000 && r <= 0x6fffd {
		return true
	}
	if r >= 0x70000 && r <= 0x7fffd {
		return true
	}
	if r >= 0x80000 && r <= 0x8fffd {
		return true
	}
	if r >= 0x90000 && r <= 0x9fffd {
		return true
	}
	if r >= 0xa0000 && r <= 0xafffd {
		return true
	}
	if r >= 0xb0000 && r <= 0xbfffd {
		return true
	}
	if r >= 0xc0000 && r <= 0xcfffd {
		return true
	}
	if r >= 0xd0000 && r <= 0xdfffd {
		return true
	}
	if r >= 0xe0000 && r <= 0xefffd {
		return true
	}
	if r >= 0xf0000 && r <= 0xffffd {
		return true
	}
	if r >= 0x100000 && r <= 0x10fffd {
		return true
	}

	return false
}

// IsNameStartChar checks if a character can start a name
func IsNameStartChar(r rune) bool {
	// Cannot start with -, ., or digits
	if r == '-' || r == '.' || (r >= '0' && r <= '9') {
		return false
	}
	return isNameChar(r)
}
