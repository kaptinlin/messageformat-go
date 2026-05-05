package functions

import (
	"errors"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/kaptinlin/messageformat-go/pkg/bidi"
	pkgerrors "github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageFunctionContextCarriesMetadata(t *testing.T) {
	t.Parallel()

	ctx := NewMessageFunctionContext(
		[]string{"ar", "en"},
		"{$value :number}",
		"lookup",
		nil,
		nil,
		"rtl",
		"expr-id",
	)

	assert.Equal(t, "rtl", ctx.Dir())
	assert.Equal(t, "expr-id", ctx.ID())
	assert.Equal(t, "{$value :number}", ctx.Source())
	assert.Equal(t, "lookup", ctx.LocaleMatcher())
	if diff := cmp.Diff([]string{"ar", "en"}, ctx.Locales()); diff != "" {
		t.Errorf("Locales() mismatch (-want +got):\n%s", diff)
	}
	require.NotNil(t, ctx.LiteralOptionKeys())
	assert.False(t, ctx.LiteralOptionKeys()["select"])

	ctx.OnError(errors.New("ignored"))
}

func TestNumberFunctionValidatesSelectionOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		literalOptionKeys map[string]bool
		options           map[string]any
		operand           any
	}{
		{
			name:              "select option must be literal",
			literalOptionKeys: map[string]bool{"select": false},
			options:           map[string]any{"select": "exact"},
			operand:           42,
		},
		{
			name:              "select value must be recognized",
			literalOptionKeys: map[string]bool{"select": true},
			options:           map[string]any{"select": "range"},
			operand:           42,
		},
		{
			name:              "select option must be string",
			literalOptionKeys: map[string]bool{"select": true},
			operand: map[string]any{
				"valueOf": 42,
				"options": map[string]any{"select": 7},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var errs []error
			ctx := NewMessageFunctionContext(
				[]string{"en"},
				"test source",
				"best fit",
				func(err error) { errs = append(errs, err) },
				tc.literalOptionKeys,
				"",
				"",
			)

			result := NumberFunction(ctx, tc.options, tc.operand)
			require.NotNil(t, result)
			assert.Equal(t, "number", result.Type())
			require.Len(t, errs, 1)
			assertResolutionErrorType(t, errs[0], pkgerrors.ErrorTypeBadOption)
		})
	}
}

func TestNumberFunctionUsesLocaleDirectionWhenContextHasNoDirection(t *testing.T) {
	t.Parallel()

	ctx := NewMessageFunctionContext(
		[]string{"ar-SA"},
		"test source",
		"best fit",
		nil,
		nil,
		"",
		"",
	)

	result := NumberFunction(ctx, nil, 42)
	require.NotNil(t, result)
	assert.Equal(t, bidi.DirRTL, result.Dir())
}

func TestReadNumericOperandRejectsFallbackAndValueErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		{
			name:  "fallback value",
			value: messagevalue.NewFallbackValue("missing", "en"),
		},
		{
			name:  "message value whose value cannot be read",
			value: failingMessageValue{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := readNumericOperand(tc.value, "test source")
			require.Error(t, err)
			assert.Nil(t, result)
			assertResolutionErrorType(t, err, pkgerrors.ErrorTypeBadOperand)
		})
	}
}

func TestIntegerFunctionRoundsSupportedOperandsAndReportsBadOptions(t *testing.T) {
	t.Parallel()

	t.Run("rounds finite floats", func(t *testing.T) {
		t.Parallel()

		ctx := newTestContext(nil)
		result := IntegerFunction(ctx, nil, float32(3.6))
		require.NotNil(t, result)
		got, err := result.ToString()
		require.NoError(t, err)
		assert.Equal(t, "4", got)
	})

	t.Run("converts finite big floats to integers", func(t *testing.T) {
		t.Parallel()

		ctx := newTestContext(nil)
		result := IntegerFunction(ctx, nil, big.NewFloat(3.9))
		require.NotNil(t, result)
		got, err := result.ToString()
		require.NoError(t, err)
		assert.Equal(t, "3", got)
	})

	t.Run("preserves infinite floats", func(t *testing.T) {
		t.Parallel()

		ctx := newTestContext(nil)
		result := IntegerFunction(ctx, nil, math.Inf(1))
		require.NotNil(t, result)
		got, err := result.ValueOf()
		require.NoError(t, err)
		assert.Equal(t, math.Inf(1), got)
	})

	t.Run("reports invalid option values", func(t *testing.T) {
		t.Parallel()

		var errs []error
		ctx := newTestContext(func(err error) { errs = append(errs, err) })
		result := IntegerFunction(ctx, map[string]any{
			"minimumIntegerDigits": -1,
			"signDisplay":          123,
		}, 42)
		require.NotNil(t, result)
		assert.Equal(t, "number", result.Type())
		require.Len(t, errs, 2)
		for _, err := range errs {
			assertResolutionErrorType(t, err, pkgerrors.ErrorTypeBadOption)
		}
	})
}

func TestCurrencyFunctionReportsTypedOptionErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		options       map[string]any
		wantErrorType string
	}{
		{
			name:          "string option rejects non-string value",
			options:       map[string]any{"currency": "USD", "currencySign": 42},
			wantErrorType: pkgerrors.ErrorTypeBadOption,
		},
		{
			name:          "positive integer option rejects negative value",
			options:       map[string]any{"currency": "USD", "minimumIntegerDigits": -1},
			wantErrorType: pkgerrors.ErrorTypeBadOption,
		},
		{
			name:          "currency display rejects non-string value",
			options:       map[string]any{"currency": "USD", "currencyDisplay": 99},
			wantErrorType: pkgerrors.ErrorTypeBadOption,
		},
		{
			name:          "currency display never is unsupported",
			options:       map[string]any{"currency": "USD", "currencyDisplay": "never"},
			wantErrorType: pkgerrors.ErrorTypeUnsupportedOperation,
		},
		{
			name:          "fraction digits rejects non-integer string",
			options:       map[string]any{"currency": "USD", "fractionDigits": "1.5"},
			wantErrorType: pkgerrors.ErrorTypeBadOption,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var errs []error
			ctx := newTestContext(func(err error) { errs = append(errs, err) })
			result := CurrencyFunction(ctx, tc.options, 42)
			require.NotNil(t, result)
			assert.Equal(t, "number", result.Type())
			require.Len(t, errs, 1)
			assertResolutionErrorType(t, errs[0], tc.wantErrorType)
		})
	}
}

func TestCurrencyFunctionReportsMissingCurrencyAsTypedError(t *testing.T) {
	t.Parallel()

	var errs []error
	ctx := newTestContext(func(err error) { errs = append(errs, err) })
	result := CurrencyFunction(ctx, nil, 42)
	require.NotNil(t, result)
	assert.Equal(t, "fallback", result.Type())
	require.Len(t, errs, 1)
	assertResolutionErrorType(t, errs[0], pkgerrors.ErrorTypeBadOperand)
}

func TestMathFunctionAcceptsNumericOperandFamilies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		operand any
		want    string
	}{
		{name: "int8", operand: int8(5), want: "7"},
		{name: "int16", operand: int16(5), want: "7"},
		{name: "int32", operand: int32(5), want: "7"},
		{name: "uint", operand: uint(5), want: "7"},
		{name: "uint8", operand: uint8(5), want: "7"},
		{name: "uint16", operand: uint16(5), want: "7"},
		{name: "uint32", operand: uint32(5), want: "7"},
		{name: "uint64", operand: uint64(5), want: "7"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := newTestContext(nil)
			result := MathFunction(ctx, map[string]any{"add": 2}, tc.operand)
			require.NotNil(t, result)
			got, err := result.ToString()
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestDateTimeFunctionsAcceptNumericAndMessageValueOperands(t *testing.T) {
	t.Parallel()

	t.Run("unix timestamp integer", func(t *testing.T) {
		t.Parallel()

		ctx := newTestContext(nil)
		result := DatetimeFunction(ctx, nil, int64(1136214245))
		require.NotNil(t, result)
		got, err := result.ValueOf()
		require.NoError(t, err)
		assert.Equal(t, time.Unix(1136214245, 0), got)
	})

	t.Run("unix timestamp float", func(t *testing.T) {
		t.Parallel()

		ctx := newTestContext(nil)
		result := DatetimeFunction(ctx, nil, float64(1136214245))
		require.NotNil(t, result)
		got, err := result.ValueOf()
		require.NoError(t, err)
		assert.Equal(t, time.Unix(1136214245, 0), got)
	})

	t.Run("message value falls back to string representation", func(t *testing.T) {
		t.Parallel()

		ctx := newTestContext(nil)
		result := DateFunction(ctx, nil, dateStringMessageValue{value: "2006-01-02"})
		require.NotNil(t, result)
		got, err := result.ValueOf()
		require.NoError(t, err)
		assert.Equal(t, time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC), got)
	})
}

func TestDateTimeFunctionsReportBadOptions(t *testing.T) {
	t.Parallel()

	var errs []error
	ctx := newTestContext(func(err error) { errs = append(errs, err) })
	result := DatetimeFunction(ctx, map[string]any{
		"dateFields":    "bad",
		"timePrecision": 7,
		"timeZoneStyle": "invalid",
		"calendar":      3,
		"hour12":        "sometimes",
		"timeZone":      5,
	}, "2006-01-02T15:04:05Z")
	require.NotNil(t, result)
	assert.Equal(t, "datetime", result.Type())
	require.Len(t, errs, 6)
	for _, err := range errs {
		assertResolutionErrorType(t, err, pkgerrors.ErrorTypeBadOption)
	}
}

type failingMessageValue struct{}

func (failingMessageValue) Type() string                                 { return "failing" }
func (failingMessageValue) Source() string                               { return "failing source" }
func (failingMessageValue) Dir() bidi.Direction                          { return bidi.DirAuto }
func (failingMessageValue) Locale() string                               { return "en" }
func (failingMessageValue) Options() map[string]any                      { return nil }
func (failingMessageValue) ToString() (string, error)                    { return "failing", nil }
func (failingMessageValue) ToParts() ([]messagevalue.MessagePart, error) { return nil, nil }
func (failingMessageValue) ValueOf() (any, error)                        { return nil, errors.New("cannot read value") }
func (failingMessageValue) SelectKeys([]string) ([]string, error)        { return nil, nil }

type dateStringMessageValue struct {
	value string
}

func (v dateStringMessageValue) Type() string                                 { return "date-string" }
func (v dateStringMessageValue) Source() string                               { return "date source" }
func (v dateStringMessageValue) Dir() bidi.Direction                          { return bidi.DirAuto }
func (v dateStringMessageValue) Locale() string                               { return "en" }
func (v dateStringMessageValue) Options() map[string]any                      { return nil }
func (v dateStringMessageValue) ToString() (string, error)                    { return v.value, nil }
func (v dateStringMessageValue) ToParts() ([]messagevalue.MessagePart, error) { return nil, nil }
func (v dateStringMessageValue) ValueOf() (any, error)                        { return nil, errors.New("use string") }
func (v dateStringMessageValue) SelectKeys([]string) ([]string, error)        { return nil, nil }
