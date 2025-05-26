package functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringFunction(t *testing.T) {
	ctx := NewMessageFunctionContext(
		[]string{"en"},
		"test source",
		"best fit",
		nil,
		nil,
		"auto",
		"",
	)

	tests := []struct {
		name     string
		operand  interface{}
		options  map[string]interface{}
		expected string
	}{
		{"string input", "hello", nil, "hello"},
		{"nil input", nil, nil, ""},
		{"number input", 42, nil, "42"},
		{"boolean input", true, nil, "true"},
		{"with locale option", "test", map[string]interface{}{"locale": "fr"}, "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringFunction(ctx, tt.options, tt.operand)

			assert.Equal(t, "string", result.Type())
			assert.Equal(t, "test source", result.Source())

			// Test string conversion
			str, err := result.ToString()
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, str)
		})
	}
}

func TestStringFunctionWithDirection(t *testing.T) {
	tests := []struct {
		name        string
		contextDir  string
		expectedDir string
	}{
		{"ltr direction", "ltr", "ltr"},
		{"rtl direction", "rtl", "rtl"},
		{"auto direction", "auto", "auto"},
		{"empty direction", "", "auto"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewMessageFunctionContext(
				[]string{"en"},
				"test source",
				"best fit",
				nil,
				nil,
				tt.contextDir,
				"",
			)

			result := StringFunction(ctx, nil, "test")
			assert.Equal(t, "string", result.Type())
		})
	}
}
