// Package cst provides name parsing utilities for CST
// TypeScript original code: cst/names.ts module
package cst

import (
	"regexp"
	"strings"

	"golang.org/x/text/unicode/norm"
)

// bidiCharsRegex matches bidirectional control characters
// TypeScript original code:
// const bidiChars = /^[\u061c\u200e\u200f\u2066-\u2069]+/;
var bidiCharsRegex = regexp.MustCompile("^[\u061C\u200E\u200F\u2066-\u2069]+")

// nameCharsRegex matches valid name characters
// TypeScript original code:
// const nameChars = /^[-.+0-9A-Z_a-z\u{a1}-\u{61b}\u{61d}-\u{167f}\u{1681}-\u{1fff}\u{200b}-\u{200d}\u{2010}-\u{2027}\u{2030}-\u{205e}\u{2060}-\u{2065}\u{206a}-\u{2fff}\u{3001}-\u{d7ff}\u{e000}-\u{fdcf}\u{fdf0}-\u{fffd}\u{10000}-\u{1fffd}\u{20000}-\u{2fffd}\u{30000}-\u{3fffd}\u{40000}-\u{4fffd}\u{50000}-\u{5fffd}\u{60000}-\u{6fffd}\u{70000}-\u{7fffd}\u{80000}-\u{8fffd}\u{90000}-\u{9fffd}\u{a0000}-\u{afffd}\u{b0000}-\u{bfffd}\u{c0000}-\u{cfffd}\u{d0000}-\u{dfffd}\u{e0000}-\u{efffd}\u{f0000}-\u{ffffd}\u{100000}-\u{10fffd}]+/u;
// Note: Go regex doesn't support Unicode code points > \uFFFF in character classes, so we use the BMP range
var nameCharsRegex = regexp.MustCompile("^[-.+0-9A-Z_a-z\u00A1-\u061B\u061D-\u167F\u1681-\u1FFF\u200B-\u200D\u2010-\u2027\u2030-\u205E\u2060-\u2065\u206A-\u2FFF\u3001-\uD7FF\uE000-\uFDCF\uFDF0-\uFFFD]+")

// notNameStartRegex matches characters that cannot start a name
// TypeScript original code:
// const notNameStart = /^[-.0-9]/;
var notNameStartRegex = regexp.MustCompile(`^[-.0-9]`)

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

	// Skip initial bidi characters
	if match := bidiCharsRegex.FindString(source[pos:]); match != "" {
		pos += len(match)
	}

	if pos >= len(source) {
		return nil
	}

	// Match name characters
	match := nameCharsRegex.FindString(source[pos:])
	if match == "" {
		return nil
	}

	// Check if name starts with invalid characters
	if notNameStartRegex.MatchString(match) {
		return nil
	}

	name := match
	pos += len(match)

	// Skip ending bidi characters
	if pos < len(source) {
		if endMatch := bidiCharsRegex.FindString(source[pos:]); endMatch != "" {
			pos += len(endMatch)
		}
	}

	// Normalize the name (Unicode NFC normalization)
	// TypeScript: name.normalize() - applies NFC normalization
	normalizedName := norm.NFC.String(name)

	return &NameValue{
		Value: normalizedName,
		End:   pos,
	}
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
		if !isValidNameChar(r) {
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

	// Check character by character to handle both BMP and non-BMP characters
	var result strings.Builder
	for _, r := range source[start:] {
		if isValidNameChar(r) {
			result.WriteRune(r)
		} else {
			break
		}
	}

	return result.String()
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

// For characters outside BMP, we'll use a fallback function
func isValidNameChar(r rune) bool {
	// First check if it matches the basic regex pattern
	if r <= 0xFFFF {
		return nameCharsRegex.MatchString(string(r))
	}

	// For characters outside BMP, use the isNameChar function
	return isNameChar(r)
}
