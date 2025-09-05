package messagevalue

import (
	"testing"

	"github.com/kaptinlin/messageformat-go/pkg/bidi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnknownValue(t *testing.T) {
	tests := []struct {
		name             string
		source           string
		value            interface{}
		locale           string
		expectedType     string
		expectedDir      bidi.Direction
		expectedToString string
		expectedValueOf  interface{}
	}{
		{
			name:             "string value",
			source:           "test",
			value:            "hello",
			locale:           "en",
			expectedType:     "unknown",
			expectedDir:      bidi.DirAuto,
			expectedToString: "hello",
			expectedValueOf:  "hello",
		},
		{
			name:             "number value",
			source:           "test",
			value:            42,
			locale:           "en",
			expectedType:     "unknown",
			expectedDir:      bidi.DirAuto,
			expectedToString: "42",
			expectedValueOf:  42,
		},
		{
			name:             "nil value",
			source:           "test",
			value:            nil,
			locale:           "en",
			expectedType:     "unknown",
			expectedDir:      bidi.DirAuto,
			expectedToString: "<nil>",
			expectedValueOf:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uv := NewUnknownValue(tt.source, tt.value, tt.locale)

			// Test basic properties
			assert.Equal(t, tt.expectedType, uv.Type())
			assert.Equal(t, tt.source, uv.Source())
			assert.Equal(t, tt.locale, uv.Locale())
			assert.Equal(t, tt.expectedDir, uv.Dir())
			assert.Nil(t, uv.Options())

			// Test ToString
			str, err := uv.ToString()
			require.NoError(t, err)
			assert.Equal(t, tt.expectedToString, str)

			// Test ValueOf
			value, err := uv.ValueOf()
			require.NoError(t, err)
			assert.Equal(t, tt.expectedValueOf, value)

			// Test SelectKeys (should return empty)
			keys, err := uv.SelectKeys([]string{"one", "other"})
			require.NoError(t, err)
			assert.Empty(t, keys)

			// Test ToParts
			parts, err := uv.ToParts()
			require.NoError(t, err)
			require.Len(t, parts, 1)

			part := parts[0]
			assert.Equal(t, "unknown", part.Type())
			assert.Equal(t, tt.value, part.Value())
			assert.Equal(t, tt.source, part.Source())
			assert.Equal(t, tt.locale, part.Locale())
			assert.Equal(t, bidi.DirAuto, part.Dir())
		})
	}
}

func TestUnknownPart(t *testing.T) {
	source := "test"
	value := "hello"
	locale := "en"

	up := NewUnknownPart(source, value, locale)

	assert.Equal(t, "unknown", up.Type())
	assert.Equal(t, value, up.Value())
	assert.Equal(t, source, up.Source())
	assert.Equal(t, locale, up.Locale())
	assert.Equal(t, bidi.DirAuto, up.Dir())
}
