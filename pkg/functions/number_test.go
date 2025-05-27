package functions

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// TestNumberSoftFailForIntegerOptions tests soft fail for integer options
// TypeScript original code:
//
//	test('soft fail for integer options', () => {
//	  const mf = new MessageFormat('en', '{42 :number minimumFractionDigits=foo}');
//	  const onError = jest.fn();
//	  expect(mf.format(undefined, onError)).toEqual('42');
//	  expect(onError.mock.calls).toMatchObject([[{ type: 'bad-option' }]]);
//	});
func TestNumberSoftFailForIntegerOptions(t *testing.T) {
	// TypeScript original code: const onError = jest.fn();
	var errors []error
	onError := func(err error) {
		errors = append(errors, err)
	}

	// TypeScript original code: const mf = new MessageFormat('en', '{42 :number minimumFractionDigits=foo}');
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
		"minimumFractionDigits": "foo", // Invalid value
	}

	// TypeScript original code: expect(mf.format(undefined, onError)).toEqual('42');
	result := NumberFunction(ctx, options, 42)
	require.NotNil(t, result)
	assert.Equal(t, "number", result.Type())

	// Should still format the number despite the bad option
	str, err := result.ToString()
	require.NoError(t, err)
	assert.Contains(t, str, "42")

	// TypeScript original code: expect(onError.mock.calls).toMatchObject([[{ type: 'bad-option' }]]);
	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0].Error(), "bad-option")
}

// TestNumberSelection tests number function in selection context
// TypeScript original code:
//
//	test('selection', () => {
//	  const mf = new MessageFormat(
//	    'en',
//	    '.local $exact = {exact} .local $n = {42 :number select=$exact} .match $n 42 {{exact}} * {{other}}'
//	  );
//	  const onError = jest.fn();
//	  expect(mf.format(undefined, onError)).toEqual('other');
//	  expect(onError.mock.calls).toMatchObject([
//	    [{ type: 'bad-option' }],
//	    [{ type: 'bad-selector' }]
//	  ]);
//	});
func TestNumberSelection(t *testing.T) {
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
		map[string]bool{
			"select": false, // select option is not set by literal value
		},
		"",
		"",
	)

	options := map[string]interface{}{
		"select": "exact", // This should cause a bad-option error since it's not literal
	}

	// Test that number function returns a MessageValue
	result := NumberFunction(ctx, options, 42)
	require.NotNil(t, result)
	assert.Equal(t, "number", result.Type())

	// TypeScript original code: expect(onError.mock.calls).toMatchObject([
	//   [{ type: 'bad-option' }],
	//   [{ type: 'bad-selector' }]
	// ]);
	// Note: In the Go implementation, we only get the bad-option error
	// The bad-selector error would come from the matching logic, not the function itself
	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0].Error(), "bad-option")
}

// TestNumberBasicFunctionality tests basic number function behavior
func TestNumberBasicFunctionality(t *testing.T) {
	t.Run("basic number formatting", func(t *testing.T) {
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
			"style": "decimal",
		}

		result := NumberFunction(ctx, options, 42)
		require.NotNil(t, result)
		assert.Equal(t, "number", result.Type())

		str, err := result.ToString()
		require.NoError(t, err)
		assert.Equal(t, "42", str)
	})

	t.Run("with fraction digits", func(t *testing.T) {
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
			"minimumFractionDigits": 2,
			"maximumFractionDigits": 2,
		}

		result := NumberFunction(ctx, options, 42)
		require.NotNil(t, result)
		assert.Equal(t, "number", result.Type())

		str, err := result.ToString()
		require.NoError(t, err)
		assert.Contains(t, str, "42.00")
	})

	t.Run("with sign display", func(t *testing.T) {
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
			"signDisplay": "always",
		}

		result := NumberFunction(ctx, options, 42)
		require.NotNil(t, result)
		assert.Equal(t, "number", result.Type())

		str, err := result.ToString()
		require.NoError(t, err)
		assert.Contains(t, str, "+42")
	})

	t.Run("invalid option values", func(t *testing.T) {
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

		options := map[string]interface{}{
			"minimumFractionDigits": "invalid",
			"signDisplay":           123, // Should be string
		}

		result := NumberFunction(ctx, options, 42)
		require.NotNil(t, result)
		assert.Equal(t, "number", result.Type())

		// Should have errors for invalid options
		assert.Len(t, errors, 2)
		for _, err := range errors {
			assert.Contains(t, err.Error(), "bad-option")
		}
	})
}
