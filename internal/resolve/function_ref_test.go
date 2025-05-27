package resolve

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/messageformat-go/pkg/bidi"
	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	pkgErrors "github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// TestCustomFunction tests custom function implementation
// TypeScript original code:
//
//	test('Custom function', () => {
//	  const custom: MessageFunction<'custom'> = (
//	    { dir, source, locales: [locale] },
//	    _opt,
//	    input
//	  ) => ({
//	    type: 'custom',
//	    source,
//	    dir: dir ?? 'auto',
//	    locale,
//	    toParts: () => [{ type: 'custom', locale, value: `part:${input}` }],
//	    toString: () => `str:${input}`
//	  });
//	  const mf = new MessageFormat('en', '{$var :custom}', {
//	    functions: { custom }
//	  });
//	  expect(mf.format({ var: 42 })).toEqual('\u2068str:42\u2069');
//	  expect(mf.formatToParts({ var: 42 })).toEqual([
//	    { type: 'bidiIsolation', value: '\u2068' },
//	    { type: 'custom', locale: 'en', value: 'part:42' },
//	    { type: 'bidiIsolation', value: '\u2069' }
//	  ]);
//	});
func TestCustomFunction(t *testing.T) {
	// Create a custom function that mimics the TypeScript behavior
	customFunc := func(
		ctx functions.MessageFunctionContext,
		options map[string]interface{},
		input interface{},
	) messagevalue.MessageValue {
		locale := "en"
		if len(ctx.Locales()) > 0 {
			locale = ctx.Locales()[0]
		}

		dir := ctx.Dir()
		if dir == "" {
			dir = "auto"
		}

		return &customMessageValue{
			typ:    "custom",
			source: ctx.Source(),
			dir:    dir,
			locale: locale,
			input:  input,
		}
	}

	// Create context with custom function
	ctx := NewContext(
		[]string{"en"},
		map[string]functions.MessageFunction{
			"custom": customFunc,
		},
		map[string]interface{}{
			"var": 42,
		},
		nil,
	)

	// Create variable reference and function reference
	varRef := datamodel.NewVariableRef("var")
	funcRef := datamodel.NewFunctionRef("custom", nil)

	// Resolve the function reference
	result := ResolveFunctionRef(ctx, varRef, funcRef)

	// Test the result
	assert.Equal(t, "custom", result.Type())
	assert.Equal(t, "$var", result.Source())

	// Test toString
	str, err := result.ToString()
	require.NoError(t, err)
	assert.Equal(t, "str:42", str)

	// Test toParts
	parts, err := result.ToParts()
	require.NoError(t, err)
	require.Len(t, parts, 1)
	assert.Equal(t, "custom", parts[0].Type())
	assert.Equal(t, "en", parts[0].Locale())
	assert.Equal(t, "part:42", parts[0].Value())
}

// customMessageValue implements a custom message value for testing
type customMessageValue struct {
	typ    string
	source string
	dir    string
	locale string
	input  interface{}
}

func (cv *customMessageValue) Type() string   { return cv.typ }
func (cv *customMessageValue) Source() string { return cv.source }
func (cv *customMessageValue) Dir() bidi.Direction {
	switch cv.dir {
	case "ltr":
		return bidi.DirLTR
	case "rtl":
		return bidi.DirRTL
	default:
		return bidi.DirAuto
	}
}
func (cv *customMessageValue) Locale() string                             { return cv.locale }
func (cv *customMessageValue) Options() map[string]interface{}            { return nil }
func (cv *customMessageValue) ValueOf() (interface{}, error)              { return cv.input, nil }
func (cv *customMessageValue) SelectKeys(keys []string) ([]string, error) { return nil, nil }

func (cv *customMessageValue) ToString() (string, error) {
	return "str:" + fmt.Sprintf("%v", cv.input), nil
}

func (cv *customMessageValue) ToParts() ([]messagevalue.MessagePart, error) {
	return []messagevalue.MessagePart{
		&customMessagePart{
			typ:    "custom",
			locale: cv.locale,
			value:  "part:" + fmt.Sprintf("%v", cv.input),
		},
	}, nil
}

// customMessagePart implements a custom message part for testing
type customMessagePart struct {
	typ    string
	locale string
	value  string
}

func (cp *customMessagePart) Type() string        { return cp.typ }
func (cp *customMessagePart) Value() interface{}  { return cp.value }
func (cp *customMessagePart) Source() string      { return "" }
func (cp *customMessagePart) Locale() string      { return cp.locale }
func (cp *customMessagePart) Dir() bidi.Direction { return bidi.DirAuto }

// TestInputsWithOptions tests inputs with options
// TypeScript original code:
//
//	describe('inputs with options', () => {
//	  test('local variable with :number expression', () => {
//	    const mf = new MessageFormat(
//	      'en',
//	      `.local $val = {12345678 :number useGrouping=never}
//	      {{{$val :number minimumFractionDigits=2}}}`
//	    );
//	    const msg = mf.formatToParts();
//	    const { parts } = msg[0] as MessageNumberPart;
//
//	    const nf = new Intl.NumberFormat('en', {
//	      minimumFractionDigits: 2,
//	      useGrouping: false
//	    });
//	    expect(parts).toEqual(nf.formatToParts(12345678));
//	  });
func TestInputsWithOptions(t *testing.T) {
	t.Run("local variable with :number expression", func(t *testing.T) {
		// This test simulates a local variable with number formatting options
		// We'll test that options from the operand are preserved and merged with expression options

		// Create a NumberValue with specific options (simulating the local variable)
		operandOptions := map[string]interface{}{
			"useGrouping": "never",
		}
		numberValue := messagevalue.NewNumberValue(12345678, "en", "test", operandOptions)

		// Create context with number function
		ctx := NewContext(
			[]string{"en"},
			map[string]functions.MessageFunction{
				"number": functions.NumberFunction,
			},
			map[string]interface{}{
				"val": numberValue,
			},
			nil,
		)

		// Create function reference with additional options
		options := make(datamodel.Options)
		options["minimumFractionDigits"] = datamodel.NewLiteral("2")
		funcRef := datamodel.NewFunctionRef("number", options)

		// Create variable reference
		varRef := datamodel.NewVariableRef("val")

		// Resolve the function reference
		result := ResolveFunctionRef(ctx, varRef, funcRef)

		// Test the result
		assert.Equal(t, "number", result.Type())

		// Test that the result has the expected formatting
		str, err := result.ToString()
		require.NoError(t, err)
		// Should have minimum 2 fraction digits and no grouping
		assert.Contains(t, str, "12345678.00")
	})

	t.Run("value with options", func(t *testing.T) {
		// TypeScript original code:
		//   test('value with options', () => {
		//     const mf = new MessageFormat(
		//       'en',
		//       '{$val :number minimumFractionDigits=2}'
		//     );
		//     const val = Object.assign(new Number(12345678), {
		//       options: { minimumFractionDigits: 4, useGrouping: false }
		//     });
		//     const msg = mf.formatToParts({ val });
		//     const { parts } = msg[0] as MessageNumberPart;
		//
		//     const nf = new Intl.NumberFormat('en', {
		//       minimumFractionDigits: 2,
		//       useGrouping: false
		//     });
		//     expect(parts).toEqual(nf.formatToParts(12345678));
		//   });

		// Create a NumberValue with operand options
		operandOptions := map[string]interface{}{
			"minimumFractionDigits": 4,
			"useGrouping":           false,
		}
		numberValue := messagevalue.NewNumberValue(12345678, "en", "test", operandOptions)

		// Create context with number function
		ctx := NewContext(
			[]string{"en"},
			map[string]functions.MessageFunction{
				"number": functions.NumberFunction,
			},
			map[string]interface{}{
				"val": numberValue,
			},
			nil,
		)

		// Create function reference with expression options that override operand options
		options := make(datamodel.Options)
		options["minimumFractionDigits"] = datamodel.NewLiteral("2")
		funcRef := datamodel.NewFunctionRef("number", options)

		// Create variable reference
		varRef := datamodel.NewVariableRef("val")

		// Resolve the function reference
		result := ResolveFunctionRef(ctx, varRef, funcRef)

		// Test the result
		assert.Equal(t, "number", result.Type())

		// Test that expression options override operand options
		// minimumFractionDigits should be 2 (from expression), not 4 (from operand)
		// useGrouping should be false (from operand)
		str, err := result.ToString()
		require.NoError(t, err)
		// Should have minimum 2 fraction digits (not 4) and no grouping
		assert.Contains(t, str, "12345678.00")
	})
}

// TestTypeCastsBasedOnRuntime tests type casts based on runtime values
// TypeScript original code:
//
//	describe('Type casts based on runtime', () => {
//	  const date = '2000-01-01T15:00:00';
//
//	  test('boolean function option with literal value', () => {
//	    const mfTrue = new MessageFormat(
//	      'en',
//	      '{$date :datetime timeStyle=short hour12=true}',
//	      { functions: DraftFunctions }
//	    );
//	    expect(mfTrue.format({ date })).toMatch(/3:00/);
//	    const mfFalse = new MessageFormat(
//	      'en',
//	      '{$date :datetime timeStyle=short hour12=false}',
//	      { functions: DraftFunctions }
//	    );
//	    expect(mfFalse.format({ date })).toMatch(/15:00/);
//	  });
func TestTypeCastsBasedOnRuntime(t *testing.T) {
	date := "2000-01-01T15:00:00"

	t.Run("boolean function option with literal value", func(t *testing.T) {
		// Create a mock datetime function for testing
		datetimeFunc := func(
			ctx functions.MessageFunctionContext,
			options map[string]interface{},
			operand interface{},
		) messagevalue.MessageValue {
			// Mock implementation that respects hour12 option
			hour12 := options["hour12"]

			var result string
			if hour12 == "true" || hour12 == true {
				result = "3:00 PM" // 12-hour format
			} else {
				result = "15:00" // 24-hour format
			}

			return messagevalue.NewStringValue(result, "en", ctx.Source())
		}

		// Test with hour12=true
		ctx1 := NewContext(
			[]string{"en"},
			map[string]functions.MessageFunction{
				"datetime": datetimeFunc,
			},
			map[string]interface{}{
				"date": date,
			},
			nil,
		)

		options1 := make(datamodel.Options)
		options1["timeStyle"] = datamodel.NewLiteral("short")
		options1["hour12"] = datamodel.NewLiteral("true")
		funcRef1 := datamodel.NewFunctionRef("datetime", options1)
		varRef1 := datamodel.NewVariableRef("date")

		result1 := ResolveFunctionRef(ctx1, varRef1, funcRef1)
		str1, err := result1.ToString()
		require.NoError(t, err)
		assert.Contains(t, str1, "3:00")

		// Test with hour12=false
		ctx2 := NewContext(
			[]string{"en"},
			map[string]functions.MessageFunction{
				"datetime": datetimeFunc,
			},
			map[string]interface{}{
				"date": date,
			},
			nil,
		)

		options2 := make(datamodel.Options)
		options2["timeStyle"] = datamodel.NewLiteral("short")
		options2["hour12"] = datamodel.NewLiteral("false")
		funcRef2 := datamodel.NewFunctionRef("datetime", options2)
		varRef2 := datamodel.NewVariableRef("date")

		result2 := ResolveFunctionRef(ctx2, varRef2, funcRef2)
		str2, err := result2.ToString()
		require.NoError(t, err)
		assert.Contains(t, str2, "15:00")
	})

	t.Run("boolean function option with variable value", func(t *testing.T) {
		// TypeScript original code:
		//   test('boolean function option with variable value', () => {
		//     const mf = new MessageFormat(
		//       'en',
		//       '{$date :datetime timeStyle=short hour12=$hour12}',
		//       { functions: DraftFunctions }
		//     );
		//     expect(mf.format({ date, hour12: 'false' })).toMatch(/15:00/);
		//     expect(mf.format({ date, hour12: false })).toMatch(/15:00/);
		//   });

		// Create a mock datetime function that handles variable hour12 values
		datetimeFunc := func(
			ctx functions.MessageFunctionContext,
			options map[string]interface{},
			operand interface{},
		) messagevalue.MessageValue {
			hour12 := options["hour12"]

			var result string
			// Handle both string "false" and boolean false
			if hour12 == "false" || hour12 == false {
				result = "15:00" // 24-hour format
			} else {
				result = "3:00 PM" // 12-hour format
			}

			return messagevalue.NewStringValue(result, "en", ctx.Source())
		}

		// Test with hour12="false" (string)
		ctx1 := NewContext(
			[]string{"en"},
			map[string]functions.MessageFunction{
				"datetime": datetimeFunc,
			},
			map[string]interface{}{
				"date":   date,
				"hour12": "false",
			},
			nil,
		)

		options1 := make(datamodel.Options)
		options1["timeStyle"] = datamodel.NewLiteral("short")
		options1["hour12"] = datamodel.NewVariableRef("hour12")
		funcRef1 := datamodel.NewFunctionRef("datetime", options1)
		varRef1 := datamodel.NewVariableRef("date")

		result1 := ResolveFunctionRef(ctx1, varRef1, funcRef1)
		str1, err := result1.ToString()
		require.NoError(t, err)
		assert.Contains(t, str1, "15:00")

		// Test with hour12=false (boolean)
		ctx2 := NewContext(
			[]string{"en"},
			map[string]functions.MessageFunction{
				"datetime": datetimeFunc,
			},
			map[string]interface{}{
				"date":   date,
				"hour12": false,
			},
			nil,
		)

		options2 := make(datamodel.Options)
		options2["timeStyle"] = datamodel.NewLiteral("short")
		options2["hour12"] = datamodel.NewVariableRef("hour12")
		funcRef2 := datamodel.NewFunctionRef("datetime", options2)
		varRef2 := datamodel.NewVariableRef("date")

		result2 := ResolveFunctionRef(ctx2, varRef2, funcRef2)
		str2, err := result2.ToString()
		require.NoError(t, err)
		assert.Contains(t, str2, "15:00")
	})
}

// TestFunctionReturnIsNotMessageValue tests error handling when function returns invalid values
// TypeScript original code:
//
//	describe('Function return is not a MessageValue', () => {
//	  test('object with type, but no source', () => {
//	    const functions = { fail: () => ({ type: 'fail' }) as any };
//	    const mf = new MessageFormat('en', '{:fail}', { functions });
//	    const onError = jest.fn();
//	    expect(mf.format(undefined, onError)).toEqual('\u2068{:fail}\u2069');
//	    expect(mf.formatToParts(undefined, onError)).toEqual([
//	      { type: 'bidiIsolation', value: '\u2068' },
//	      { type: 'fallback', source: ':fail' },
//	      { type: 'bidiIsolation', value: '\u2069' }
//	    ]);
//	    expect(onError).toHaveBeenCalledTimes(2);
//	  });
func TestFunctionReturnIsNotMessageValue(t *testing.T) {
	t.Run("object with type, but no source", func(t *testing.T) {
		// Create a function that returns an invalid MessageValue (missing required methods)
		failFunc := func(
			ctx functions.MessageFunctionContext,
			options map[string]interface{},
			operand interface{},
		) messagevalue.MessageValue {
			// Return nil to simulate invalid return
			return nil
		}

		var errorCalled bool
		onError := func(err error) {
			errorCalled = true
		}

		ctx := NewContext(
			[]string{"en"},
			map[string]functions.MessageFunction{
				"fail": failFunc,
			},
			map[string]interface{}{},
			onError,
		)

		// Create function reference without operand
		funcRef := datamodel.NewFunctionRef("fail", nil)

		// This should trigger error handling and return fallback
		result := ResolveFunctionRef(ctx, nil, funcRef)

		// Should return fallback value
		assert.Equal(t, "fallback", result.Type())
		assert.Equal(t, ":fail", result.Source())

		// Error handler should have been called
		assert.True(t, errorCalled)
	})

	t.Run("null", func(t *testing.T) {
		// TypeScript original code:
		//   test('null', () => {
		//     const functions = { fail: () => null as any };
		//     const mf = new MessageFormat('en', '{42 :fail}', { functions });
		//     const onError = jest.fn();
		//     expect(mf.format(undefined, onError)).toEqual('\u2068{|42|}\u2069');
		//     expect(mf.formatToParts(undefined, onError)).toEqual([
		//       { type: 'bidiIsolation', value: '\u2068' },
		//       { type: 'fallback', source: '|42|' },
		//       { type: 'bidiIsolation', value: '\u2069' }
		//     ]);
		//     expect(onError).toHaveBeenCalledTimes(2);
		//   });

		// Create a function that returns nil
		failFunc := func(
			ctx functions.MessageFunctionContext,
			options map[string]interface{},
			operand interface{},
		) messagevalue.MessageValue {
			return nil
		}

		var errorCalled bool
		onError := func(err error) {
			errorCalled = true
		}

		ctx := NewContext(
			[]string{"en"},
			map[string]functions.MessageFunction{
				"fail": failFunc,
			},
			map[string]interface{}{},
			onError,
		)

		// Create function reference with literal operand
		literal := datamodel.NewLiteral("42")
		funcRef := datamodel.NewFunctionRef("fail", nil)

		// This should trigger error handling and return fallback
		result := ResolveFunctionRef(ctx, literal, funcRef)

		// Should return fallback value with operand source
		assert.Equal(t, "fallback", result.Type())
		assert.Equal(t, "|42|", result.Source())

		// Error handler should have been called
		assert.True(t, errorCalled)
	})

	t.Run("Object.p.toString used as function", func(t *testing.T) {
		// TypeScript original code:
		//   test('Object.p.toString used as function', () => {
		//     const mf = new MessageFormat('en', '{13 :toString}', { functions: {} });
		//     const onError = jest.fn();
		//     expect(mf.format(undefined, onError)).toEqual('\u2068{|13|}\u2069');
		//     expect(onError.mock.calls).toMatchObject([[{ type: 'unknown-function' }]]);
		//   });

		var errorCalled bool
		var errorType string
		onError := func(err error) {
			errorCalled = true
			// Check if it's a resolution error with unknown-function type
			var resErr *pkgErrors.MessageResolutionError
			if errors.As(err, &resErr) {
				errorType = resErr.GetType()
			}
		}

		ctx := NewContext(
			[]string{"en"},
			map[string]functions.MessageFunction{}, // Empty functions map
			map[string]interface{}{},
			onError,
		)

		// Create function reference with literal operand and unknown function
		literal := datamodel.NewLiteral("13")
		funcRef := datamodel.NewFunctionRef("toString", nil)

		// This should trigger unknown-function error and return fallback
		result := ResolveFunctionRef(ctx, literal, funcRef)

		// Should return fallback value with operand source
		assert.Equal(t, "fallback", result.Type())
		assert.Equal(t, "|13|", result.Source())

		// Error handler should have been called with unknown-function error
		assert.True(t, errorCalled)
		assert.Equal(t, "unknown-function", errorType)
	})
}
