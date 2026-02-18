package functions

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ErrNotBoolean indicates the value cannot be converted to a boolean.
var ErrNotBoolean = errors.New("not a boolean")

// ErrNotPositiveInteger indicates the value cannot be converted to a non-negative integer.
var ErrNotPositiveInteger = errors.New("not a positive integer")

// ErrNotString indicates the value cannot be converted to a string.
var ErrNotString = errors.New("not a string")

// Pre-compiled regex for positive integer validation
var positiveIntegerRegex = regexp.MustCompile(`^(0|[1-9][0-9]*)$`)

// asBoolean casts a value as a Boolean, unwrapping objects using their valueOf() methods
// TypeScript original code:
//
//	export function asBoolean(value: unknown): boolean {
//	  if (value && typeof value === 'object') value = value.valueOf();
//	  if (typeof value === 'boolean') return value;
//	  if (value && typeof value === 'object') value = String(value);
//	  if (value === 'true') return true;
//	  if (value === 'false') return false;
//	  throw new RangeError('Not a boolean');
//	}
func asBoolean(value any) (bool, error) {
	// Unwrap objects with valueOf method
	if obj, ok := value.(map[string]any); ok {
		if valueOf, hasValueOf := obj["valueOf"]; hasValueOf {
			value = valueOf
		}
	}

	// Check boolean type
	if b, ok := value.(bool); ok {
		return b, nil
	}

	// Convert to string and check
	str := fmt.Sprintf("%v", value)
	if str == "true" {
		return true, nil
	}
	if str == "false" {
		return false, nil
	}

	return false, ErrNotBoolean
}

// asPositiveInteger casts a value as a non-negative integer
// TypeScript original code:
//
//	export function asPositiveInteger(value: unknown): number {
//	  if (value && typeof value === 'object') value = value.valueOf();
//	  if (value && typeof value === 'object') value = String(value);
//	  if (typeof value === 'string' && /^(0|[1-9][0-9]*)$/.test(value)) {
//	    value = Number(value);
//	  }
//	  if (typeof value === 'number' && value >= 0 && Number.isInteger(value)) {
//	    return value;
//	  }
//	  throw new RangeError('Not a positive integer');
//	}
func asPositiveInteger(value any) (int, error) {
	// Unwrap objects with valueOf method
	if obj, ok := value.(map[string]any); ok {
		if valueOf, hasValueOf := obj["valueOf"]; hasValueOf {
			value = valueOf
		}
	}

	// Handle different numeric types
	switch v := value.(type) {
	case int:
		if v >= 0 {
			return v, nil
		}
	case int64:
		if v >= 0 {
			return int(v), nil
		}
	case float64:
		if v >= 0 && v == float64(int(v)) {
			return int(v), nil
		}
	case string:
		// Check if string matches positive integer pattern
		matched := positiveIntegerRegex.MatchString(v)
		if matched {
			if intVal, err := strconv.Atoi(v); err == nil && intVal >= 0 {
				return intVal, nil
			}
		}
	}

	return 0, ErrNotPositiveInteger
}

// asString casts a value as a string, unwrapping objects using their valueOf() methods
// TypeScript original code:
//
//	export function asString(value: unknown): string {
//	  if (value && typeof value === 'object') value = value.valueOf();
//	  if (typeof value === 'string') return value;
//	  if (value && typeof value === 'object') return String(value);
//	  throw new RangeError('Not a string');
//	}
func asString(value any) (string, error) {
	// Unwrap objects with valueOf method
	if obj, ok := value.(map[string]any); ok {
		if valueOf, hasValueOf := obj["valueOf"]; hasValueOf {
			value = valueOf
		}
	}

	// Check string type
	if str, ok := value.(string); ok {
		return str, nil
	}

	// For non-string types, return error to match TypeScript behavior
	return "", ErrNotString
}

// getStringOption safely gets a string option with a default value
func getStringOption(options map[string]any, name, defaultValue string) string {
	if val, ok := options[name]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

// getIntOption safely gets an integer option with a default value
func getIntOption(options map[string]any, name string, defaultValue int) int {
	if val, ok := options[name]; ok {
		if intVal, err := asPositiveInteger(val); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getBoolOption safely gets a boolean option with a default value
func getBoolOption(options map[string]any, name string, defaultValue bool) bool {
	if val, ok := options[name]; ok {
		if boolVal, err := asBoolean(val); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

// GetFirstLocale returns the first locale from a list, or "en" as fallback
func GetFirstLocale(locales []string) string {
	if len(locales) > 0 {
		return locales[0]
	}
	return "en"
}

// normalizeLocale normalizes a locale string by taking the primary language tag
func normalizeLocale(locale string) string {
	// Handle empty string
	if locale == "" {
		return "en"
	}

	// Use strings.Cut to extract language part before hyphen (Go 1.20+)
	lang, _, _ := strings.Cut(locale, "-")
	if lang != "" {
		return strings.ToLower(lang)
	}
	return "en"
}
