package messagevalue

import (
	"testing"
	"time"

	"github.com/kaptinlin/messageformat-go/pkg/bidi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDateTimeValue(t *testing.T) {
	// Create a test time
	testTime := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)

	t.Run("basic datetime value", func(t *testing.T) {
		options := map[string]any{
			"dateStyle": "medium",
			"timeStyle": "short",
		}

		dtv := mustDateTimeValue(t, testTime, "en-US", "test", options)

		assert.Equal(t, "datetime", dtv.Type())
		assert.Equal(t, "test", dtv.Source())
		assert.Equal(t, "en-US", dtv.Locale())
		assert.Equal(t, bidi.DirAuto, dtv.Dir())
		assert.Equal(t, options, dtv.Options())
		assert.Equal(t, testTime, dtv.Time())

		// Test ValueOf
		val, err := dtv.ValueOf()
		require.NoError(t, err)
		assert.Equal(t, testTime, val)
	})

	t.Run("datetime value with explicit direction", func(t *testing.T) {
		options := map[string]any{
			"dateStyle": "long",
		}

		dtv := mustDateTimeValueWithDir(t, testTime, "ar", "test", bidi.DirRTL, options)

		assert.Equal(t, "ar", dtv.Locale())
		assert.Equal(t, bidi.DirRTL, dtv.Dir())
	})

	t.Run("toString formatting", func(t *testing.T) {
		options := map[string]any{
			"dateStyle": "medium",
			"timeStyle": "short",
		}

		dtv := mustDateTimeValue(t, testTime, "en", "test", options)

		str, err := dtv.ToString()
		require.NoError(t, err)
		assert.NotEmpty(t, str)
		// Should contain both date and time elements
		assert.Contains(t, str, "Jan")
		assert.Contains(t, str, "2006")
	})

	t.Run("toParts", func(t *testing.T) {
		options := map[string]any{
			"dateStyle": "short",
		}

		dtv := mustDateTimeValue(t, testTime, "en", "test", options)

		parts, err := dtv.ToParts()
		require.NoError(t, err)
		assert.Len(t, parts, 1)

		part := parts[0]
		assert.Equal(t, "datetime", part.Type())
		assert.Equal(t, "test", part.Source())
		assert.Equal(t, "en", part.Locale())
		dateTimePart, ok := part.(*DateTimePart)
		require.True(t, ok)
		assert.Equal(t, dateTimePart.Value(), dateTimePart.Text())
	})

	t.Run("named time zone formatting", func(t *testing.T) {
		dtv := mustDateTimeValue(t, testTime, "en-US", "test", map[string]any{
			"timePrecision": "minute",
			"timeZone":      "America/New_York",
		})

		parts, err := dtv.ToParts()
		require.NoError(t, err)
		require.Len(t, parts, 1)

		dateTimePart, ok := parts[0].(*DateTimePart)
		require.True(t, ok)
		assert.Equal(t, dateTimePart.Value(), dateTimePart.Text())
		require.NotEmpty(t, dateTimePart.Parts())

		subPart, ok := dateTimePart.Parts()[0].(*DateTimeSubPart)
		require.True(t, ok)
		assert.Equal(t, subPart.Value(), subPart.Text())
		assert.Equal(t, "test", subPart.Source())
		assert.Equal(t, "en-US", subPart.Locale())
		assert.Equal(t, bidi.DirAuto, subPart.Dir())

		valuesByType := make(map[string]any)
		for _, part := range dateTimePart.Parts() {
			valuesByType[part.Type()] = part.Value()
		}
		assert.Equal(t, "10", valuesByType["hour"])
		assert.Equal(t, "04", valuesByType["minute"])
	})

	t.Run("not selectable", func(t *testing.T) {
		dtv := mustDateTimeValue(t, testTime, "en", "test", nil)

		_, ok := any(dtv).(Selector)
		assert.False(t, ok)
	})
}

// TestDateTimeValueRejectsInvalidStyles proves the option bridge preserves invalid style semantics for validation.
// TypeScript original code:
// new Intl.DateTimeFormat(locales, { dateStyle: 'full', year: 'numeric' });
func TestDateTimeValueRejectsInvalidStyles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		options map[string]any
	}{
		{name: "invalid style", options: map[string]any{"dateStyle": "bad"}},
		{
			name: "style and fields conflict",
			options: map[string]any{
				"dateStyle":  "full",
				"dateFields": "year-month-day",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			value, err := NewDateTimeValue(
				time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
				"en",
				"source",
				tc.options,
			)
			assert.Nil(t, value)
			assert.ErrorIs(t, err, ErrInvalidDateTimeOptions)
		})
	}
}

// TestDateTimeValueUsesResolvedMetadataAndConsistentProjection proves all public projections share one Intl plan.
// TypeScript original code:
// const locale = formatter.resolvedOptions().locale;
// const parts = formatter.formatToParts(value);
func TestDateTimeValueUsesResolvedMetadataAndConsistentProjection(t *testing.T) {
	t.Parallel()

	dtv := mustDateTimeValue(
		t,
		time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
		"en-us",
		"source",
		map[string]any{
			"calendar":      "gregory",
			"timePrecision": "second",
			"timeZone":      "America/New_York",
		},
	)
	assert.Equal(t, "en-US", dtv.Locale())
	assert.Equal(t, "gregory", dtv.Calendar())
	assert.Equal(t, "America/New_York", dtv.TimeZone())

	formatted, err := dtv.ToString()
	require.NoError(t, err)
	parts, err := dtv.ToParts()
	require.NoError(t, err)
	require.Len(t, parts, 1)

	dateTimePart, ok := parts[0].(*DateTimePart)
	require.True(t, ok)
	assert.Equal(t, dtv.Locale(), dateTimePart.Locale())
	assert.Equal(t, dtv.Calendar(), dateTimePart.Calendar())
	assert.Equal(t, dtv.TimeZone(), dateTimePart.TimeZone())

	var joined string
	for _, part := range dateTimePart.Parts() {
		assert.Equal(t, dtv.Locale(), part.Locale())
		textPart, ok := part.(interface{ Text() string })
		require.True(t, ok)
		joined += textPart.Text()
	}
	assert.Equal(t, formatted, joined)
}

// TestDateTimeValuePreservesInputTimeZone proves implicit zones keep the input wall clock.
// TypeScript original code:
// options.timeZone = input.options?.timeZone;
func TestDateTimeValuePreservesInputTimeZone(t *testing.T) {
	namedLocation, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	localTime := time.Date(2006, 1, 2, 15, 4, 5, 0, time.Local)
	localTimeZone, ok := timeZoneFromValue(localTime)
	require.True(t, ok)

	tests := []struct {
		name     string
		value    time.Time
		timeZone string
	}{
		{
			name:     "UTC",
			value:    time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			timeZone: "UTC",
		},
		{
			name:     "named zone",
			value:    time.Date(2006, 1, 2, 15, 4, 5, 0, namedLocation),
			timeZone: "America/New_York",
		},
		{
			name:     "fixed offset",
			value:    time.Date(2006, 1, 2, 15, 4, 5, 0, time.FixedZone("custom", 5*60*60+30*60)),
			timeZone: "+05:30",
		},
		{
			name:     "local",
			value:    localTime,
			timeZone: localTimeZone,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dtv := mustDateTimeValue(t, tc.value, "en-US", "source", map[string]any{
				"timePrecision": "minute",
			})
			assert.Equal(t, tc.timeZone, dtv.TimeZone())

			formatted, err := dtv.ToString()
			require.NoError(t, err)
			assert.Equal(t, "3:04 PM", formatted)
		})
	}
}

func TestDateTimePart(t *testing.T) {
	part := &DateTimePart{
		value:  "Jan 2, 2006",
		source: "test",
		locale: "en",
		dir:    bidi.DirLTR,
	}

	assert.Equal(t, "datetime", part.Type())
	assert.Equal(t, "Jan 2, 2006", part.Value())
	assert.Equal(t, "test", part.Source())
	assert.Equal(t, "en", part.Locale())
	assert.Equal(t, bidi.DirLTR, part.Dir())
	assert.Equal(t, "Jan 2, 2006", part.Text())
}
