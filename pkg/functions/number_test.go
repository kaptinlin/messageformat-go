package functions

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadNumericOperand(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		source    string
		expectErr bool
		expected  interface{}
	}{
		{"integer", 42, "test", false, 42},
		{"float", 3.14, "test", false, 3.14},
		{"string number", "123", "test", false, int64(123)},
		{"string float", "3.14", "test", false, 3.14},
		{"invalid string", "abc", "test", true, nil},
		{"nil", nil, "test", true, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := readNumericOperand(tt.input, tt.source)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected, result.Value)
			}
		})
	}
}

func TestNumberFunction(t *testing.T) {
	ctx := NewMessageFunctionContext(
		[]string{"en"},
		"test source",
		"best fit",
		nil,
		nil,
		"auto",
		"",
	)

	options := map[string]interface{}{
		"style": "decimal",
	}

	result := NumberFunction(ctx, options, 42)

	assert.Equal(t, "number", result.Type())
	assert.Equal(t, "test source", result.Source())
}

func TestIntegerFunction(t *testing.T) {
	ctx := NewMessageFunctionContext(
		[]string{"en"},
		"test source",
		"best fit",
		nil,
		nil,
		"auto",
		"",
	)

	options := map[string]interface{}{}

	t.Run("integer input", func(t *testing.T) {
		result := IntegerFunction(ctx, options, 42)
		assert.Equal(t, "number", result.Type())
	})

	t.Run("float input", func(t *testing.T) {
		result := IntegerFunction(ctx, options, 3.7)
		assert.Equal(t, "number", result.Type())
	})

	t.Run("invalid input", func(t *testing.T) {
		result := IntegerFunction(ctx, options, "invalid")
		assert.Equal(t, "fallback", result.Type())
	})
}

func TestMergeNumberOptions(t *testing.T) {
	operandOptions := map[string]interface{}{
		"style": "currency",
	}

	exprOptions := map[string]interface{}{
		"minimumFractionDigits": 2,
		"style":                 "decimal", // Should override operand
	}

	result := mergeNumberOptions(operandOptions, exprOptions, "best fit")

	assert.Equal(t, "best fit", result["localeMatcher"])
	assert.Equal(t, "decimal", result["style"]) // Expression option overrides
	assert.Equal(t, 2, result["minimumFractionDigits"])
}

func TestParseJSONNumber(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{"integer", "123", false},
		{"float", "3.14", false},
		{"negative", "-42", false},
		{"invalid", "abc", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseJSONNumber(tt.input)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestIsFinite(t *testing.T) {
	assert.True(t, isFinite(42.0))
	assert.True(t, isFinite(-3.14))
	assert.True(t, isFinite(0.0))
	assert.False(t, isFinite(math.Inf(1)))  // +Inf
	assert.False(t, isFinite(math.Inf(-1))) // -Inf
	assert.False(t, isFinite(math.NaN()))   // NaN
}
