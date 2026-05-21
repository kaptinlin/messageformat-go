package intlbridge

import (
	"testing"

	"github.com/agentable/go-intl/datetimeformat"
	"github.com/stretchr/testify/assert"
)

func TestDateTimeOptions_Empty(t *testing.T) {
	assert.Equal(t, datetimeformat.Options{}, DateTimeOptions(nil))
	assert.Equal(t, datetimeformat.Options{}, DateTimeOptions(map[string]any{}))
}

func TestDateTimeOptions_LegacyStyle(t *testing.T) {
	got := DateTimeOptions(map[string]any{
		"dateStyle": "full",
		"timeStyle": "short",
	})
	assert.Equal(t, datetimeformat.FullDateTimeStyle, got.DateStyle)
	assert.Equal(t, datetimeformat.ShortDateTimeStyle, got.TimeStyle)
	assert.Equal(t, datetimeformat.MonthStyle(""), got.Month)
	assert.Equal(t, datetimeformat.FieldStyle(""), got.Weekday)
}

func TestDateTimeOptions_LegacyTakesPrecedence(t *testing.T) {
	got := DateTimeOptions(map[string]any{
		"dateStyle":  "medium",
		"dateFields": "year-month-day",
		"dateLength": "long",
	})
	assert.Equal(t, datetimeformat.MediumDateTimeStyle, got.DateStyle)
	assert.Equal(t, datetimeformat.NumericStyle(""), got.Year)
	assert.Equal(t, datetimeformat.MonthStyle(""), got.Month)
	assert.Equal(t, datetimeformat.NumericStyle(""), got.Day)
}

func TestDateTimeOptions_DateFields(t *testing.T) {
	cases := []struct {
		name      string
		fields    string
		length    string
		wantYear  datetimeformat.NumericStyle
		wantMonth datetimeformat.MonthStyle
		wantDay   datetimeformat.NumericStyle
		wantWk    datetimeformat.FieldStyle
	}{
		{"year-month-day medium", "year-month-day", "medium", datetimeformat.NumericFieldStyle, datetimeformat.ShortMonthStyle, datetimeformat.NumericFieldStyle, ""},
		{"year-month-day long", "year-month-day", "long", datetimeformat.NumericFieldStyle, datetimeformat.LongMonthStyle, datetimeformat.NumericFieldStyle, ""},
		{"year-month-day short", "year-month-day", "short", datetimeformat.NumericFieldStyle, datetimeformat.NumericMonthStyle, datetimeformat.NumericFieldStyle, ""},
		{"weekday-year-month-day long", "weekday-year-month-day", "long", datetimeformat.NumericFieldStyle, datetimeformat.LongMonthStyle, datetimeformat.NumericFieldStyle, datetimeformat.LongFieldStyle},
		{"weekday only", "weekday", "long", "", "", "", datetimeformat.LongFieldStyle},
		{"default length", "year-month-day", "", datetimeformat.NumericFieldStyle, datetimeformat.ShortMonthStyle, datetimeformat.NumericFieldStyle, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts := map[string]any{"dateFields": tc.fields}
			if tc.length != "" {
				opts["dateLength"] = tc.length
			}
			got := DateTimeOptions(opts)
			assert.Equal(t, tc.wantYear, got.Year)
			assert.Equal(t, tc.wantMonth, got.Month)
			assert.Equal(t, tc.wantDay, got.Day)
			assert.Equal(t, tc.wantWk, got.Weekday)
		})
	}
}

func TestDateTimeOptions_TimePrecision(t *testing.T) {
	cases := []struct {
		precision  string
		wantHour   datetimeformat.NumericStyle
		wantMinute datetimeformat.NumericStyle
		wantSecond datetimeformat.NumericStyle
	}{
		{"hour", datetimeformat.NumericFieldStyle, "", ""},
		{"minute", datetimeformat.NumericFieldStyle, datetimeformat.NumericFieldStyle, ""},
		{"second", datetimeformat.NumericFieldStyle, datetimeformat.NumericFieldStyle, datetimeformat.NumericFieldStyle},
	}
	for _, tc := range cases {
		t.Run(tc.precision, func(t *testing.T) {
			got := DateTimeOptions(map[string]any{"timePrecision": tc.precision})
			assert.Equal(t, tc.wantHour, got.Hour)
			assert.Equal(t, tc.wantMinute, got.Minute)
			assert.Equal(t, tc.wantSecond, got.Second)
		})
	}
}

func TestDateTimeOptions_TimeZoneStyle(t *testing.T) {
	cases := map[string]datetimeformat.TimeZoneName{
		"long":  datetimeformat.LongTimeZoneName,
		"short": datetimeformat.ShortTimeZoneName,
		"none":  "",
	}
	for in, want := range cases {
		t.Run(in, func(t *testing.T) {
			got := DateTimeOptions(map[string]any{"timeZoneStyle": in})
			assert.Equal(t, want, got.TimeZoneName)
		})
	}
}

func TestDateTimeOptions_Scalars(t *testing.T) {
	got := DateTimeOptions(map[string]any{
		"calendar":        "buddhist",
		"numberingSystem": "arab",
		"localeMatcher":   "lookup",
		"timeZone":        "Asia/Shanghai",
		"hour12":          true,
	})
	assert.Equal(t, "buddhist", got.Calendar)
	assert.Equal(t, "arab", got.NumberingSystem)
	assert.Equal(t, datetimeformat.LookupLocaleMatcher, got.LocaleMatcher)
	assert.Equal(t, "Asia/Shanghai", got.TimeZone)
	if assert.NotNil(t, got.Hour12) {
		assert.True(t, *got.Hour12)
	}
}

func TestDateTimeOptions_InvalidTypesDropped(t *testing.T) {
	got := DateTimeOptions(map[string]any{
		"dateFields":    123,
		"timePrecision": 123,
		"timeZoneStyle": "long",
	})
	assert.Equal(t, datetimeformat.NumericStyle(""), got.Year)
	assert.Equal(t, datetimeformat.NumericStyle(""), got.Hour)
	assert.Equal(t, datetimeformat.LongTimeZoneName, got.TimeZoneName)
}
