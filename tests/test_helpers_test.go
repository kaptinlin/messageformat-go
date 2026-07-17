package tests

import (
	"testing"

	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	"github.com/stretchr/testify/require"
)

// mustPattern constructs a valid pattern for public feature tests.
// TypeScript original code:
// const pattern: Pattern = elements;
func mustPattern(t *testing.T, elements []datamodel.PatternElement) datamodel.Pattern {
	t.Helper()

	pattern, err := datamodel.NewPattern(elements)
	require.NoError(t, err)
	return pattern
}

// mustVariant constructs a valid variant for public feature tests.
// TypeScript original code:
// const variant: Variant = { keys, value };
func mustVariant(t *testing.T, keys []datamodel.VariantKey, value datamodel.Pattern) *datamodel.Variant {
	t.Helper()

	variant, err := datamodel.NewVariant(keys, value)
	require.NoError(t, err)
	return variant
}

// mustPatternMessage constructs a valid pattern message for public feature tests.
// TypeScript original code:
// const message: PatternMessage = { type: 'message', declarations, pattern };
func mustPatternMessage(t *testing.T, declarations []datamodel.Declaration, pattern datamodel.Pattern, comment string) *datamodel.PatternMessage {
	t.Helper()

	message, err := datamodel.NewPatternMessage(declarations, pattern, comment)
	require.NoError(t, err)
	return message
}

// mustSelectMessage constructs a valid select message for public feature tests.
// TypeScript original code:
// const message: SelectMessage = { type: 'select', declarations, selectors, variants };
func mustSelectMessage(t *testing.T, declarations []datamodel.Declaration, selectors []datamodel.VariableRef, variants []datamodel.Variant, comment string) *datamodel.SelectMessage {
	t.Helper()

	message, err := datamodel.NewSelectMessage(declarations, selectors, variants, comment)
	require.NoError(t, err)
	return message
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
