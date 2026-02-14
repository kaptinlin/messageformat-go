package functions

import (
	"testing"

	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPercentSelectionMultiplies verifies that :percent selection uses value * 100
// This test is based on the specification requirement that percent selection
// should match on the multiplied value
// Reference: spec/functions/number.md lines 515-524
func TestPercentSelectionMultiplies(t *testing.T) {
	tests := []struct {
		name          string
		value         float64
		keys          []string
		expectedMatch string
		description   string
	}{
		{
			name:          "0.01 matches key '1'",
			value:         0.01,
			keys:          []string{"1", "100", "other"},
			expectedMatch: "1",
			description:   "Value 0.01 * 100 = 1, should match key '1'",
		},
		{
			name:          "1.0 matches key '100'",
			value:         1.0,
			keys:          []string{"1", "100", "other"},
			expectedMatch: "100",
			description:   "Value 1.0 * 100 = 100, should match key '100'",
		},
		{
			name:          "0.5 matches key other",
			value:         0.5,
			keys:          []string{"1", "100", "other"},
			expectedMatch: "other",
			description:   "Value 0.5 * 100 = 50, no exact match, should fall back to 'other'",
		},
		{
			name:          "1.0 matches 'other' in plural rules",
			value:         1.0,
			keys:          []string{"one", "other"},
			expectedMatch: "other",
			description:   "Value 1.0 * 100 = 100, plural category for 100 is 'other' in English",
		},
		{
			name:          "0.01 matches 'one' in plural rules",
			value:         0.01,
			keys:          []string{"one", "other"},
			expectedMatch: "one",
			description:   "Value 0.01 * 100 = 1, plural category for 1 is 'one' in English",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create context using the correct constructor
			ctx := NewMessageFunctionContext(
				[]string{"en"},
				"test source",
				"best fit",
				nil,
				nil,
				"",
				"",
			)

			// Create percent value
			options := make(map[string]any)
			nv := PercentFunction(ctx, options, tt.value)

			// Verify it's a number value
			require.NotNil(t, nv)
			numberVal, ok := nv.(*messagevalue.NumberValue)
			require.True(t, ok, "Expected NumberValue")

			// Perform selection
			selectedKeys, err := numberVal.SelectKeys(tt.keys)
			require.NoError(t, err, tt.description)

			// Verify the match
			if tt.expectedMatch == "" {
				assert.Empty(t, selectedKeys, tt.description)
			} else {
				require.Len(t, selectedKeys, 1, tt.description)
				assert.Equal(t, tt.expectedMatch, selectedKeys[0], tt.description)
			}
		})
	}
}

// TestPercentResolvedValueNotMultiplied verifies that the resolved value
// retains the original value (not multiplied by 100)
// Reference: spec/functions/number.md lines 498-506
func TestPercentResolvedValueNotMultiplied(t *testing.T) {
	ctx := NewMessageFunctionContext(
		[]string{"en"},
		"test source",
		"best fit",
		nil,
		nil,
		"",
		"",
	)

	value := 0.5
	options := make(map[string]any)
	nv := PercentFunction(ctx, options, value)

	// Get the resolved value
	resolvedValue, err := nv.ValueOf()
	require.NoError(t, err)

	// Convert to float64
	floatVal, ok := resolvedValue.(float64)
	require.True(t, ok, "Expected float64")

	// Verify it's the original value, not multiplied
	assert.Equal(t, 0.5, floatVal, "Resolved value should be original (0.5), not multiplied (50)")
}

// TestCurrencyCannotSelect verifies that :currency does not support selection
// Reference: TypeScript implementation - currency uses getMessageNumber(..., false)
func TestCurrencyCannotSelect(t *testing.T) {
	ctx := NewMessageFunctionContext(
		[]string{"en"},
		"test source",
		"best fit",
		nil,
		nil,
		"",
		"",
	)

	value := 42.0
	options := map[string]any{
		"currency": "USD",
	}
	nv := CurrencyFunction(ctx, options, value)

	// Verify it's a number value
	require.NotNil(t, nv)
	numberVal, ok := nv.(*messagevalue.NumberValue)
	require.True(t, ok, "Expected NumberValue")

	// Attempt selection - should return empty or error
	selectedKeys, err := numberVal.SelectKeys([]string{"42", "one", "other"})

	// Currency should not support selection
	// Either returns error or empty result
	if err == nil {
		assert.Empty(t, selectedKeys, "Currency should not support selection")
	}
}
