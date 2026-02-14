package functions

import (
	"math/big"
	"testing"

	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOffsetFunction(t *testing.T) {
	tests := []struct {
		name        string
		operand     any
		options     map[string]any
		expectError bool
		expectValue any
		description string
	}{
		{
			name:        "add integer to number",
			operand:     5,
			options:     map[string]any{"add": 3},
			expectError: false,
			expectValue: 8,
			description: "Should add 3 to 5 to get 8",
		},
		{
			name:        "subtract integer from number",
			operand:     10,
			options:     map[string]any{"subtract": 4},
			expectError: false,
			expectValue: 6,
			description: "Should subtract 4 from 10 to get 6",
		},
		{
			name:        "add to float64",
			operand:     5.5,
			options:     map[string]any{"add": 2},
			expectError: false,
			expectValue: 7.5,
			description: "Should add 2 to 5.5 to get 7.5",
		},
		{
			name:        "subtract from float64",
			operand:     10.7,
			options:     map[string]any{"subtract": 3},
			expectError: false,
			expectValue: 7.7,
			description: "Should subtract 3 from 10.7 to get 7.7",
		},
		{
			name:        "both add and subtract provided",
			operand:     5,
			options:     map[string]any{"add": 3, "subtract": 2},
			expectError: true,
			description: "Should error when both add and subtract are provided",
		},
		{
			name:        "neither add nor subtract provided",
			operand:     5,
			options:     map[string]any{},
			expectError: true,
			description: "Should error when neither add nor subtract are provided",
		},
		{
			name:        "invalid add value",
			operand:     5,
			options:     map[string]any{"add": "invalid"},
			expectError: true,
			description: "Should error when add value is not a positive integer",
		},
		{
			name:        "invalid subtract value",
			operand:     5,
			options:     map[string]any{"subtract": -1},
			expectError: true,
			description: "Should error when subtract value is negative",
		},
		{
			name:        "string operand that parses to number",
			operand:     "10",
			options:     map[string]any{"add": 5},
			expectError: false,
			expectValue: int64(15),
			description: "Should parse string operand to number and add 5",
		},
		{
			name:        "object with valueOf method",
			operand:     map[string]any{"valueOf": 8, "options": map[string]any{"style": "decimal"}},
			options:     map[string]any{"subtract": 3},
			expectError: false,
			expectValue: 5,
			description: "Should extract value from object and subtract 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test context
			var errors []error
			ctx := NewMessageFunctionContext(
				[]string{"en"},
				"test",
				"best fit",
				func(err error) { errors = append(errors, err) },
				make(map[string]bool),
				"ltr",
				"",
			)

			// Call the offset function
			result := OffsetFunction(ctx, tt.options, tt.operand)

			if tt.expectError {
				// For error cases, result should be a fallback value
				fallbackVal, ok := result.(*messagevalue.FallbackValue)
				require.True(t, ok, "Expected fallback value for error case")
				assert.NotEmpty(t, fallbackVal.Source(), "Fallback should have source")
				assert.True(t, len(errors) > 0, "Context should have errors recorded")
			} else {
				// For success cases, result should be a NumberValue
				require.False(t, result == nil, "Result should not be nil")

				// The result should be a NumberValue since offset calls NumberFunction
				numVal, ok := result.(*messagevalue.NumberValue)
				require.True(t, ok, "Expected NumberValue, got %T", result)

				// Check that the underlying value matches expected
				actualValue, err := numVal.ValueOf()
				require.NoError(t, err, "Failed to get value from NumberValue")

				// Compare values with tolerance for floats
				switch expected := tt.expectValue.(type) {
				case float64:
					actualFloat, ok := actualValue.(float64)
					require.True(t, ok, "Expected float64, got %T", actualValue)
					assert.InDelta(t, expected, actualFloat, 0.0001, "Float values should match within tolerance")
				default:
					assert.Equal(t, tt.expectValue, actualValue, "Values should match")
				}

				// Should not have errors
				assert.Empty(t, errors, "Should not have errors for success case")
			}
		})
	}
}

func TestOffsetFunctionEdgeCases(t *testing.T) {
	t.Run("zero offset", func(t *testing.T) {
		var errors []error
		ctx := NewMessageFunctionContext(
			[]string{"en"},
			"test",
			"best fit",
			func(err error) { errors = append(errors, err) },
			make(map[string]bool),
			"ltr",
			"",
		)
		result := OffsetFunction(ctx, map[string]any{"add": 0}, 10)

		numVal, ok := result.(*messagevalue.NumberValue)
		require.True(t, ok)

		actualValue, err := numVal.ValueOf()
		require.NoError(t, err)
		assert.Equal(t, 10, actualValue)
	})

	t.Run("large numbers", func(t *testing.T) {
		var errors []error
		ctx := NewMessageFunctionContext(
			[]string{"en"},
			"test",
			"best fit",
			func(err error) { errors = append(errors, err) },
			make(map[string]bool),
			"ltr",
			"",
		)
		result := OffsetFunction(ctx, map[string]any{"add": 1000000}, 2000000)

		numVal, ok := result.(*messagevalue.NumberValue)
		require.True(t, ok)

		actualValue, err := numVal.ValueOf()
		require.NoError(t, err)
		assert.Equal(t, 3000000, actualValue)
	})

	// Add TypeScript compatibility test cases based on reference implementation
	t.Run("TypeScript compatibility - BigInt operand", func(t *testing.T) {
		var errors []error
		ctx := NewMessageFunctionContext(
			[]string{"en"},
			"test",
			"best fit",
			func(err error) { errors = append(errors, err) },
			make(map[string]bool),
			"ltr",
			"",
		)

		// Test with big.Int to match TypeScript BigInt behavior
		bigInt := big.NewInt(9223372036854775807) // Max int64
		result := OffsetFunction(ctx, map[string]any{"add": 1}, bigInt)

		_, ok := result.(*messagevalue.NumberValue)
		require.True(t, ok)
		assert.Empty(t, errors)
	})

	t.Run("TypeScript compatibility - exactly one option required", func(t *testing.T) {
		var errors []error
		ctx := NewMessageFunctionContext(
			[]string{"en"},
			"test",
			"best fit",
			func(err error) { errors = append(errors, err) },
			make(map[string]bool),
			"ltr",
			"",
		)

		// This matches TypeScript: if (add < 0 === sub < 0)
		result := OffsetFunction(ctx, map[string]any{}, 10)

		_, ok := result.(*messagevalue.FallbackValue)
		require.True(t, ok)
		assert.NotEmpty(t, errors)
		assert.Contains(t, errors[0].Error(), "Exactly one")
	})

	t.Run("TypeScript compatibility - negative value handling", func(t *testing.T) {
		var errors []error
		ctx := NewMessageFunctionContext(
			[]string{"en"},
			"test",
			"best fit",
			func(err error) { errors = append(errors, err) },
			make(map[string]bool),
			"ltr",
			"",
		)

		result := OffsetFunction(ctx, map[string]any{"subtract": 5}, -10)

		numVal, ok := result.(*messagevalue.NumberValue)
		require.True(t, ok)

		actualValue, err := numVal.ValueOf()
		require.NoError(t, err)
		assert.Equal(t, -15, actualValue) // -10 - 5 = -15
		assert.Empty(t, errors)
	})
}
