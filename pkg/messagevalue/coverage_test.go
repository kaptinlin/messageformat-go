package messagevalue

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/dromara/carbon/v2"
	"github.com/kaptinlin/messageformat-go/pkg/bidi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubStringer string

func (s stubStringer) String() string {
	return string(s)
}

func TestHelperConversionsAndConstructors(t *testing.T) {
	t.Parallel()

	t.Run("ToString handles supported inputs", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name  string
			value any
			want  string
		}{
			{name: "nil", value: nil, want: ""},
			{name: "string", value: "hello", want: "hello"},
			{name: "int", value: 42, want: "42"},
			{name: "uint", value: uint16(7), want: "7"},
			{name: "float", value: 12.5, want: "12.5"},
			{name: "bool true", value: true, want: "true"},
			{name: "bool false", value: false, want: "false"},
			{name: "stringer", value: stubStringer("custom"), want: "custom"},
		}

		for _, tt := range tests {
			tc := tt
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, tc.want, ToString(tc.value))
			})
		}
	})

	t.Run("ToNumber handles supported inputs", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name  string
			value any
			want  float64
		}{
			{name: "nil", value: nil, want: 0},
			{name: "float64", value: 12.5, want: 12.5},
			{name: "float32", value: float32(3.5), want: 3.5},
			{name: "int64", value: int64(42), want: 42},
			{name: "uint", value: uint(9), want: 9},
			{name: "string number", value: "7.25", want: 7.25},
			{name: "invalid string", value: "nope", want: 0},
			{name: "bool true", value: true, want: 1},
			{name: "bool false", value: false, want: 0},
			{name: "stringer fallback", value: stubStringer("8.5"), want: 8.5},
		}

		for _, tt := range tests {
			tc := tt
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				assert.InDelta(t, tc.want, ToNumber(tc.value), 0.000001)
			})
		}
	})

	t.Run("String and Number constructors return usable values", func(t *testing.T) {
		t.Parallel()

		sv := String("hello")
		require.NotNil(t, sv)
		assert.Equal(t, "string", sv.Type())
		assert.Equal(t, "", sv.Locale())
		assert.Equal(t, "", sv.Source())

		str, err := sv.ToString()
		require.NoError(t, err)
		assert.Equal(t, "hello", str)

		nv := Number(42)
		require.NotNil(t, nv)
		assert.Equal(t, "number", nv.Type())
		assert.Equal(t, "", nv.Locale())
		assert.Equal(t, "", nv.Source())
		assert.NotNil(t, nv.Options())

		value, err := nv.ValueOf()
		require.NoError(t, err)
		assert.Equal(t, 42, value)
	})
}

func TestStringValueSelectKeysNormalizesUnicode(t *testing.T) {
	t.Parallel()

	sv := NewStringValue("Café", "en", "source")
	keys, err := sv.SelectKeys([]string{"Café", "Other"})
	require.NoError(t, err)
	require.Len(t, keys, 1)
	assert.Equal(t, "Café", keys[0])

	parts, err := NewStringValueWithDir("مرحبا", "ar", "source", bidi.DirRTL).ToParts()
	require.NoError(t, err)
	require.Len(t, parts, 1)
	assert.Equal(t, "string", parts[0].Type())
	assert.Equal(t, "source", parts[0].Source())
	assert.Equal(t, "ar", parts[0].Locale())
	assert.Equal(t, bidi.DirRTL, parts[0].Dir())
}

func TestDateTimeValueAdditionalBehaviors(t *testing.T) {
	t.Parallel()

	testTime := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)

	t.Run("new style options format date time and timezone", func(t *testing.T) {
		t.Parallel()

		dtv := NewDateTimeValue(testTime, "en_US", "source", map[string]any{
			"dateFields":    "weekday-year-month-day",
			"dateLength":    "long",
			"timePrecision": "second",
			"timeZoneStyle": "short",
		})

		formatted, err := dtv.ToString()
		require.NoError(t, err)
		assert.Contains(t, formatted, "2006")
		assert.Contains(t, formatted, "January")
		assert.Contains(t, formatted, "3:04:05 PM")
		assert.Contains(t, formatted, "3:04:05 PM T")

		parts, err := dtv.ToParts()
		require.NoError(t, err)
		require.Len(t, parts, 1)
		assert.Equal(t, "datetime", parts[0].Type())
		assert.Equal(t, formatted, fmt.Sprint(parts[0].Value()))
		assert.Equal(t, "source", parts[0].Source())
		assert.Equal(t, "en_US", parts[0].Locale())
	})

	t.Run("explicit direction with nil options preserves metadata", func(t *testing.T) {
		t.Parallel()

		dtv := NewDateTimeValueWithDir(testTime, "ar", "source", bidi.DirRTL, nil)
		assert.NotNil(t, dtv.Options())
		assert.Equal(t, bidi.DirRTL, dtv.Dir())

		parts, err := dtv.ToParts()
		require.NoError(t, err)
		require.Len(t, parts, 1)
		assert.Equal(t, bidi.DirRTL, parts[0].Dir())
	})

	t.Run("locale normalization supports region formats", func(t *testing.T) {
		t.Parallel()

		for _, locale := range []string{"pt-BR", "pt_BR", "es-ES", "fr_FR", "de-DE", "ja_JP", "ko-KR", "ru-RU", "ar_SA", "it-IT", "en"} {
			t.Run(locale, func(t *testing.T) {
				t.Parallel()

				dtv := NewDateTimeValue(testTime, locale, "source", map[string]any{"dateStyle": "short"})
				formatted, err := dtv.ToString()
				require.NoError(t, err)
				assert.NotEmpty(t, formatted)
			})
		}
	})

	t.Run("new style date and time options fall back independently", func(t *testing.T) {
		t.Parallel()

		dateOnly := NewDateTimeValue(testTime, "en-US", "source", map[string]any{"dateFields": "year-month-day", "dateLength": "short"})
		formatted, err := dateOnly.ToString()
		require.NoError(t, err)
		assert.Equal(t, "2006 1 2", formatted)

		timeOnly := NewDateTimeValue(testTime, "en-US", "source", map[string]any{"timePrecision": "minute", "timeZoneStyle": "none"})
		formatted, err = timeOnly.ToString()
		require.NoError(t, err)
		assert.Equal(t, "3:04 PM", formatted)

		defaulted := NewDateTimeValue(testTime, "en-US", "source", map[string]any{"dateFields": 123, "timePrecision": 123, "timeZoneStyle": "long"})
		formatted, err = defaulted.ToString()
		require.NoError(t, err)
		assert.Equal(t, "T", formatted)
	})

	t.Run("old style formatting supports date only time only and default", func(t *testing.T) {
		t.Parallel()

		dateOnly := NewDateTimeValue(testTime, "en-US", "source", map[string]any{"dateStyle": "full"})
		formatted, err := dateOnly.ToString()
		require.NoError(t, err)
		assert.Equal(t, "Monday, January 2, 2006", formatted)

		timeOnly := NewDateTimeValue(testTime, "en-US", "source", map[string]any{"timeStyle": "long"})
		formatted, err = timeOnly.ToString()
		require.NoError(t, err)
		assert.Equal(t, "3:04:05 PM T", formatted)

		defaulted := NewDateTimeValue(testTime, "en-US", "source", nil)
		formatted, err = defaulted.ToString()
		require.NoError(t, err)
		assert.Equal(t, "2006-01-02 15:04:05", formatted)
	})
}

func TestFormatTimeWithStyle(t *testing.T) {
	t.Parallel()

	c := carbon.CreateFromStdTime(time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC))

	assert.Equal(t, "3:04 PM", FormatTimeWithStyle(*c, "short"))
	assert.Equal(t, "3:04:05 PM", FormatTimeWithStyle(*c, "medium"))
	assert.Equal(t, "3:04:05 PM T", FormatTimeWithStyle(*c, "full"))

	hourValue := NewDateTimeValue(time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC), "en-US", "source", map[string]any{"timePrecision": "hour"})
	formatted, err := hourValue.formatDateTime()
	require.NoError(t, err)
	assert.Equal(t, "3 PM", formatted)
}

func TestFallbackValueWithDir(t *testing.T) {
	t.Parallel()

	fv := NewFallbackValueWithDir("$name", "ar", bidi.DirRTL)
	assert.Equal(t, "fallback", fv.Type())
	assert.Equal(t, "$name", fv.Source())
	assert.Equal(t, "ar", fv.Locale())
	assert.Equal(t, bidi.DirRTL, fv.Dir())

	formatted, err := fv.ToString()
	require.NoError(t, err)
	assert.Equal(t, "{$name}", formatted)

	parts, err := fv.ToParts()
	require.NoError(t, err)
	require.Len(t, parts, 1)
	assert.Equal(t, "fallback", parts[0].Type())
	assert.Equal(t, "{$name}", fmt.Sprint(parts[0].Value()))
	assert.Equal(t, bidi.DirRTL, parts[0].Dir())
}

func TestNumberValueSelectionBehaviors(t *testing.T) {
	t.Parallel()

	t.Run("returns sentinel error when selection is disabled", func(t *testing.T) {
		t.Parallel()

		nv := NewNumberValueWithSelection(42, "en", "source", bidi.DirLTR, nil, false)
		assert.NotNil(t, nv.Options())

		keys, err := nv.SelectKeys([]string{"42", "other"})
		require.ErrorIs(t, err, ErrNumberNotSelectable)
		assert.Nil(t, keys)
	})

	t.Run("exact selection disables plural fallback", func(t *testing.T) {
		t.Parallel()

		nv := NewNumberValueWithSelection(2, "en", "source", bidi.DirAuto, map[string]any{"select": "exact"}, true)
		keys, err := nv.SelectKeys([]string{"one", "other"})
		require.NoError(t, err)
		assert.Empty(t, keys)
	})

	t.Run("ordinal selection uses ordinal categories", func(t *testing.T) {
		t.Parallel()

		nv := NewNumberValueWithSelection(3, "en", "source", bidi.DirAuto, map[string]any{"select": "ordinal"}, true)
		keys, err := nv.SelectKeys([]string{"one", "two", "few", "other"})
		require.NoError(t, err)
		require.Len(t, keys, 1)
		assert.Equal(t, "few", keys[0])
	})

	t.Run("percent selection matches multiplied value", func(t *testing.T) {
		t.Parallel()

		nv := NewNumberValue(0.01, "en", "source", map[string]any{"style": "percent"})
		keys, err := nv.SelectKeys([]string{"=bad", "=1", "1", "other"})
		require.NoError(t, err)
		require.Len(t, keys, 1)
		assert.Equal(t, "=1", keys[0])
	})

	t.Run("exact numeric key has precedence over plural category", func(t *testing.T) {
		t.Parallel()

		nv := NewNumberValueWithSelection(1, "en", "source", bidi.DirAuto, nil, true)
		keys, err := nv.SelectKeys([]string{"=1", "one", "other"})
		require.NoError(t, err)
		require.Len(t, keys, 1)
		assert.Equal(t, "=1", keys[0])
	})

	t.Run("ordinal teen values use other category", func(t *testing.T) {
		t.Parallel()

		nv := NewNumberValueWithSelection(11, "en", "source", bidi.DirAuto, map[string]any{"select": "ordinal"}, true)
		keys, err := nv.SelectKeys([]string{"one", "two", "few", "other"})
		require.NoError(t, err)
		require.Len(t, keys, 1)
		assert.Equal(t, "other", keys[0])
	})

	t.Run("ordinal values select one two and other categories", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name  string
			value int
			want  string
		}{
			{name: "one", value: 1, want: "one"},
			{name: "two", value: 2, want: "two"},
			{name: "teen other", value: 13, want: "other"},
			{name: "twenty two", value: 22, want: "two"},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				nv := NewNumberValueWithSelection(tc.value, "en", "source", bidi.DirAuto, map[string]any{"select": "ordinal"}, true)
				keys, err := nv.SelectKeys([]string{"one", "two", "few", "other"})
				require.NoError(t, err)
				require.Len(t, keys, 1)
				assert.Equal(t, tc.want, keys[0])
			})
		}
	})

	t.Run("unsupported value type is not selectable", func(t *testing.T) {
		t.Parallel()

		nv := NewNumberValueWithSelection("not-a-number", "en", "source", bidi.DirAuto, nil, true)
		keys, err := nv.SelectKeys([]string{"other"})
		require.NoError(t, err)
		assert.Empty(t, keys)
	})

	t.Run("fractional numbers match formatted string keys", func(t *testing.T) {
		t.Parallel()

		nv := NewNumberValueWithSelection(1.5, "en", "source", bidi.DirAuto, nil, true)
		keys, err := nv.SelectKeys([]string{"=bad", "1.5", "other"})
		require.NoError(t, err)
		require.Len(t, keys, 1)
		assert.Equal(t, "1.5", keys[0])
	})
}

func TestNumberValueFormattingAndParts(t *testing.T) {
	t.Parallel()

	t.Run("decimal formatting handles grouping fraction and unsupported values", func(t *testing.T) {
		t.Parallel()

		ungrouped := NewNumberValue(1234.5, "en", "source", map[string]any{
			"minimumFractionDigits": 2,
			"maximumFractionDigits": 1,
			"useGrouping":           "never",
		})
		formatted, err := ungrouped.ToString()
		require.NoError(t, err)
		assert.Equal(t, "1234.50", formatted)

		unsupported := NewNumberValue(struct{ Name string }{Name: "Ada"}, "en", "source", nil)
		formatted, err = unsupported.ToString()
		require.NoError(t, err)
		assert.Equal(t, "{Ada}", formatted)

		noGrouping := NewNumberValue(1234, "en", "source", map[string]any{"useGrouping": false})
		formatted, err = noGrouping.ToString()
		require.NoError(t, err)
		assert.Equal(t, "1234", formatted)
	})

	t.Run("decimal parts preserve sign grouping and metadata", func(t *testing.T) {
		t.Parallel()

		nv := NewNumberValueWithDir(-1234.5, "en", "number-source", bidi.DirRTL, map[string]any{
			"minimumFractionDigits": 1,
			"maximumFractionDigits": 1,
		})

		formatted, err := nv.ToString()
		require.NoError(t, err)
		assert.Equal(t, "-1,234.5", formatted)

		parts, err := nv.ToParts()
		require.NoError(t, err)
		require.Len(t, parts, 1)

		numberPart, ok := parts[0].(*NumberPart)
		require.True(t, ok)
		assert.Equal(t, "number", numberPart.Type())
		assert.Equal(t, formatted, numberPart.Value())
		assert.Equal(t, "number-source", numberPart.Source())
		assert.Equal(t, "en", numberPart.Locale())
		assert.Equal(t, bidi.DirRTL, numberPart.Dir())

		subParts := numberPart.Parts()
		require.Len(t, subParts, 4)
		assert.Equal(t, "minusSign", subParts[0].Type())
		assert.Equal(t, "-", fmt.Sprint(subParts[0].Value()))
		assert.Equal(t, "integer", subParts[1].Type())
		assert.Equal(t, "1,234", fmt.Sprint(subParts[1].Value()))
		assert.Equal(t, "decimal", subParts[2].Type())
		assert.Equal(t, ".", fmt.Sprint(subParts[2].Value()))
		assert.Equal(t, "fraction", subParts[3].Type())
		assert.Equal(t, "5", fmt.Sprint(subParts[3].Value()))
		assert.Equal(t, "number-source", subParts[0].Source())
		assert.Equal(t, "en", subParts[0].Locale())
		assert.Equal(t, bidi.DirRTL, subParts[0].Dir())
	})

	t.Run("currency formatting supports code and accounting name display", func(t *testing.T) {
		t.Parallel()

		codeValue := NewNumberValue(42, "en", "money-source", map[string]any{
			"style":           "currency",
			"currency":        "USD",
			"currencyDisplay": "code",
		})
		codeFormatted, err := codeValue.ToString()
		require.NoError(t, err)
		assert.Contains(t, codeFormatted, "USD")
		assert.NotContains(t, codeFormatted, "$")

		nameValue := NewNumberValue(-42, "en", "money-source", map[string]any{
			"style":           "currency",
			"currency":        "USD",
			"currencyDisplay": "name",
			"currencySign":    "accounting",
		})
		nameFormatted, err := nameValue.ToString()
		require.NoError(t, err)
		assert.Contains(t, nameFormatted, "US dollars")
		assert.Contains(t, nameFormatted, "(")
		assert.Contains(t, nameFormatted, ")")
		assert.NotContains(t, nameFormatted, "$")
	})

	t.Run("currency formatting falls back for missing invalid and unknown currency", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name    string
			options map[string]any
			want    string
		}{
			{name: "missing currency", options: map[string]any{"style": "currency"}, want: "42"},
			{name: "non string currency", options: map[string]any{"style": "currency", "currency": 123}, want: "42"},
			{name: "unsupported currency", options: map[string]any{"style": "currency", "currency": "ZZZ"}, want: "42.00ZZZ"},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				nv := NewNumberValue(42, "en", "money-source", tc.options)
				formatted, err := nv.ToString()
				require.NoError(t, err)
				assert.Equal(t, tc.want, formatted)
			})
		}
	})

	t.Run("currency name display covers common regional names", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name     string
			currency string
			want     string
		}{
			{name: "euro", currency: "EUR", want: "euros"},
			{name: "pound", currency: "GBP", want: "British pounds"},
			{name: "yen", currency: "JPY", want: "Japanese yen"},
			{name: "yuan", currency: "CNY", want: "Chinese yuan"},
			{name: "rupee", currency: "INR", want: "Indian rupees"},
			{name: "shekel", currency: "ILS", want: "Israeli shekels"},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				nv := NewNumberValue(42, "en", "money-source", map[string]any{
					"style":           "currency",
					"currency":        tc.currency,
					"currencyDisplay": "name",
				})
				formatted, err := nv.ToString()
				require.NoError(t, err)
				assert.Contains(t, formatted, tc.want)
			})
		}
	})

	t.Run("currency parts expose literals currency and numeric pieces", func(t *testing.T) {
		t.Parallel()

		nv := NewNumberValue(-42, "en", "money-source", map[string]any{
			"style":        "currency",
			"currency":     "USD",
			"currencySign": "accounting",
		})

		parts, err := nv.ToParts()
		require.NoError(t, err)
		require.Len(t, parts, 1)

		numberPart, ok := parts[0].(*NumberPart)
		require.True(t, ok)
		subParts := numberPart.Parts()
		require.NotEmpty(t, subParts)
		assert.Equal(t, "literal", subParts[0].Type())
		assert.Equal(t, "(", fmt.Sprint(subParts[0].Value()))
		assert.Equal(t, "literal", subParts[len(subParts)-1].Type())
		assert.Equal(t, ")", fmt.Sprint(subParts[len(subParts)-1].Value()))

		var foundCurrency bool
		for _, part := range subParts {
			if part.Type() == "currency" {
				foundCurrency = true
				assert.Equal(t, "$", fmt.Sprint(part.Value()))
			}
		}
		assert.True(t, foundCurrency)
	})

	t.Run("currency code parts fall back to numeric parsing", func(t *testing.T) {
		t.Parallel()

		nv := NewNumberValue(42, "en", "money-source", map[string]any{
			"style":           "currency",
			"currency":        "USD",
			"currencyDisplay": "code",
		})
		formatted, err := nv.ToString()
		require.NoError(t, err)
		assert.Contains(t, formatted, "USD")

		parts, err := nv.ToParts()
		require.NoError(t, err)
		require.Len(t, parts, 1)

		numberPart, ok := parts[0].(*NumberPart)
		require.True(t, ok)
		assert.Equal(t, formatted, numberPart.Value())
		subParts := numberPart.Parts()
		require.NotEmpty(t, subParts)
		for _, part := range subParts {
			assert.NotEqual(t, "currency", part.Type())
		}
	})

	t.Run("percent parts expose sign decimal fraction and percent symbol", func(t *testing.T) {
		t.Parallel()

		nv := NewNumberValue(0.125, "en", "percent-source", map[string]any{
			"style":                 "percent",
			"maximumFractionDigits": 1,
			"signDisplay":           "always",
		})

		formatted, err := nv.ToString()
		require.NoError(t, err)
		assert.Equal(t, "+12.5%", formatted)

		parts, err := nv.ToParts()
		require.NoError(t, err)
		require.Len(t, parts, 1)

		numberPart, ok := parts[0].(*NumberPart)
		require.True(t, ok)
		subParts := numberPart.Parts()
		require.Len(t, subParts, 5)
		assert.Equal(t, "plusSign", subParts[0].Type())
		assert.Equal(t, "integer", subParts[1].Type())
		assert.Equal(t, "decimal", subParts[2].Type())
		assert.Equal(t, "fraction", subParts[3].Type())
		assert.Equal(t, "percentSign", subParts[4].Type())
		assert.Equal(t, "%", fmt.Sprint(subParts[4].Value()))
	})

	t.Run("percent formatting chooses integer and fractional defaults", func(t *testing.T) {
		t.Parallel()

		integerPercent := NewNumberValue(0.12, "en", "percent-source", map[string]any{"style": "percent"})
		formatted, err := integerPercent.ToString()
		require.NoError(t, err)
		assert.Equal(t, "12%", formatted)

		fractionPercent := NewNumberValue(0.125, "en", "percent-source", map[string]any{"style": "percent"})
		formatted, err = fractionPercent.ToString()
		require.NoError(t, err)
		assert.Equal(t, "12.5%", formatted)

		negativeZero := NewNumberValue(math.Copysign(0, -1), "en", "percent-source", map[string]any{"style": "percent", "signDisplay": "negative"})
		formatted, err = negativeZero.ToString()
		require.NoError(t, err)
		assert.Equal(t, "0%", formatted)
	})

	t.Run("sign display options handle zero positive and negative numbers", func(t *testing.T) {
		t.Parallel()

		zero := NewNumberValue(0, "en", "zero-source", map[string]any{"signDisplay": "exceptZero"})
		formatted, err := zero.ToString()
		require.NoError(t, err)
		assert.Equal(t, "0", formatted)

		positive := NewNumberValue(7, "en", "positive-source", map[string]any{"signDisplay": "exceptZero"})
		formatted, err = positive.ToString()
		require.NoError(t, err)
		assert.Equal(t, "+7", formatted)

		negative := NewNumberValue(-7, "en", "negative-source", map[string]any{"signDisplay": "never"})
		formatted, err = negative.ToString()
		require.NoError(t, err)
		assert.Equal(t, "7", formatted)
	})

	t.Run("sign display always and auto preserve existing signs", func(t *testing.T) {
		t.Parallel()

		positive := NewNumberValue(7, "en", "positive-source", map[string]any{"signDisplay": "always"})
		formatted, err := positive.ToString()
		require.NoError(t, err)
		assert.Equal(t, "+7", formatted)

		negative := NewNumberValue(-7, "en", "negative-source", map[string]any{"signDisplay": "auto"})
		formatted, err = negative.ToString()
		require.NoError(t, err)
		assert.Equal(t, "-7", formatted)
	})

	t.Run("unit formatting falls back for missing and invalid units", func(t *testing.T) {
		t.Parallel()

		missing := NewNumberValue(42, "en", "unit-source", map[string]any{"style": "unit"})
		formatted, err := missing.ToString()
		require.NoError(t, err)
		assert.Equal(t, "42", formatted)

		invalid := NewNumberValue(42, "en", "unit-source", map[string]any{"style": "unit", "unit": 123})
		formatted, err = invalid.ToString()
		require.NoError(t, err)
		assert.Equal(t, "42", formatted)
	})

	t.Run("unit fallback parts parse fallback numbers", func(t *testing.T) {
		t.Parallel()

		nv := NewNumberValue(42, "en", "unit-source", map[string]any{"style": "unit"})
		parts, err := nv.ToParts()
		require.NoError(t, err)
		require.Len(t, parts, 1)

		numberPart, ok := parts[0].(*NumberPart)
		require.True(t, ok)
		subParts := numberPart.Parts()
		require.Len(t, subParts, 1)
		assert.Equal(t, "integer", subParts[0].Type())
		assert.Equal(t, "42", fmt.Sprint(subParts[0].Value()))
	})

	t.Run("unit symbols support default narrow and unknown displays", func(t *testing.T) {
		t.Parallel()

		longUnits := []struct {
			name string
			unit string
			want string
		}{
			{name: "meter", unit: "meter", want: "1 meters"},
			{name: "kilometer", unit: "kilometer", want: "1 kilometers"},
			{name: "gram", unit: "gram", want: "1 grams"},
			{name: "kilogram", unit: "kilogram", want: "1 kilograms"},
			{name: "second", unit: "second", want: "1 seconds"},
			{name: "minute", unit: "minute", want: "1 minutes"},
			{name: "hour", unit: "hour", want: "1 hours"},
		}

		for _, tt := range longUnits {
			t.Run("long "+tt.name, func(t *testing.T) {
				t.Parallel()

				nv := NewNumberValue(1, "en", "unit-source", map[string]any{"style": "unit", "unit": tt.unit, "unitDisplay": "long"})
				formatted, err := nv.ToString()
				require.NoError(t, err)
				assert.Equal(t, tt.want, formatted)
			})
		}

		defaults := []struct {
			name string
			unit string
			want string
		}{
			{name: "kilometer", unit: "kilometer", want: "1 km"},
			{name: "gram", unit: "gram", want: "1 g"},
			{name: "kilogram", unit: "kilogram", want: "1 kg"},
			{name: "second", unit: "second", want: "1 s"},
			{name: "minute", unit: "minute", want: "1 min"},
			{name: "hour", unit: "hour", want: "1 h"},
			{name: "unknown", unit: "lightyear", want: "1 lightyear"},
		}

		for _, tt := range defaults {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				nv := NewNumberValue(1, "en", "unit-source", map[string]any{"style": "unit", "unit": tt.unit})
				formatted, err := nv.ToString()
				require.NoError(t, err)
				assert.Equal(t, tt.want, formatted)
			})
		}

		narrow := NewNumberValue(1, "en", "unit-source", map[string]any{"style": "unit", "unit": "kilometer", "unitDisplay": "narrow"})
		formatted, err := narrow.ToString()
		require.NoError(t, err)
		assert.Equal(t, "1 km", formatted)

		longUnknown := NewNumberValue(1, "en", "unit-source", map[string]any{"style": "unit", "unit": "lightyear", "unitDisplay": "long"})
		formatted, err = longUnknown.ToString()
		require.NoError(t, err)
		assert.Equal(t, "1 lightyear", formatted)
	})

	t.Run("unit parts expose literal separator and unit name", func(t *testing.T) {
		t.Parallel()

		nv := NewNumberValue(42, "en", "unit-source", map[string]any{
			"style":       "unit",
			"unit":        "meter",
			"unitDisplay": "long",
		})

		formatted, err := nv.ToString()
		require.NoError(t, err)
		assert.Contains(t, formatted, "42")
		assert.Contains(t, formatted, "meters")

		parts, err := nv.ToParts()
		require.NoError(t, err)
		require.Len(t, parts, 1)

		numberPart, ok := parts[0].(*NumberPart)
		require.True(t, ok)
		subParts := numberPart.Parts()
		require.Len(t, subParts, 3)
		assert.Equal(t, "integer", subParts[0].Type())
		assert.Equal(t, "literal", subParts[1].Type())
		assert.Equal(t, " ", fmt.Sprint(subParts[1].Value()))
		assert.Equal(t, "unit", subParts[2].Type())
		assert.Equal(t, "meters", fmt.Sprint(subParts[2].Value()))
	})
}
