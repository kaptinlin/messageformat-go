package messagevalue

import (
	"fmt"
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

		for _, locale := range []string{"pt-BR", "it-IT"} {
			t.Run(locale, func(t *testing.T) {
				t.Parallel()

				dtv := NewDateTimeValue(testTime, locale, "source", map[string]any{"dateStyle": "short"})
				formatted, err := dtv.ToString()
				require.NoError(t, err)
				assert.NotEmpty(t, formatted)
			})
		}
	})
}

func TestFormatTimeWithStyle(t *testing.T) {
	t.Parallel()

	c := carbon.CreateFromStdTime(time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC))

	assert.Equal(t, "3:04 PM", FormatTimeWithStyle(*c, "short"))
	assert.Equal(t, "3:04:05 PM", FormatTimeWithStyle(*c, "medium"))
	assert.Equal(t, "3:04:05 PM T", FormatTimeWithStyle(*c, "full"))
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
		keys, err := nv.SelectKeys([]string{"1", "other"})
		require.NoError(t, err)
		require.Len(t, keys, 1)
		assert.Equal(t, "1", keys[0])
	})
}

func TestNumberValueFormattingAndParts(t *testing.T) {
	t.Parallel()

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
