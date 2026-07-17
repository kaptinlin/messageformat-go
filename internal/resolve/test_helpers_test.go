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

// mustExpression constructs a valid data-model expression for resolver tests.
// TypeScript original code:
// const expression = { type: 'expression', arg, functionRef, attributes };
func mustExpression(t *testing.T, arg datamodel.ExpressionArg, functionRef *datamodel.FunctionRef, attributes datamodel.Attributes) *datamodel.Expression {
	t.Helper()

	expression, err := datamodel.NewExpression(arg, functionRef, attributes)
	require.NoError(t, err)
	return expression
}

// mustFunctionRef constructs a valid function reference for resolver tests.
// TypeScript original code:
// const functionRef: FunctionRef = { type: 'function', name, options };
func mustFunctionRef(t *testing.T, name string, options datamodel.Options) *datamodel.FunctionRef {
	t.Helper()

	functionRef, err := datamodel.NewFunctionRef(name, options)
	require.NoError(t, err)
	return functionRef
}
