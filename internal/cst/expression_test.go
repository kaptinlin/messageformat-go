package cst

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseExpression_VariableReferences tests basic variable reference parsing
func TestParseExpression_VariableReferences(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		expectedType   string // "variable" or "literal"
		expectedValue  string
		shouldError    bool
	}{
		{
			name:         "simple variable reference",
			source:       "{name}",
			expectedType: "variable",
			expectedValue: "name",
		},
		{
			name:         "variable with underscore",
			source:       "{user_name}",
			expectedType: "variable", 
			expectedValue: "user_name",
		},
		{
			name:         "variable with number",
			source:       "{count123}",
			expectedType: "variable",
			expectedValue: "count123",
		},
		{
			name:         "explicit variable reference",
			source:       "{$name}",
			expectedType: "variable",
			expectedValue: "name",
		},
		{
			name:         "quoted literal",
			source:       "{|literal text|}",
			expectedType: "literal",
			expectedValue: "literal text",
		},
		// Note: MessageFormat 2.0 spec only supports |quoted| literals, not "double quotes"
		{
			name:         "numeric literal positive",
			source:       "{123}",
			expectedType: "literal",
			expectedValue: "123",
		},
		{
			name:         "numeric literal negative",
			source:       "{-456}",
			expectedType: "literal",
			expectedValue: "-456",
		},
		{
			name:         "numeric literal with plus",
			source:       "{+789}",
			expectedType: "literal",
			expectedValue: "+789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewParseContext(tt.source, false)
			result := parseExpression(ctx, 0)

			require.NotNil(t, result, "parseExpression should not return nil")

			if tt.shouldError {
				assert.True(t, len(ctx.errors) > 0, "Expected parsing errors")
				return
			}

			assert.Equal(t, 0, len(ctx.errors), "No parsing errors expected, got: %v", ctx.errors)

			// Check the argument type and value
			arg := result.Arg()
			require.NotNil(t, arg, "Expression argument should not be nil")

			switch tt.expectedType {
			case "variable":
				varRef, ok := arg.(*VariableRef)
				require.True(t, ok, "Expected VariableRef, got %T", arg)
				assert.Equal(t, tt.expectedValue, varRef.Name(), "Variable name mismatch")
			case "literal":
				literal, ok := arg.(*Literal)
				require.True(t, ok, "Expected Literal, got %T", arg)
				assert.Equal(t, tt.expectedValue, literal.Value(), "Literal value mismatch")
			default:
				t.Fatalf("Unknown expected type: %s", tt.expectedType)
			}
		})
	}
}

// TestParseExpression_FunctionCalls tests function call parsing
func TestParseExpression_FunctionCalls(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		expectedVar    string
		expectedFunc   string
	}{
		{
			name:         "integer function",
			source:       "{count :integer}",
			expectedVar:  "count",
			expectedFunc: "integer",
		},
		{
			name:         "number function",
			source:       "{price :number}",
			expectedVar:  "price", 
			expectedFunc: "number",
		},
		{
			name:         "string function",
			source:       "{name :string}",
			expectedVar:  "name",
			expectedFunc: "string",
		},
		{
			name:         "explicit variable in function",
			source:       "{$count :integer}",
			expectedVar:  "count",
			expectedFunc: "integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewParseContext(tt.source, false)
			result := parseExpression(ctx, 0)

			require.NotNil(t, result, "parseExpression should not return nil")
			assert.Equal(t, 0, len(ctx.errors), "No parsing errors expected, got: %v", ctx.errors)

			// Check the argument (variable)
			arg := result.Arg()
			require.NotNil(t, arg, "Expression argument should not be nil")
			
			varRef, ok := arg.(*VariableRef)
			require.True(t, ok, "Expected VariableRef, got %T", arg)
			assert.Equal(t, tt.expectedVar, varRef.Name(), "Variable name mismatch")

			// Check the function reference
			funcRef := result.FunctionRef()
			require.NotNil(t, funcRef, "Function reference should not be nil")
			
			if fr, ok := funcRef.(*FunctionRef); ok {
				require.True(t, len(fr.Name()) > 0, "Function name should not be empty")
				// Get function name from identifier parts
				funcName := ""
				for _, part := range fr.Name() {
					funcName += part.Value()
				}
				assert.Equal(t, tt.expectedFunc, funcName, "Function name mismatch")
			} else {
				t.Fatalf("Expected FunctionRef, got %T", funcRef)
			}
		})
	}
}

// TestParseExpression_ComplexCases tests more complex parsing scenarios
func TestParseExpression_ComplexCases(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		expectedType   string
		shouldError    bool
	}{
		{
			name:         "empty expression should error",
			source:       "{}",
			shouldError:  true,
		},
		{
			name:         "malformed expression", 
			source:       "{",
			shouldError:  true,
		},
		{
			name:         "variable with whitespace",
			source:       "{ name }",
			expectedType: "variable",
		},
		{
			name:         "function with whitespace",
			source:       "{ count :integer }",
			expectedType: "variable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewParseContext(tt.source, false)
			result := parseExpression(ctx, 0)

			if tt.shouldError {
				assert.True(t, len(ctx.errors) > 0, "Expected parsing errors")
				return
			}

			require.NotNil(t, result, "parseExpression should not return nil")
			assert.Equal(t, 0, len(ctx.errors), "No parsing errors expected, got: %v", ctx.errors)

			if tt.expectedType == "variable" {
				arg := result.Arg()
				require.NotNil(t, arg, "Expression argument should not be nil")
				_, ok := arg.(*VariableRef)
				assert.True(t, ok, "Expected VariableRef, got %T", arg)
			}
		})
	}
}

// TestHelperFunctions tests the new helper functions
func TestHelperFunctions(t *testing.T) {
	t.Run("isDigit", func(t *testing.T) {
		assert.True(t, isDigit('0'))
		assert.True(t, isDigit('5'))
		assert.True(t, isDigit('9'))
		assert.False(t, isDigit('a'))
		assert.False(t, isDigit('-'))
		assert.False(t, isDigit('+'))
	})

	t.Run("isIdentifierStart", func(t *testing.T) {
		assert.True(t, isIdentifierStart('a'))
		assert.True(t, isIdentifierStart('A'))
		assert.True(t, isIdentifierStart('_'))
		assert.False(t, isIdentifierStart('0'))  // digits cannot start identifiers
		assert.False(t, isIdentifierStart('-'))  // - cannot start identifiers  
		assert.False(t, isIdentifierStart('.'))  // . cannot start identifiers
	})
}

// TestParseVariableRef tests the new parseVariableRef function
func TestParseVariableRef(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		start         int
		expectedName  string
		expectedEnd   int
		shouldError   bool
	}{
		{
			name:         "simple name",
			source:       "name",
			start:        0,
			expectedName: "name",
			expectedEnd:  4,
		},
		{
			name:         "name with underscore",
			source:       "user_name",
			start:        0,
			expectedName: "user_name",
			expectedEnd:  9,
		},
		{
			name:         "name in middle of string",
			source:       "hello name world",
			start:        6,
			expectedName: "name", 
			expectedEnd:  10,
		},
		{
			name:         "empty name should error",
			source:       "",
			start:        0,
			shouldError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewParseContext(tt.source, false)
			result := parseVariableRef(ctx, tt.start)

			require.NotNil(t, result, "parseVariableRef should not return nil")

			if tt.shouldError {
				assert.True(t, len(ctx.errors) > 0, "Expected parsing errors")
				return
			}

			assert.Equal(t, 0, len(ctx.errors), "No parsing errors expected")
			assert.Equal(t, tt.expectedName, result.Name(), "Variable name mismatch")
			assert.Equal(t, tt.expectedEnd, result.End(), "End position mismatch")
		})
	}
}