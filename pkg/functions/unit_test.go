// Package functions provides tests for unit function
package functions

import (
	"testing"

	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUnitSelection tests unit function in selection context
// TypeScript original code:
//
//	test('selection', () => {
//	  const mf = new MessageFormat(
//	    'en',
//	    '.local $n = {42 :unit unit=meter} .match $n 42 {{exact}} * {{other}}',
//	    { functions: { unit } }
//	  );
//	  const onError = jest.fn();
//	  expect(mf.format(undefined, onError)).toEqual('other');
//	  expect(onError.mock.calls).toMatchObject([[{ type: 'bad-selector' }]]);
//	});
func TestUnitSelection(t *testing.T) {
	// TypeScript original code: const onError = jest.fn();
	var errors []error
	onError := func(err error) {
		errors = append(errors, err)
	}

	// TypeScript original code: const mf = new MessageFormat(
	ctx := NewMessageFunctionContext(
		[]string{"en"},
		"test source",
		"best fit",
		onError,
		nil,
		"",
		"",
	)

	options := map[string]interface{}{
		"unit": "meter",
	}

	// Test that unit function returns a MessageValue
	result := UnitFunction(ctx, options, 42)
	require.NotNil(t, result)
	assert.Equal(t, "number", result.Type())

	// Test selection behavior - unit values should not be selectable
	if numberValue, ok := result.(*messagevalue.NumberValue); ok {
		keys, err := numberValue.SelectKeys([]string{"42"})
		// Unit values should not match exact numeric keys in selection
		assert.Error(t, err) // Should error because unit is not selectable
		assert.Nil(t, keys)
	}
}

// TestUnitComplexOperand tests complex operand scenarios
// TypeScript original code:
//
//	describe('complex operand', () => {
//	  test(':currency result', () => {
//	    const mf = new MessageFormat(
//	      'en',
//	      '.local $n = {42 :unit unit=meter trailingZeroDisplay=stripIfInteger} {{{$n :unit signDisplay=always}}}',
//	      { functions: { unit } }
//	    );
//	    const nf = new Intl.NumberFormat('en', {
//	      signDisplay: 'always',
//	      style: 'unit',
//	      // @ts-expect-error TS doesn't know about trailingZeroDisplay
//	      trailingZeroDisplay: 'stripIfInteger',
//	      unit: 'meter'
//	    });
//	    expect(mf.format()).toEqual(nf.format(42));
//	    expect(mf.formatToParts()).toMatchObject([{ parts: nf.formatToParts(42) }]);
//	  });
func TestUnitComplexOperand(t *testing.T) {
	t.Run("unit result", func(t *testing.T) {
		// TypeScript original code: const mf = new MessageFormat(
		ctx := NewMessageFunctionContext(
			[]string{"en"},
			"test source",
			"best fit",
			nil,
			nil,
			"",
			"",
		)

		options := map[string]interface{}{
			"unit":                "meter",
			"signDisplay":         "always",
			"trailingZeroDisplay": "stripIfInteger",
		}

		// TypeScript original code: expect(mf.format()).toEqual(nf.format(42));
		result := UnitFunction(ctx, options, 42)
		require.NotNil(t, result)
		assert.Equal(t, "number", result.Type())

		// Verify it contains unit formatting
		str, err := result.ToString()
		require.NoError(t, err)
		assert.Contains(t, str, "m") // meter unit symbol should be present
		assert.Contains(t, str, "+") // signDisplay=always should show + for positive numbers
	})

	// TypeScript original code:
	//   test('external variable', () => {
	//     const mf = new MessageFormat('en', '{$n :unit}', { functions: { unit } });
	//     const nf = new Intl.NumberFormat('en', { style: 'unit', unit: 'meter' });
	//     const n = { valueOf: () => 42, options: { unit: 'meter' } };
	//     expect(mf.format({ n })).toEqual(nf.format(42));
	//     expect(mf.formatToParts({ n })).toMatchObject([
	//       { parts: nf.formatToParts(42) }
	//     ]);
	//   });
	t.Run("external variable", func(t *testing.T) {
		// TypeScript original code: const mf = new MessageFormat('en', '{$n :unit}', {
		ctx := NewMessageFunctionContext(
			[]string{"en"},
			"test source",
			"best fit",
			nil,
			nil,
			"",
			"",
		)

		// TypeScript original code: const n = { valueOf: () => 42, options: { unit: 'meter' } };
		// In Go, we simulate this with a map containing valueOf and options
		operand := map[string]interface{}{
			"valueOf": 42,
			"options": map[string]interface{}{"unit": "meter"},
		}

		options := map[string]interface{}{} // No explicit options, should use operand options

		// TypeScript original code: expect(mf.format({ n })).toEqual(nf.format(42));
		result := UnitFunction(ctx, options, operand)
		require.NotNil(t, result)
		assert.Equal(t, "number", result.Type())

		// Verify it contains unit formatting
		str, err := result.ToString()
		require.NoError(t, err)
		assert.Contains(t, str, "m") // meter unit symbol should be present
	})
}

// TestUnitBasicFunctionality tests basic unit function behavior
func TestUnitBasicFunctionality(t *testing.T) {
	t.Run("basic unit formatting", func(t *testing.T) {
		ctx := NewMessageFunctionContext(
			[]string{"en"},
			"test source",
			"best fit",
			nil,
			nil,
			"",
			"",
		)

		options := map[string]interface{}{
			"unit": "meter",
		}

		result := UnitFunction(ctx, options, 42)
		require.NotNil(t, result)
		assert.Equal(t, "number", result.Type())

		str, err := result.ToString()
		require.NoError(t, err)
		assert.Contains(t, str, "42")
		assert.Contains(t, str, "m") // meter unit symbol
	})

	t.Run("missing unit code", func(t *testing.T) {
		var errors []error
		onError := func(err error) {
			errors = append(errors, err)
		}

		ctx := NewMessageFunctionContext(
			[]string{"en"},
			"test source",
			"best fit",
			onError,
			nil,
			"",
			"",
		)

		options := map[string]interface{}{} // No unit option

		result := UnitFunction(ctx, options, 42)
		require.NotNil(t, result)
		assert.Equal(t, "fallback", result.Type()) // Should return fallback

		assert.Len(t, errors, 1)
		assert.Contains(t, errors[0].Error(), "unit identifier is required")
	})

	t.Run("different units", func(t *testing.T) {
		ctx := NewMessageFunctionContext(
			[]string{"en"},
			"test source",
			"best fit",
			nil,
			nil,
			"",
			"",
		)

		testCases := []struct {
			unit     string
			expected string
		}{
			{"meter", "m"},
			{"kilometer", "km"},
			{"gram", "g"},
			{"kilogram", "kg"},
			{"second", "s"},
			{"minute", "min"},
			{"hour", "h"},
		}

		for _, tc := range testCases {
			t.Run("unit="+tc.unit, func(t *testing.T) {
				options := map[string]interface{}{
					"unit": tc.unit,
				}

				result := UnitFunction(ctx, options, 42)
				require.NotNil(t, result)
				assert.Equal(t, "number", result.Type())

				str, err := result.ToString()
				require.NoError(t, err)
				assert.Contains(t, str, "42")
				assert.Contains(t, str, tc.expected)
			})
		}
	})

	t.Run("unit display options", func(t *testing.T) {
		ctx := NewMessageFunctionContext(
			[]string{"en"},
			"test source",
			"best fit",
			nil,
			nil,
			"",
			"",
		)

		testCases := []struct {
			unitDisplay string
			expected    string
		}{
			{"short", "m"},    // short form
			{"narrow", "m"},   // narrow form
			{"long", "meter"}, // long form
		}

		for _, tc := range testCases {
			t.Run("unitDisplay="+tc.unitDisplay, func(t *testing.T) {
				options := map[string]interface{}{
					"unit":        "meter",
					"unitDisplay": tc.unitDisplay,
				}

				result := UnitFunction(ctx, options, 42)
				require.NotNil(t, result)
				assert.Equal(t, "number", result.Type())

				str, err := result.ToString()
				require.NoError(t, err)
				assert.Contains(t, str, "42")
				assert.Contains(t, str, tc.expected)
			})
		}
	})
}
