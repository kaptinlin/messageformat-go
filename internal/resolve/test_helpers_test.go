package resolve

import (
	"testing"

	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	"github.com/stretchr/testify/require"
)

func mustMarkup(t *testing.T, kind datamodel.MarkupKind, name string, options datamodel.Options, attributes datamodel.Attributes) *datamodel.Markup {
	t.Helper()

	markup, err := datamodel.NewMarkup(kind, name, options, attributes)
	require.NoError(t, err)
	return markup
}
