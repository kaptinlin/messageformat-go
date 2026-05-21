package datamodel

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func mustMarkup(t *testing.T, kind MarkupKind, name string, options Options, attributes Attributes) *Markup {
	t.Helper()

	markup, err := NewMarkup(kind, name, options, attributes)
	require.NoError(t, err)
	return markup
}
