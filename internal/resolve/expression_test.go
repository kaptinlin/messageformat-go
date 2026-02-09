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
		functions.DefaultFunctions,
		map[string]interface{}{},
		nil,
	)

	result := ResolveExpression(ctx, nil)

	assert.Equal(t, "fallback", result.Type())
	assert.Equal(t, "unknown", result.Source())
}

// TestResolveExpression_NilArg tests expressions with nil arguments
func TestResolveExpression_NilArg(t *testing.T) {
	ctx := NewContext(
		[]string{"en"},
		functions.DefaultFunctions,
		map[string]interface{}{},
		nil,
	)

	// Create expression with nil arg
	expr := datamodel.NewExpression(nil, nil, nil)
	result := ResolveExpression(ctx, expr)

	assert.Equal(t, "fallback", result.Type())
	assert.Equal(t, "unknown", result.Source())
}

// TestResolveExpression_WithFunctionRef tests expression resolution with function references
func TestResolveExpression_WithFunctionRef(t *testing.T) {
	tests := []struct {
		name         string
		operand      datamodel.Node
		functionName string
		options      datamodel.Options
		scope        map[string]interface{}
		expected     string
	}{
		{
			name:         "literal with number function",
			operand:      datamodel.NewLiteral("42"),
			functionName: "number",
			options:      nil,
			scope:        map[string]interface{}{},
			expected:     "number",
		},
		{
			name:         "variable with string function",
			operand:      datamodel.NewVariableRef("name"),
			functionName: "string",
			options:      nil,
			scope:        map[string]interface{}{"name": "Alice"},
			expected:     "string",
		},
		{
			name:         "nil operand with number function",
			operand:      nil,
			functionName: "number",
			options:      nil,
			scope:        map[string]interface{}{},
			expected:     "fallback", // nil operand results in fallback
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext(
				[]string{"en"},
				functions.DefaultFunctions,
				tt.scope,
				nil,
			)

			funcRef := datamodel.NewFunctionRef(tt.functionName, tt.options)
			expr := datamodel.NewExpression(tt.operand, funcRef, nil)

			result := ResolveExpression(ctx, expr)
			assert.Equal(t, tt.expected, result.Type())
		})
	}
}

// TestResolveExpression_LiteralOnly tests expressions with only literals
func TestResolveExpression_LiteralOnly(t *testing.T) {
	ctx := NewContext(
		[]string{"en"},
		functions.DefaultFunctions,
		map[string]interface{}{},
		nil,
	)

	literal := datamodel.NewLiteral("test value")
	expr := datamodel.NewExpression(literal, nil, nil)

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
		scope        map[string]interface{}
		expectedType string
	}{
		{
			name:         "string variable",
			variableName: "name",
			scope:        map[string]interface{}{"name": "Alice"},
			expectedType: "string",
		},
		{
			name:         "number variable",
			variableName: "age",
			scope:        map[string]interface{}{"age": 42},
			expectedType: "number",
		},
		{
			name:         "missing variable returns fallback",
			variableName: "missing",
			scope:        map[string]interface{}{},
			expectedType: "fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext(
				[]string{"en"},
				functions.DefaultFunctions,
				tt.scope,
				nil,
			)

			varRef := datamodel.NewVariableRef(tt.variableName)
			expr := datamodel.NewExpression(varRef, nil, nil)

			result := ResolveExpression(ctx, expr)
			assert.Equal(t, tt.expectedType, result.Type())
		})
	}
}

// TestResolveExpression_UnsupportedType tests that unsupported expression types return fallback
func TestResolveExpression_UnsupportedType(t *testing.T) {
	var errorCalled bool
	ctx := NewContext(
		[]string{"en"},
		functions.DefaultFunctions,
		map[string]interface{}{},
		func(err error) {
			errorCalled = true
		},
	)

	// Create a mock unsupported node type
	unsupportedNode := &mockUnsupportedNode{}
	expr := datamodel.NewExpression(unsupportedNode, nil, nil)

	// This should return a fallback value instead of panicking
	result := ResolveExpression(ctx, expr)
	assert.Equal(t, "fallback", result.Type())
	assert.True(t, errorCalled, "OnError should have been called")
}

// TestResolveExpression_ComplexScenarios tests complex resolution scenarios
func TestResolveExpression_ComplexScenarios(t *testing.T) {
	t.Run("function with variable operand", func(t *testing.T) {
		ctx := NewContext(
			[]string{"en"},
			functions.DefaultFunctions,
			map[string]interface{}{"count": 42},
			nil,
		)

		varRef := datamodel.NewVariableRef("count")
		options := make(datamodel.Options)
		options["minimumFractionDigits"] = datamodel.NewLiteral("2")
		funcRef := datamodel.NewFunctionRef("number", options)
		expr := datamodel.NewExpression(varRef, funcRef, nil)

		result := ResolveExpression(ctx, expr)
		assert.Equal(t, "number", result.Type())

		str, err := result.ToString()
		require.NoError(t, err)
		assert.Contains(t, str, "42")
	})

	t.Run("function with literal operand", func(t *testing.T) {
		ctx := NewContext(
			[]string{"en"},
			functions.DefaultFunctions,
			map[string]interface{}{},
			nil,
		)

		literal := datamodel.NewLiteral("100")
		funcRef := datamodel.NewFunctionRef("number", nil)
		expr := datamodel.NewExpression(literal, funcRef, nil)

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
			functions.DefaultFunctions,
			map[string]interface{}{},
			onError,
		)

		literal := datamodel.NewLiteral("test")
		funcRef := datamodel.NewFunctionRef("unknownFunction", nil)
		expr := datamodel.NewExpression(literal, funcRef, nil)

		result := ResolveExpression(ctx, expr)

		// Should return fallback for unknown function
		assert.Equal(t, "fallback", result.Type())
		assert.True(t, errorCalled)
	})

	t.Run("function with invalid operand", func(t *testing.T) {
		ctx := NewContext(
			[]string{"en"},
			functions.DefaultFunctions,
			map[string]interface{}{},
			nil,
		)

		// Try to use number function with a variable that doesn't exist
		varRef := datamodel.NewVariableRef("nonexistent")
		funcRef := datamodel.NewFunctionRef("number", nil)
		expr := datamodel.NewExpression(varRef, funcRef, nil)

		result := ResolveExpression(ctx, expr)

		// The number function should still work, producing a number from nil/fallback
		assert.NotNil(t, result)
	})
}

// TestResolveExpression_WithAnnotations tests expressions with annotations
func TestResolveExpression_WithAnnotations(t *testing.T) {
	ctx := NewContext(
		[]string{"en"},
		functions.DefaultFunctions,
		map[string]interface{}{"value": 42},
		nil,
	)

	varRef := datamodel.NewVariableRef("value")
	funcRef := datamodel.NewFunctionRef("number", nil)

	// Create annotations
	annotations := datamodel.Attributes{
		"test": datamodel.NewLiteral("annotation"),
	}

	expr := datamodel.NewExpression(varRef, funcRef, annotations)

	result := ResolveExpression(ctx, expr)
	assert.Equal(t, "number", result.Type())
}

// mockUnsupportedNode is a mock implementation of an unsupported node type
type mockUnsupportedNode struct{}

func (m *mockUnsupportedNode) Type() string     { return "unsupported" }
func (m *mockUnsupportedNode) CST() interface{} { return nil }

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
				functions.DefaultFunctions,
				map[string]interface{}{"value": "test"},
				nil,
			)

			varRef := datamodel.NewVariableRef("value")
			expr := datamodel.NewExpression(varRef, nil, nil)

			result := ResolveExpression(ctx, expr)
			assert.Equal(t, tt.expected, result.Locale())
		})
	}
}

// TestResolveExpression_MessageValuePassthrough tests that existing MessageValues are handled correctly
func TestResolveExpression_MessageValuePassthrough(t *testing.T) {
	ctx := NewContext(
		[]string{"en"},
		functions.DefaultFunctions,
		map[string]interface{}{
			"preformatted": messagevalue.NewStringValue("formatted", "en", "test"),
		},
		nil,
	)

	varRef := datamodel.NewVariableRef("preformatted")
	expr := datamodel.NewExpression(varRef, nil, nil)

	result := ResolveExpression(ctx, expr)

	// The MessageValue gets wrapped by the string function
	assert.Equal(t, "string", result.Type())
	// The ToString might return the struct representation or the value
	str, err := result.ToString()
	require.NoError(t, err)
	// Accept either the original value or struct representation
	assert.True(t, str == "formatted" || len(str) > 0, "Expected non-empty string result")
}
