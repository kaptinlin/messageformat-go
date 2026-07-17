package intlbridge

import (
	"strconv"
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
	assertStringPtr(t, string(numberformat.CurrencyStyle), got.Style)
	assertStringPtr(t, "USD", got.Currency)
	assertStringPtr(t, string(numberformat.CurrencyDisplayName), got.CurrencyDisplay)
	assertStringPtr(t, string(numberformat.AccountingCurrencySign), got.CurrencySign)
	assertStringPtr(t, "kilometer", got.Unit)
	assertStringPtr(t, string(numberformat.LongUnitDisplay), got.UnitDisplay)
	assertStringPtr(t, string(numberformat.CompactNotation), got.Notation)
	assertStringPtr(t, string(numberformat.ShortCompactDisplay), got.CompactDisplay)
	assertStringPtr(t, string(numberformat.ExceptZeroSignDisplay), got.SignDisplay)
	assertStringPtr(t, string(numberformat.HalfEvenRoundingMode), got.RoundingMode)
	assertStringPtr(t, string(numberformat.MorePrecisionRoundingPriority), got.RoundingPriority)
	assertStringPtr(t, string(numberformat.StripIfIntegerTrailingZeroDisplay), got.TrailingZeroDisplay)
	assertStringPtr(t, "arab", got.NumberingSystem)
	assertStringPtr(t, string(numberformat.LookupLocaleMatcher), got.LocaleMatcher)
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
			assertOptionalStringPtr(t, string(tc.want), got.UseGrouping)
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

func TestNumberOptions_PreservesInvalidDigitRangesForValidation(t *testing.T) {
	t.Run("fraction max below min remains unchanged", func(t *testing.T) {
		got := NumberOptions(map[string]any{
			"minimumFractionDigits": 2,
			"maximumFractionDigits": 1,
		})
		assertIntPtr(t, 2, got.MinimumFractionDigits)
		assertIntPtr(t, 1, got.MaximumFractionDigits)
	})
	t.Run("significant max below min remains unchanged", func(t *testing.T) {
		got := NumberOptions(map[string]any{
			"minimumSignificantDigits": 4,
			"maximumSignificantDigits": 2,
		})
		assertIntPtr(t, 4, got.MinimumSignificantDigits)
		assertIntPtr(t, 2, got.MaximumSignificantDigits)
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

func TestNumberOptions_IntCoercionAcceptedTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   any
	}{
		{name: "int8", in: int8(4)},
		{name: "int16", in: int16(4)},
		{name: "int32", in: int32(4)},
		{name: "uint", in: uint(4)},
		{name: "uint8", in: uint8(4)},
		{name: "uint16", in: uint16(4)},
		{name: "uint32", in: uint32(4)},
		{name: "uint64", in: uint64(4)},
		{name: "float32", in: float32(4)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NumberOptions(map[string]any{"minimumIntegerDigits": tc.in})
			assertIntPtr(t, 4, got.MinimumIntegerDigits)
		})
	}
}

func TestNumberOptions_IntCoercionRejectsMalformedValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   any
	}{
		{name: "string with trailing junk", in: "4px"},
		{name: "string with leading space", in: " 4"},
		{name: "empty string", in: ""},
		{name: "bool", in: true},
		{name: "uint64 over int max", in: uint64(1) << (strconv.IntSize - 1)},
		{name: "integral float64 over int max", in: float64(uint64(1) << (strconv.IntSize - 1))},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NumberOptions(map[string]any{"minimumIntegerDigits": tc.in})
			assert.Nil(t, got.MinimumIntegerDigits)
		})
	}
}

func assertIntPtr(t *testing.T, want int, got *int) {
	t.Helper()
	if assert.NotNil(t, got) {
		assert.Equal(t, want, *got)
	}
}

func assertOptionalStringPtr(t *testing.T, want string, got *string) {
	t.Helper()
	if want == "" {
		assert.Nil(t, got)
		return
	}
	assertStringPtr(t, want, got)
}

func assertStringPtr(t *testing.T, want string, got *string) {
	t.Helper()
	if assert.NotNil(t, got) {
		assert.Equal(t, want, *got)
	}
}
