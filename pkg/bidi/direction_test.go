package bidi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDirection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		want Direction
	}{
		{name: "ltr", text: "ltr", want: DirLTR},
		{name: "rtl", text: "rtl", want: DirRTL},
		{name: "auto", text: "auto", want: DirAuto},
		{name: "unknown defaults to auto", text: "sideways", want: DirAuto},
		{name: "empty defaults to auto", text: "", want: DirAuto},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, ParseDirection(tt.text))
		})
	}
}

func TestGetDirection(t *testing.T) {
	t.Parallel()

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
			name:     "Latin supplement text",
			text:     "Éclair",
			expected: DirLTR,
		},
		{
			name:     "Cyrillic text",
			text:     "Привет",
			expected: DirLTR,
		},
		{
			name:     "CJK text",
			text:     "世界",
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
			name:     "Syriac text",
			text:     "ܫܠܡܐ",
			expected: DirRTL,
		},
		{
			name:     "Thaana text",
			text:     "ދިވެހި",
			expected: DirRTL,
		},
		{
			name:     "NKo text",
			text:     "ߒߞߏ",
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
		{
			name:     "Neutral prefix before LTR",
			text:     "123 ! Hello",
			expected: DirLTR,
		},
		{
			name:     "Neutral prefix before RTL",
			text:     "… مرحبا",
			expected: DirRTL,
		},
		{
			name:     "Isolation prefix before LTR",
			text:     string(FSI) + "Hello",
			expected: DirLTR,
		},
		{
			name:     "RTL text wrapped in punctuation",
			text:     "(שלום)",
			expected: DirRTL,
		},
		{
			name:     "LTR text wrapped in punctuation",
			text:     "[Hello]",
			expected: DirLTR,
		},
		{
			name:     "Arabic supplement text",
			text:     "ݐ",
			expected: DirRTL,
		},
		{
			name:     "Samaritan text",
			text:     "ࠀ",
			expected: DirRTL,
		},
		{
			name:     "Hebrew presentation form text",
			text:     "יִ",
			expected: DirRTL,
		},
		{
			name:     "Arabic presentation form A text",
			text:     "ﭐ",
			expected: DirRTL,
		},
		{
			name:     "Arabic presentation form B text",
			text:     "ﹰ",
			expected: DirRTL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, GetDirection(tt.text))
		})
	}
}

func TestGetLocaleDirection(t *testing.T) {
	t.Parallel()

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
			name:     "Arabic Saudi locale with uppercase language",
			locale:   "AR-SA",
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
			name:     "Persian Afghanistan locale",
			locale:   "fa-AF",
			expected: DirRTL,
		},
		{
			name:     "Urdu India locale",
			locale:   "ur-IN",
			expected: DirRTL,
		},
		{
			name:     "Yiddish region locale",
			locale:   "yi-001",
			expected: DirRTL,
		},
		{
			name:     "English script region locale",
			locale:   "en-Latn-US",
			expected: DirLTR,
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
			t.Parallel()

			assert.Equal(t, tt.expected, GetLocaleDirection(tt.locale))
		})
	}
}

func TestWrapWithIsolation(t *testing.T) {
	t.Parallel()

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
		{
			name:     "Empty text with RTL",
			text:     "",
			dir:      DirRTL,
			expected: string(RLI) + string(PDI),
		},
		{
			name:     "Empty text with auto",
			text:     "",
			dir:      DirAuto,
			expected: string(FSI) + string(PDI),
		},
		{
			name:     "Invalid direction preserves bidi text",
			text:     "مرحبا",
			dir:      Direction("invalid"),
			expected: "مرحبا",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, WrapWithIsolation(tt.text, tt.dir))
		})
	}
}

func TestIsIsolationChar(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			assert.Equal(t, tt.expected, IsIsolationChar(tt.char))
		})
	}
}
