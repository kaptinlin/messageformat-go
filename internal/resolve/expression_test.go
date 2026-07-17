package resolve

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// TestResolveExpression_NilExpression tests that nil expressions return fallback
// TypeScript original code:
// if (!expression) return fallback('unknown');
func TestResolveExpression_NilExpression(t *testing.T) {
	ctx := NewContext(
		[]string{"en"},
		functions.DefaultFunctionMap(),
		map[string]any{},
		nil, "best fit")

	result := ResolveExpression(ctx, nil)

	assert.Equal(t, "fallback", result.Type())
	assert.Equal(t, "unknown", result.Source())
}

// TestResolveExpression_WithFunctionRef tests expression resolution with function references
func TestResolveExpression_WithFunctionRef(t *testing.T) {
	tests := []struct {
		name         string
		operand      datamodel.ExpressionArg
		functionName string
		options      datamodel.Options
		scope        map[string]any
		expected     string
	}{
		{
			name:         "literal with number function",
			operand:      datamodel.NewLiteral("42"),
			functionName: "number",
			options:      nil,
			scope:        map[string]any{},
			expected:     "number",
		},
		{
			name:         "variable with string function",
			operand:      datamodel.NewVariableRef("name"),
			functionName: "string",
			options:      nil,
			scope:        map[string]any{"name": "Alice"},
			expected:     "string",
		},
		{
			name:         "nil operand with number function",
			operand:      nil,
			functionName: "number",
			options:      nil,
			scope:        map[string]any{},
			expected:     "fallback", // nil operand results in fallback
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext(
				[]string{"en"},
				functions.DefaultFunctionMap(),
				tt.scope,
				nil, "best fit")

			funcRef := mustFunctionRef(t, tt.functionName, tt.options)
			expr := mustExpression(t, tt.operand, funcRef, nil)

			result := ResolveExpression(ctx, expr)
			assert.Equal(t, tt.expected, result.Type())
		})
	}
}

// TestResolveExpression_LiteralOnly tests expressions with only literals
func TestResolveExpression_LiteralOnly(t *testing.T) {
	ctx := NewContext(
		[]string{"en"},
		functions.DefaultFunctionMap(),
		map[string]any{},
		nil, "best fit")

	literal := datamodel.NewLiteral("test value")
	expr := mustExpression(t, literal, nil, nil)

	result := ResolveExpression(ctx, expr)

	assert.Equal(t, "string", result.Type())
	str, err := result.ToString()
	require.NoError(t, err)
	assert.Equal(t, "test value", str)
}

// TestResolveExpression_VariableOnly tests expressions with only variable references
func TestResolveExpression_VariableOnly(t *testing.T) {
	tests := []struct {
		name         string
		variableName string
		scope        map[string]any
		expectedType string
	}{
		{
			name:         "string variable",
			variableName: "name",
			scope:        map[string]any{"name": "Alice"},
			expectedType: "string",
		},
		{
			name:         "number variable",
			variableName: "age",
			scope:        map[string]any{"age": 42},
			expectedType: "number",
		},
		{
			name:         "missing variable returns fallback",
			variableName: "missing",
			scope:        map[string]any{},
			expectedType: "fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext(
				[]string{"en"},
				functions.DefaultFunctionMap(),
				tt.scope,
				nil, "best fit")

			varRef := datamodel.NewVariableRef(tt.variableName)
			expr := mustExpression(t, varRef, nil, nil)

			result := ResolveExpression(ctx, expr)
			assert.Equal(t, tt.expectedType, result.Type())
		})
	}
}

// TestResolveExpression_ComplexScenarios tests complex resolution scenarios
func TestResolveExpression_ComplexScenarios(t *testing.T) {
	t.Run("function with variable operand", func(t *testing.T) {
		ctx := NewContext(
			[]string{"en"},
			functions.DefaultFunctionMap(),
			map[string]any{"count": 42},
			nil, "best fit")

		varRef := datamodel.NewVariableRef("count")
		options := make(datamodel.Options)
		options["minimumFractionDigits"] = datamodel.NewLiteral("2")
		funcRef := mustFunctionRef(t, "number", options)
		expr := mustExpression(t, varRef, funcRef, nil)

		result := ResolveExpression(ctx, expr)
		assert.Equal(t, "number", result.Type())

		str, err := result.ToString()
		require.NoError(t, err)
		assert.Contains(t, str, "42")
	})

	t.Run("function with literal operand", func(t *testing.T) {
		ctx := NewContext(
			[]string{"en"},
			functions.DefaultFunctionMap(),
			map[string]any{},
			nil, "best fit")

		literal := datamodel.NewLiteral("100")
		funcRef := mustFunctionRef(t, "number", nil)
		expr := mustExpression(t, literal, funcRef, nil)

		result := ResolveExpression(ctx, expr)
		assert.Equal(t, "number", result.Type())

		str, err := result.ToString()
		require.NoError(t, err)
		assert.Equal(t, "100", str)
	})
}

// TestResolveExpression_ErrorHandling tests error handling in expression resolution
func TestResolveExpression_ErrorHandling(t *testing.T) {
	t.Run("unknown function triggers error", func(t *testing.T) {
		var errorCalled bool
		onError := func(err error) {
			errorCalled = true
		}

		ctx := NewContext(
			[]string{"en"},
			functions.DefaultFunctionMap(),
			map[string]any{},
			onError, "best fit")

		literal := datamodel.NewLiteral("test")
		funcRef := mustFunctionRef(t, "unknownFunction", nil)
		expr := mustExpression(t, literal, funcRef, nil)

		result := ResolveExpression(ctx, expr)

		// Should return fallback for unknown function
		assert.Equal(t, "fallback", result.Type())
		assert.True(t, errorCalled)
	})

	t.Run("function with invalid operand", func(t *testing.T) {
		ctx := NewContext(
			[]string{"en"},
			functions.DefaultFunctionMap(),
			map[string]any{},
			nil, "best fit")

		// Try to use number function with a variable that doesn't exist
		varRef := datamodel.NewVariableRef("nonexistent")
		funcRef := mustFunctionRef(t, "number", nil)
		expr := mustExpression(t, varRef, funcRef, nil)

		result := ResolveExpression(ctx, expr)

		// The number function should still work, producing a number from nil/fallback
		assert.NotNil(t, result)
	})
}

// TestResolveExpression_WithAnnotations tests expressions with annotations
func TestResolveExpression_WithAnnotations(t *testing.T) {
	ctx := NewContext(
		[]string{"en"},
		functions.DefaultFunctionMap(),
		map[string]any{"value": 42},
		nil, "best fit")

	varRef := datamodel.NewVariableRef("value")
	funcRef := mustFunctionRef(t, "number", nil)

	// Create annotations
	annotations := datamodel.Attributes{
		"test": datamodel.NewLiteral("annotation"),
	}

	expr := mustExpression(t, varRef, funcRef, annotations)

	result := ResolveExpression(ctx, expr)
	assert.Equal(t, "number", result.Type())
}

// TestResolveExpression_LocaleHandling tests locale handling in expression resolution
func TestResolveExpression_LocaleHandling(t *testing.T) {
	tests := []struct {
		name     string
		locales  []string
		expected string
	}{
		{
			name:     "single locale",
			locales:  []string{"en"},
			expected: "en",
		},
		{
			name:     "multiple locales",
			locales:  []string{"fr", "en"},
			expected: "fr",
		},
		{
			name:     "empty locales defaults to en",
			locales:  []string{},
			expected: "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext(
				tt.locales,
				functions.DefaultFunctionMap(),
				map[string]any{"value": "test"},
				nil, "best fit")

			varRef := datamodel.NewVariableRef("value")
			expr := mustExpression(t, varRef, nil, nil)

			result := ResolveExpression(ctx, expr)
			assert.Equal(t, tt.expected, result.Locale())
		})
	}
}

// TestResolveExpression_ExternalMessageValuePreservedAsUnknown tests that externally supplied MessageValues
// are preserved as unknown input values instead of being treated as locally resolved values.
func TestResolveExpression_ExternalMessageValuePreservedAsUnknown(t *testing.T) {
	preformatted := messagevalue.NewStringValue("formatted", "en", "test")
	ctx := NewContext(
		[]string{"en"},
		functions.DefaultFunctionMap(),
		map[string]any{
			"preformatted": preformatted,
		},
		nil, "best fit")

	varRef := datamodel.NewVariableRef("preformatted")
	expr := mustExpression(t, varRef, nil, nil)

	result := ResolveExpression(ctx, expr)

	assert.Equal(t, "unknown", result.Type())
	value, err := result.ValueOf()
	require.NoError(t, err)
	assert.Same(t, preformatted, value)
}
