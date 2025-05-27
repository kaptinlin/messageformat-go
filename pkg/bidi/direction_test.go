package bidi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDirection(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected Direction
	}{
		{
			name:     "English text",
			text:     "Hello World",
			expected: DirLTR,
		},
		{
			name:     "Arabic text",
			text:     "مرحبا",
			expected: DirRTL,
		},
		{
			name:     "Hebrew text",
			text:     "שלום",
			expected: DirRTL,
		},
		{
			name:     "Mixed text with LTR first",
			text:     "Hello مرحبا",
			expected: DirLTR,
		},
		{
			name:     "Mixed text with RTL first",
			text:     "مرحبا Hello",
			expected: DirRTL,
		},
		{
			name:     "Numbers only",
			text:     "12345",
			expected: DirAuto,
		},
		{
			name:     "Empty text",
			text:     "",
			expected: DirAuto,
		},
		{
			name:     "Punctuation only",
			text:     "!@#$%",
			expected: DirAuto,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDirection(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetLocaleDirection(t *testing.T) {
	tests := []struct {
		name     string
		locale   string
		expected Direction
	}{
		{
			name:     "English locale",
			locale:   "en",
			expected: DirLTR,
		},
		{
			name:     "English US locale",
			locale:   "en-US",
			expected: DirLTR,
		},
		{
			name:     "Arabic locale",
			locale:   "ar",
			expected: DirRTL,
		},
		{
			name:     "Arabic Saudi locale",
			locale:   "ar-SA",
			expected: DirRTL,
		},
		{
			name:     "Hebrew locale",
			locale:   "he",
			expected: DirRTL,
		},
		{
			name:     "Hebrew Israel locale",
			locale:   "he-IL",
			expected: DirRTL,
		},
		{
			name:     "Persian locale",
			locale:   "fa",
			expected: DirRTL,
		},
		{
			name:     "Urdu locale",
			locale:   "ur",
			expected: DirRTL,
		},
		{
			name:     "Yiddish locale",
			locale:   "yi",
			expected: DirRTL,
		},
		{
			name:     "French locale",
			locale:   "fr",
			expected: DirLTR,
		},
		{
			name:     "German locale",
			locale:   "de",
			expected: DirLTR,
		},
		{
			name:     "Empty locale",
			locale:   "",
			expected: DirLTR,
		},
		{
			name:     "Unknown locale",
			locale:   "xx",
			expected: DirLTR,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetLocaleDirection(tt.locale)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWrapWithIsolation(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		dir      Direction
		expected string
	}{
		{
			name:     "LTR isolation",
			text:     "Hello",
			dir:      DirLTR,
			expected: string(LRI) + "Hello" + string(PDI),
		},
		{
			name:     "RTL isolation",
			text:     "مرحبا",
			dir:      DirRTL,
			expected: string(RLI) + "مرحبا" + string(PDI),
		},
		{
			name:     "Auto isolation",
			text:     "Hello",
			dir:      DirAuto,
			expected: string(FSI) + "Hello" + string(PDI),
		},
		{
			name:     "Invalid direction",
			text:     "Hello",
			dir:      Direction("invalid"),
			expected: "Hello",
		},
		{
			name:     "Empty text with LTR",
			text:     "",
			dir:      DirLTR,
			expected: string(LRI) + string(PDI),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapWithIsolation(tt.text, tt.dir)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsIsolationChar(t *testing.T) {
	tests := []struct {
		name     string
		char     rune
		expected bool
	}{
		{
			name:     "LRI character",
			char:     LRI,
			expected: true,
		},
		{
			name:     "RLI character",
			char:     RLI,
			expected: true,
		},
		{
			name:     "FSI character",
			char:     FSI,
			expected: true,
		},
		{
			name:     "PDI character",
			char:     PDI,
			expected: true,
		},
		{
			name:     "Regular letter",
			char:     'A',
			expected: false,
		},
		{
			name:     "Arabic letter",
			char:     'ا',
			expected: false,
		},
		{
			name:     "Number",
			char:     '1',
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsIsolationChar(tt.char)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDirectionConstants(t *testing.T) {
	// Test that direction constants are properly defined
	assert.Equal(t, "ltr", string(DirLTR))
	assert.Equal(t, "rtl", string(DirRTL))
	assert.Equal(t, "auto", string(DirAuto))
}

func TestUnicodeConstants(t *testing.T) {
	// Test that Unicode constants are properly defined
	assert.Equal(t, '\u2066', LRI)
	assert.Equal(t, '\u2067', RLI)
	assert.Equal(t, '\u2068', FSI)
	assert.Equal(t, '\u2069', PDI)
}
