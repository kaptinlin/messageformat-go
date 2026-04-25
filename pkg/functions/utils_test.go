package functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAsBoolean(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
		hasError bool
	}{
		{"true boolean", true, true, false},
		{"false boolean", false, false, false},
		{"true string", "true", true, false},
		{"false string", "false", false, false},
		{"invalid string", "invalid", false, true},
		{"number", 42, false, true},
		{"nil", nil, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := asBoolean(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestAsPositiveInteger(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected int
		hasError bool
	}{
		{"positive int", 42, 42, false},
		{"zero", 0, 0, false},
		{"negative int", -5, 0, true},
		{"positive float", 3.0, 3, false},
		{"positive string", "123", 123, false},
		{"zero string", "0", 0, false},
		{"invalid string", "abc", 0, true},
		{"negative string", "-5", 0, true},
		{"float string", "3.14", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := asPositiveInteger(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestAsString(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
		hasError bool
	}{
		{"string", "hello", "hello", false},
		{"empty string", "", "", false},
		{"number", 42, "", true},
		{"boolean", true, "", true},
		{"nil", nil, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := asString(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetFirstLocale(t *testing.T) {
	assert.Equal(t, "en-US", GetFirstLocale([]string{"en-US", "fr"}))
	assert.Equal(t, "en", GetFirstLocale([]string{}))
	assert.Equal(t, "en", GetFirstLocale(nil))
}
