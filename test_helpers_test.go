package messageformat

import (
	"testing"

	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	"github.com/stretchr/testify/require"
)

// mustPattern constructs a valid pattern for root-package tests.
// TypeScript original code:
// const pattern: Pattern = elements;
func mustPattern(t *testing.T, elements []datamodel.PatternElement) datamodel.Pattern {
	t.Helper()

	pattern, err := datamodel.NewPattern(elements)
	require.NoError(t, err)
	return pattern
}

// mustPatternMessage constructs a valid pattern message for root-package tests.
// TypeScript original code:
// const message: PatternMessage = { type: 'message', declarations, pattern };
func mustPatternMessage(t *testing.T, declarations []datamodel.Declaration, pattern datamodel.Pattern, comment string) *datamodel.PatternMessage {
	t.Helper()

	message, err := datamodel.NewPatternMessage(declarations, pattern, comment)
	require.NoError(t, err)
	return message
}

// mustVariant constructs a valid variant for root-package tests.
// TypeScript original code:
// const variant: Variant = { keys, value };
func mustVariant(t *testing.T, keys []datamodel.VariantKey, value datamodel.Pattern) *datamodel.Variant {
	t.Helper()

	variant, err := datamodel.NewVariant(keys, value)
	require.NoError(t, err)
	return variant
}

// mustSelectMessage constructs a valid select message for root-package tests.
// TypeScript original code:
// const message: SelectMessage = { type: 'select', declarations, selectors, variants };
func mustSelectMessage(t *testing.T, declarations []datamodel.Declaration, selectors []datamodel.VariableRef, variants []datamodel.Variant, comment string) *datamodel.SelectMessage {
	t.Helper()

	message, err := datamodel.NewSelectMessage(declarations, selectors, variants, comment)
	require.NoError(t, err)
	return message
}

// mustFunctionRef constructs a valid function reference for root-package tests.
// TypeScript original code:
// const functionRef: FunctionRef = { type: 'function', name, options };
func mustFunctionRef(t *testing.T, name string, options datamodel.Options) *datamodel.FunctionRef {
	t.Helper()

	functionRef, err := datamodel.NewFunctionRef(name, options)
	require.NoError(t, err)
	return functionRef
}

// diagnosticsFromError expands joined diagnostics in encounter order.
// TypeScript original code:
// const diagnostics = errors;
func diagnosticsFromError(err error) []error {
	if err == nil {
		return nil
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		var diagnostics []error
		for _, child := range joined.Unwrap() {
			diagnostics = append(diagnostics, diagnosticsFromError(child)...)
		}
		return diagnostics
	}
	return []error{err}
}
