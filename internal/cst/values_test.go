// Package cst provides comprehensive tests for values.go
package cst

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseText_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		start         int
		resource      bool
		expectedValue string
		expectedEnd   int
	}{
		{
			name:          "simple text",
			source:        "Hello, world!",
			start:         0,
			resource:      false,
			expectedValue: "Hello, world!",
			expectedEnd:   13,
		},
		{
			name:          "text with escape",
			source:        "Hello\\{world",
			start:         0,
			resource:      false,
			expectedValue: "Hello{world",
			expectedEnd:   12,
		},
		{
			name:          "text stops at {",
			source:        "Hello{expression}",
			start:         0,
			resource:      false,
			expectedValue: "Hello",
			expectedEnd:   5,
		},
		{
			name:          "text stops at }",
			source:        "Hello}",
			start:         0,
			resource:      false,
			expectedValue: "Hello",
			expectedEnd:   5,
		},
		{
			name:          "resource mode with newline",
			source:        "Line1\n  \t Line2",
			start:         0,
			resource:      true,
			expectedValue: "Line1\nLine2",
			expectedEnd:   15,
		},
		{
			name:          "non-resource mode with newline",
			source:        "Line1\n  \t Line2",
			start:         0,
			resource:      false,
			expectedValue: "Line1\n  \t Line2",
			expectedEnd:   15,
		},
		{
			name:          "empty text",
			source:        "{expr}",
			start:         0,
			resource:      false,
			expectedValue: "",
			expectedEnd:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewParseContext(tt.source, tt.resource)
			text := ParseText(ctx, tt.start)

			assert.NotNil(t, text)
			assert.Equal(t, tt.expectedValue, text.Value())
			assert.Equal(t, tt.expectedEnd, text.End())
		})
	}
}

func TestParseSimpleText_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		start         int
		resource      bool
		expectedValue string
		expectedEnd   int
	}{
		{
			name:          "text with {{ escape",
			source:        "Hello{{world",
			start:         0,
			resource:      false,
			expectedValue: "Hello{world",
			expectedEnd:   12,
		},
		{
			name:          "text with }} escape",
			source:        "Hello}}world",
			start:         0,
			resource:      false,
			expectedValue: "Hello}world",
			expectedEnd:   12,
		},
		{
			name:          "multiple {{ escapes",
			source:        "{{{{test",
			start:         0,
			resource:      false,
			expectedValue: "{{test",
			expectedEnd:   8,
		},
		{
			name:          "mixed escapes",
			source:        "a{{b}}c",
			start:         0,
			resource:      false,
			expectedValue: "a{b}c",
			expectedEnd:   7,
		},
		{
			name:          "backslash escape",
			source:        "test\\{value",
			start:         0,
			resource:      false,
			expectedValue: "test{value",
			expectedEnd:   11,
		},
		{
			name:          "resource mode newline",
			source:        "Line1\n  \tLine2",
			start:         0,
			resource:      true,
			expectedValue: "Line1\nLine2",
			expectedEnd:   14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewParseContext(tt.source, tt.resource)
			text := ParseSimpleText(ctx, tt.start)

			assert.NotNil(t, text)
			assert.Equal(t, tt.expectedValue, text.Value())
			assert.Equal(t, tt.expectedEnd, text.End())
		})
	}
}

func TestParseLiteral_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		start       int
		required    bool
		shouldBeNil bool
		expectedVal string
		quoted      bool
	}{
		{
			name:        "unquoted literal",
			source:      "hello",
			start:       0,
			required:    true,
			expectedVal: "hello",
			quoted:      false,
		},
		{
			name:        "quoted literal",
			source:      "|hello world|",
			start:       0,
			required:    true,
			expectedVal: "hello world",
			quoted:      true,
		},
		{
			name:        "empty unquoted not required",
			source:      " ",
			start:       0,
			required:    false,
			shouldBeNil: true,
		},
		{
			name:        "empty unquoted required",
			source:      " ",
			start:       0,
			required:    true,
			expectedVal: "",
			quoted:      false,
		},
		{
			name:        "numeric literal",
			source:      "123",
			start:       0,
			required:    true,
			expectedVal: "123",
			quoted:      false,
		},
		{
			name:        "negative numeric",
			source:      "-456",
			start:       0,
			required:    true,
			expectedVal: "-456",
			quoted:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewParseContext(tt.source, false)
			literal := ParseLiteral(ctx, tt.start, tt.required)

			if tt.shouldBeNil {
				assert.Nil(t, literal)
			} else {
				assert.NotNil(t, literal)
				assert.Equal(t, tt.expectedVal, literal.Value())
				assert.Equal(t, tt.quoted, literal.Quoted())
			}
		})
	}
}

func TestParseQuotedLiteral_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		start         int
		resource      bool
		expectedValue string
		shouldError   bool
	}{
		{
			name:          "simple quoted",
			source:        "|hello|",
			start:         0,
			resource:      false,
			expectedValue: "hello",
		},
		{
			name:          "quoted with spaces",
			source:        "|hello world|",
			start:         0,
			resource:      false,
			expectedValue: "hello world",
		},
		{
			name:          "quoted with escapes",
			source:        "|hello\\|world|",
			start:         0,
			resource:      false,
			expectedValue: "hello|world",
		},
		{
			name:          "missing closing quote",
			source:        "|unclosed",
			start:         0,
			resource:      false,
			expectedValue: "unclosed",
			shouldError:   true,
		},
		{
			name:          "quoted with newline resource mode",
			source:        "|line1\n  \tline2|",
			start:         0,
			resource:      true,
			expectedValue: "line1\nline2",
		},
		{
			name:          "empty quoted literal",
			source:        "||",
			start:         0,
			resource:      false,
			expectedValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewParseContext(tt.source, tt.resource)
			literal := parseQuotedLiteral(ctx, tt.start)

			assert.NotNil(t, literal)
			assert.Equal(t, tt.expectedValue, literal.Value())
			assert.True(t, literal.Quoted())

			if tt.shouldError {
				assert.NotEmpty(t, ctx.Errors())
			}
		})
	}
}

func TestParseVariable_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		start        int
		expectedName string
		shouldError  bool
	}{
		{
			name:         "simple variable",
			source:       "$name",
			start:        0,
			expectedName: "name",
		},
		{
			name:         "variable with underscore",
			source:       "$user_name",
			start:        0,
			expectedName: "user_name",
		},
		{
			name:         "variable with numbers",
			source:       "$var123",
			start:        0,
			expectedName: "var123",
		},
		{
			name:         "empty variable name",
			source:       "$ ",
			start:        0,
			expectedName: "",
			shouldError:  true,
		},
		{
			name:         "variable at offset",
			source:       "prefix$name",
			start:        6,
			expectedName: "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewParseContext(tt.source, false)
			variable := ParseVariable(ctx, tt.start)

			assert.NotNil(t, variable)
			assert.Equal(t, tt.expectedName, variable.Name())

			if tt.shouldError {
				assert.NotEmpty(t, ctx.Errors())
			}
		})
	}
}

func TestParseEscape_AllSequences(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		start         int
		resource      bool
		expectedValue string
		expectedLen   int
		shouldBeNil   bool
	}{
		// Basic escapes (always valid)
		{
			name:          "escape backslash",
			source:        "\\\\",
			start:         0,
			resource:      false,
			expectedValue: "\\",
			expectedLen:   1,
		},
		{
			name:          "escape left brace",
			source:        "\\{",
			start:         0,
			resource:      false,
			expectedValue: "{",
			expectedLen:   1,
		},
		{
			name:          "escape pipe",
			source:        "\\|",
			start:         0,
			resource:      false,
			expectedValue: "|",
			expectedLen:   1,
		},
		{
			name:          "escape right brace",
			source:        "\\}",
			start:         0,
			resource:      false,
			expectedValue: "}",
			expectedLen:   1,
		},
		// Resource-only escapes
		{
			name:          "escape tab (resource)",
			source:        "\\t",
			start:         0,
			resource:      true,
			expectedValue: "\t",
			expectedLen:   1,
		},
		{
			name:          "escape tab (non-resource)",
			source:        "\\t",
			start:         0,
			resource:      false,
			shouldBeNil:   true,
		},
		{
			name:          "escape space (resource)",
			source:        "\\ ",
			start:         0,
			resource:      true,
			expectedValue: " ",
			expectedLen:   1,
		},
		{
			name:          "escape newline",
			source:        "\\n",
			start:         0,
			resource:      true,
			expectedValue: "\n",
			expectedLen:   1,
		},
		{
			name:          "escape carriage return",
			source:        "\\r",
			start:         0,
			resource:      true,
			expectedValue: "\r",
			expectedLen:   1,
		},
		{
			name:          "escape \\t literal",
			source:        "\\\t",
			start:         0,
			resource:      true,
			expectedValue: "\t",
			expectedLen:   1,
		},
		// Hex escapes
		{
			name:          "hex escape \\x",
			source:        "\\x41",
			start:         0,
			resource:      true,
			expectedValue: "A",
			expectedLen:   3,
		},
		{
			name:          "unicode escape \\u",
			source:        "\\u0041",
			start:         0,
			resource:      true,
			expectedValue: "A",
			expectedLen:   5,
		},
		{
			name:          "unicode escape \\U",
			source:        "\\U000041",
			start:         0,
			resource:      true,
			expectedValue: "A",
			expectedLen:   7,
		},
		// Invalid escapes
		{
			name:        "invalid escape sequence",
			source:      "\\z",
			start:       0,
			resource:    false,
			shouldBeNil: true,
		},
		{
			name:        "escape at end of string",
			source:      "\\",
			start:       0,
			resource:    false,
			shouldBeNil: true,
		},
		{
			name:        "invalid hex escape",
			source:      "\\xZZ",
			start:       0,
			resource:    true,
			shouldBeNil: true,
		},
		{
			name:        "incomplete hex escape",
			source:      "\\x4",
			start:       0,
			resource:    true,
			shouldBeNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewParseContext(tt.source, tt.resource)
			result := parseEscape(ctx, tt.start)

			if tt.shouldBeNil {
				assert.Nil(t, result)
				assert.NotEmpty(t, ctx.Errors())
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedValue, result.Value)
				assert.Equal(t, tt.expectedLen, result.Length)
			}
		})
	}
}

func TestParseHexEscape_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		start         int
		hexLen        int
		expectedValue string
		shouldBeNil   bool
	}{
		{
			name:          "valid 2-digit hex",
			source:        "\\x41rest",
			start:         0,
			hexLen:        2,
			expectedValue: "A",
		},
		{
			name:          "valid 4-digit hex",
			source:        "\\u0041rest",
			start:         0,
			hexLen:        4,
			expectedValue: "A",
		},
		{
			name:          "valid 6-digit hex",
			source:        "\\U000041rest",
			start:         0,
			hexLen:        6,
			expectedValue: "A",
		},
		{
			name:        "incomplete hex",
			source:      "\\x4",
			start:       0,
			hexLen:      2,
			shouldBeNil: true,
		},
		{
			name:        "invalid hex characters",
			source:      "\\xGG",
			start:       0,
			hexLen:      2,
			shouldBeNil: true,
		},
		{
			name:        "hex beyond string length",
			source:      "\\x",
			start:       0,
			hexLen:      2,
			shouldBeNil: true,
		},
		{
			name:          "hex with lowercase",
			source:        "\\x61",
			start:         0,
			hexLen:        2,
			expectedValue: "a",
		},
		{
			name:          "unicode emoji",
			source:        "\\U01F600",
			start:         0,
			hexLen:        6,
			expectedValue: "\U0001F600",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewParseContext(tt.source, true)
			result := parseHexEscape(ctx, tt.start, tt.hexLen)

			if tt.shouldBeNil {
				assert.Nil(t, result)
				assert.NotEmpty(t, ctx.Errors())
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedValue, result.Value)
				assert.Equal(t, 1+tt.hexLen, result.Length)
			}
		})
	}
}

func TestLiteral_Accessors(t *testing.T) {
	t.Run("unquoted literal accessors", func(t *testing.T) {
		literal := NewLiteral(0, 5, false, nil, "hello", nil)

		assert.Equal(t, "literal", literal.Type())
		assert.Equal(t, 0, literal.Start())
		assert.Equal(t, 5, literal.End())
		assert.False(t, literal.Quoted())
		assert.Nil(t, literal.Open())
		assert.Equal(t, "hello", literal.Value())
		assert.Nil(t, literal.Close())
	})

	t.Run("quoted literal accessors", func(t *testing.T) {
		open := NewSyntax(0, 1, "|")
		close := NewSyntax(6, 7, "|")
		literal := NewLiteral(0, 7, true, &open, "hello", &close)

		assert.True(t, literal.Quoted())
		assert.NotNil(t, literal.Open())
		assert.NotNil(t, literal.Close())
	})
}

func TestVariableRef_Accessors(t *testing.T) {
	open := NewSyntax(0, 1, "$")
	varRef := NewVariableRef(0, 5, open, "name")

	assert.Equal(t, "variable", varRef.Type())
	assert.Equal(t, 0, varRef.Start())
	assert.Equal(t, 5, varRef.End())
	assert.NotNil(t, varRef.Open())
	assert.Equal(t, "name", varRef.Name())
}

func TestText_Accessor(t *testing.T) {
	text := NewText(0, 5, "hello")

	assert.Equal(t, "text", text.Type())
	assert.Equal(t, 0, text.Start())
	assert.Equal(t, 5, text.End())
	assert.Equal(t, "hello", text.Value())
}

func TestParseText_StopsAtBraces(t *testing.T) {
	t.Run("stops at opening brace", func(t *testing.T) {
		ctx := NewParseContext("text{expr}", false)
		text := ParseText(ctx, 0)

		assert.Equal(t, "text", text.Value())
		assert.Equal(t, 4, text.End())
	})

	t.Run("stops at closing brace", func(t *testing.T) {
		ctx := NewParseContext("text}end", false)
		text := ParseText(ctx, 0)

		assert.Equal(t, "text", text.Value())
		assert.Equal(t, 4, text.End())
	})
}
