package functions

import (
	"fmt"
	"strings"
)

// MaxOptionKeyLength defines the maximum allowed length for an option key
const MaxOptionKeyLength = 100

// MaxOptionsCount defines the maximum number of options allowed to prevent DoS
const MaxOptionsCount = 50

// ValidateOptionKey validates an option key name to prevent security issues.
// This function prevents potential injection attacks or malformed keys.
//
// Security checks performed:
// 1. Key length validation (max 100 characters)
// 2. Character whitelist (alphanumeric, underscore, hyphen only)
// 3. Forbidden key names (dangerous JavaScript-like keys)
//
// Reference: Inspired by TypeScript fix for prototype pollution (commit 82cd10b4)
// https://github.com/messageformat/messageformat/commit/82cd10b40e3f922f990bbcf88a6d14b70c0a3ce0
func ValidateOptionKey(key string) error {
	// Check key length
	if len(key) > MaxOptionKeyLength {
		return fmt.Errorf("option key too long: %d characters (max: %d)", len(key), MaxOptionKeyLength)
	}

	if len(key) == 0 {
		return fmt.Errorf("option key cannot be empty")
	}

	// Check character whitelist
	for i, ch := range key {
		// Allow: a-z, A-Z, 0-9, underscore, hyphen
		// Disallow: special characters, control characters, etc.
		if !isValidOptionKeyChar(ch) {
			return fmt.Errorf("invalid character '%c' at position %d in option key '%s'", ch, i, key)
		}
	}

	// Check for forbidden key names (similar to prototype pollution prevention)
	// While Go doesn't have prototype pollution, we still want to prevent
	// any potential confusion with reserved keywords or internal fields
	forbiddenKeys := []string{
		"__proto__",
		"constructor",
		"prototype",
		"__defineGetter__",
		"__defineSetter__",
		"__lookupGetter__",
		"__lookupSetter__",
	}

	lowerKey := strings.ToLower(key)
	for _, forbidden := range forbiddenKeys {
		if lowerKey == forbidden {
			return fmt.Errorf("forbidden option key: '%s'", key)
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
		ch == '-'
}

// ValidateOptions validates an entire options map.
// This prevents DoS attacks through excessive options and validates all keys.
//
// Reference: Based on security best practices from TypeScript implementation
func ValidateOptions(options map[string]interface{}) error {
	// Check options count to prevent DoS
	if len(options) > MaxOptionsCount {
		return fmt.Errorf("too many options: %d (max: %d)", len(options), MaxOptionsCount)
	}

	// Validate each key
	for key := range options {
		if err := ValidateOptionKey(key); err != nil {
			return fmt.Errorf("invalid option: %w", err)
		}
	}

	return nil
}

// SanitizeOptions creates a sanitized copy of options map, filtering out invalid keys.
// This is useful when you want to be permissive but still protect against malicious input.
//
// Returns a new map containing only valid options.
func SanitizeOptions(options map[string]interface{}) map[string]interface{} {
	if options == nil {
		return nil
	}

	sanitized := make(map[string]interface{})
	for key, value := range options {
		if ValidateOptionKey(key) == nil {
			sanitized[key] = value
		}
		// Silently skip invalid keys
	}

	return sanitized
}
