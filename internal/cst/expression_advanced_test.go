// Package cst provides comprehensive tests for expression.go advanced cases
package cst

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseExpression_Attributes(t *testing.T) {
	tests := []struct {
		name            string
		source          string
		expectedAttrCnt int
	}{
		{
			name:            "single attribute",
			source:          "{$var :func @attr}",
			expectedAttrCnt: 1,
		},
		{
			name:            "multiple attributes",
			source:          "{$var :func @attr1 @attr2}",
			expectedAttrCnt: 2,
		},
		{
			name:            "attribute with value",
			source:          "{$var :func @attr=val}",
			expectedAttrCnt: 1,
		},
		{
			name:            "attribute without function",
			source:          "{$var @attr}",
			expectedAttrCnt: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewParseContext(tt.source, false)
			expr := parseExpression(ctx, 0)

			assert.NotNil(t, expr)
			assert.Len(t, expr.Attributes(), tt.expectedAttrCnt)
		})
	}
}

func TestParseAttribute(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		start         int
		expectedName  string
		expectedValue string
		hasValue      bool
	}{
		{
			name:         "simple attribute",
			source:       "@attr",
			start:        0,
			expectedName: "attr",
			hasValue:     false,
		},
		{
			name:          "attribute with value",
			source:        "@attr=value",
			start:         0,
			expectedName:  "attr",
			expectedValue: "value",
			hasValue:      true,
		},
		{
			name:          "attribute with quoted value",
			source:        "@attr=|quoted value|",
			start:         0,
			expectedName:  "attr",
			expectedValue: "quoted value",
			hasValue:      true,
		},
		{
			name:         "attribute with namespace",
			source:       "@ns:attr",
			start:        0,
			expectedName: "ns:attr",
			hasValue:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewParseContext(tt.source, false)
			attr := parseAttribute(ctx, tt.start)

			assert.NotNil(t, attr)

			// Build name from identifier parts
			var name string
			for _, part := range attr.Name() {
				name += part.Value()
			}
			assert.Equal(t, tt.expectedName, name)

			if tt.hasValue {
				assert.NotNil(t, attr.Value())
				assert.Equal(t, tt.expectedValue, attr.Value().Value())
				assert.NotNil(t, attr.Equals())
			} else {
				// When no value, both should be nil
				if attr.Value() != nil {
					assert.Nil(t, attr.Equals())
				}
			}
		})
	}
}

func TestParseAttribute_Accessors(t *testing.T) {
	source := "@attr=value rest"
	ctx := NewParseContext(source, false)
	attr := parseAttribute(ctx, 0)

	assert.NotNil(t, attr)
	assert.Greater(t, attr.Start(), -1)
	assert.Greater(t, attr.End(), attr.Start())
	assert.NotNil(t, attr.Open())
	assert.NotEmpty(t, attr.Name())
	assert.NotNil(t, attr.Equals())
	assert.NotNil(t, attr.Value())
}

func TestParseIdentifier_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		start        int
		expectedParts int
		shouldError  bool
	}{
		{
			name:          "simple identifier",
			source:        "name",
			start:         0,
			expectedParts: 1,
		},
		{
			name:          "namespaced identifier",
			source:        "ns:name",
			start:         0,
			expectedParts: 3, // ns, :, name
		},
		{
			name:          "namespace without name",
			source:        "ns:",
			start:         0,
			expectedParts: 2, // ns, :
			shouldError:   true,
		},
		{
			name:          "empty identifier",
			source:        " ",
			start:         0,
			expectedParts: 1,
			shouldError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewParseContext(tt.source, false)
			result := parseIdentifier(ctx, tt.start)

			assert.NotNil(t, result)
			assert.Len(t, result.Parts, tt.expectedParts)

			if tt.shouldError {
				assert.NotEmpty(t, ctx.Errors())
			}
		})
	}
}

func TestParseOption_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		start       int
		shouldError bool
	}{
		{
			name:        "option with literal value",
			source:      "key=value",
			start:       0,
			shouldError: false,
		},
		{
			name:        "option with variable value",
			source:      "key=$var",
			start:       0,
			shouldError: false,
		},
		{
			name:        "option missing equals",
			source:      "key value",
			start:       0,
			shouldError: true,
		},
		{
			name:        "option with quoted value",
			source:      "key=|quoted|",
			start:       0,
			shouldError: false,
		},
		{
			name:        "option with namespace",
			source:      "ns:key=val",
			start:       0,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewParseContext(tt.source, false)
			opt := parseOption(ctx, tt.start)

			assert.NotNil(t, opt)
			assert.NotNil(t, opt.Value())

			if tt.shouldError {
				assert.NotEmpty(t, ctx.Errors())
			} else {
				assert.NotNil(t, opt.Equals())
			}
		})
	}
}

func TestParseOption_Accessors(t *testing.T) {
	source := "key=value"
	ctx := NewParseContext(source, false)
	opt := parseOption(ctx, 0)

	assert.NotNil(t, opt)
	assert.Greater(t, opt.Start(), -1)
	assert.Greater(t, opt.End(), opt.Start())
	assert.NotEmpty(t, opt.Name())
	assert.NotNil(t, opt.Equals())
	assert.NotNil(t, opt.Value())
}

func TestParseFunctionRefOrMarkup_Function(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		start        int
		nodeType     string
		expectedFunc string
		optionCount  int
	}{
		{
			name:         "simple function",
			source:       ":integer",
			start:        0,
			nodeType:     "function",
			expectedFunc: "integer",
			optionCount:  0,
		},
		{
			name:         "function with option",
			source:       ":number minimumFractionDigits=2",
			start:        0,
			nodeType:     "function",
			expectedFunc: "number",
			optionCount:  1,
		},
		{
			name:         "function with multiple options",
			source:       ":datetime year=numeric month=long",
			start:        0,
			nodeType:     "function",
			expectedFunc: "datetime",
			optionCount:  2,
		},
		{
			name:         "namespaced function",
			source:       ":ns:func",
			start:        0,
			nodeType:     "function",
			expectedFunc: "ns:func",
			optionCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewParseContext(tt.source, false)
			result := parseFunctionRefOrMarkup(ctx, tt.start, tt.nodeType)

			assert.NotNil(t, result)

			if funcRef, ok := result.(*FunctionRef); ok {
				assert.Equal(t, "function", funcRef.Type())

				var funcName string
				for _, part := range funcRef.Name() {
					funcName += part.Value()
				}
				assert.Equal(t, tt.expectedFunc, funcName)
				assert.Len(t, funcRef.Options(), tt.optionCount)
			} else {
				t.Fatalf("Expected FunctionRef, got %T", result)
			}
		})
	}
}

func TestParseFunctionRefOrMarkup_Markup(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		start       int
		nodeType    string
		expectedTag string
		hasClose    bool
	}{
		{
			name:        "open markup",
			source:      "#tag",
			start:       0,
			nodeType:    "markup",
			expectedTag: "tag",
			hasClose:    false,
		},
		{
			name:        "close markup",
			source:      "/tag",
			start:       0,
			nodeType:    "markup",
			expectedTag: "tag",
			hasClose:    false,
		},
		{
			name:        "self-closing markup",
			source:      "#tag /",
			start:       0,
			nodeType:    "markup",
			expectedTag: "tag",
			hasClose:    true,
		},
		{
			name:        "markup with options",
			source:      "#tag attr=value",
			start:       0,
			nodeType:    "markup",
			expectedTag: "tag",
			hasClose:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewParseContext(tt.source, false)
			result := parseFunctionRefOrMarkup(ctx, tt.start, tt.nodeType)

			assert.NotNil(t, result)

			if markup, ok := result.(*Markup); ok {
				assert.Equal(t, "markup", markup.Type())

				var tagName string
				for _, part := range markup.Name() {
					tagName += part.Value()
				}
				assert.Equal(t, tt.expectedTag, tagName)

				if tt.hasClose {
					assert.NotNil(t, markup.Close())
				}
			} else {
				t.Fatalf("Expected Markup, got %T", result)
			}
		})
	}
}

func TestParseFunctionRefOrMarkup_DuplicateOptions(t *testing.T) {
	source := ":func opt=1 opt=2"
	ctx := NewParseContext(source, false)
	result := parseFunctionRefOrMarkup(ctx, 0, "function")

	assert.NotNil(t, result)
	// Should have error for duplicate option name
	assert.NotEmpty(t, ctx.Errors())
	assert.Equal(t, "duplicate-option-name", ctx.Errors()[0].Type)
}

func TestParseFunctionRefOrMarkup_MissingWhitespace(t *testing.T) {
	source := ":funcopt=value"
	ctx := NewParseContext(source, false)
	result := parseFunctionRefOrMarkup(ctx, 0, "function")

	assert.NotNil(t, result)
	// Should have error for missing whitespace before option
	assert.NotEmpty(t, ctx.Errors())
}

func TestParseExpression_Markup(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		hasArg   bool
		hasError bool
	}{
		{
			name:     "open markup",
			source:   "{#tag}",
			hasArg:   false,
			hasError: false,
		},
		{
			name:     "close markup",
			source:   "{/tag}",
			hasArg:   false,
			hasError: false,
		},
		{
			name:     "markup with arg (error)",
			source:   "{$var #tag}",
			hasArg:   true,
			hasError: true,
		},
		{
			name:     "self-closing markup",
			source:   "{#tag /}",
			hasArg:   false,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewParseContext(tt.source, false)
			expr := parseExpression(ctx, 0)

			assert.NotNil(t, expr)
			assert.NotNil(t, expr.Markup())

			if tt.hasError {
				assert.NotEmpty(t, ctx.Errors())
			}
		})
	}
}

func TestParseExpression_ExtraContent(t *testing.T) {
	source := "{$var extra content}"
	ctx := NewParseContext(source, false)
	expr := parseExpression(ctx, 0)

	assert.NotNil(t, expr)
	assert.NotEmpty(t, ctx.Errors())
	assert.Equal(t, "extra-content", ctx.Errors()[0].Type)
}

func TestParseExpression_MissingClosingBrace(t *testing.T) {
	source := "{$var"
	ctx := NewParseContext(source, false)
	expr := parseExpression(ctx, 0)

	assert.NotNil(t, expr)
	assert.NotEmpty(t, ctx.Errors())
	assert.Equal(t, "missing-syntax", ctx.Errors()[0].Type)
}

func TestGetOptionName(t *testing.T) {
	tests := []struct {
		name     string
		parts    Identifier
		expected string
	}{
		{
			name: "simple name",
			parts: Identifier{
				NewSyntax(0, 3, "key"),
			},
			expected: "key",
		},
		{
			name: "namespaced name",
			parts: Identifier{
				NewSyntax(0, 2, "ns"),
				NewSyntax(2, 3, ":"),
				NewSyntax(3, 6, "key"),
			},
			expected: "ns:key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getOptionName(tt.parts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFunctionRef_Accessors(t *testing.T) {
	open := NewSyntax(0, 1, ":")
	name := Identifier{NewSyntax(1, 5, "func")}
	opt := NewOption(6, 12, Identifier{NewSyntax(6, 9, "key")}, nil, NewLiteral(10, 13, false, nil, "val", nil))
	options := []Option{*opt}

	funcRef := NewFunctionRef(0, 12, open, name, options)

	assert.Equal(t, "function", funcRef.Type())
	assert.Equal(t, 0, funcRef.Start())
	assert.Equal(t, 12, funcRef.End())
	assert.NotNil(t, funcRef.Open())
	assert.Equal(t, name, funcRef.Name())
	assert.Len(t, funcRef.Options(), 1)
}

func TestMarkup_Accessors(t *testing.T) {
	open := NewSyntax(0, 1, "#")
	name := Identifier{NewSyntax(1, 4, "tag")}
	close := NewSyntax(10, 11, "/")

	markup := NewMarkup(0, 11, open, name, []Option{}, &close)

	assert.Equal(t, "markup", markup.Type())
	assert.Equal(t, 0, markup.Start())
	assert.Equal(t, 11, markup.End())
	assert.NotNil(t, markup.Open())
	assert.Equal(t, name, markup.Name())
	assert.Empty(t, markup.Options())
	assert.NotNil(t, markup.Close())
}

func TestIdentifier_Methods(t *testing.T) {
	t.Run("String method", func(t *testing.T) {
		id := Identifier{
			NewSyntax(0, 2, "ns"),
			NewSyntax(2, 3, ":"),
			NewSyntax(3, 7, "name"),
		}

		result := id.String()
		assert.Equal(t, "ns:name", result)
	})

	t.Run("Namespace method", func(t *testing.T) {
		id := Identifier{
			NewSyntax(0, 2, "ns"),
			NewSyntax(2, 3, ":"),
			NewSyntax(3, 7, "name"),
		}

		ns := id.Namespace()
		assert.NotNil(t, ns)
		assert.Equal(t, "ns", ns.Value())
	})

	t.Run("Namespace method - no namespace", func(t *testing.T) {
		id := Identifier{
			NewSyntax(0, 4, "name"),
		}

		ns := id.Namespace()
		assert.Nil(t, ns)
	})

	t.Run("Name method", func(t *testing.T) {
		id := Identifier{
			NewSyntax(0, 2, "ns"),
			NewSyntax(2, 3, ":"),
			NewSyntax(3, 7, "name"),
		}

		name := id.Name()
		assert.NotNil(t, name)
		assert.Equal(t, "name", name.Value())
	})

	t.Run("Separator method", func(t *testing.T) {
		id := Identifier{
			NewSyntax(0, 2, "ns"),
			NewSyntax(2, 3, ":"),
			NewSyntax(3, 7, "name"),
		}

		sep := id.Separator()
		assert.NotNil(t, sep)
		assert.Equal(t, ":", sep.Value())
	})

	t.Run("Separator method - no separator", func(t *testing.T) {
		id := Identifier{
			NewSyntax(0, 4, "name"),
		}

		sep := id.Separator()
		assert.Nil(t, sep)
	})
}

func TestExpression_Accessors(t *testing.T) {
	open := NewSyntax(0, 1, "{")
	close := NewSyntax(10, 11, "}")
	braces := []Syntax{open, close}

	varOpen := NewSyntax(1, 2, "$")
	arg := NewVariableRef(1, 5, varOpen, "var")

	funcOpen := NewSyntax(6, 7, ":")
	funcName := Identifier{NewSyntax(7, 10, "int")}
	funcRef := NewFunctionRef(6, 10, funcOpen, funcName, []Option{})

	expr := NewExpression(0, 11, braces, arg, funcRef, nil, []Attribute{})

	assert.Equal(t, "expression", expr.Type())
	assert.Equal(t, 0, expr.Start())
	assert.Equal(t, 11, expr.End())
	assert.Len(t, expr.Braces(), 2)
	assert.NotNil(t, expr.Arg())
	assert.NotNil(t, expr.FunctionRef())
	assert.Nil(t, expr.Markup())
	assert.Empty(t, expr.Attributes())
}
