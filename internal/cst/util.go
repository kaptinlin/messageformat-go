// Package cst provides utility functions for CST parsing
// TypeScript original code: cst/util.ts module
package cst

import (
	"unicode"
	"unicode/utf8"
)

// bidiChars represents bidirectional text control characters
// TypeScript original code:
// const bidiChars = new Set('\u061C\u200E\u200F\u2066\u2067\u2068\u2069');
var bidiChars = map[rune]bool{
	'\u061C': true, // Arabic Letter Mark
	'\u200E': true, // Left-to-Right Mark
	'\u200F': true, // Right-to-Left Mark
	'\u2066': true, // Left-to-Right Isolate
	'\u2067': true, // Right-to-Left Isolate
	'\u2068': true, // First Strong Isolate
	'\u2069': true, // Pop Directional Isolate
}

// whitespaceChars represents whitespace characters
// TypeScript original code:
// const whitespaceChars = new Set('\t\n\r \u3000');
var whitespaceChars = map[rune]bool{
	'\t':     true, // Tab
	'\n':     true, // Line Feed
	'\r':     true, // Carriage Return
	' ':      true, // Space
	'\u3000': true, // Ideographic Space
}

// WhitespaceResult represents the result of whitespace parsing
type WhitespaceResult struct {
	HasWS bool
	End   int
}

// Whitespaces parses whitespace and bidi characters from the given position
// TypeScript original code:
// export function whitespaces(
//
//	src: string,
//	start: number
//
//	): { hasWS: boolean; end: number } {
//	  let hasWS = false;
//	  let pos = start;
//	  let ch = src[pos];
//	  while (bidiChars.has(ch)) ch = src[++pos];
//	  while (whitespaceChars.has(ch)) {
//	    hasWS = true;
//	    ch = src[++pos];
//	  }
//	  while (bidiChars.has(ch) || whitespaceChars.has(ch)) ch = src[++pos];
//	  return { hasWS, end: pos };
//	}
func Whitespaces(src string, start int) WhitespaceResult {
	hasWS := false
	pos := start

	// Use byte-based iteration to maintain correct byte positions
	for pos < len(src) {
		r, size := utf8.DecodeRuneInString(src[pos:])
		if r == utf8.RuneError {
			break
		}

		if bidiChars[r] {
			pos += size
			continue
		}

		if whitespaceChars[r] {
			hasWS = true
			pos += size
			continue
		}

		// If it's neither bidi nor whitespace, we're done
		break
	}

	return WhitespaceResult{HasWS: hasWS, End: pos}
}

// IsBidiChar checks if a character is a bidirectional control character
func IsBidiChar(ch rune) bool {
	return bidiChars[ch]
}

// IsWhitespaceChar checks if a character is a whitespace character
func IsWhitespaceChar(ch rune) bool {
	return whitespaceChars[ch] || unicode.IsSpace(ch)
}
