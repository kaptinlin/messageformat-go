package functions

import (
	"maps"
	"strconv"
)

const maxIntValue = uint64(1<<(strconv.IntSize-1) - 1)

// Options provides read-oriented helpers for resolved function options.
// TypeScript original code: Record<string, unknown>
type Options map[string]any

// NewOptions returns a detached options wrapper.
func NewOptions(options map[string]any) Options {
	if options == nil {
		return nil
	}
	return Options(maps.Clone(options))
}

// Has reports whether name is present.
func (o Options) Has(name string) bool {
	_, ok := o[name]
	return ok
}

// Value returns the raw option value.
func (o Options) Value(name string) (any, bool) {
	value, ok := o[name]
	return value, ok
}

// String returns a string option.
func (o Options) String(name string) (string, bool) {
	value, ok := o[name].(string)
	return value, ok
}

// Int returns an int option.
func (o Options) Int(name string) (int, bool) {
	switch value := o[name].(type) {
	case int:
		return value, true
	case int8:
		return int(value), true
	case int16:
		return int(value), true
	case int32:
		return int(value), true
	case int64:
		return int(value), true
	case uint:
		if uint64(value) > maxIntValue {
			return 0, false
		}
		return int(value), true
	case uint8:
		return int(value), true
	case uint16:
		return int(value), true
	case uint32:
		if uint64(value) > maxIntValue {
			return 0, false
		}
		return int(value), true
	case uint64:
		if value > maxIntValue {
			return 0, false
		}
		return int(value), true
	case float32:
		return int(value), true
	case float64:
		return int(value), true
	default:
		return 0, false
	}
}

// Bool returns a bool option.
func (o Options) Bool(name string) (bool, bool) {
	value, ok := o[name].(bool)
	return value, ok
}

// Map returns a detached map copy.
func (o Options) Map() map[string]any {
	if o == nil {
		return nil
	}
	return maps.Clone(map[string]any(o))
}
