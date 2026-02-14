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

		dtv := NewDateTimeValue(testTime, "en-US", "test", options)

		assert.Equal(t, "datetime", dtv.Type())
		assert.Equal(t, "test", dtv.Source())
		assert.Equal(t, "en-US", dtv.Locale())
		assert.Equal(t, bidi.DirAuto, dtv.Dir())
		assert.Equal(t, options, dtv.Options())

		// Test ValueOf
		val, err := dtv.ValueOf()
		require.NoError(t, err)
		assert.Equal(t, testTime, val)
	})

	t.Run("datetime value with explicit direction", func(t *testing.T) {
		options := map[string]any{
			"dateStyle": "long",
		}

		dtv := NewDateTimeValueWithDir(testTime, "ar", "test", bidi.DirRTL, options)

		assert.Equal(t, "ar", dtv.Locale())
		assert.Equal(t, bidi.DirRTL, dtv.Dir())
	})

	t.Run("toString formatting", func(t *testing.T) {
		options := map[string]any{
			"dateStyle": "medium",
			"timeStyle": "short",
		}

		dtv := NewDateTimeValue(testTime, "en", "test", options)

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

		dtv := NewDateTimeValue(testTime, "en", "test", options)

		parts, err := dtv.ToParts()
		require.NoError(t, err)
		assert.Len(t, parts, 1)

		part := parts[0]
		assert.Equal(t, "datetime", part.Type())
		assert.Equal(t, "test", part.Source())
		assert.Equal(t, "en", part.Locale())
	})

	t.Run("selectKeys", func(t *testing.T) {
		dtv := NewDateTimeValue(testTime, "en", "test", nil)

		keys, err := dtv.SelectKeys([]string{"one", "other"})
		require.NoError(t, err)
		assert.Empty(t, keys) // DateTime values don't support selection
	})
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
}

func TestFormatFunctions(t *testing.T) {
	t.Run("GetDateFormat", func(t *testing.T) {
		assert.Equal(t, "l, F j, Y", GetDateFormat("full"))
		assert.Equal(t, "F j, Y", GetDateFormat("long"))
		assert.Equal(t, "M j, Y", GetDateFormat("medium"))
		assert.Equal(t, "n/j/y", GetDateFormat("short"))
		assert.Equal(t, "M j, Y", GetDateFormat("unknown")) // default
	})

	t.Run("GetTimeFormat", func(t *testing.T) {
		assert.Equal(t, "g:i:s A T", GetTimeFormat("full"))
		assert.Equal(t, "g:i:s A T", GetTimeFormat("long"))
		assert.Equal(t, "g:i:s A", GetTimeFormat("medium"))
		assert.Equal(t, "g:i A", GetTimeFormat("short"))
		assert.Equal(t, "g:i A", GetTimeFormat("unknown")) // default
	})
}
