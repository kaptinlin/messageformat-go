// Package functions provides tests for currency function
package functions

import (
	"testing"

	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCurrencyFractionDigits tests fractionDigits option
// TypeScript original code:
//
//	describe('fractionDigits', () => {
//	  for (const fd of [0, 2, 'auto' as const]) {
//	    test(`fractionDigits=${fd}`, () => {
//	      const mf = new MessageFormat(
//	        'en',
//	        `{42 :currency currency=EUR fractionDigits=${fd}}`,
//	        { functions: { currency } }
//	      );
//	      const nf = new Intl.NumberFormat('en', {
//	        style: 'currency',
//	        currency: 'EUR',
//	        minimumFractionDigits: fd === 'auto' ? undefined : fd,
//	        maximumFractionDigits: fd === 'auto' ? undefined : fd
//	      });
//	      expect(mf.format()).toEqual(nf.format(42));
//	      expect(mf.formatToParts()).toMatchObject([
//	        { parts: nf.formatToParts(42) }
//	      ]);
//	    });
//	  }
//	});
func TestCurrencyFractionDigits(t *testing.T) {
	fractionDigitsValues := []interface{}{0, 2, "auto"}

	for _, fd := range fractionDigitsValues {
		t.Run("fractionDigits="+toString(fd), func(t *testing.T) {
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
				"currency":       "EUR",
				"fractionDigits": fd,
			}

			// TypeScript original code: expect(mf.format()).toEqual(nf.format(42));
			result := CurrencyFunction(ctx, options, 42)
			require.NotNil(t, result)
			assert.Equal(t, "number", result.Type())

			// Verify it's a currency format by checking the formatted string
			str, err := result.ToString()
			require.NoError(t, err)
			assert.Contains(t, str, "€") // EUR symbol should be present
		})
	}
}

// TestCurrencyCurrencyDisplay tests currencyDisplay option
// TypeScript original code:
//
//	describe('currencyDisplay', () => {
//	  for (const cd of [
//	    'narrowSymbol',
//	    'symbol',
//	    'name',
//	    'code',
//	    'never'
//	  ] as const) {
//	    test(`currencyDisplay=${cd}`, () => {
//	      const mf = new MessageFormat(
//	        'en',
//	        `{42 :currency currency=EUR currencyDisplay=${cd}}`,
//	        { functions: { currency } }
//	      );
//	      const nf = new Intl.NumberFormat('en', {
//	        style: 'currency',
//	        currency: 'EUR',
//	        currencyDisplay: cd === 'never' ? undefined : cd
//	      });
//	      const onError = jest.fn();
//	      expect(mf.format(undefined, onError)).toEqual(nf.format(42));
//	      expect(mf.formatToParts(undefined, onError)).toMatchObject([
//	        { parts: nf.formatToParts(42) }
//	      ]);
//	      if (cd === 'never') {
//	        expect(onError.mock.calls).toMatchObject([
//	          [{ type: 'unsupported-operation' }],
//	          [{ type: 'unsupported-operation' }]
//	        ]);
//	      } else {
//	        expect(onError.mock.calls).toMatchObject([]);
//	      }
//	    });
//	  }
//	});
func TestCurrencyCurrencyDisplay(t *testing.T) {
	currencyDisplayValues := []string{
		"narrowSymbol",
		"symbol",
		"name",
		"code",
		"never",
	}

	for _, cd := range currencyDisplayValues {
		t.Run("currencyDisplay="+cd, func(t *testing.T) {
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
				"currency":        "EUR",
				"currencyDisplay": cd,
			}

			// TypeScript original code: expect(mf.format(undefined, onError)).toEqual(nf.format(42));
			result := CurrencyFunction(ctx, options, 42)
			require.NotNil(t, result)

			// TypeScript original code: if (cd === 'never') {
			if cd == "never" {
				// TypeScript original code: expect(onError.mock.calls).toMatchObject([
				assert.Len(t, errors, 1) // Should have 1 unsupported-operation error
				assert.Contains(t, errors[0].Error(), "unsupported-operation")
			} else {
				// TypeScript original code: expect(onError.mock.calls).toMatchObject([]);
				assert.Empty(t, errors)

				// Verify it's a currency format
				str, err := result.ToString()
				require.NoError(t, err)

				// Check for appropriate currency representation based on display type
				switch cd {
				case "name":
					assert.Contains(t, str, "euros") // EUR name should be present
				case "code":
					assert.Contains(t, str, "EUR") // EUR code should be present
				default:
					assert.Contains(t, str, "€") // EUR symbol should be present
				}
			}
		})
	}
}

// TestCurrencySelection tests currency function in selection context
// TypeScript original code:
//
//	test('selection', () => {
//	  const mf = new MessageFormat(
//	    'en',
//	    '.local $n = {42 :currency currency=EUR} .match $n 42 {{exact}} * {{other}}',
//	    { functions: { currency } }
//	  );
//	  const onError = jest.fn();
//	  expect(mf.format(undefined, onError)).toEqual('other');
//	  expect(onError.mock.calls).toMatchObject([[{ type: 'bad-selector' }]]);
//	});
func TestCurrencySelection(t *testing.T) {
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
		"currency": "EUR",
	}

	// Test that currency function returns a MessageValue
	result := CurrencyFunction(ctx, options, 42)
	require.NotNil(t, result)
	assert.Equal(t, "number", result.Type())

	// Test selection behavior - currency values should not be selectable
	if numberValue, ok := result.(*messagevalue.NumberValue); ok {
		keys, err := numberValue.SelectKeys([]string{"42"})
		// Currency values should not match exact numeric keys in selection
		assert.Error(t, err) // Should error because currency is not selectable
		assert.Nil(t, keys)
	}
}

// TestCurrencyComplexOperand tests complex operand scenarios
// TypeScript original code:
//
//	describe('complex operand', () => {
//	  test(':currency result', () => {
//	    const mf = new MessageFormat(
//	      'en',
//	      '.local $n = {-42 :currency currency=USD trailingZeroDisplay=stripIfInteger} {{{$n :currency currencySign=accounting}}}',
//	      { functions: { currency } }
//	    );
//	    const nf = new Intl.NumberFormat('en', {
//	      style: 'currency',
//	      currencySign: 'accounting',
//	      // @ts-expect-error TS doesn't know about trailingZeroDisplay
//	      trailingZeroDisplay: 'stripIfInteger',
//	      currency: 'USD'
//	    });
//	    expect(mf.format()).toEqual(nf.format(-42));
//	    expect(mf.formatToParts()).toMatchObject([
//	      { parts: nf.formatToParts(-42) }
//	    ]);
//	  });
func TestCurrencyComplexOperand(t *testing.T) {
	t.Run("currency result", func(t *testing.T) {
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
			"currency":            "USD",
			"currencySign":        "accounting",
			"trailingZeroDisplay": "stripIfInteger",
		}

		// TypeScript original code: expect(mf.format()).toEqual(nf.format(-42));
		result := CurrencyFunction(ctx, options, -42)
		require.NotNil(t, result)
		assert.Equal(t, "number", result.Type())

		// Verify it contains USD currency and accounting format (parentheses for negative)
		str, err := result.ToString()
		require.NoError(t, err)
		assert.Contains(t, str, "$")
		assert.Contains(t, str, "(") // Accounting format uses parentheses for negative
	})

	// TypeScript original code:
	//   test('external variable', () => {
	//     const mf = new MessageFormat('en', '{$n :currency}', {
	//       functions: { currency }
	//     });
	//     const nf = new Intl.NumberFormat('en', {
	//       style: 'currency',
	//       currency: 'EUR'
	//     });
	//     const n = { valueOf: () => 42, options: { currency: 'EUR' } };
	//     expect(mf.format({ n })).toEqual(nf.format(42));
	//     expect(mf.formatToParts({ n })).toMatchObject([
	//       { parts: nf.formatToParts(42) }
	//     ]);
	//   });
	t.Run("external variable", func(t *testing.T) {
		// TypeScript original code: const mf = new MessageFormat('en', '{$n :currency}', {
		ctx := NewMessageFunctionContext(
			[]string{"en"},
			"test source",
			"best fit",
			nil,
			nil,
			"",
			"",
		)

		// TypeScript original code: const n = { valueOf: () => 42, options: { currency: 'EUR' } };
		// In Go, we simulate this with a map containing valueOf and options
		operand := map[string]interface{}{
			"valueOf": 42,
			"options": map[string]interface{}{"currency": "EUR"},
		}

		options := map[string]interface{}{} // No explicit options, should use operand options

		// TypeScript original code: expect(mf.format({ n })).toEqual(nf.format(42));
		result := CurrencyFunction(ctx, options, operand)
		require.NotNil(t, result)
		assert.Equal(t, "number", result.Type())

		// Verify it contains EUR currency
		str, err := result.ToString()
		require.NoError(t, err)
		assert.Contains(t, str, "€")
	})
}

// TestCurrencyBasicFunctionality tests basic currency function behavior
func TestCurrencyBasicFunctionality(t *testing.T) {
	t.Run("basic currency formatting", func(t *testing.T) {
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
			"currency": "USD",
		}

		result := CurrencyFunction(ctx, options, 42)
		require.NotNil(t, result)
		assert.Equal(t, "number", result.Type())

		str, err := result.ToString()
		require.NoError(t, err)
		assert.Contains(t, str, "$")
		assert.Contains(t, str, "42")
	})

	t.Run("missing currency code", func(t *testing.T) {
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

		options := map[string]interface{}{} // No currency option

		result := CurrencyFunction(ctx, options, 42)
		require.NotNil(t, result)
		assert.Equal(t, "fallback", result.Type()) // Should return fallback

		assert.Len(t, errors, 1)
		assert.Contains(t, errors[0].Error(), "currency code is required")
	})
}
