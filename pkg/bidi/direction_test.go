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
			expected: DirectionLTR,
		},
		{
			name:     "Arabic text",
			text:     "مرحبا",
			expected: DirectionRTL,
		},
		{
			name:     "Hebrew text",
			text:     "שלום",
			expected: DirectionRTL,
		},
		{
			name:     "Mixed text with LTR first",
			text:     "Hello مرحبا",
			expected: DirectionLTR,
		},
		{
			name:     "Mixed text with RTL first",
			text:     "مرحبا Hello",
			expected: DirectionRTL,
		},
		{
			name:     "Numbers only",
			text:     "12345",
			expected: DirectionAuto,
		},
		{
			name:     "Empty text",
			text:     "",
			expected: DirectionAuto,
		},
		{
			name:     "Punctuation only",
			text:     "!@#$%",
			expected: DirectionAuto,
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
			expected: DirectionLTR,
		},
		{
			name:     "English US locale",
			locale:   "en-US",
			expected: DirectionLTR,
		},
		{
			name:     "Arabic locale",
			locale:   "ar",
			expected: DirectionRTL,
		},
		{
			name:     "Arabic Saudi locale",
			locale:   "ar-SA",
			expected: DirectionRTL,
		},
		{
			name:     "Hebrew locale",
			locale:   "he",
			expected: DirectionRTL,
		},
		{
			name:     "Hebrew Israel locale",
			locale:   "he-IL",
			expected: DirectionRTL,
		},
		{
			name:     "Persian locale",
			locale:   "fa",
			expected: DirectionRTL,
		},
		{
			name:     "Urdu locale",
			locale:   "ur",
			expected: DirectionRTL,
		},
		{
			name:     "Yiddish locale",
			locale:   "yi",
			expected: DirectionRTL,
		},
		{
			name:     "French locale",
			locale:   "fr",
			expected: DirectionLTR,
		},
		{
			name:     "German locale",
			locale:   "de",
			expected: DirectionLTR,
		},
		{
			name:     "Empty locale",
			locale:   "",
			expected: DirectionLTR,
		},
		{
			name:     "Unknown locale",
			locale:   "xx",
			expected: DirectionLTR,
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
			dir:      DirectionLTR,
			expected: string(LRI) + "Hello" + string(PDI),
		},
		{
			name:     "RTL isolation",
			text:     "مرحبا",
			dir:      DirectionRTL,
			expected: string(RLI) + "مرحبا" + string(PDI),
		},
		{
			name:     "Auto isolation",
			text:     "Hello",
			dir:      DirectionAuto,
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
			dir:      DirectionLTR,
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
	assert.Equal(t, "ltr", string(DirectionLTR))
	assert.Equal(t, "rtl", string(DirectionRTL))
	assert.Equal(t, "auto", string(DirectionAuto))
}

func TestUnicodeConstants(t *testing.T) {
	// Test that Unicode constants are properly defined
	assert.Equal(t, '\u2066', LRI)
	assert.Equal(t, '\u2067', RLI)
	assert.Equal(t, '\u2068', FSI)
	assert.Equal(t, '\u2069', PDI)
}
