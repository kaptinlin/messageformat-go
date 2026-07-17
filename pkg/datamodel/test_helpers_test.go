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

// mustExpression constructs a valid expression for behavior tests.
// TypeScript original code:
// const expression = { type: 'expression', arg, functionRef, attributes };
func mustExpression(t *testing.T, arg ExpressionArg, functionRef *FunctionRef, attributes Attributes) *Expression {
	t.Helper()

	expression, err := NewExpression(arg, functionRef, attributes)
	require.NoError(t, err)
	return expression
}

// mustInputDeclaration constructs an input declaration for behavior tests.
// TypeScript original code:
// const declaration = { type: 'input', name: value.arg.name, value };
func mustInputDeclaration(t *testing.T, value *Expression) *InputDeclaration {
	t.Helper()

	declaration, err := NewInputDeclaration(value)
	require.NoError(t, err)
	return declaration
}

// mustPattern constructs a valid pattern for behavior tests.
// TypeScript original code:
// const pattern: Pattern = elements;
func mustPattern(t *testing.T, elements []PatternElement) Pattern {
	t.Helper()

	pattern, err := NewPattern(elements)
	require.NoError(t, err)
	return pattern
}

// mustVariant constructs a valid variant for behavior tests.
// TypeScript original code:
// const variant: Variant = { keys, value };
func mustVariant(t *testing.T, keys []VariantKey, value Pattern) *Variant {
	t.Helper()

	variant, err := NewVariant(keys, value)
	require.NoError(t, err)
	return variant
}

// mustPatternMessage constructs a valid pattern message for behavior tests.
// TypeScript original code:
// const message: PatternMessage = { type: 'message', declarations, pattern };
func mustPatternMessage(t *testing.T, declarations []Declaration, pattern Pattern, comment string) *PatternMessage {
	t.Helper()

	message, err := NewPatternMessage(declarations, pattern, comment)
	require.NoError(t, err)
	return message
}

// mustSelectMessage constructs a valid select message for behavior tests.
// TypeScript original code:
// const message: SelectMessage = { type: 'select', declarations, selectors, variants };
func mustSelectMessage(t *testing.T, declarations []Declaration, selectors []VariableRef, variants []Variant, comment string) *SelectMessage {
	t.Helper()

	message, err := NewSelectMessage(declarations, selectors, variants, comment)
	require.NoError(t, err)
	return message
}

// mustFunctionRef constructs a valid function reference for behavior tests.
// TypeScript original code:
// const functionRef: FunctionRef = { type: 'function', name, options };
func mustFunctionRef(t *testing.T, name string, options Options) *FunctionRef {
	t.Helper()

	functionRef, err := NewFunctionRef(name, options)
	require.NoError(t, err)
	return functionRef
}
