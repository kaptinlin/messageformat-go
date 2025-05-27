package messagevalue

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/messageformat-go/pkg/bidi"
)

func TestStringValue(t *testing.T) {
	sv := NewStringValue("hello", "en", "test")

	assert.Equal(t, "string", sv.Type())
	assert.Equal(t, "test", sv.Source())
	assert.Equal(t, "en", sv.Locale())
	assert.Equal(t, bidi.DirAuto, sv.Dir())
	assert.Nil(t, sv.Options())

	str, err := sv.ToString()
	require.NoError(t, err)
	assert.Equal(t, "hello", str)

	value, err := sv.ValueOf()
	require.NoError(t, err)
	assert.Equal(t, "hello", value)

	parts, err := sv.ToParts()
	require.NoError(t, err)
	require.Len(t, parts, 1)
	assert.Equal(t, "string", parts[0].Type())
	assert.Equal(t, "hello", parts[0].Value())

	keys, err := sv.SelectKeys([]string{"hello", "world"})
	require.NoError(t, err)
	assert.Equal(t, []string{"hello"}, keys)

	keys, err = sv.SelectKeys([]string{"world", "foo"})
	require.NoError(t, err)
	assert.Empty(t, keys)
}

func TestStringValueWithDir(t *testing.T) {
	sv := NewStringValueWithDir("مرحبا", "ar", "test", bidi.DirRTL)

	assert.Equal(t, "string", sv.Type())
	assert.Equal(t, bidi.DirRTL, sv.Dir())
	assert.Equal(t, "ar", sv.Locale())
}

func TestNumberValue(t *testing.T) {
	nv := NewNumberValue(42, "en", "test", nil)

	assert.Equal(t, "number", nv.Type())
	assert.Equal(t, "test", nv.Source())
	assert.Equal(t, "en", nv.Locale())
	assert.Equal(t, bidi.DirAuto, nv.Dir())
	assert.NotNil(t, nv.Options())

	str, err := nv.ToString()
	require.NoError(t, err)
	assert.Equal(t, "42", str)

	value, err := nv.ValueOf()
	require.NoError(t, err)
	assert.Equal(t, 42, value)

	parts, err := nv.ToParts()
	require.NoError(t, err)
	require.Len(t, parts, 1)
	assert.Equal(t, "number", parts[0].Type())
	assert.Equal(t, "42", parts[0].Value())
}

func TestNumberValueTypes(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"int", 42, "42"},
		{"int64", int64(42), "42"},
		{"float64", 42.5, "42.5"},
		{"float32", float32(42.5), "42.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nv := NewNumberValue(tt.value, "en", "test", nil)
			str, err := nv.ToString()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, str)
		})
	}
}

func TestNumberValueSelectKeys(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		keys     []string
		expected []string
	}{
		{"zero exact match", 0, []string{"0", "one", "other"}, []string{"0"}},
		{"one", 1, []string{"zero", "one", "other"}, []string{"one"}},
		{"two exact match", 2, []string{"2", "one", "other"}, []string{"2"}},
		{"other fallback for 0", 0, []string{"one", "other"}, []string{"other"}},
		{"other fallback for 2", 2, []string{"one", "other"}, []string{"other"}},
		{"other fallback for 5", 5, []string{"one", "other"}, []string{"other"}},
		{"other fallback for 10", 10, []string{"one", "other"}, []string{"other"}},
		{"other fallback", 42, []string{"other"}, []string{"other"}},
		{"no match", 42, []string{"zero", "one"}, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nv := NewNumberValue(tt.value, "en", "test", nil)
			keys, err := nv.SelectKeys(tt.keys)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, keys)
		})
	}
}

func TestFallbackValue(t *testing.T) {
	fv := NewFallbackValue("$name", "en")

	assert.Equal(t, "fallback", fv.Type())
	assert.Equal(t, "$name", fv.Source())
	assert.Equal(t, "en", fv.Locale())
	assert.Equal(t, bidi.DirAuto, fv.Dir())
	assert.Nil(t, fv.Options())

	str, err := fv.ToString()
	require.NoError(t, err)
	assert.Equal(t, "{$name}", str)

	value, err := fv.ValueOf()
	require.NoError(t, err)
	assert.Equal(t, "$name", value)

	parts, err := fv.ToParts()
	require.NoError(t, err)
	require.Len(t, parts, 1)
	assert.Equal(t, "fallback", parts[0].Type())
	assert.Equal(t, "{$name}", parts[0].Value())

	keys, err := fv.SelectKeys([]string{"zero", "one", "other"})
	require.NoError(t, err)
	assert.Empty(t, keys)
}

func TestTextPart(t *testing.T) {
	tp := NewTextPart("hello", "source", "en")

	assert.Equal(t, "text", tp.Type())
	assert.Equal(t, "hello", tp.Value())
	assert.Equal(t, "source", tp.Source())
	assert.Equal(t, "en", tp.Locale())
	assert.Equal(t, bidi.DirAuto, tp.Dir())
}

func TestBidiIsolationPart(t *testing.T) {
	bip := NewBidiIsolationPart(string(bidi.LRI))

	assert.Equal(t, "bidiIsolation", bip.Type())
	assert.Equal(t, string(bidi.LRI), bip.Value())
	assert.Empty(t, bip.Source())
	assert.Empty(t, bip.Locale())
	assert.Equal(t, bidi.DirAuto, bip.Dir())
}

func TestMarkupPart(t *testing.T) {
	options := map[string]interface{}{"class": "bold"}
	mp := NewMarkupPart("open", "b", "<b>", options)

	assert.Equal(t, "markup", mp.Type())
	assert.Equal(t, "b", mp.Value())
	assert.Equal(t, "<b>", mp.Source())
	assert.Empty(t, mp.Locale())
	assert.Equal(t, bidi.DirAuto, mp.Dir())
	assert.Equal(t, "open", mp.Kind())
	assert.Equal(t, "b", mp.Name())
	assert.Equal(t, options, mp.Options())
}

func TestMarkupPartNilOptions(t *testing.T) {
	mp := NewMarkupPart("standalone", "br", "<br/>", nil)

	assert.NotNil(t, mp.Options())
	assert.Empty(t, mp.Options())
}

func TestFallbackPart(t *testing.T) {
	fp := NewFallbackPart("$name", "en")

	assert.Equal(t, "fallback", fp.Type())
	assert.Equal(t, "{$name}", fp.Value())
	assert.Equal(t, "$name", fp.Source())
	assert.Equal(t, "en", fp.Locale())
	assert.Equal(t, bidi.DirAuto, fp.Dir())
}
