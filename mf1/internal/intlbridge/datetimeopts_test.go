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
	assertStringPtr(t, string(datetimeformat.FullDateTimeStyle), got.DateStyle)
	assertStringPtr(t, string(datetimeformat.ShortDateTimeStyle), got.TimeStyle)
	assert.Nil(t, got.Month)
	assert.Nil(t, got.Weekday)
}

func TestDateTimeOptions_LegacyTakesPrecedence(t *testing.T) {
	got := DateTimeOptions(map[string]any{
		"dateStyle":  "medium",
		"dateFields": "year-month-day",
		"dateLength": "long",
	})
	assertStringPtr(t, string(datetimeformat.MediumDateTimeStyle), got.DateStyle)
	assert.Nil(t, got.Year)
	assert.Nil(t, got.Month)
	assert.Nil(t, got.Day)
}

func TestDateTimeOptions_DateFields(t *testing.T) {
	cases := []struct {
		name      string
		fields    string
		length    string
		wantYear  string
		wantMonth string
		wantDay   string
		wantWk    string
	}{
		{"year-month-day medium", "year-month-day", "medium", string(datetimeformat.NumericFieldStyle), string(datetimeformat.ShortMonthStyle), string(datetimeformat.NumericFieldStyle), ""},
		{"year-month-day long", "year-month-day", "long", string(datetimeformat.NumericFieldStyle), string(datetimeformat.LongMonthStyle), string(datetimeformat.NumericFieldStyle), ""},
		{"year-month-day short", "year-month-day", "short", string(datetimeformat.NumericFieldStyle), string(datetimeformat.NumericMonthStyle), string(datetimeformat.NumericFieldStyle), ""},
		{"weekday-year-month-day long", "weekday-year-month-day", "long", string(datetimeformat.NumericFieldStyle), string(datetimeformat.LongMonthStyle), string(datetimeformat.NumericFieldStyle), string(datetimeformat.LongFieldStyle)},
		{"weekday only", "weekday", "long", "", "", "", string(datetimeformat.LongFieldStyle)},
		{"default length", "year-month-day", "", string(datetimeformat.NumericFieldStyle), string(datetimeformat.ShortMonthStyle), string(datetimeformat.NumericFieldStyle), ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts := map[string]any{"dateFields": tc.fields}
			if tc.length != "" {
				opts["dateLength"] = tc.length
			}
			got := DateTimeOptions(opts)
			assertOptionalStringPtr(t, tc.wantYear, got.Year)
			assertOptionalStringPtr(t, tc.wantMonth, got.Month)
			assertOptionalStringPtr(t, tc.wantDay, got.Day)
			assertOptionalStringPtr(t, tc.wantWk, got.Weekday)
		})
	}
}

func TestDateTimeOptions_TimePrecision(t *testing.T) {
	cases := []struct {
		precision  string
		wantHour   string
		wantMinute string
		wantSecond string
	}{
		{"hour", string(datetimeformat.NumericFieldStyle), "", ""},
		{"minute", string(datetimeformat.NumericFieldStyle), string(datetimeformat.NumericFieldStyle), ""},
		{"second", string(datetimeformat.NumericFieldStyle), string(datetimeformat.NumericFieldStyle), string(datetimeformat.NumericFieldStyle)},
	}
	for _, tc := range cases {
		t.Run(tc.precision, func(t *testing.T) {
			got := DateTimeOptions(map[string]any{"timePrecision": tc.precision})
			assertOptionalStringPtr(t, tc.wantHour, got.Hour)
			assertOptionalStringPtr(t, tc.wantMinute, got.Minute)
			assertOptionalStringPtr(t, tc.wantSecond, got.Second)
		})
	}
}

func TestDateTimeOptions_TimeZoneStyle(t *testing.T) {
	cases := map[string]string{
		"long":  string(datetimeformat.LongTimeZoneName),
		"short": string(datetimeformat.ShortTimeZoneName),
		"none":  "",
	}
	for in, want := range cases {
		t.Run(in, func(t *testing.T) {
			got := DateTimeOptions(map[string]any{"timeZoneStyle": in})
			assertOptionalStringPtr(t, want, got.TimeZoneName)
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
	assertStringPtr(t, "buddhist", got.Calendar)
	assertStringPtr(t, "arab", got.NumberingSystem)
	assertStringPtr(t, string(datetimeformat.LookupLocaleMatcher), got.LocaleMatcher)
	assertStringPtr(t, "Asia/Shanghai", got.TimeZone)
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
	assert.Nil(t, got.Year)
	assert.Nil(t, got.Hour)
	assertStringPtr(t, string(datetimeformat.LongTimeZoneName), got.TimeZoneName)
}

func TestDateTimeOptions_ScalarsDropInvalidValues(t *testing.T) {
	t.Parallel()

	got := DateTimeOptions(map[string]any{
		"calendar":        123,
		"numberingSystem": false,
		"localeMatcher":   123,
		"timeZone":        []string{"UTC"},
		"hour12":          "true",
	})

	assert.Nil(t, got.Calendar)
	assert.Nil(t, got.NumberingSystem)
	assert.Nil(t, got.LocaleMatcher)
	assert.Nil(t, got.TimeZone)
	assert.Nil(t, got.Hour12)
}
