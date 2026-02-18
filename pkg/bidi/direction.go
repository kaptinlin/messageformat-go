// Package bidi provides bidirectional text direction utilities for MessageFormat 2.0
// TypeScript original code: dir-utils.ts module
package bidi

import (
	"strings"
	"unicode"
)

// Direction represents text direction
// TypeScript original code:
// export type Direction = 'ltr' | 'rtl' | 'auto';
type Direction string

const (
	DirLTR  Direction = "ltr"  // Left-to-right text direction
	DirRTL  Direction = "rtl"  // Right-to-left text direction
	DirAuto Direction = "auto" // Automatic direction detection
)

// Unicode bidirectional control characters
// TypeScript original code: Unicode constants
// TypeScript original code:
// export const LRI = '\u2066'; // Left-to-Right Isolate
// export const RLI = '\u2067'; // Right-to-Left Isolate
// export const FSI = '\u2068'; // First Strong Isolate
// export const PDI = '\u2069'; // Pop Directional Isolate
const (
	LRI = '\u2066' // Left-to-Right Isolate
	RLI = '\u2067' // Right-to-Left Isolate
	FSI = '\u2068' // First Strong Isolate
	PDI = '\u2069' // Pop Directional Isolate
)

// ParseDirection converts a string direction to a Direction type.
// Returns DirAuto for unrecognized values.
func ParseDirection(s string) Direction {
	switch s {
	case "ltr":
		return DirLTR
	case "rtl":
		return DirRTL
	case "auto":
		return DirAuto
	default:
		return DirAuto
	}
}

// GetDirection determines text direction from text content
// TypeScript original code:
//
//	export function getDirection(text: string): Direction {
//	  for (const char of text) {
//	    const dir = getCharDirection(char);
//	    if (dir === 'ltr' || dir === 'rtl') {
//	      return dir;
//	    }
//	  }
//	  return 'auto';
//	}
func GetDirection(text string) Direction {
	for _, r := range text {
		if isRTLChar(r) {
			return DirRTL
		}
		if isLTRChar(r) {
			return DirLTR
		}
	}
	return DirAuto
}

// GetLocaleDirection determines text direction from locale
// TypeScript original code:
//
//	export function getLocaleDirection(locale: string): Direction {
//	  const rtlLanguages = new Set(['ar', 'he', 'fa', 'ur', 'yi']);
//	  const lang = locale.split('-')[0];
//	  return rtlLanguages.has(lang) ? 'rtl' : 'ltr';
//	}
func GetLocaleDirection(locale string) Direction {
	// Extract language code from locale (e.g., "en-US" -> "en")
	parts := strings.Split(locale, "-")
	if len(parts) == 0 {
		return DirLTR
	}

	lang := strings.ToLower(parts[0])

	// Check for RTL languages
	rtlLanguages := map[string]bool{
		"ar": true, // Arabic
		"he": true, // Hebrew
		"fa": true, // Persian/Farsi
		"ur": true, // Urdu
		"yi": true, // Yiddish
	}

	if rtlLanguages[lang] {
		return DirRTL
	}

	return DirLTR
}

// WrapWithIsolation wraps text with appropriate isolation characters
// TypeScript original code:
//
//	export function wrapWithIsolation(text: string, dir: Direction): string {
//	  switch (dir) {
//	    case 'ltr':
//	      return LRI + text + PDI;
//	    case 'rtl':
//	      return RLI + text + PDI;
//	    case 'auto':
//	      return FSI + text + PDI;
//	    default:
//	      return text;
//	  }
//	}
func WrapWithIsolation(text string, dir Direction) string {
	switch dir {
	case DirLTR:
		return string(LRI) + text + string(PDI)
	case DirRTL:
		return string(RLI) + text + string(PDI)
	case DirAuto:
		return string(FSI) + text + string(PDI)
	default:
		return text
	}
}

// isRTLChar checks if a character is right-to-left
func isRTLChar(r rune) bool {
	// Check for Arabic, Hebrew, and other RTL scripts
	switch {
	case r >= 0x0590 && r <= 0x05FF: // Hebrew
		return true
	case r >= 0x0600 && r <= 0x06FF: // Arabic
		return true
	case r >= 0x0700 && r <= 0x074F: // Syriac
		return true
	case r >= 0x0750 && r <= 0x077F: // Arabic Supplement
		return true
	case r >= 0x0780 && r <= 0x07BF: // Thaana
		return true
	case r >= 0x07C0 && r <= 0x07FF: // NKo
		return true
	case r >= 0x0800 && r <= 0x083F: // Samaritan
		return true
	case r >= 0x08A0 && r <= 0x08FF: // Arabic Extended-A
		return true
	case r >= 0xFB1D && r <= 0xFB4F: // Hebrew Presentation Forms
		return true
	case r >= 0xFB50 && r <= 0xFDFF: // Arabic Presentation Forms-A
		return true
	case r >= 0xFE70 && r <= 0xFEFF: // Arabic Presentation Forms-B
		return true
	default:
		return false
	}
}

// isLTRChar checks if a character is left-to-right
func isLTRChar(r rune) bool {
	// Check for Latin, Cyrillic, and other LTR scripts
	switch {
	case r >= 0x0041 && r <= 0x005A: // Basic Latin uppercase
		return true
	case r >= 0x0061 && r <= 0x007A: // Basic Latin lowercase
		return true
	case r >= 0x00C0 && r <= 0x00FF: // Latin-1 Supplement
		return true
	case r >= 0x0100 && r <= 0x017F: // Latin Extended-A
		return true
	case r >= 0x0180 && r <= 0x024F: // Latin Extended-B
		return true
	case r >= 0x0400 && r <= 0x04FF: // Cyrillic
		return true
	case unicode.IsLetter(r): // Other letters are assumed LTR
		return true
	default:
		return false
	}
}

// IsIsolationChar checks if a character is a bidirectional isolation character
func IsIsolationChar(r rune) bool {
	return r == LRI || r == RLI || r == FSI || r == PDI
}
