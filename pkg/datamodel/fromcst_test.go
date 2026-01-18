package datamodel

import (
	"testing"

	"github.com/kaptinlin/messageformat-go/internal/cst"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromCSTSimpleMessage(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantType      string
		wantPattern   int // number of pattern elements
		wantError     bool
		errorContains string
	}{
		{
			name:        "simple text message",
			input:       "Hello, world!",
			wantType:    "message",
			wantPattern: 1,
			wantError:   false,
		},
		{
			name:        "message with expression",
			input:       "Hello {name}!",
			wantType:    "message",
			wantPattern: 3, // "Hello ", expression, "!"
			wantError:   false,
		},
		{
			name:        "message with literal expression",
			input:       "Price: {|42|}",
			wantType:    "message",
			wantPattern: 2, // "Price: ", expression
			wantError:   false,
		},
		{
			name:        "empty message",
			input:       "",
			wantType:    "message",
			wantPattern: 0,
			wantError:   false,
		},
		{
			name:        "message with function",
			input:       "{count :number}",
			wantType:    "message",
			wantPattern: 1,
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse CST
			cstMsg := cst.ParseCST(tt.input, false)

			// Convert to data model
			msg, err := FromCST(cstMsg)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, msg)

			assert.Equal(t, tt.wantType, msg.Type())

			if patternMsg, ok := msg.(*PatternMessage); ok {
				assert.Len(t, patternMsg.Pattern().Elements(), tt.wantPattern)
			}
		})
	}
}

func TestFromCSTComplexMessage(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantType      string
		wantDecls     int
		wantError     bool
		errorContains string
	}{
		{
			name:      "input declaration",
			input:     ".input {$count :integer}\n{{Hello}}",
			wantType:  "message",
			wantDecls: 1,
			wantError: false,
		},
		{
			name:      "local declaration",
			input:     ".local $x = {|42|}\n{{Value: {$x}}}",
			wantType:  "message",
			wantDecls: 1,
			wantError: false,
		},
		{
			name:      "multiple declarations",
			input:     ".input {$count :integer}\n.local $formatted = {$count :number}\n{{Count: {$formatted}}}",
			wantType:  "message",
			wantDecls: 2,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cstMsg := cst.ParseCST(tt.input, false)
			msg, err := FromCST(cstMsg)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, msg)

			assert.Equal(t, tt.wantType, msg.Type())
			assert.Len(t, msg.Declarations(), tt.wantDecls)
		})
	}
}

func TestFromCSTSelectMessage(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantSelectors int
		wantVariants  int
		wantError     bool
		errorContains string
	}{
		{
			name: "simple select",
			input: `.match $count
one {{One item}}
* {{Many items}}`,
			wantSelectors: 1,
			wantVariants:  2,
			wantError:     false,
		},
		{
			name: "multi-selector",
			input: `.match $count $gender
one male {{He has one item}}
one female {{She has one item}}
* * {{They have items}}`,
			wantSelectors: 2,
			wantVariants:  3,
			wantError:     false,
		},
		{
			name: "select with declarations",
			input: `.input {$count :integer}
.match $count
one {{One}}
* {{Other}}`,
			wantSelectors: 1,
			wantVariants:  2,
			wantError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cstMsg := cst.ParseCST(tt.input, false)
			msg, err := FromCST(cstMsg)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, msg)

			assert.Equal(t, "select", msg.Type())

			selectMsg, ok := msg.(*SelectMessage)
			require.True(t, ok)

			assert.Len(t, selectMsg.Selectors(), tt.wantSelectors)
			assert.Len(t, selectMsg.Variants(), tt.wantVariants)
		})
	}
}

func TestFromCSTExpressions(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		checkFunc  func(t *testing.T, msg Message)
		wantError  bool
	}{
		{
			name:  "variable reference",
			input: "Hello {$name}",
			checkFunc: func(t *testing.T, msg Message) {
				patternMsg := msg.(*PatternMessage)
				elem := patternMsg.Pattern().Elements()[1]
				expr, ok := elem.(*Expression)
				require.True(t, ok)
				assert.True(t, IsVariableRef(expr.Arg()))
			},
			wantError: false,
		},
		{
			name:  "literal argument",
			input: "Price: {|42| :number}",
			checkFunc: func(t *testing.T, msg Message) {
				patternMsg := msg.(*PatternMessage)
				elem := patternMsg.Pattern().Elements()[1]
				expr, ok := elem.(*Expression)
				require.True(t, ok)
				assert.True(t, IsLiteral(expr.Arg()))
			},
			wantError: false,
		},
		{
			name:  "function with options",
			input: "{$price :number style=currency currency=USD}",
			checkFunc: func(t *testing.T, msg Message) {
				patternMsg := msg.(*PatternMessage)
				elem := patternMsg.Pattern().Elements()[0]
				expr, ok := elem.(*Expression)
				require.True(t, ok)
				require.NotNil(t, expr.FunctionRef())
				assert.Equal(t, "number", expr.FunctionRef().Name())
				assert.NotNil(t, expr.FunctionRef().Options())
			},
			wantError: false,
		},
		{
			name:  "expression with attributes",
			input: "{$value @id=test}",
			checkFunc: func(t *testing.T, msg Message) {
				patternMsg := msg.(*PatternMessage)
				elem := patternMsg.Pattern().Elements()[0]
				expr, ok := elem.(*Expression)
				require.True(t, ok)
				assert.NotNil(t, expr.Attributes())
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cstMsg := cst.ParseCST(tt.input, false)
			msg, err := FromCST(cstMsg)

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, msg)

			if tt.checkFunc != nil {
				tt.checkFunc(t, msg)
			}
		})
	}
}

func TestFromCSTMarkup(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		checkFunc func(t *testing.T, msg Message)
		wantError bool
	}{
		{
			name:  "open markup",
			input: "Text with {#b}bold{/b} markup",
			checkFunc: func(t *testing.T, msg Message) {
				patternMsg := msg.(*PatternMessage)
				elems := patternMsg.Pattern().Elements()

				// Check for markup elements
				hasOpenMarkup := false
				hasCloseMarkup := false
				for _, elem := range elems {
					if markup, ok := elem.(*Markup); ok {
						if markup.Kind() == "open" && markup.Name() == "b" {
							hasOpenMarkup = true
						}
						if markup.Kind() == "close" && markup.Name() == "b" {
							hasCloseMarkup = true
						}
					}
				}

				assert.True(t, hasOpenMarkup, "Should have open markup")
				assert.True(t, hasCloseMarkup, "Should have close markup")
			},
			wantError: false,
		},
		{
			name:  "standalone markup",
			input: "Line 1{#br /}Line 2",
			checkFunc: func(t *testing.T, msg Message) {
				patternMsg := msg.(*PatternMessage)
				elems := patternMsg.Pattern().Elements()

				hasStandaloneMarkup := false
				for _, elem := range elems {
					if markup, ok := elem.(*Markup); ok {
						if markup.Kind() == "standalone" && markup.Name() == "br" {
							hasStandaloneMarkup = true
						}
					}
				}

				assert.True(t, hasStandaloneMarkup, "Should have standalone markup")
			},
			wantError: false,
		},
		{
			name:  "markup with options",
			input: "{#link href=|https://example.com|}Click{/link}",
			checkFunc: func(t *testing.T, msg Message) {
				patternMsg := msg.(*PatternMessage)
				elems := patternMsg.Pattern().Elements()

				for _, elem := range elems {
					if markup, ok := elem.(*Markup); ok {
						if markup.Kind() == "open" && markup.Name() == "link" {
							assert.NotNil(t, markup.Options())
							return
						}
					}
				}

				t.Error("Should have markup with options")
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cstMsg := cst.ParseCST(tt.input, false)
			msg, err := FromCST(cstMsg)

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, msg)

			if tt.checkFunc != nil {
				tt.checkFunc(t, msg)
			}
		})
	}
}

func TestFromCSTVariants(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		checkFunc func(t *testing.T, msg Message)
		wantError bool
	}{
		{
			name: "literal keys",
			input: `.match $count
one {{One}}
two {{Two}}
* {{Other}}`,
			checkFunc: func(t *testing.T, msg Message) {
				selectMsg := msg.(*SelectMessage)
				require.Len(t, selectMsg.Variants(), 3)

				variant := selectMsg.Variants()[0]
				require.Len(t, variant.Keys(), 1)
				assert.True(t, IsLiteral(variant.Keys()[0]))

				lastVariant := selectMsg.Variants()[2]
				assert.True(t, IsCatchallKey(lastVariant.Keys()[0]))
			},
			wantError: false,
		},
		{
			name: "multi-key variants",
			input: `.match $count $gender
one male {{He has one}}
* * {{Default}}`,
			checkFunc: func(t *testing.T, msg Message) {
				selectMsg := msg.(*SelectMessage)
				require.Len(t, selectMsg.Variants(), 2)

				variant := selectMsg.Variants()[0]
				require.Len(t, variant.Keys(), 2)
				assert.True(t, IsLiteral(variant.Keys()[0]))
				assert.True(t, IsLiteral(variant.Keys()[1]))

				lastVariant := selectMsg.Variants()[1]
				assert.True(t, IsCatchallKey(lastVariant.Keys()[0]))
				assert.True(t, IsCatchallKey(lastVariant.Keys()[1]))
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cstMsg := cst.ParseCST(tt.input, false)
			msg, err := FromCST(cstMsg)

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, msg)

			if tt.checkFunc != nil {
				tt.checkFunc(t, msg)
			}
		})
	}
}

func TestFromCSTDeclarations(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		checkFunc func(t *testing.T, msg Message)
		wantError bool
	}{
		{
			name:  "input declaration with function",
			input: ".input {$count :integer}\n{{Value: {$count}}}",
			checkFunc: func(t *testing.T, msg Message) {
				require.Len(t, msg.Declarations(), 1)

				decl := msg.Declarations()[0]
				assert.Equal(t, "input", decl.Type())
				assert.Equal(t, "count", decl.Name())

				inputDecl, ok := decl.(*InputDeclaration)
				require.True(t, ok)
				require.NotNil(t, inputDecl.value)
				assert.NotNil(t, inputDecl.value.FunctionRef())
			},
			wantError: false,
		},
		{
			name:  "local declaration with expression",
			input: ".local $formatted = {$count :number}\n{{Result: {$formatted}}}",
			checkFunc: func(t *testing.T, msg Message) {
				require.Len(t, msg.Declarations(), 1)

				decl := msg.Declarations()[0]
				assert.Equal(t, "local", decl.Type())
				assert.Equal(t, "formatted", decl.Name())

				localDecl, ok := decl.(*LocalDeclaration)
				require.True(t, ok)
				require.NotNil(t, localDecl.value)
				assert.True(t, IsVariableRef(localDecl.value.Arg()))
			},
			wantError: false,
		},
		{
			name:  "local declaration with literal",
			input: ".local $x = {|42|}\n{{X = {$x}}}",
			checkFunc: func(t *testing.T, msg Message) {
				require.Len(t, msg.Declarations(), 1)

				decl := msg.Declarations()[0]
				localDecl, ok := decl.(*LocalDeclaration)
				require.True(t, ok)
				require.NotNil(t, localDecl.value)
				assert.True(t, IsLiteral(localDecl.value.Arg()))
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cstMsg := cst.ParseCST(tt.input, false)
			msg, err := FromCST(cstMsg)

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, msg)

			if tt.checkFunc != nil {
				tt.checkFunc(t, msg)
			}
		})
	}
}

func TestFromCSTErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantError     bool
		errorContains string
	}{
		{
			name:      "valid message",
			input:     "Hello, world!",
			wantError: false,
		},
		{
			name:          "CST with errors is propagated",
			input:         "{unclosed",
			wantError:     true,
			errorContains: "parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cstMsg := cst.ParseCST(tt.input, false)
			msg, err := FromCST(cstMsg)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, msg)
		})
	}
}

func TestFromCSTComplexScenarios(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		checkFunc func(t *testing.T, msg Message)
	}{
		{
			name: "nested expressions in variants",
			input: `.match $count
one {{You have {$count} item}}
* {{You have {$count} items}}`,
			checkFunc: func(t *testing.T, msg Message) {
				selectMsg := msg.(*SelectMessage)
				require.Len(t, selectMsg.Variants(), 2)

				// Check first variant pattern has expressions
				variant := selectMsg.Variants()[0]
				pattern := variant.Value()
				hasExpression := false
				for _, elem := range pattern.Elements() {
					if IsExpression(elem) {
						hasExpression = true
						break
					}
				}
				assert.True(t, hasExpression, "Variant should contain expression")
			},
		},
		{
			name: "multiple declarations with dependencies",
			input: `.input {$price :number}
.local $tax = {$price :number style=percent}
.local $total = {$price}
{{Total: {$total} (tax: {$tax})}}`,
			checkFunc: func(t *testing.T, msg Message) {
				assert.Len(t, msg.Declarations(), 3)

				// Check types
				assert.Equal(t, "input", msg.Declarations()[0].Type())
				assert.Equal(t, "local", msg.Declarations()[1].Type())
				assert.Equal(t, "local", msg.Declarations()[2].Type())

				// Check names
				assert.Equal(t, "price", msg.Declarations()[0].Name())
				assert.Equal(t, "tax", msg.Declarations()[1].Name())
				assert.Equal(t, "total", msg.Declarations()[2].Name())
			},
		},
		{
			name:  "function with multiple options",
			input: "{$date :datetime year=numeric month=long day=numeric}",
			checkFunc: func(t *testing.T, msg Message) {
				patternMsg := msg.(*PatternMessage)
				elem := patternMsg.Pattern().Elements()[0]
				expr, ok := elem.(*Expression)
				require.True(t, ok)
				require.NotNil(t, expr.FunctionRef())

				funcRef := expr.FunctionRef()
				assert.Equal(t, "datetime", funcRef.Name())
				assert.NotNil(t, funcRef.Options())
				assert.Len(t, funcRef.Options(), 3)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cstMsg := cst.ParseCST(tt.input, false)
			msg, err := FromCST(cstMsg)

			require.NoError(t, err)
			require.NotNil(t, msg)

			if tt.checkFunc != nil {
				tt.checkFunc(t, msg)
			}
		})
	}
}

func TestFromCSTNamespaces(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		checkFunc func(t *testing.T, msg Message)
	}{
		{
			name:  "namespaced function",
			input: "{$value :custom:format}",
			checkFunc: func(t *testing.T, msg Message) {
				patternMsg := msg.(*PatternMessage)
				elem := patternMsg.Pattern().Elements()[0]
				expr, ok := elem.(*Expression)
				require.True(t, ok)
				require.NotNil(t, expr.FunctionRef())
				assert.Equal(t, "custom:format", expr.FunctionRef().Name())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cstMsg := cst.ParseCST(tt.input, false)
			msg, err := FromCST(cstMsg)

			require.NoError(t, err)
			require.NotNil(t, msg)

			if tt.checkFunc != nil {
				tt.checkFunc(t, msg)
			}
		})
	}
}
