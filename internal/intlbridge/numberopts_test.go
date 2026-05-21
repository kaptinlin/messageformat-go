package intlbridge

import (
	"testing"

	"github.com/agentable/go-intl/numberformat"
	"github.com/stretchr/testify/assert"
)

func TestNumberOptions_Empty(t *testing.T) {
	got := NumberOptions(nil)
	assert.Equal(t, numberformat.Options{}, got)

	got = NumberOptions(map[string]any{})
	assert.Equal(t, numberformat.Options{}, got)
}

func TestNumberOptions_Strings(t *testing.T) {
	got := NumberOptions(map[string]any{
		"style":               "currency",
		"currency":            "usd",
		"currencyDisplay":     "name",
		"currencySign":        "accounting",
		"unit":                "kilometer",
		"unitDisplay":         "long",
		"notation":            "compact",
		"compactDisplay":      "short",
		"signDisplay":         "exceptZero",
		"roundingMode":        "halfEven",
		"roundingPriority":    "morePrecision",
		"trailingZeroDisplay": "stripIfInteger",
		"numberingSystem":     "arab",
		"localeMatcher":       "lookup",
	})
	assert.Equal(t, numberformat.CurrencyStyle, got.Style)
	assert.Equal(t, numberformat.Currency("USD"), got.Currency)
	assert.Equal(t, numberformat.CurrencyDisplayName, got.CurrencyDisplay)
	assert.Equal(t, numberformat.AccountingCurrencySign, got.CurrencySign)
	assert.Equal(t, numberformat.Unit("kilometer"), got.Unit)
	assert.Equal(t, numberformat.LongUnitDisplay, got.UnitDisplay)
	assert.Equal(t, numberformat.CompactNotation, got.Notation)
	assert.Equal(t, numberformat.ShortCompactDisplay, got.CompactDisplay)
	assert.Equal(t, numberformat.ExceptZeroSignDisplay, got.SignDisplay)
	assert.Equal(t, numberformat.HalfEvenRoundingMode, got.RoundingMode)
	assert.Equal(t, numberformat.MorePrecisionRoundingPriority, got.RoundingPriority)
	assert.Equal(t, numberformat.StripIfIntegerTrailingZeroDisplay, got.TrailingZeroDisplay)
	assert.Equal(t, "arab", got.NumberingSystem)
	assert.Equal(t, numberformat.LookupLocaleMatcher, got.LocaleMatcher)
}

func TestNumberOptions_IntegerCounts(t *testing.T) {
	got := NumberOptions(map[string]any{
		"minimumIntegerDigits":     int64(3),
		"roundingIncrement":        5,
		"minimumFractionDigits":    2,
		"maximumFractionDigits":    4,
		"minimumSignificantDigits": 1,
		"maximumSignificantDigits": 6,
	})
	assertIntPtr(t, 3, got.MinimumIntegerDigits)
	assertIntPtr(t, 5, got.RoundingIncrement)
	assertIntPtr(t, 2, got.MinimumFractionDigits)
	assertIntPtr(t, 4, got.MaximumFractionDigits)
	assertIntPtr(t, 1, got.MinimumSignificantDigits)
	assertIntPtr(t, 6, got.MaximumSignificantDigits)
}

func TestNumberOptions_PartialDigits(t *testing.T) {
	t.Run("only min fraction", func(t *testing.T) {
		got := NumberOptions(map[string]any{"minimumFractionDigits": 2})
		assertIntPtr(t, 2, got.MinimumFractionDigits)
		assert.Nil(t, got.MaximumFractionDigits)
	})
	t.Run("only max fraction", func(t *testing.T) {
		got := NumberOptions(map[string]any{"maximumFractionDigits": 3})
		assert.Nil(t, got.MinimumFractionDigits)
		assertIntPtr(t, 3, got.MaximumFractionDigits)
	})
	t.Run("only min significant", func(t *testing.T) {
		got := NumberOptions(map[string]any{"minimumSignificantDigits": 2})
		assertIntPtr(t, 2, got.MinimumSignificantDigits)
		assert.Nil(t, got.MaximumSignificantDigits)
	})
	t.Run("only max significant", func(t *testing.T) {
		got := NumberOptions(map[string]any{"maximumSignificantDigits": 4})
		assert.Nil(t, got.MinimumSignificantDigits)
		assertIntPtr(t, 4, got.MaximumSignificantDigits)
	})
}

func TestNumberOptions_UseGrouping(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want numberformat.UseGrouping
	}{
		{"bool true", true, numberformat.UseGroupingAlways},
		{"bool false", false, numberformat.UseGroupingFalse},
		{"string never", "never", numberformat.UseGroupingFalse},
		{"string false", "false", numberformat.UseGroupingFalse},
		{"string true", "true", numberformat.UseGroupingAlways},
		{"string always", "always", numberformat.UseGroupingAlways},
		{"string min2", "min2", numberformat.UseGroupingMin2},
		{"string auto", "auto", numberformat.UseGroupingAuto},
		{"unknown drops to default", "weird", numberformat.UseGrouping("")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NumberOptions(map[string]any{"useGrouping": tc.in})
			assert.Equal(t, tc.want, got.UseGrouping)
		})
	}
}

func TestNumberOptions_NilAndUnknown(t *testing.T) {
	got := NumberOptions(map[string]any{
		"style":    nil,
		"unknown":  "ignored",
		"currency": "",
	})
	assert.Equal(t, numberformat.Options{}, got)
}

func TestNumberOptions_FractionDigitsClamp(t *testing.T) {
	t.Run("max below min clamps to min", func(t *testing.T) {
		got := NumberOptions(map[string]any{
			"minimumFractionDigits": 2,
			"maximumFractionDigits": 1,
		})
		assertIntPtr(t, 2, got.MinimumFractionDigits)
		assertIntPtr(t, 2, got.MaximumFractionDigits)
	})
	t.Run("significant max below min clamps", func(t *testing.T) {
		got := NumberOptions(map[string]any{
			"minimumSignificantDigits": 4,
			"maximumSignificantDigits": 2,
		})
		assertIntPtr(t, 4, got.MinimumSignificantDigits)
		assertIntPtr(t, 4, got.MaximumSignificantDigits)
	})
}

func TestNumberOptions_IntCoercion(t *testing.T) {
	t.Run("float coerces when integral", func(t *testing.T) {
		got := NumberOptions(map[string]any{"minimumIntegerDigits": 2.0})
		assertIntPtr(t, 2, got.MinimumIntegerDigits)
	})
	t.Run("non-integral float is dropped", func(t *testing.T) {
		got := NumberOptions(map[string]any{"minimumIntegerDigits": 2.5})
		assert.Nil(t, got.MinimumIntegerDigits)
	})
	t.Run("string digits coerce", func(t *testing.T) {
		got := NumberOptions(map[string]any{"minimumIntegerDigits": "4"})
		assertIntPtr(t, 4, got.MinimumIntegerDigits)
	})
}

func assertIntPtr(t *testing.T, want int, got *int) {
	t.Helper()
	if assert.NotNil(t, got) {
		assert.Equal(t, want, *got)
	}
}
