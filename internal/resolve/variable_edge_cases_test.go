package resolve

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	"github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// TestResolveVariableRef_CircularReference tests circular reference detection
// TypeScript original code:
// Circular references should be detected and return fallback
func TestResolveVariableRef_CircularReference(t *testing.T) {
	t.Run("simple circular reference", func(t *testing.T) {
		var errorCalled bool
		var errorMsg string
		onError := func(err error) {
			errorCalled = true
			errorMsg = err.Error()
		}

		// Create context where variable "a" references itself
		expr := datamodel.NewExpression(
			datamodel.NewVariableRef("a"),
			nil,
			nil,
		)

		ctx := NewContext(
			[]string{"en"},
			functions.DefaultFunctions,
			map[string]interface{}{
				"a": NewUnresolvedExpression(expr, nil),
			},
			onError,
		)

		ref := datamodel.NewVariableRef("a")
		result := ResolveVariableRef(ctx, ref)

		// Should return fallback due to circular reference
		assert.Equal(t, "fallback", result.Type())
		assert.True(t, errorCalled)
		assert.Contains(t, errorMsg, "circular reference")
	})

	t.Run("indirect circular reference A->B->A", func(t *testing.T) {
		var errorCalled bool
		onError := func(err error) {
			errorCalled = true
		}

		// Create expressions for A and B that reference each other
		exprA := datamodel.NewExpression(
			datamodel.NewVariableRef("b"),
			nil,
			nil,
		)
		exprB := datamodel.NewExpression(
			datamodel.NewVariableRef("a"),
			nil,
			nil,
		)

		scope := map[string]interface{}{
			"a": NewUnresolvedExpression(exprA, nil),
			"b": NewUnresolvedExpression(exprB, nil),
		}

		ctx := NewContext(
			[]string{"en"},
			functions.DefaultFunctions,
			scope,
			onError,
		)

		// Try to resolve "a", which references "b", which references "a"
		ref := datamodel.NewVariableRef("a")
		result := ResolveVariableRef(ctx, ref)

		// Should detect circular reference and return fallback
		assert.Equal(t, "fallback", result.Type())
		assert.True(t, errorCalled)
	})

	t.Run("no circular reference with proper resolution", func(t *testing.T) {
		// Create expression where A references B, and B is a concrete value
		exprA := datamodel.NewExpression(
			datamodel.NewVariableRef("b"),
			nil,
			nil,
		)

		scope := map[string]interface{}{
			"a": NewUnresolvedExpression(exprA, nil),
			"b": "concrete value",
		}

		ctx := NewContext(
			[]string{"en"},
			functions.DefaultFunctions,
			scope,
			nil,
		)

		ref := datamodel.NewVariableRef("a")
		result := ResolveVariableRef(ctx, ref)

		// Should successfully resolve to "concrete value"
		assert.Equal(t, "string", result.Type())
		str, err := result.ToString()
		require.NoError(t, err)
		assert.Equal(t, "concrete value", str)
	})
}

// TestResolveVariableRef_NestedPaths tests nested variable path resolution
func TestResolveVariableRef_NestedPaths(t *testing.T) {
	tests := []struct {
		name         string
		variablePath string
		scope        map[string]interface{}
		expectedType string
		expectedVal  string
	}{
		{
			name:         "one level deep - a.b",
			variablePath: "a.b",
			scope: map[string]interface{}{
				"a": map[string]interface{}{
					"b": "value",
				},
			},
			expectedType: "string",
			expectedVal:  "value",
		},
		{
			name:         "two levels deep - a.b.c",
			variablePath: "a.b.c",
			scope: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": 42,
					},
				},
			},
			expectedType: "number",
			expectedVal:  "42",
		},
		{
			name:         "three levels deep - user.profile.name",
			variablePath: "user.profile.name",
			scope: map[string]interface{}{
				"user": map[string]interface{}{
					"profile": map[string]interface{}{
						"name": "Alice",
					},
				},
			},
			expectedType: "string",
			expectedVal:  "Alice",
		},
		{
			name:         "partial path exists as key - user.name when 'user.name' key exists",
			variablePath: "user.name",
			scope: map[string]interface{}{
				"user.name": "direct",
				"user": map[string]interface{}{
					"name": "nested",
				},
			},
			expectedType: "string",
			expectedVal:  "direct", // Direct key takes precedence
		},
		{
			name:         "missing nested path",
			variablePath: "a.b.c",
			scope: map[string]interface{}{
				"a": map[string]interface{}{
					"x": "value",
				},
			},
			expectedType: "fallback",
			expectedVal:  "$a.b.c",
		},
		{
			name:         "path with numeric value at intermediate level",
			variablePath: "a.b.c",
			scope: map[string]interface{}{
				"a": map[string]interface{}{
					"b": 42, // b is a number, not an object
				},
			},
			expectedType: "fallback",
			expectedVal:  "$a.b.c",
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

			ref := datamodel.NewVariableRef(tt.variablePath)
			result := ResolveVariableRef(ctx, ref)

			assert.Equal(t, tt.expectedType, result.Type())

			if tt.expectedType != "fallback" {
				str, err := result.ToString()
				require.NoError(t, err)
				assert.Equal(t, tt.expectedVal, str)
			} else {
				assert.Equal(t, tt.expectedVal, result.Source())
			}
		})
	}
}

// TestResolveVariableRef_MissingVariables tests handling of missing variables
func TestResolveVariableRef_MissingVariables(t *testing.T) {
	var errorCalled bool
	var errorType string
	onError := func(err error) {
		errorCalled = true
		var resErr *errors.MessageResolutionError
		if e, ok := err.(*errors.MessageResolutionError); ok {
			resErr = e
			errorType = resErr.ErrorType()
		}
	}

	ctx := NewContext(
		[]string{"en"},
		functions.DefaultFunctions,
		map[string]interface{}{},
		onError,
	)

	ref := datamodel.NewVariableRef("nonexistent")
	result := ResolveVariableRef(ctx, ref)

	assert.Equal(t, "fallback", result.Type())
	assert.Equal(t, "$nonexistent", result.Source())
	assert.True(t, errorCalled)
	assert.Equal(t, "unresolved-variable", errorType)
}

// TestLookupVariableRef_UnresolvedExpression tests unresolved expression handling
func TestLookupVariableRef_UnresolvedExpression(t *testing.T) {
	t.Run("basic unresolved expression", func(t *testing.T) {
		// Create an unresolved expression that references a literal
		expr := datamodel.NewExpression(
			datamodel.NewLiteral("42"),
			datamodel.NewFunctionRef("number", nil),
			nil,
		)

		ctx := NewContext(
			[]string{"en"},
			functions.DefaultFunctions,
			map[string]interface{}{
				"x": NewUnresolvedExpression(expr, nil),
			},
			nil,
		)

		ref := datamodel.NewVariableRef("x")
		result := lookupVariableRef(ctx, ref)

		// Should resolve to a MessageValue
		assert.NotNil(t, result)
		mv, ok := result.(messagevalue.MessageValue)
		assert.True(t, ok)
		assert.Equal(t, "number", mv.Type())
	})

	t.Run("unresolved expression with custom scope", func(t *testing.T) {
		// Create an unresolved expression with its own scope
		expr := datamodel.NewExpression(
			datamodel.NewVariableRef("y"),
			nil,
			nil,
		)

		customScope := map[string]interface{}{
			"y": "scoped value",
		}

		ctx := NewContext(
			[]string{"en"},
			functions.DefaultFunctions,
			map[string]interface{}{
				"x": NewUnresolvedExpression(expr, customScope),
				"y": "outer value", // This should be ignored
			},
			nil,
		)

		ref := datamodel.NewVariableRef("x")
		result := lookupVariableRef(ctx, ref)

		// Should resolve using the custom scope
		assert.NotNil(t, result)
		// The result could be a MessageValue or the raw value "scoped value"
		if mv, ok := result.(messagevalue.MessageValue); ok {
			str, err := mv.ToString()
			require.NoError(t, err)
			assert.Equal(t, "scoped value", str)
		} else {
			// If it's returned as the raw value, check directly
			assert.Equal(t, "scoped value", result)
		}
	})

	t.Run("input declaration without function", func(t *testing.T) {
		// Test .input declaration case where we have a simple variable reference
		// without a function - should return the original parameter value
		expr := datamodel.NewExpression(
			datamodel.NewVariableRef("param"),
			nil,
			nil,
		)

		scopeWithParam := map[string]interface{}{
			"param": "original value",
		}

		ctx := NewContext(
			[]string{"en"},
			functions.DefaultFunctions,
			map[string]interface{}{
				"local": NewUnresolvedExpression(expr, scopeWithParam),
			},
			nil,
		)

		ref := datamodel.NewVariableRef("local")
		result := lookupVariableRef(ctx, ref)

		// Should return the original parameter value directly
		assert.Equal(t, "original value", result)
	})
}

// TestGetValue_EdgeCases tests edge cases in getValue function
func TestGetValue_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		scope    interface{}
		varName  string
		expected interface{}
	}{
		{
			name:     "nil scope",
			scope:    nil,
			varName:  "test",
			expected: nil,
		},
		{
			name:     "string scope (not a valid scope)",
			scope:    "not a scope",
			varName:  "test",
			expected: nil,
		},
		{
			name:     "number scope (not a valid scope)",
			scope:    42,
			varName:  "test",
			expected: nil,
		},
		{
			name:     "empty map",
			scope:    map[string]interface{}{},
			varName:  "test",
			expected: nil,
		},
		{
			name: "map[interface{}]interface{}",
			scope: map[interface{}]interface{}{
				"key": "value",
			},
			varName:  "key",
			expected: "value",
		},
		{
			name: "dotted key with no match",
			scope: map[string]interface{}{
				"a": "value",
			},
			varName:  "a.b.c",
			expected: nil,
		},
		{
			name: "dotted key with partial match only",
			scope: map[string]interface{}{
				"a.b": "partial",
			},
			varName:  "a.b.c",
			expected: nil,
		},
		{
			name: "complex nested path resolution",
			scope: map[string]interface{}{
				"a.b": map[string]interface{}{
					"c": "found",
				},
			},
			varName:  "a.b.c",
			expected: "found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getValue(tt.scope, tt.varName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsScope_AllTypes tests isScope with various types
func TestIsScope_AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		// Valid scopes
		{name: "map[string]interface{}", value: map[string]interface{}{}, expected: true},
		{name: "map[interface{}]interface{}", value: map[interface{}]interface{}{}, expected: true},
		{name: "struct", value: struct{}{}, expected: true},
		{name: "pointer to struct", value: &struct{}{}, expected: true},
		{name: "function", value: func() {}, expected: true},

		// Invalid scopes
		{name: "nil", value: nil, expected: false},
		{name: "string", value: "test", expected: false},
		{name: "int", value: 42, expected: false},
		{name: "float", value: 3.14, expected: false},
		{name: "bool", value: true, expected: false},
		{name: "slice", value: []string{}, expected: false},
		{name: "array", value: [3]int{}, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isScope(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetValueType tests getValueType function
func TestGetValueType(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{name: "nil", value: nil, expected: "undefined"},
		{name: "bool", value: true, expected: "boolean"},
		{name: "int", value: 42, expected: "number"},
		{name: "int64", value: int64(42), expected: "number"},
		{name: "float64", value: 3.14, expected: "number"},
		{name: "string", value: "test", expected: "string"},
		{name: "function", value: func(...interface{}) interface{} { return nil }, expected: "function"},
		{name: "map", value: map[string]interface{}{}, expected: "object"},
		{name: "struct", value: struct{}{}, expected: "object"},
		{name: "slice", value: []string{}, expected: "object"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getValueType(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestResolveVariableRef_PointerTypes tests variable resolution with pointer types
func TestResolveVariableRef_PointerTypes(t *testing.T) {
	t.Run("pointer to int", func(t *testing.T) {
		value := 42
		ctx := NewContext(
			[]string{"en"},
			functions.DefaultFunctions,
			map[string]interface{}{
				"val": &value,
			},
			nil,
		)

		ref := datamodel.NewVariableRef("val")
		result := ResolveVariableRef(ctx, ref)

		// Pointers to primitive types are treated as objects, not numbers
		// The type will be "string" since unknown objects are converted to strings
		assert.True(t, result.Type() == "string" || result.Type() == "fallback",
			"Expected string or fallback, got: %s", result.Type())
	})

	t.Run("pointer to string", func(t *testing.T) {
		value := "test"
		ctx := NewContext(
			[]string{"en"},
			functions.DefaultFunctions,
			map[string]interface{}{
				"val": &value,
			},
			nil,
		)

		ref := datamodel.NewVariableRef("val")
		result := ResolveVariableRef(ctx, ref)

		// Should treat pointer to string as string type
		assert.Equal(t, "string", result.Type())
	})
}

// TestResolveVariableRef_LocalVarsTracking tests local variable tracking
func TestResolveVariableRef_LocalVarsTracking(t *testing.T) {
	expr := datamodel.NewExpression(
		datamodel.NewLiteral("test"),
		nil,
		nil,
	)

	ctx := NewContext(
		[]string{"en"},
		functions.DefaultFunctions,
		map[string]interface{}{
			"local": NewUnresolvedExpression(expr, nil),
		},
		nil,
	)

	// Initially, LocalVars should be empty
	assert.Empty(t, ctx.LocalVars)

	ref := datamodel.NewVariableRef("local")
	result := ResolveVariableRef(ctx, ref)

	// After resolution, the result should be tracked in LocalVars
	assert.True(t, ctx.LocalVars[result])

	// Resolving again should return the tracked value
	result2 := ResolveVariableRef(ctx, ref)
	assert.Equal(t, result, result2)
}

// TestResolveVariableRef_FallbackType tests handling of fallback message values
func TestResolveVariableRef_FallbackType(t *testing.T) {
	fallbackValue := messagevalue.NewFallbackValue("test", "en")

	ctx := NewContext(
		[]string{"en"},
		functions.DefaultFunctions,
		map[string]interface{}{
			"val": fallbackValue,
		},
		nil,
	)

	ref := datamodel.NewVariableRef("val")
	result := ResolveVariableRef(ctx, ref)

	// Should return a new fallback, not the original
	assert.Equal(t, "fallback", result.Type())
	assert.Equal(t, "$val", result.Source())
}

// TestResolveVariableRef_NoErrorHandler tests that resolution works without error handler
func TestResolveVariableRef_NoErrorHandler(t *testing.T) {
	ctx := NewContext(
		[]string{"en"},
		functions.DefaultFunctions,
		map[string]interface{}{},
		nil, // No error handler
	)

	ref := datamodel.NewVariableRef("missing")
	result := ResolveVariableRef(ctx, ref)

	// Should still return fallback even without error handler
	assert.Equal(t, "fallback", result.Type())
	assert.Equal(t, "$missing", result.Source())
}

// TestResolveVariableRef_ComplexObject tests resolution of complex object types
func TestResolveVariableRef_ComplexObject(t *testing.T) {
	type CustomStruct struct {
		Field1 string
		Field2 int
	}

	customObj := CustomStruct{
		Field1: "value",
		Field2: 42,
	}

	ctx := NewContext(
		[]string{"en"},
		functions.DefaultFunctions,
		map[string]interface{}{
			"obj": customObj,
		},
		nil,
	)

	ref := datamodel.NewVariableRef("obj")
	result := ResolveVariableRef(ctx, ref)

	// Complex objects should be converted to string representation
	assert.Equal(t, "string", result.Type())
	str, err := result.ToString()
	require.NoError(t, err)
	assert.Contains(t, str, fmt.Sprintf("%v", customObj))
}
