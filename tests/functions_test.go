package tests

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kaptinlin/messageformat-go/pkg/bidi"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type valueOfOperand struct {
	value any
	err   error
}

func (v valueOfOperand) ValueOf() (any, error) {
	return v.value, v.err
}

func newTestFunctionContext(locales []string) (functions.MessageFunctionContext, *[]error) {
	var gotErrors []error
	ctx := functions.NewMessageFunctionContext(
		locales,
		"test-source",
		"best fit",
		func(err error) { gotErrors = append(gotErrors, err) },
		map[string]bool{"decimalPlaces": true},
		"ltr",
		"test-id",
	)
	return ctx, &gotErrors
}

func TestTestFunctionsReturnsExpectedFunctions(t *testing.T) {
	t.Parallel()

	got := TestFunctions()

	assert.Len(t, got, 5)
	for _, name := range []string{"test", "test:function", "test:select", "test:format", "placeholder"} {
		assert.Contains(t, got, name)
		assert.NotNil(t, got[name])
	}
}

func TestTestValueFormatsAndSelects(t *testing.T) {
	t.Parallel()

	value := &testValue{
		source:        "source",
		input:         -12.4,
		canFormat:     true,
		selectable:    true,
		decimalPlaces: 1,
		locale:        "fr",
		dir:           bidi.DirLTR,
	}

	assert.Equal(t, "test", value.Type())
	assert.Equal(t, "source", value.Source())
	assert.Equal(t, bidi.DirLTR, value.Dir())
	assert.Equal(t, "fr", value.Locale())
	assert.Nil(t, value.Options())

	underlying, err := value.ValueOf()
	require.NoError(t, err)
	assert.InDelta(t, -12.4, underlying, 0.000001)

	gotString, err := value.ToString()
	require.NoError(t, err)
	assert.Equal(t, "-12.4", gotString)

	gotParts, err := value.ToParts()
	require.NoError(t, err)
	simplifiedParts := []map[string]any{
		{
			"type":   gotParts[0].Type(),
			"value":  gotParts[0].Value(),
			"source": gotParts[0].Source(),
			"locale": gotParts[0].Locale(),
		},
	}
	wantParts := []map[string]any{
		{
			"type":   "text",
			"value":  "-12.4",
			"source": "source",
			"locale": "und",
		},
	}
	if diff := cmp.Diff(wantParts, simplifiedParts); diff != "" {
		t.Fatalf("parts mismatch (-want +got):\n%s", diff)
	}
}

func TestTestValueSelectKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   *testValue
		keys    []string
		want    []string
		wantErr error
	}{
		{
			name: "prefers one decimal key",
			value: &testValue{
				input:         1,
				selectable:    true,
				decimalPlaces: 1,
			},
			keys: []string{"1", "1.0"},
			want: []string{"1.0"},
		},
		{
			name: "falls back to integer key",
			value: &testValue{
				input:      1,
				selectable: true,
			},
			keys: []string{"other", "1"},
			want: []string{"1"},
		},
		{
			name: "returns empty selection when no key matches",
			value: &testValue{
				input:      2,
				selectable: true,
			},
			keys: []string{"1"},
			want: []string{},
		},
		{
			name:    "rejects non-selectable value",
			value:   &testValue{},
			keys:    []string{"1"},
			wantErr: ErrNotSelectable,
		},
		{
			name: "rejects bad option",
			value: &testValue{
				selectable: true,
				badOption:  true,
			},
			keys:    []string{"1"},
			wantErr: ErrBadOption,
		},
		{
			name: "reports selection failure",
			value: &testValue{
				selectable:  true,
				failsSelect: true,
			},
			keys:    []string{"1"},
			wantErr: ErrSelectionFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.value.SelectKeys(tc.keys)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Fatalf("selection mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTestValueFormattingFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   *testValue
		want    string
		wantErr error
	}{
		{
			name:    "requires formatting capability",
			value:   &testValue{},
			wantErr: ErrNotFormattable,
		},
		{
			name: "reports formatting failure",
			value: &testValue{
				canFormat:   true,
				failsFormat: true,
			},
			wantErr: ErrFormattingFailed,
		},
		{
			name: "formats bad option marker",
			value: &testValue{
				canFormat: true,
				badOption: true,
			},
			want: "bad-option-value",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.value.ToString()
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)

				_, partsErr := tc.value.ToParts()
				require.ErrorIs(t, partsErr, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestCreateTestValueCapabilitiesAndOperands(t *testing.T) {
	t.Parallel()

	t.Run("test function formats and selects", func(t *testing.T) {
		t.Parallel()

		ctx, gotErrors := newTestFunctionContext([]string{"de"})
		value := testFunction(ctx, map[string]any{"decimalPlaces": 1}, 1)

		assert.Empty(t, *gotErrors)
		assert.Equal(t, "test", value.Type())
		assert.Equal(t, "test-source", value.Source())
		assert.Equal(t, "de", value.Locale())
		assert.Equal(t, bidi.DirAuto, value.Dir())

		gotString, err := value.ToString()
		require.NoError(t, err)
		assert.Equal(t, "1.0", gotString)

		selector := value.(messagevalue.Selector)
		gotKeys, err := selector.SelectKeys([]string{"1", "1.0"})
		require.NoError(t, err)
		if diff := cmp.Diff([]string{"1.0"}, gotKeys); diff != "" {
			t.Fatalf("selection mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("select function does not format", func(t *testing.T) {
		t.Parallel()

		ctx, _ := newTestFunctionContext(nil)
		value := testSelectFunction(ctx, map[string]any{}, int32(1))

		_, err := value.ToString()
		require.ErrorIs(t, err, ErrNotFormattable)

		selector := value.(messagevalue.Selector)
		gotKeys, err := selector.SelectKeys([]string{"1"})
		require.NoError(t, err)
		if diff := cmp.Diff([]string{"1"}, gotKeys); diff != "" {
			t.Fatalf("selection mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("format function does not select", func(t *testing.T) {
		t.Parallel()

		ctx, _ := newTestFunctionContext(nil)
		value := testFormatFunction(ctx, map[string]any{}, int64(3))

		gotString, err := value.ToString()
		require.NoError(t, err)
		assert.Equal(t, "3", gotString)

		selector := value.(messagevalue.Selector)
		_, err = selector.SelectKeys([]string{"3"})
		require.ErrorIs(t, err, ErrNotSelectable)
	})

	t.Run("fallback operand stays fallback", func(t *testing.T) {
		t.Parallel()

		ctx, _ := newTestFunctionContext([]string{"fr"})
		value := testFunction(ctx, map[string]any{}, messagevalue.NewFallbackValue("missing", "en"))

		assert.Equal(t, "fallback", value.Type())
		assert.Equal(t, "test-source", value.Source())
		assert.Equal(t, "fr", value.Locale())
	})

	t.Run("inherited test value propagates behavior", func(t *testing.T) {
		t.Parallel()

		ctx, gotErrors := newTestFunctionContext(nil)
		parent := &testValue{
			input:         1,
			decimalPlaces: 1,
			badOption:     true,
			failsFormat:   true,
			failsSelect:   true,
		}

		value := testFunction(ctx, map[string]any{}, parent)

		assert.ErrorIs(t, errors.Join(*gotErrors...), ErrBadOption)

		_, err := value.ToString()
		require.ErrorIs(t, err, ErrFormattingFailed)

		selector := value.(messagevalue.Selector)
		_, err = selector.SelectKeys([]string{"1", "1.0"})
		require.ErrorIs(t, err, ErrBadOption)
	})
}

func TestCreateTestValueOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		options       map[string]any
		wantErrors    []error
		wantSelectErr error
		wantFormatErr error
	}{
		{
			name:          "invalid decimal places reports bad option",
			options:       map[string]any{"decimalPlaces": 2},
			wantErrors:    []error{ErrInvalidDecimalPlaces, ErrBadOption},
			wantSelectErr: ErrBadOption,
		},
		{
			name:          "invalid decimal places type reports bad option",
			options:       map[string]any{"decimalPlaces": "many"},
			wantErrors:    []error{ErrInvalidDecimalPlaces, ErrBadOption},
			wantSelectErr: ErrBadOption,
		},
		{
			name:          "fails select affects selection only",
			options:       map[string]any{"fails": "select"},
			wantSelectErr: ErrSelectionFailed,
		},
		{
			name:          "fails format affects formatting only",
			options:       map[string]any{"fails": "format"},
			wantFormatErr: ErrFormattingFailed,
		},
		{
			name:          "fails always affects both operations",
			options:       map[string]any{"fails": "always"},
			wantSelectErr: ErrSelectionFailed,
			wantFormatErr: ErrFormattingFailed,
		},
		{
			name:       "invalid fails value reports option error",
			options:    map[string]any{"fails": "sometimes"},
			wantErrors: []error{ErrInvalidFailsOption},
		},
		{
			name:       "invalid fails type reports option error",
			options:    map[string]any{"fails": 1},
			wantErrors: []error{ErrInvalidFailsOption},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, gotErrors := newTestFunctionContext(nil)
			value := testFunction(ctx, tc.options, 1)

			joinedErrors := errors.Join(*gotErrors...)
			for _, wantErr := range tc.wantErrors {
				assert.ErrorIs(t, joinedErrors, wantErr)
			}
			if len(tc.wantErrors) == 0 {
				assert.Empty(t, *gotErrors)
			}

			selector := value.(messagevalue.Selector)
			_, selectErr := selector.SelectKeys([]string{"1"})
			if tc.wantSelectErr != nil {
				assert.ErrorIs(t, selectErr, tc.wantSelectErr)
			} else if !assert.NoError(t, selectErr) {
				return
			}

			_, formatErr := value.ToString()
			if tc.wantFormatErr != nil {
				assert.ErrorIs(t, formatErr, tc.wantFormatErr)
			} else if !assert.NoError(t, formatErr) {
				return
			}
		})
	}
}

func TestParseNumericInputStrict(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		want    float64
		wantErr error
	}{
		{name: "string", input: "42.5", want: 42.5},
		{name: "int", input: int(3), want: 3},
		{name: "int32", input: int32(4), want: 4},
		{name: "int64", input: int64(5), want: 5},
		{name: "float32", input: float32(6.5), want: 6.5},
		{name: "float64", input: float64(7.5), want: 7.5},
		{name: "valueOf", input: valueOfOperand{value: "8.5"}, want: 8.5},
		{name: "invalid string", input: "nope", wantErr: ErrInvalidNumeric},
		{name: "unsupported type", input: struct{}{}, wantErr: ErrInvalidNumeric},
		{name: "valueOf error", input: valueOfOperand{err: ErrInvalidNumeric}, wantErr: ErrInvalidNumeric},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseNumericInputStrict(tc.input)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.InDelta(t, tc.want, got, 0.000001)
		})
	}
}

func TestOptionCoercionHelpers(t *testing.T) {
	t.Parallel()

	t.Run("positive integers", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name    string
			input   any
			want    int
			wantErr error
		}{
			{name: "zero", input: 0, want: 0},
			{name: "positive int", input: 2, want: 2},
			{name: "integral float", input: 3.0, want: 3},
			{name: "numeric string", input: "4", want: 4},
			{name: "negative int", input: -1, wantErr: ErrNotPositiveInt},
			{name: "fractional float", input: 1.5, wantErr: ErrNotPositiveInt},
			{name: "non-numeric string", input: "x", wantErr: ErrNotPositiveInt},
			{name: "unsupported type", input: uint(1), wantErr: ErrNotPositiveInt},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				got, err := asPositiveInteger(tc.input)
				if tc.wantErr != nil {
					require.ErrorIs(t, err, tc.wantErr)
					return
				}
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			})
		}
	})

	t.Run("strings", func(t *testing.T) {
		t.Parallel()

		got, err := asString("always")
		require.NoError(t, err)
		assert.Equal(t, "always", got)

		_, err = asString(1)
		require.ErrorIs(t, err, ErrNotString)
	})
}

func TestPlaceholderFunction(t *testing.T) {
	t.Parallel()

	ctx, _ := newTestFunctionContext([]string{"es"})
	value := placeholderFunction(ctx, nil, nil)

	assert.Equal(t, "string", value.Type())
	assert.Equal(t, "test-source", value.Source())
	assert.Equal(t, "es", value.Locale())

	got, err := value.ToString()
	require.NoError(t, err)
	assert.Equal(t, "placeholder", got)
}
