package intlbridge

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLocale(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		tag  string
		want string
	}{
		{name: "empty falls back to English", tag: "", want: "en"},
		{name: "valid tag preserved", tag: "fr-CA", want: "fr-CA"},
		{name: "underscore normalized", tag: "en_US", want: "en-US"},
		{name: "invalid tag falls back to English", tag: "not a locale", want: "en"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := ParseLocale(tc.tag)
			require.Len(t, got, 1)
			assert.Equal(t, tc.want, got[0].String())
		})
	}
}

func TestFirstLocale(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		tags []string
		want string
	}{
		{name: "uses first non-empty locale", tags: []string{"", "de-DE", "fr-FR"}, want: "de-DE"},
		{name: "normalizes selected locale", tags: []string{"", "pt_BR"}, want: "pt-BR"},
		{name: "all empty falls back to English", tags: []string{"", ""}, want: "en"},
		{name: "nil falls back to English", tags: nil, want: "en"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := FirstLocale(tc.tags)
			require.Len(t, got, 1)
			assert.Equal(t, tc.want, got[0].String())
		})
	}
}
