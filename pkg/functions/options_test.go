package functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptionsAccessorsAndCopySemantics(t *testing.T) {
	t.Parallel()

	assert.Nil(t, NewOptions(nil))

	var nilOptions Options
	assert.False(t, nilOptions.Has("missing"))
	value, ok := nilOptions.Value("missing")
	assert.False(t, ok)
	assert.Nil(t, value)
	stringValue, ok := nilOptions.String("missing")
	assert.False(t, ok)
	assert.Empty(t, stringValue)
	intValue, ok := nilOptions.Int("missing")
	assert.False(t, ok)
	assert.Zero(t, intValue)
	boolValue, ok := nilOptions.Bool("missing")
	assert.False(t, ok)
	assert.False(t, boolValue)
	assert.Nil(t, nilOptions.Map())

	source := map[string]any{
		"count":   3,
		"enabled": true,
		"name":    "Ada",
		"present": nil,
	}
	options := NewOptions(source)
	source["name"] = "Grace"

	assert.True(t, options.Has("present"))
	value, ok = options.Value("present")
	require.True(t, ok)
	assert.Nil(t, value)

	value, ok = options.Value("name")
	require.True(t, ok)
	assert.Equal(t, "Ada", value)

	stringValue, ok = options.String("name")
	require.True(t, ok)
	assert.Equal(t, "Ada", stringValue)
	_, ok = options.String("enabled")
	assert.False(t, ok)

	boolValue, ok = options.Bool("enabled")
	require.True(t, ok)
	assert.True(t, boolValue)
	_, ok = options.Bool("name")
	assert.False(t, ok)

	intValue, ok = options.Int("count")
	require.True(t, ok)
	assert.Equal(t, 3, intValue)
	_, ok = options.Int("name")
	assert.False(t, ok)

	copied := options.Map()
	copied["name"] = "Lin"
	value, ok = options.Value("name")
	require.True(t, ok)
	assert.Equal(t, "Ada", value)
}

func TestOptionsIntConvertsSupportedNumericTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		options map[string]any
		want    int
		wantOK  bool
	}{
		{name: "missing", options: nil},
		{name: "int", options: map[string]any{"value": int(3)}, want: 3, wantOK: true},
		{name: "int8", options: map[string]any{"value": int8(4)}, want: 4, wantOK: true},
		{name: "int16", options: map[string]any{"value": int16(5)}, want: 5, wantOK: true},
		{name: "int32", options: map[string]any{"value": int32(6)}, want: 6, wantOK: true},
		{name: "int64", options: map[string]any{"value": int64(7)}, want: 7, wantOK: true},
		{name: "uint", options: map[string]any{"value": uint(8)}, want: 8, wantOK: true},
		{name: "uint8", options: map[string]any{"value": uint8(9)}, want: 9, wantOK: true},
		{name: "uint16", options: map[string]any{"value": uint16(10)}, want: 10, wantOK: true},
		{name: "uint32", options: map[string]any{"value": uint32(11)}, want: 11, wantOK: true},
		{name: "uint64", options: map[string]any{"value": uint64(12)}, want: 12, wantOK: true},
		{name: "float32", options: map[string]any{"value": float32(13.9)}, want: 13, wantOK: true},
		{name: "float64", options: map[string]any{"value": 14.9}, want: 14, wantOK: true},
		{name: "oversized uint64", options: map[string]any{"value": maxIntValue + 1}},
		{name: "string", options: map[string]any{"value": "15"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := NewOptions(tc.options).Int("value")
			assert.Equal(t, tc.wantOK, ok)
			assert.Equal(t, tc.want, got)
		})
	}
}
