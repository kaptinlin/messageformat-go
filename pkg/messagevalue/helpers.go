package messagevalue

import (
	"fmt"
	"strconv"
)

// ToString converts any value to a string representation
// This is a helper function for custom function implementations
func ToString(value any) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ToNumber converts any value to a float64
// This is a helper function for custom function implementations
func ToNumber(value any) float64 {
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int8:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint8:
		return float64(v)
	case uint16:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
		return 0
	case bool:
		if v {
			return 1
		}
		return 0
	default:
		// Try to convert via string
		str := ToString(v)
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			return f
		}
		return 0
	}
}

// String creates a new StringValue for use in custom functions
func String(value string) *StringValue {
	return NewStringValue(value, "", "")
}

// Number creates a new NumberValue for use in custom functions
func Number(value any) *NumberValue {
	return NewNumberValue(value, "", "", nil)
}
