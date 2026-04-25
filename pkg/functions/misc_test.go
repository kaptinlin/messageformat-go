package functions

import (
	"errors"
	"testing"
	"time"

	pkgerrors "github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFallbackFunction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		source     string
		wantSource string
		wantString string
	}{
		{name: "uses provided source", source: "missing", wantSource: "missing", wantString: "{missing}"},
		{name: "defaults empty source", source: "", wantSource: "�", wantString: "{�}"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := FallbackFunction(tc.source, "en")

			assert.Equal(t, "fallback", result.Type())
			assert.Equal(t, tc.wantSource, result.Source())
			assert.Equal(t, "en", result.Locale())
			str, err := result.ToString()
			require.NoError(t, err)
			assert.Equal(t, tc.wantString, str)
		})
	}
}

func TestUnknownFunctionPreservesInput(t *testing.T) {
	t.Parallel()

	result := UnknownFunction("test source", 42, "en")

	assert.Equal(t, "unknown", result.Type())
	assert.Equal(t, "test source", result.Source())
	assert.Equal(t, "en", result.Locale())
	str, err := result.ToString()
	require.NoError(t, err)
	assert.Equal(t, "42", str)
	value, err := result.ValueOf()
	require.NoError(t, err)
	assert.Equal(t, 42, value)
}

func TestMathFunctionAppliesDelta(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		options map[string]any
		operand any
		want    any
	}{
		{name: "adds integer", options: map[string]any{"add": 3}, operand: 7, want: 10},
		{name: "subtracts integer", options: map[string]any{"subtract": 4}, operand: int64(10), want: int64(6)},
		{name: "adds float", options: map[string]any{"add": 2}, operand: 1.5, want: 3.5},
		{name: "adds int32 operand", options: map[string]any{"add": 2}, operand: int32(5), want: 7.0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var errs []error
			result := MathFunction(newTestContext(func(err error) {
				errs = append(errs, err)
			}), tc.options, tc.operand)

			require.NotNil(t, result)
			assert.Equal(t, "number", result.Type())
			value, err := result.ValueOf()
			require.NoError(t, err)
			assert.Equal(t, tc.want, value)
			assert.Empty(t, errs)
		})
	}
}

func TestMathFunctionReportsOptionErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		options map[string]any
	}{
		{name: "missing operation", options: map[string]any{}},
		{name: "conflicting operations", options: map[string]any{"add": 1, "subtract": 1}},
		{name: "invalid add", options: map[string]any{"add": "bad"}},
		{name: "invalid subtract", options: map[string]any{"subtract": "bad"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var errs []error
			result := MathFunction(newTestContext(func(err error) {
				errs = append(errs, err)
			}), tc.options, 10)

			require.NotNil(t, result)
			assert.Equal(t, "fallback", result.Type())
			require.Len(t, errs, 1)
			assertResolutionErrorType(t, errs[0], pkgerrors.ErrorTypeBadOption)
		})
	}
}

func TestIntegerFunctionAppliesOptions(t *testing.T) {
	t.Parallel()

	var errs []error
	result := IntegerFunction(newTestContext(func(err error) {
		errs = append(errs, err)
	}), map[string]any{
		"minimumIntegerDigits": 3,
		"signDisplay":          "always",
	}, 3.7)

	require.NotNil(t, result)
	assert.Equal(t, "number", result.Type())
	value, err := result.ValueOf()
	require.NoError(t, err)
	assert.Equal(t, int64(4), value)
	assert.Equal(t, 3, result.Options()["minimumIntegerDigits"])
	assert.Equal(t, "always", result.Options()["signDisplay"])
	assert.Empty(t, errs)
}

func TestIntegerFunctionReportsOptionErrors(t *testing.T) {
	t.Parallel()

	var errs []error
	result := IntegerFunction(newTestContext(func(err error) {
		errs = append(errs, err)
	}), map[string]any{
		"minimumIntegerDigits": "bad",
		"signDisplay":          10,
	}, 4)

	require.NotNil(t, result)
	assert.Equal(t, "number", result.Type())
	require.Len(t, errs, 2)
	for _, err := range errs {
		assertResolutionErrorType(t, err, pkgerrors.ErrorTypeBadOption)
	}
}

func TestPercentFunctionAppliesOptions(t *testing.T) {
	t.Parallel()

	var errs []error
	result := PercentFunction(newTestContext(func(err error) {
		errs = append(errs, err)
	}), map[string]any{
		"minimumFractionDigits": 1,
		"signDisplay":           "always",
	}, 0.5)

	require.NotNil(t, result)
	assert.Equal(t, "number", result.Type())
	assert.Equal(t, 1, result.Options()["minimumFractionDigits"])
	assert.Equal(t, "always", result.Options()["signDisplay"])
	str, err := result.ToString()
	require.NoError(t, err)
	assert.Equal(t, "+50.0%", str)
	assert.Empty(t, errs)
}

func TestPercentFunctionReportsOptionErrors(t *testing.T) {
	t.Parallel()

	var errs []error
	result := PercentFunction(newTestContext(func(err error) {
		errs = append(errs, err)
	}), map[string]any{
		"minimumFractionDigits": "bad",
		"signDisplay":           10,
	}, 0.5)

	require.NotNil(t, result)
	assert.Equal(t, "number", result.Type())
	require.Len(t, errs, 2)
	for _, err := range errs {
		assertResolutionErrorType(t, err, pkgerrors.ErrorTypeBadOption)
	}
}

func TestUnitFunctionReportsOptionErrors(t *testing.T) {
	t.Parallel()

	var errs []error
	result := UnitFunction(newTestContext(func(err error) {
		errs = append(errs, err)
	}), map[string]any{
		"unit":                  "meter",
		"minimumFractionDigits": "bad",
		"signDisplay":           10,
	}, 2)

	require.NotNil(t, result)
	assert.Equal(t, "number", result.Type())
	require.Len(t, errs, 2)
	for _, err := range errs {
		assertResolutionErrorType(t, err, pkgerrors.ErrorTypeBadOption)
	}
}

func TestDateAndTimeFunctions(t *testing.T) {
	t.Parallel()

	instant := time.Date(2026, 4, 26, 15, 30, 45, 0, time.UTC)
	tests := []struct {
		name        string
		format      func(MessageFunctionContext, map[string]any, any) any
		operand     any
		options     map[string]any
		wantOptions map[string]any
	}{
		{
			name: "date uses fields aliases",
			format: func(ctx MessageFunctionContext, opts map[string]any, operand any) any {
				return DateFunction(ctx, opts, operand)
			},
			operand: map[string]any{
				"valueOf": instant,
				"options": map[string]any{
					"calendar": "gregory",
					"hour12":   true,
					"timeZone": "UTC",
				},
			},
			options: map[string]any{
				"fields":   "month-day",
				"length":   "short",
				"calendar": "iso8601",
				"hour12":   false,
				"timeZone": "input",
			},
			wantOptions: map[string]any{
				"dateFields":    "month-day",
				"dateLength":    "short",
				"localeMatcher": "best fit",
				"calendar":      "iso8601",
				"timeZone":      "UTC",
			},
		},
		{
			name: "time uses precision alias",
			format: func(ctx MessageFunctionContext, opts map[string]any, operand any) any {
				return TimeFunction(ctx, opts, operand)
			},
			operand: instant,
			options: map[string]any{
				"precision":     "second",
				"timeZoneStyle": "short",
				"hour12":        true,
			},
			wantOptions: map[string]any{
				"timePrecision": "second",
				"timeZoneStyle": "short",
				"localeMatcher": "best fit",
				"hour12":        true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var errs []error
			result, ok := tc.format(newTestContext(func(err error) {
				errs = append(errs, err)
			}), tc.options, tc.operand).(interface {
				Type() string
				ValueOf() (any, error)
				Options() map[string]any
			})
			require.True(t, ok)
			assert.Equal(t, "datetime", result.Type())
			value, err := result.ValueOf()
			require.NoError(t, err)
			assert.Equal(t, instant, value)
			for key, want := range tc.wantOptions {
				assert.Equal(t, want, result.Options()[key])
			}
			assert.Empty(t, errs)
		})
	}
}

func TestDateTimeFunctionReportsBadOperand(t *testing.T) {
	t.Parallel()

	var errs []error
	result := DateFunction(newTestContext(func(err error) {
		errs = append(errs, err)
	}), nil, struct{}{})

	require.NotNil(t, result)
	assert.Equal(t, "fallback", result.Type())
	require.Len(t, errs, 1)
	assertFunctionErrorType(t, errs[0], pkgerrors.ErrorTypeBadOperand)
}

func TestDateTimeFunctionReportsOptionErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		format    func(MessageFunctionContext, map[string]any, any) any
		operand   any
		options   map[string]any
		wantTypes []string
	}{
		{
			name: "datetime rejects invalid date and time options",
			format: func(ctx MessageFunctionContext, opts map[string]any, operand any) any {
				return DatetimeFunction(ctx, opts, operand)
			},
			operand: time.Date(2026, 4, 26, 15, 30, 45, 0, time.UTC),
			options: map[string]any{
				"dateFields":    "bad",
				"dateLength":    "bad",
				"timePrecision": "bad",
				"timeZoneStyle": "bad",
			},
			wantTypes: []string{
				pkgerrors.ErrorTypeBadOption,
				pkgerrors.ErrorTypeBadOption,
				pkgerrors.ErrorTypeBadOption,
				pkgerrors.ErrorTypeBadOption,
			},
		},
		{
			name: "date rejects missing input timezone",
			format: func(ctx MessageFunctionContext, opts map[string]any, operand any) any {
				return DateFunction(ctx, opts, operand)
			},
			operand: time.Date(2026, 4, 26, 15, 30, 45, 0, time.UTC),
			options: map[string]any{
				"timeZone": "input",
			},
			wantTypes: []string{pkgerrors.ErrorTypeBadOperand},
		},
		{
			name: "time rejects timezone conversion",
			format: func(ctx MessageFunctionContext, opts map[string]any, operand any) any {
				return TimeFunction(ctx, opts, operand)
			},
			operand: map[string]any{
				"valueOf": time.Date(2026, 4, 26, 15, 30, 45, 0, time.UTC),
				"options": map[string]any{
					"timeZone": "UTC",
				},
			},
			options: map[string]any{
				"timeZone": "America/New_York",
			},
			wantTypes: []string{pkgerrors.ErrorTypeBadOption},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var errs []error
			result := tc.format(newTestContext(func(err error) {
				errs = append(errs, err)
			}), tc.options, tc.operand)

			require.NotNil(t, result)
			require.Len(t, errs, len(tc.wantTypes))
			for i, wantType := range tc.wantTypes {
				var functionErr *pkgerrors.MessageFunctionError
				if errors.As(errs[i], &functionErr) {
					assertFunctionErrorType(t, errs[i], wantType)
				} else {
					assertResolutionErrorType(t, errs[i], wantType)
				}
			}
		})
	}
}
