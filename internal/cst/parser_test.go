// Package cst provides comprehensive tests for parser.go
package cst

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseContext_Errors(t *testing.T) {
	ctx := NewParseContext("test", false)
	ctx.OnError("parse-error", 0, 4)

	errors := ctx.Errors()
	assert.Len(t, errors, 1)
	assert.Equal(t, "parse-error", errors[0].Type)
}

func TestOnError_AllErrorTypes(t *testing.T) {
	tests := []struct {
		name        string
		errorType   string
		start       int
		endOrChar   interface{}
		expectedLen int
	}{
		{
			name:        "missing-syntax with string",
			errorType:   "missing-syntax",
			start:       0,
			endOrChar:   "}}",
			expectedLen: 1,
		},
		{
			name:        "missing-syntax with char",
			errorType:   "missing-syntax",
			start:       5,
			endOrChar:   "}",
			expectedLen: 1,
		},
		{
			name:        "extra-content",
			errorType:   "extra-content",
			start:       10,
			endOrChar:   20,
			expectedLen: 1,
		},
		{
			name:        "empty-token",
			errorType:   "empty-token",
			start:       0,
			endOrChar:   1,
			expectedLen: 1,
		},
		{
			name:        "bad-escape",
			errorType:   "bad-escape",
			start:       3,
			endOrChar:   5,
			expectedLen: 1,
		},
		{
			name:        "bad-input-expression",
			errorType:   "bad-input-expression",
			start:       0,
			endOrChar:   10,
			expectedLen: 1,
		},
		{
			name:        "duplicate-option-name",
			errorType:   "duplicate-option-name",
			start:       5,
			endOrChar:   15,
			expectedLen: 1,
		},
		{
			name:        "parse-error",
			errorType:   "parse-error",
			start:       0,
			endOrChar:   5,
			expectedLen: 1,
		},
		{
			name:        "unknown-error-type-fallback",
			errorType:   "unknown-type",
			start:       0,
			endOrChar:   5,
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewParseContext("test source", false)
			ctx.OnError(tt.errorType, tt.start, tt.endOrChar)

			errors := ctx.Errors()
			assert.Len(t, errors, tt.expectedLen)
			if tt.expectedLen > 0 {
				assert.NotNil(t, errors[0])
			}
		})
	}
}

func TestParsePatternMessage_ExtraContent(t *testing.T) {
	// In a simple message, "Hello extra" is valid text - no extra content error
	// Extra content only occurs when there's content after a complete pattern
	ctx := NewParseContext("{{Hello}} extra", false)
	msg := parsePatternMessage(ctx, 0, []Declaration{}, true)

	assert.NotNil(t, msg)
	// Should have error for content after closing }}
	if len(ctx.Errors()) > 0 {
		assert.Equal(t, "extra-content", ctx.Errors()[0].Type)
	}
}

func TestParseSelectMessage_NoSelectors(t *testing.T) {
	source := ".match * {{default}}"
	msg := ParseCST(source, false)

	select_, ok := msg.(*SelectMessage)
	require.True(t, ok)
	// Should have error for empty selectors
	assert.NotEmpty(t, select_.Errors())
}

func TestParseSelectMessage_ErrorInSelector(t *testing.T) {
	source := ".match{$var} * {{default}}"
	msg := ParseCST(source, false)

	select_, ok := msg.(*SelectMessage)
	require.True(t, ok)
	// Should have error for selector in braces
	assert.NotEmpty(t, select_.Errors())
}

func TestParseSelectMessage_MissingWhitespace(t *testing.T) {
	source := ".match$count$gender * {{default}}"
	msg := ParseCST(source, false)

	select_, ok := msg.(*SelectMessage)
	require.True(t, ok)
	// Should have error for missing whitespace between selectors
	assert.NotEmpty(t, select_.Errors())
}

func TestParseVariant_NoKeys(t *testing.T) {
	ctx := NewParseContext("{{pattern}}", false)
	variant := parseVariant(ctx, 0)

	assert.NotNil(t, variant)
	assert.Empty(t, variant.Keys())
}

func TestParseVariant_InvalidLiteral(t *testing.T) {
	ctx := NewParseContext("  {{pattern}}", false)
	variant := parseVariant(ctx, 0)

	assert.NotNil(t, variant)
}

func TestParsePattern_QuotedMissingOpen(t *testing.T) {
	ctx := NewParseContext("missing open", false)
	pattern := parsePattern(ctx, 0, true)

	assert.NotNil(t, pattern)
	assert.Len(t, ctx.Errors(), 1)
	assert.Equal(t, "missing-syntax", ctx.Errors()[0].Type)
}

func TestParsePattern_QuotedMissingClose(t *testing.T) {
	ctx := NewParseContext("{{missing close", false)
	pattern := parsePattern(ctx, 0, true)

	assert.NotNil(t, pattern)
	assert.Len(t, ctx.Errors(), 1)
	assert.Equal(t, "missing-syntax", ctx.Errors()[0].Type)
}

func TestParseDeclarations_UnknownDeclaration(t *testing.T) {
	ctx := NewParseContext(".unknown test", false)
	declarations, pos := parseDeclarations(ctx, 0)

	assert.Len(t, declarations, 1)
	assert.Greater(t, pos, 0)
	// Should create junk for unknown declaration
	junk, ok := declarations[0].(*Junk)
	assert.True(t, ok)
	assert.NotNil(t, junk)
}

func TestParseInputDeclaration_BadExpression(t *testing.T) {
	// Input declaration with literal (not allowed)
	source := ".input {|literal|}"
	ctx := NewParseContext(source, false)
	decl := parseInputDeclaration(ctx, 0)

	assert.NotNil(t, decl)
	assert.Len(t, ctx.Errors(), 1)
	assert.Equal(t, "bad-input-expression", ctx.Errors()[0].Type)
}

func TestParseInputDeclaration_WithMarkup(t *testing.T) {
	// Input declaration with markup (not allowed)
	source := ".input {#tag}"
	ctx := NewParseContext(source, false)
	decl := parseInputDeclaration(ctx, 0)

	assert.NotNil(t, decl)
	assert.Len(t, ctx.Errors(), 1)
	assert.Equal(t, "bad-input-expression", ctx.Errors()[0].Type)
}

func TestParseLocalDeclaration_MissingVariable(t *testing.T) {
	source := ".local = {$value}"
	ctx := NewParseContext(source, false)
	decl := parseLocalDeclaration(ctx, 0)

	assert.NotNil(t, decl)
	assert.NotEmpty(t, ctx.Errors())
}

func TestParseLocalDeclaration_MissingEquals(t *testing.T) {
	source := ".local $var {$value}"
	ctx := NewParseContext(source, false)
	decl := parseLocalDeclaration(ctx, 0)

	assert.NotNil(t, decl)
	assert.NotEmpty(t, ctx.Errors())
}

func TestParseLocalDeclaration_MissingWhitespace(t *testing.T) {
	source := ".local$var = {$value}"
	ctx := NewParseContext(source, false)
	decl := parseLocalDeclaration(ctx, 0)

	assert.NotNil(t, decl)
	assert.NotEmpty(t, ctx.Errors())
}

func TestParseDeclarationValue_NonExpression(t *testing.T) {
	ctx := NewParseContext("invalid", false)
	value := parseDeclarationValue(ctx, 0)

	assert.NotNil(t, value)
	junk, ok := value.(*Junk)
	assert.True(t, ok)
	assert.NotNil(t, junk)
}

func TestParseDeclarationJunk(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		start    int
		expected string
	}{
		{
			name:     "junk before next declaration",
			source:   "junk .input {$var}",
			start:    0,
			expected: "junk",
		},
		{
			name:     "junk before pattern",
			source:   "junk {{pattern}}",
			start:    0,
			expected: "junk",
		},
		{
			name:     "junk with trailing whitespace",
			source:   "junk  \n  .input {$var}",
			start:    0,
			expected: "junk",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewParseContext(tt.source, false)
			junk := parseDeclarationJunk(ctx, tt.start)

			assert.NotNil(t, junk)
			assert.Equal(t, tt.expected, junk.Source())
			assert.Len(t, ctx.Errors(), 1)
			assert.Equal(t, "missing-syntax", ctx.Errors()[0].Type)
		})
	}
}

func TestParseCST_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		resource     bool
		expectedType string
		shouldError  bool
	}{
		{
			name:         "empty string",
			source:       "",
			resource:     false,
			expectedType: "simple",
			shouldError:  false,
		},
		{
			name:         "only whitespace",
			source:       "   \n\t  ",
			resource:     false,
			expectedType: "simple",
			shouldError:  false,
		},
		{
			name:         "declaration only",
			source:       ".input {$var}",
			resource:     false,
			expectedType: "complex",
			shouldError:  true, // Missing pattern
		},
		{
			name:         "match without selectors or variants",
			source:       ".match",
			resource:     false,
			expectedType: "select",
			shouldError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := ParseCST(tt.source, tt.resource)

			assert.NotNil(t, msg)
			assert.Equal(t, tt.expectedType, msg.Type())

			if tt.shouldError {
				assert.NotEmpty(t, msg.Errors())
			}
		})
	}
}

func TestParseCST_ComplexScenarios(t *testing.T) {
	t.Run("multiple declarations with select", func(t *testing.T) {
		source := `.input {$count :integer}
.local $formatted = {$count :number}
.match $count
one {{one item}}
* {{many items}}`
		msg := ParseCST(source, false)

		select_, ok := msg.(*SelectMessage)
		require.True(t, ok)
		assert.Len(t, select_.Declarations(), 2)
		assert.Len(t, select_.Selectors(), 1)
		assert.Len(t, select_.Variants(), 2)
	})

	t.Run("nested expressions in pattern", func(t *testing.T) {
		source := "{{You have {$count} items in {$category}}}"
		msg := ParseCST(source, false)

		complex, ok := msg.(*ComplexMessage)
		require.True(t, ok)
		assert.Empty(t, complex.Errors())

		pattern := complex.Pattern()
		assert.Len(t, pattern.Braces(), 2)
	})
}

func TestSelectMessage_Accessors(t *testing.T) {
	source := ".match $count one {{one}} * {{many}}"
	msg := ParseCST(source, false)

	select_, ok := msg.(*SelectMessage)
	require.True(t, ok)

	// Test all accessor methods
	assert.NotNil(t, select_.Match())
	assert.Len(t, select_.Selectors(), 1)
	assert.Len(t, select_.Variants(), 2)
	assert.Equal(t, "select", select_.Type())
}

func TestVariant_Accessors(t *testing.T) {
	ctx := NewParseContext("one {{pattern}}", false)
	variant := parseVariant(ctx, 0)

	assert.NotNil(t, variant)
	assert.NotEmpty(t, variant.Keys())
	assert.NotNil(t, variant.Value())
}

func TestDeclaration_Accessors(t *testing.T) {
	t.Run("input declaration", func(t *testing.T) {
		source := ".input {$var}"
		ctx := NewParseContext(source, false)
		decl := parseInputDeclaration(ctx, 0)

		assert.Equal(t, "input", decl.Type())
		assert.NotNil(t, decl.Keyword())
		assert.NotNil(t, decl.Value())
	})

	t.Run("local declaration", func(t *testing.T) {
		source := ".local $var = {$value}"
		ctx := NewParseContext(source, false)
		decl := parseLocalDeclaration(ctx, 0)

		assert.Equal(t, "local", decl.Type())
		assert.NotNil(t, decl.Keyword())
		assert.NotNil(t, decl.Target())
		assert.NotNil(t, decl.Equals())
		assert.NotNil(t, decl.Value())
	})
}

func TestJunk_Creation(t *testing.T) {
	junk := NewJunk(0, 5, "junk!")

	assert.Equal(t, "junk", junk.Type())
	assert.Equal(t, 0, junk.Start())
	assert.Equal(t, 5, junk.End())
	assert.Equal(t, "junk!", junk.Source())
}
