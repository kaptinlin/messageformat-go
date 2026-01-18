// Package cst provides comprehensive tests for util.go
package cst

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhitespaces_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		start       int
		expectedWS  bool
		expectedEnd int
	}{
		{
			name:        "no whitespace",
			source:      "hello",
			start:       0,
			expectedWS:  false,
			expectedEnd: 0,
		},
		{
			name:        "single space",
			source:      " hello",
			start:       0,
			expectedWS:  true,
			expectedEnd: 1,
		},
		{
			name:        "multiple spaces",
			source:      "   hello",
			start:       0,
			expectedWS:  true,
			expectedEnd: 3,
		},
		{
			name:        "tab",
			source:      "\thello",
			start:       0,
			expectedWS:  true,
			expectedEnd: 1,
		},
		{
			name:        "newline",
			source:      "\nhello",
			start:       0,
			expectedWS:  true,
			expectedEnd: 1,
		},
		{
			name:        "carriage return",
			source:      "\rhello",
			start:       0,
			expectedWS:  true,
			expectedEnd: 1,
		},
		{
			name:        "ideographic space",
			source:      "\u3000hello",
			start:       0,
			expectedWS:  true,
			expectedEnd: 3, // U+3000 is 3 bytes in UTF-8
		},
		{
			name:        "mixed whitespace",
			source:      " \t\n\rhello",
			start:       0,
			expectedWS:  true,
			expectedEnd: 4,
		},
		{
			name:        "bidi chars only",
			source:      "\u061Chello",
			start:       0,
			expectedWS:  false,
			expectedEnd: 2, // U+061C is 2 bytes in UTF-8
		},
		{
			name:        "bidi then whitespace",
			source:      "\u061C hello",
			start:       0,
			expectedWS:  true,
			expectedEnd: 3, // 2 bytes for U+061C + 1 for space
		},
		{
			name:        "whitespace then bidi",
			source:      " \u200Ehello",
			start:       0,
			expectedWS:  true,
			expectedEnd: 4, // 1 for space + 3 for U+200E
		},
		{
			name:        "all bidi chars",
			source:      "\u061C\u200E\u200F\u2066\u2067\u2068\u2069hello",
			start:       0,
			expectedWS:  false,
			expectedEnd: 20, // 2 + 3 + 3 + 3 + 3 + 3 + 3 = 20 bytes
		},
		{
			name:        "start beyond string",
			source:      "test",
			start:       10,
			expectedWS:  false,
			expectedEnd: 10,
		},
		{
			name:        "empty string",
			source:      "",
			start:       0,
			expectedWS:  false,
			expectedEnd: 0,
		},
		{
			name:        "whitespace at offset",
			source:      "hello   world",
			start:       5,
			expectedWS:  true,
			expectedEnd: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Whitespaces(tt.source, tt.start)

			assert.Equal(t, tt.expectedWS, result.HasWS)
			assert.Equal(t, tt.expectedEnd, result.End)
		})
	}
}

func TestIsBidiChar(t *testing.T) {
	tests := []struct {
		name     string
		char     rune
		expected bool
	}{
		{
			name:     "Arabic Letter Mark",
			char:     '\u061C',
			expected: true,
		},
		{
			name:     "Left-to-Right Mark",
			char:     '\u200E',
			expected: true,
		},
		{
			name:     "Right-to-Left Mark",
			char:     '\u200F',
			expected: true,
		},
		{
			name:     "Left-to-Right Isolate",
			char:     '\u2066',
			expected: true,
		},
		{
			name:     "Right-to-Left Isolate",
			char:     '\u2067',
			expected: true,
		},
		{
			name:     "First Strong Isolate",
			char:     '\u2068',
			expected: true,
		},
		{
			name:     "Pop Directional Isolate",
			char:     '\u2069',
			expected: true,
		},
		{
			name:     "regular char 'a'",
			char:     'a',
			expected: false,
		},
		{
			name:     "space",
			char:     ' ',
			expected: false,
		},
		{
			name:     "tab",
			char:     '\t',
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBidiChar(tt.char)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsWhitespaceChar(t *testing.T) {
	tests := []struct {
		name     string
		char     rune
		expected bool
	}{
		{
			name:     "space",
			char:     ' ',
			expected: true,
		},
		{
			name:     "tab",
			char:     '\t',
			expected: true,
		},
		{
			name:     "newline",
			char:     '\n',
			expected: true,
		},
		{
			name:     "carriage return",
			char:     '\r',
			expected: true,
		},
		{
			name:     "ideographic space",
			char:     '\u3000',
			expected: true,
		},
		{
			name:     "regular char 'a'",
			char:     'a',
			expected: false,
		},
		{
			name:     "digit '0'",
			char:     '0',
			expected: false,
		},
		{
			name:     "unicode letter",
			char:     'é',
			expected: false,
		},
		{
			name:     "unicode whitespace (non-breaking space)",
			char:     '\u00A0',
			expected: true, // unicode.IsSpace catches this
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWhitespaceChar(tt.char)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWhitespaces_InvalidUTF8(t *testing.T) {
	// Test with invalid UTF-8 sequence
	source := "hello\xFF\xFEworld"
	result := Whitespaces(source, 5)

	// Should stop at invalid UTF-8
	assert.False(t, result.HasWS)
	assert.Equal(t, 5, result.End)
}

func TestWhitespaces_ComplexMixtures(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		start       int
		expectedWS  bool
		expectedEnd int
	}{
		{
			name:        "bidi + ws + bidi",
			source:      "\u200E \u200F",
			start:       0,
			expectedWS:  true,
			expectedEnd: 7, // 3 bytes for U+200E + 1 byte space + 3 bytes for U+200F
		},
		{
			name:        "ws + bidi + ws",
			source:      " \u061C\thello",
			start:       0,
			expectedWS:  true,
			expectedEnd: 4, // 1 + 2 + 1 = 4
		},
		{
			name:        "multiple bidi mixed with ws",
			source:      "\u2066 \u2067\t\u2068\nhello",
			start:       0,
			expectedWS:  true,
			expectedEnd: 12, // 3 + 1 + 3 + 1 + 3 + 1 = 12
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Whitespaces(tt.source, tt.start)

			assert.Equal(t, tt.expectedWS, result.HasWS)
			assert.Equal(t, tt.expectedEnd, result.End)
		})
	}
}
