package main

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalizedMessageFormatsAndReportsErrors(t *testing.T) {
	t.Parallel()

	message := NewLocalizedMessage()
	require.NoError(t, message.AddLocale("en", "Hello {name}, you have {count, plural, one {# item} other {# items}}."))

	got, err := message.Format("en", map[string]any{"name": "Ada", "count": 2})
	require.NoError(t, err)
	assert.Equal(t, "Hello Ada, you have 2 items.", got)

	_, err = message.Format("fr", map[string]any{"name": "Ada"})
	assert.ErrorIs(t, err, ErrUnsupportedLocale)

	delete(message.templates, "en")
	_, err = message.Format("en", map[string]any{"name": "Ada"})
	assert.True(t, errors.Is(err, ErrTemplateNotFound), "error %v should wrap ErrTemplateNotFound", err)
}

func TestGetCurrencyForLocale(t *testing.T) {
	t.Parallel()

	tests := []struct {
		locale string
		want   string
	}{
		{locale: "en", want: "USD"},
		{locale: "fr", want: "EUR"},
		{locale: "ja", want: "JPY"},
		{locale: "zh-CN", want: "CNY"},
		{locale: "unknown", want: "USD"},
	}

	for _, tc := range tests {
		t.Run(tc.locale, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, getCurrencyForLocale(tc.locale))
		})
	}
}

func TestMultilingualExampleRuns(t *testing.T) {
	silenceStdout(t)

	main()
}

func silenceStdout(t *testing.T) {
	t.Helper()

	originalStdout := os.Stdout
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	require.NoError(t, err)

	os.Stdout = devNull
	t.Cleanup(func() {
		os.Stdout = originalStdout
		require.NoError(t, devNull.Close())
	})
}
