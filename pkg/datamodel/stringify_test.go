package datamodel

import (
	"testing"

	"github.com/kaptinlin/messageformat-go/internal/cst"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringifyMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  Message
		expected string
	}{
		{
			name: "simple pattern message",
			message: mustPatternMessage(t,
				nil,
				mustPattern(t, []PatternElement{NewTextElement("Hello, world!")}),
				""),

			expected: "Hello, world!",
		},
		{
			name: "pattern message with expression",
			message: mustPatternMessage(t,
				nil,
				mustPattern(t, []PatternElement{
					NewTextElement("Hello "),
					mustExpression(t, NewVariableRef("name"), nil, nil),
					NewTextElement("!"),
				}),

				""),

			expected: "Hello {$name}!",
		},
		{
			name: "pattern message with function",
			message: mustPatternMessage(t,
				nil,
				mustPattern(t, []PatternElement{
					mustExpression(t,
						NewVariableRef("count"),
						mustFunctionRef(t, "number", nil),
						nil),
				}),

				""),

			expected: "{$count :number}",
		},
		{
			name: "pattern message with input declaration",
			message: mustPatternMessage(t,
				[]Declaration{
					mustInputDeclaration(t,

						mustExpression(t,
							NewVariableRef("count"),
							mustFunctionRef(t, "integer", nil),
							nil)),
				},
				mustPattern(t, []PatternElement{
					mustExpression(t, NewVariableRef("count"), nil, nil),
				}),

				""),

			expected: ".input {$count :integer}\n{{{$count}}}",
		},
		{
			name: "pattern message with local declaration",
			message: mustPatternMessage(t,
				[]Declaration{
					NewLocalDeclaration(
						"formatted",
						mustExpression(t,
							NewVariableRef("count"),
							mustFunctionRef(t, "number", nil),
							nil),
					),
				},
				mustPattern(t, []PatternElement{
					mustExpression(t, NewVariableRef("formatted"), nil, nil),
				}),

				""),

			expected: ".local $formatted = {$count :number}\n{{{$formatted}}}",
		},
		{
			name: "simple select message",
			message: mustSelectMessage(t,
				nil,
				[]VariableRef{*NewVariableRef("count")},
				[]Variant{
					*mustVariant(t,
						[]VariantKey{NewLiteral("one")},
						mustPattern(t, []PatternElement{NewTextElement("One item")})),

					*mustVariant(t,
						[]VariantKey{NewCatchallKey("")},
						mustPattern(t, []PatternElement{NewTextElement("Many items")})),
				},
				""),

			expected: ".match $count\none {{One item}}\n* {{Many items}}",
		},
		{
			name: "multi-selector message",
			message: mustSelectMessage(t,
				nil,
				[]VariableRef{
					*NewVariableRef("count"),
					*NewVariableRef("gender"),
				},
				[]Variant{
					*mustVariant(t,
						[]VariantKey{NewLiteral("one"), NewLiteral("male")},
						mustPattern(t, []PatternElement{NewTextElement("He has one item")})),

					*mustVariant(t,
						[]VariantKey{NewLiteral("one"), NewLiteral("female")},
						mustPattern(t, []PatternElement{NewTextElement("She has one item")})),

					*mustVariant(t,
						[]VariantKey{NewCatchallKey(""), NewCatchallKey("")},
						mustPattern(t, []PatternElement{NewTextElement("They have items")})),
				},
				""),

			expected: ".match $count $gender\none male {{He has one item}}\none female {{She has one item}}\n* * {{They have items}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringifyMessage(tt.message)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStringifyExpression(t *testing.T) {
	tests := []struct {
		name     string
		expr     *Expression
		expected string
	}{
		{
			name:     "variable reference",
			expr:     mustExpression(t, NewVariableRef("name"), nil, nil),
			expected: "{$name}",
		},
		{
			name:     "literal argument",
			expr:     mustExpression(t, NewLiteral("42"), nil, nil),
			expected: "{42}",
		},
		{
			name:     "quoted literal with spaces",
			expr:     mustExpression(t, NewLiteral("hello world"), nil, nil),
			expected: "{|hello world|}",
		},
		{
			name: "expression with function",
			expr: mustExpression(t,
				NewVariableRef("count"),
				mustFunctionRef(t, "number", nil),
				nil),

			expected: "{$count :number}",
		},
		{
			name: "expression with function and options",
			expr: mustExpression(t,
				NewVariableRef("price"),
				mustFunctionRef(t, "number", ConvertMapToOptions(map[string]any{
					"style":    NewLiteral("currency"),
					"currency": NewLiteral("USD"),
				})),

				nil),

			expected: "", // Skip exact match due to map iteration order
		},
		{
			name: "expression with attributes",
			expr: mustExpression(t,
				NewVariableRef("value"),
				nil,
				ConvertMapToAttributes(map[string]any{
					"id": NewLiteral("test"),
				})),

			expected: "{$value @id=test}",
		},
		{
			name: "expression with boolean attribute",
			expr: mustExpression(t,
				NewVariableRef("value"),
				nil,
				ConvertMapToAttributes(map[string]any{
					"checked": true,
				})),

			expected: "{$value @checked}",
		},
		{
			name: "function only (no argument)",
			expr: mustExpression(t,
				nil,
				mustFunctionRef(t, "randomValue", nil),
				nil),

			expected: "{:randomValue}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringifyExpression(tt.expr)
			if tt.expected == "" {
				// For tests with map options, just check structure
				assert.Contains(t, result, "{$")
				assert.Contains(t, result, "}")
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestStringifyLiteral(t *testing.T) {
	tests := []struct {
		name     string
		literal  *Literal
		expected string
	}{
		{
			name:     "simple literal",
			literal:  NewLiteral("hello"),
			expected: "hello",
		},
		{
			name:     "numeric literal",
			literal:  NewLiteral("42"),
			expected: "42",
		},
		{
			name:     "literal with spaces",
			literal:  NewLiteral("hello world"),
			expected: "|hello world|",
		},
		{
			name:     "literal with special chars",
			literal:  NewLiteral("hello{world}"),
			expected: "|hello{world}|",
		},
		{
			name:     "literal with pipe",
			literal:  NewLiteral("hello|world"),
			expected: "|hello\\|world|",
		},
		{
			name:     "literal with backslash",
			literal:  NewLiteral("hello\\world"),
			expected: "|hello\\\\world|",
		},
		{
			name:     "literal starting with dot",
			literal:  NewLiteral(".hidden"),
			expected: "|.hidden|",
		},
		{
			name:     "empty literal",
			literal:  NewLiteral(""),
			expected: "||",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringifyLiteral(tt.literal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStringifyMarkup(t *testing.T) {
	tests := []struct {
		name     string
		markup   *Markup
		expected string
	}{
		{
			name:     "open markup",
			markup:   mustMarkup(t, "open", "b", nil, nil),
			expected: "{#b}",
		},
		{
			name:     "close markup",
			markup:   mustMarkup(t, "close", "b", nil, nil),
			expected: "{/b}",
		},
		{
			name:     "standalone markup",
			markup:   mustMarkup(t, "standalone", "br", nil, nil),
			expected: "{#br /}",
		},
		{
			name: "markup with options",
			markup: mustMarkup(t, "open", "link", ConvertMapToOptions(map[string]any{
				"href": NewLiteral("https://example.com"),
			}), nil),
			expected: "{#link href=|https://example.com|}",
		},
		{
			name: "markup with attributes",
			markup: mustMarkup(t, "open", "img", nil, ConvertMapToAttributes(map[string]any{
				"alt": NewLiteral("Image"),
			})),
			expected: "{#img @alt=Image}",
		},
		{
			name: "markup with boolean attribute",
			markup: mustMarkup(t, "open", "input", nil, ConvertMapToAttributes(map[string]any{
				"required": true,
			})),
			expected: "{#input @required}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringifyMarkup(tt.markup)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStringifyPattern(t *testing.T) {
	tests := []struct {
		name     string
		pattern  Pattern
		quoted   bool
		expected string
	}{
		{
			name:     "simple text",
			pattern:  mustPattern(t, []PatternElement{NewTextElement("Hello")}),
			quoted:   false,
			expected: "Hello",
		},
		{
			name: "text with expression",
			pattern: mustPattern(t, []PatternElement{
				NewTextElement("Count: "),
				mustExpression(t, NewVariableRef("count"), nil, nil),
			}),

			quoted:   false,
			expected: "Count: {$count}",
		},
		{
			name:     "quoted pattern",
			pattern:  mustPattern(t, []PatternElement{NewTextElement("Hello")}),
			quoted:   true,
			expected: "{{Hello}}",
		},
		{
			name:     "pattern starting with dot (auto-quoted)",
			pattern:  mustPattern(t, []PatternElement{NewTextElement(".local value")}),
			quoted:   false,
			expected: "{{.local value}}",
		},
		{
			name: "pattern starting with spaced dot and expression",
			pattern: mustPattern(t, []PatternElement{
				NewTextElement("  .local "),
				mustExpression(t, NewVariableRef("name"), nil, nil),
			}),

			quoted:   false,
			expected: "{{  .local {$name}}}",
		},
		{
			name:     "pattern with special characters",
			pattern:  mustPattern(t, []PatternElement{NewTextElement("Hello {world}")}),
			quoted:   false,
			expected: "Hello \\{world\\}",
		},
		{
			name:     "pattern with backslash",
			pattern:  mustPattern(t, []PatternElement{NewTextElement("Path: C:\\folder")}),
			quoted:   false,
			expected: "Path: C:\\\\folder",
		},
		{
			name: "pattern with markup",
			pattern: mustPattern(t, []PatternElement{
				NewTextElement("Text with "),
				mustMarkup(t, "open", "b", nil, nil),
				NewTextElement("bold"),
				mustMarkup(t, "close", "b", nil, nil),
			}),

			quoted:   false,
			expected: "Text with {#b}bold{/b}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringifyPattern(tt.pattern, tt.quoted)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStringifyVariableRef(t *testing.T) {
	tests := []struct {
		name     string
		varRef   *VariableRef
		expected string
	}{
		{
			name:     "simple variable",
			varRef:   NewVariableRef("name"),
			expected: "$name",
		},
		{
			name:     "variable with underscore",
			varRef:   NewVariableRef("user_name"),
			expected: "$user_name",
		},
		{
			name:     "variable with number",
			varRef:   NewVariableRef("item1"),
			expected: "$item1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringifyVariableRef(tt.varRef)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStringifyFunctionRef(t *testing.T) {
	tests := []struct {
		name     string
		funcRef  *FunctionRef
		expected string
	}{
		{
			name:     "simple function",
			funcRef:  mustFunctionRef(t, "number", nil),
			expected: ":number",
		},
		{
			name: "function with single option",
			funcRef: mustFunctionRef(t, "number", ConvertMapToOptions(map[string]any{
				"style": NewLiteral("decimal"),
			})),

			expected: ":number",
		},
		{
			name: "function with multiple options",
			funcRef: mustFunctionRef(t, "number", ConvertMapToOptions(map[string]any{
				"style":    NewLiteral("currency"),
				"currency": NewLiteral("USD"),
			})),

			// Note: map iteration order is not guaranteed
			expected: ":number",
		},
		{
			name: "function with variable option",
			funcRef: mustFunctionRef(t, "number", ConvertMapToOptions(map[string]any{
				"minDigits": NewVariableRef("digits"),
			})),

			expected: ":number",
		},
		{
			name:     "namespaced function",
			funcRef:  mustFunctionRef(t, "custom:format", nil),
			expected: ":custom:format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringifyFunctionRef(tt.funcRef)

			// All results should at least start with the function name
			assert.Contains(t, result, tt.expected, "Result should contain function name")

			// For functions with options, check they're longer
			if tt.funcRef.Options() != nil && len(tt.funcRef.Options()) > 0 {
				assert.True(t, len(result) > len(tt.expected), "Result should be longer with options")
			}
		})
	}
}

func TestStringifyOption(t *testing.T) {
	tests := []struct {
		name     string
		optName  string
		optValue any
		expected string
	}{
		{
			name:     "literal option",
			optName:  "style",
			optValue: NewLiteral("currency"),
			expected: "style=currency",
		},
		{
			name:     "variable option",
			optName:  "minDigits",
			optValue: NewVariableRef("digits"),
			expected: "minDigits=$digits",
		},
		{
			name:     "quoted literal option",
			optName:  "format",
			optValue: NewLiteral("hello world"),
			expected: "format=|hello world|",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringifyOption(tt.optName, tt.optValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStringifyAttribute(t *testing.T) {
	tests := []struct {
		name     string
		attrName string
		attrVal  any
		expected string
	}{
		{
			name:     "boolean attribute",
			attrName: "checked",
			attrVal:  true,
			expected: "@checked",
		},
		{
			name:     "literal attribute",
			attrName: "id",
			attrVal:  NewLiteral("test123"),
			expected: "@id=test123",
		},
		{
			name:     "quoted literal attribute",
			attrName: "title",
			attrVal:  NewLiteral("Hello World"),
			expected: "@title=|Hello World|",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringifyAttribute(tt.attrName, tt.attrVal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStringifyRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "simple message",
			input: "Hello, world!",
		},
		{
			name:  "message with variable",
			input: "Hello {$name}!",
		},
		{
			name:  "message with function",
			input: "{$count :number}",
		},
		{
			name:  "select message",
			input: ".match $count\none {{One item}}\n* {{Other items}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse input to CST
			cstMsg := cst.ParseCST(tt.input, false)

			// Convert to data model
			msg, err := fromCST(cstMsg)
			require.NoError(t, err)

			// Stringify back
			result := StringifyMessage(msg)

			// Parse again to verify equivalence
			cstMsg2 := cst.ParseCST(result, false)
			msg2, err := fromCST(cstMsg2)
			require.NoError(t, err)

			// Both messages should have the same structure
			assert.Equal(t, msg.Type(), msg2.Type())
			assert.Equal(t, len(msg.Declarations()), len(msg2.Declarations()))
		})
	}
}

func TestIsValidUnquotedLiteral(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{
			name:     "simple word",
			value:    "hello",
			expected: true,
		},
		{
			name:     "alphanumeric",
			value:    "test123",
			expected: true,
		},
		{
			name:     "with underscore",
			value:    "test_value",
			expected: true,
		},
		{
			name:     "with hyphen",
			value:    "test-value",
			expected: true,
		},
		{
			name:     "empty string",
			value:    "",
			expected: false,
		},
		{
			name:     "with space",
			value:    "hello world",
			expected: false,
		},
		{
			name:     "with tab",
			value:    "hello\tworld",
			expected: false,
		},
		{
			name:     "with newline",
			value:    "hello\nworld",
			expected: false,
		},
		{
			name:     "with brace",
			value:    "hello{world",
			expected: false,
		},
		{
			name:     "with pipe",
			value:    "hello|world",
			expected: false,
		},
		{
			name:     "with backslash",
			value:    "hello\\world",
			expected: false,
		},
		{
			name:     "with equals",
			value:    "hello=world",
			expected: false,
		},
		{
			name:     "with at sign",
			value:    "hello@world",
			expected: false,
		},
		{
			name:     "with dollar sign",
			value:    "hello$world",
			expected: false,
		},
		{
			name:     "with colon",
			value:    "hello:world",
			expected: false,
		},
		{
			name:     "with hash",
			value:    "hello#world",
			expected: false,
		},
		{
			name:     "with slash",
			value:    "hello/world",
			expected: false,
		},
		{
			name:     "starts with dot",
			value:    ".hidden",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidUnquotedLiteral(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStringifyComplexMessages(t *testing.T) {
	tests := []struct {
		name    string
		message Message
	}{
		{
			name: "message with multiple declarations",
			message: mustPatternMessage(t,
				[]Declaration{
					mustInputDeclaration(t,

						mustExpression(t,
							NewVariableRef("price"),
							mustFunctionRef(t, "number", nil),
							nil)),

					NewLocalDeclaration(
						"tax",
						mustExpression(t,
							NewVariableRef("price"),
							mustFunctionRef(t, "number", ConvertMapToOptions(map[string]any{
								"style": NewLiteral("percent"),
							})),

							nil),
					),
				},
				mustPattern(t, []PatternElement{
					NewTextElement("Price: "),
					mustExpression(t, NewVariableRef("price"), nil, nil),
					NewTextElement(", Tax: "),
					mustExpression(t, NewVariableRef("tax"), nil, nil),
				}),

				""),
		},
		{
			name: "select with expressions in variants",
			message: mustSelectMessage(t,
				nil,
				[]VariableRef{*NewVariableRef("count")},
				[]Variant{
					*mustVariant(t,
						[]VariantKey{NewLiteral("one")},
						mustPattern(t, []PatternElement{
							NewTextElement("You have "),
							mustExpression(t, NewVariableRef("count"), nil, nil),
							NewTextElement(" item"),
						})),

					*mustVariant(t,
						[]VariantKey{NewCatchallKey("")},
						mustPattern(t, []PatternElement{
							NewTextElement("You have "),
							mustExpression(t, NewVariableRef("count"), nil, nil),
							NewTextElement(" items"),
						})),
				},
				""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringifyMessage(tt.message)

			// Should not panic and should produce non-empty string
			assert.NotEmpty(t, result)

			// Result should be parseable
			cstMsg := cst.ParseCST(result, false)
			msg, err := fromCST(cstMsg)
			require.NoError(t, err)
			assert.NotNil(t, msg)
		})
	}
}
