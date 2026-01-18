// Package cst provides comprehensive tests for names.go
package cst

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseNameValue_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		start       int
		expectedVal string
		expectedEnd int
		shouldBeNil bool
	}{
		{
			name:        "valid simple name",
			source:      "hello",
			start:       0,
			expectedVal: "hello",
			expectedEnd: 5,
		},
		{
			name:        "name with underscore",
			source:      "user_name",
			start:       0,
			expectedVal: "user_name",
			expectedEnd: 9,
		},
		{
			name:        "name with numbers",
			source:      "var123",
			start:       0,
			expectedVal: "var123",
			expectedEnd: 6,
		},
		{
			name:        "name starting with digit (invalid)",
			source:      "123var",
			start:       0,
			shouldBeNil: true,
		},
		{
			name:        "name starting with dash (invalid)",
			source:      "-name",
			start:       0,
			shouldBeNil: true,
		},
		{
			name:        "name starting with dot (invalid)",
			source:      ".name",
			start:       0,
			shouldBeNil: true,
		},
		{
			name:        "empty string",
			source:      "",
			start:       0,
			shouldBeNil: true,
		},
		{
			name:        "start beyond string length",
			source:      "test",
			start:       10,
			shouldBeNil: true,
		},
		{
			name:        "name with bidi characters",
			source:      "\u061Chello\u200E",
			start:       0,
			expectedVal: "hello",
			expectedEnd: 10, // includes bidi chars (3 + 5 + 3 = 11, but actually 3 + 5 + 2 = 10)
		},
		{
			name:        "only bidi characters",
			source:      "\u061C\u200E\u200F",
			start:       0,
			shouldBeNil: true,
		},
		{
			name:        "unicode name",
			source:      "café",
			start:       0,
			expectedVal: "café",
			expectedEnd: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseNameValue(tt.source, tt.start)

			if tt.shouldBeNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedVal, result.Value)
				assert.Equal(t, tt.expectedEnd, result.End)
			}
		})
	}
}

func TestIsValidUnquotedLiteral(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid simple literal",
			input:    "hello",
			expected: true,
		},
		{
			name:     "valid with underscore",
			input:    "hello_world",
			expected: true,
		},
		{
			name:     "valid with numbers",
			input:    "test123",
			expected: true,
		},
		{
			name:     "valid with dash and dot",
			input:    "some-value.key",
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "with space (invalid)",
			input:    "hello world",
			expected: false,
		},
		{
			name:     "with special chars (invalid)",
			input:    "hello@world",
			expected: false,
		},
		{
			name:     "unicode characters",
			input:    "café",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidUnquotedLiteral(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsNameChar(t *testing.T) {
	tests := []struct {
		name     string
		char     rune
		expected bool
	}{
		// ASCII alphanumeric
		{name: "lowercase a", char: 'a', expected: true},
		{name: "uppercase Z", char: 'Z', expected: true},
		{name: "digit 5", char: '5', expected: true},
		// Allowed symbols
		{name: "dash", char: '-', expected: true},
		{name: "dot", char: '.', expected: true},
		{name: "plus", char: '+', expected: true},
		{name: "underscore", char: '_', expected: true},
		// Disallowed symbols
		{name: "space", char: ' ', expected: false},
		{name: "at sign", char: '@', expected: false},
		{name: "hash", char: '#', expected: false},
		// Unicode ranges
		{name: "unicode 0x00A1", char: '\u00A1', expected: true},
		{name: "unicode 0x061B", char: '\u061B', expected: true},
		{name: "unicode 0x061D", char: '\u061D', expected: true},
		{name: "unicode 0x3001", char: '\u3001', expected: true},
		{name: "unicode 0xFFFD", char: '\uFFFD', expected: true},
		// High unicode (outside BMP)
		{name: "high unicode 0x10000", char: '\U00010000', expected: true},
		{name: "high unicode 0x1FFFD", char: '\U0001FFFD', expected: true},
		{name: "high unicode 0x10FFFD", char: '\U0010FFFD', expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNameChar(tt.char)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsNameStartChar(t *testing.T) {
	tests := []struct {
		name     string
		char     rune
		expected bool
	}{
		{name: "letter a", char: 'a', expected: true},
		{name: "letter Z", char: 'Z', expected: true},
		{name: "underscore", char: '_', expected: true},
		{name: "dash (invalid start)", char: '-', expected: false},
		{name: "dot (invalid start)", char: '.', expected: false},
		{name: "digit 0 (invalid start)", char: '0', expected: false},
		{name: "digit 9 (invalid start)", char: '9', expected: false},
		{name: "plus", char: '+', expected: true},
		{name: "unicode char", char: '\u00A1', expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNameStartChar(tt.char)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidNameChar(t *testing.T) {
	tests := []struct {
		name     string
		char     rune
		expected bool
	}{
		{name: "BMP character a", char: 'a', expected: true},
		{name: "BMP character 0", char: '0', expected: true},
		{name: "BMP unicode", char: '\u00A1', expected: true},
		{name: "invalid BMP", char: '\u0000', expected: false},
		{name: "beyond BMP valid", char: '\U00010000', expected: true},
		{name: "beyond BMP high", char: '\U0010FFFD', expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidNameChar(tt.char)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseUnquotedLiteralValue_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		start    int
		expected string
	}{
		{
			name:     "simple value",
			source:   "hello",
			start:    0,
			expected: "hello",
		},
		{
			name:     "stops at space",
			source:   "hello world",
			start:    0,
			expected: "hello",
		},
		{
			name:     "stops at special char",
			source:   "hello@world",
			start:    0,
			expected: "hello",
		},
		{
			name:     "start beyond length",
			source:   "test",
			start:    10,
			expected: "",
		},
		{
			name:     "unicode characters",
			source:   "café123",
			start:    0,
			expected: "café123",
		},
		{
			name:     "with dash and dot",
			source:   "value-1.2.3",
			start:    0,
			expected: "value-1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseUnquotedLiteralValue(tt.source, tt.start)
			assert.Equal(t, tt.expected, result)
		})
	}
}
