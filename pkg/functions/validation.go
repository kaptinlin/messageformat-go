package functions

import "fmt"

// MaxOptionKeyLength defines the maximum allowed length for an option key
const MaxOptionKeyLength = 100

// MaxOptionsCount defines the maximum number of options allowed in a function call.
const MaxOptionsCount = 50

// ValidateOptionKey validates an option key name.
// It follows the identifier shape accepted by this package's option boundary:
// ASCII alphanumeric characters plus underscore, hyphen, and namespace colon.
func ValidateOptionKey(key string) error {
	if len(key) > MaxOptionKeyLength {
		return fmt.Errorf("option key too long: %d characters (max: %d)", len(key), MaxOptionKeyLength)
	}

	if len(key) == 0 {
		return fmt.Errorf("option key cannot be empty")
	}

	for i, ch := range key {
		if !isValidOptionKeyChar(ch) {
			return fmt.Errorf("invalid character '%c' at position %d in option key '%s'", ch, i, key)
		}
	}

	return nil
}

// isValidOptionKeyChar checks if a character is valid for an option key
func isValidOptionKeyChar(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '_' ||
		ch == '-' ||
		ch == ':'
}

// ValidateOptions validates an entire options map.
// It validates key shape and caps option count at the package boundary.
func ValidateOptions(options Options) error {
	if len(options) > MaxOptionsCount {
		return fmt.Errorf("too many options: %d (max: %d)", len(options), MaxOptionsCount)
	}

	for key := range options {
		if err := ValidateOptionKey(key); err != nil {
			return fmt.Errorf("invalid option: %w", err)
		}
	}

	return nil
}

// SanitizeOptions creates a sanitized copy of options map, filtering out invalid keys.
// This is useful when you want to be permissive while preserving option-key rules.
func SanitizeOptions(options Options) map[string]any {
	if options == nil {
		return nil
	}

	sanitized := make(map[string]any)
	for key, value := range options {
		if ValidateOptionKey(key) == nil {
			sanitized[key] = value
		}
	}

	return sanitized
}
