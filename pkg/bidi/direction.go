// Package bidi provides bidirectional text direction utilities for MessageFormat 2.0.
package bidi

import (
	intllocale "github.com/agentable/go-intl/locale"

	"golang.org/x/text/language"
	unicodebidi "golang.org/x/text/unicode/bidi"
)

// Direction represents text direction.
type Direction string

const (
	DirLTR  Direction = "ltr"  // Left-to-right text direction
	DirRTL  Direction = "rtl"  // Right-to-left text direction
	DirAuto Direction = "auto" // Automatic direction detection
)

// Unicode bidirectional control characters.
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

// GetDirection returns the direction of the first strong character in text.
func GetDirection(text string) Direction {
	for _, r := range text {
		properties, _ := unicodebidi.LookupRune(r)
		switch properties.Class() {
		case unicodebidi.L:
			return DirLTR
		case unicodebidi.R, unicodebidi.AL:
			return DirRTL
		default:
		}
	}
	return DirAuto
}

// GetLocaleDirection returns the direction of a locale's explicit or inferred script.
// It returns DirAuto when the locale or its script cannot be determined.
func GetLocaleDirection(localeID string) Direction {
	loc, err := intllocale.Parse(localeID)
	if err != nil {
		return DirAuto
	}
	script, confidence := loc.Tag().Script()
	if confidence == language.No || script.String() == "Zzzz" {
		return DirAuto
	}
	direction := loc.GetTextInfo().Direction
	if direction == nil {
		return DirAuto
	}
	return ParseDirection(*direction)
}

// WrapWithIsolation wraps text with isolation characters for dir.
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

// IsIsolationChar checks if a character is a bidirectional isolation character
func IsIsolationChar(r rune) bool {
	return r == LRI || r == RLI || r == FSI || r == PDI
}
