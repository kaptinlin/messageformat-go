package datamodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateMessage(t *testing.T) {
	tests := []struct {
		name          string
		message       Message
		wantErrors    bool
		wantFuncs     []string
		wantVars      []string
		errorContains string
	}{
		{
			name: "valid simple pattern message",
			message: NewPatternMessage(
				nil,
				NewPattern([]PatternElement{NewTextElement("Hello World")}),
				"",
			),
			wantErrors: false,
			wantFuncs:  []string{},
			wantVars:   []string{},
		},
		{
			name: "valid pattern message with expression",
			message: NewPatternMessage(
				nil,
				NewPattern([]PatternElement{
					NewTextElement("Hello "),
					NewExpression(NewVariableRef("name"), nil, nil),
				}),
				"",
			),
			wantErrors: false,
			wantFuncs:  []string{},
			wantVars:   []string{"name"},
		},
		{
			name: "valid pattern message with function",
			message: NewPatternMessage(
				nil,
				NewPattern([]PatternElement{
					NewExpression(
						NewVariableRef("count"),
						NewFunctionRef("number", nil),
						nil,
					),
				}),
				"",
			),
			wantErrors: false,
			wantFuncs:  []string{"number"},
			wantVars:   []string{"count"},
		},
		{
			name: "valid input declaration",
			message: NewPatternMessage(
				[]Declaration{
					NewInputDeclaration(
						"count",
						NewVariableRefExpression(
							NewVariableRef("count"),
							NewFunctionRef("integer", nil),
							nil,
						),
					),
				},
				NewPattern([]PatternElement{
					NewExpression(NewVariableRef("count"), nil, nil),
				}),
				"",
			),
			wantErrors: false,
			wantFuncs:  []string{"integer"},
			wantVars:   []string{"count"},
		},
		{
			name: "valid local declaration",
			message: NewPatternMessage(
				[]Declaration{
					NewLocalDeclaration(
						"formatted",
						NewExpression(
							NewVariableRef("count"),
							NewFunctionRef("number", nil),
							nil,
						),
					),
				},
				NewPattern([]PatternElement{
					NewExpression(NewVariableRef("formatted"), nil, nil),
				}),
				"",
			),
			wantErrors: false,
			wantFuncs:  []string{"number"},
			wantVars:   []string{"count"},
		},
		{
			name: "duplicate declaration error",
			message: NewPatternMessage(
				[]Declaration{
					NewLocalDeclaration("x", NewExpression(NewLiteral("1"), nil, nil)),
					NewLocalDeclaration("x", NewExpression(NewLiteral("2"), nil, nil)),
				},
				NewPattern([]PatternElement{NewTextElement("test")}),
				"",
			),
			wantErrors:    true,
			errorContains: "duplicate-declaration",
		},
		{
			name: "self-reference in local declaration",
			message: NewPatternMessage(
				[]Declaration{
					NewLocalDeclaration("x", NewExpression(NewVariableRef("x"), nil, nil)),
				},
				NewPattern([]PatternElement{NewTextElement("test")}),
				"",
			),
			wantErrors:    true,
			errorContains: "duplicate-declaration",
		},
		{
			name: "valid select message with catchall",
			message: NewSelectMessage(
				nil,
				[]VariableRef{*NewVariableRef("count")},
				[]Variant{
					*NewVariant(
						[]VariantKey{NewLiteral("one")},
						NewPattern([]PatternElement{NewTextElement("One item")}),
					),
					*NewVariant(
						[]VariantKey{NewCatchallKey("")},
						NewPattern([]PatternElement{NewTextElement("Many items")}),
					),
				},
				"",
			),
			wantErrors: false,
			wantFuncs:  []string{},
			wantVars:   []string{"count"},
		},
		{
			name: "valid select message with 'other' fallback",
			message: NewSelectMessage(
				nil,
				[]VariableRef{*NewVariableRef("count")},
				[]Variant{
					*NewVariant(
						[]VariantKey{NewLiteral("one")},
						NewPattern([]PatternElement{NewTextElement("One item")}),
					),
					*NewVariant(
						[]VariantKey{NewLiteral("other")},
						NewPattern([]PatternElement{NewTextElement("Many items")}),
					),
				},
				"",
			),
			wantErrors: false,
			wantFuncs:  []string{},
			wantVars:   []string{"count"},
		},
		{
			name: "select message with key mismatch",
			message: NewSelectMessage(
				nil,
				[]VariableRef{
					*NewVariableRef("count"),
					*NewVariableRef("type"),
				},
				[]Variant{
					*NewVariant(
						[]VariantKey{NewLiteral("one")}, // Only 1 key but 2 selectors
						NewPattern([]PatternElement{NewTextElement("One item")}),
					),
					*NewVariant(
						[]VariantKey{NewLiteral("other"), NewCatchallKey("")},
						NewPattern([]PatternElement{NewTextElement("Many items")}),
					),
				},
				"",
			),
			wantErrors:    true,
			errorContains: "key-mismatch",
		},
		{
			name: "select message with duplicate variants",
			message: NewSelectMessage(
				nil,
				[]VariableRef{*NewVariableRef("count")},
				[]Variant{
					*NewVariant(
						[]VariantKey{NewLiteral("one")},
						NewPattern([]PatternElement{NewTextElement("First one")}),
					),
					*NewVariant(
						[]VariantKey{NewLiteral("one")},
						NewPattern([]PatternElement{NewTextElement("Second one")}),
					),
					*NewVariant(
						[]VariantKey{NewCatchallKey("")},
						NewPattern([]PatternElement{NewTextElement("Other")}),
					),
				},
				"",
			),
			wantErrors:    true,
			errorContains: "duplicate-variant",
		},
		{
			name: "nil message",
			message: func() Message {
				var msg Message
				return msg
			}(),
			wantErrors:    true,
			errorContains: "parse-error",
		},
		{
			name: "function in option value",
			message: NewPatternMessage(
				nil,
				NewPattern([]PatternElement{
					NewExpression(
						NewVariableRef("count"),
						NewFunctionRef("number", ConvertMapToOptions(map[string]interface{}{
							"minDigits": NewVariableRef("digits"),
						})),
						nil,
					),
				}),
				"",
			),
			wantErrors: false,
			wantFuncs:  []string{"number"},
			wantVars:   []string{"count", "digits"},
		},
		{
			name: "multiple functions and variables",
			message: NewPatternMessage(
				[]Declaration{
					NewInputDeclaration(
						"price",
						NewVariableRefExpression(
							NewVariableRef("price"),
							NewFunctionRef("number", nil),
							nil,
						),
					),
					NewLocalDeclaration(
						"date",
						NewExpression(
							NewVariableRef("timestamp"),
							NewFunctionRef("datetime", nil),
							nil,
						),
					),
				},
				NewPattern([]PatternElement{
					NewExpression(NewVariableRef("price"), nil, nil),
					NewTextElement(" on "),
					NewExpression(NewVariableRef("date"), nil, nil),
				}),
				"",
			),
			wantErrors: false,
			wantFuncs:  []string{"number", "datetime"},
			wantVars:   []string{"price", "timestamp"},
		},
		{
			name: "forward reference in local declaration",
			message: NewPatternMessage(
				[]Declaration{
					NewLocalDeclaration("a", NewExpression(NewVariableRef("b"), nil, nil)),
					NewLocalDeclaration("b", NewExpression(NewLiteral("value"), nil, nil)),
				},
				NewPattern([]PatternElement{NewTextElement("test")}),
				"",
			),
			wantErrors:    true,
			errorContains: "duplicate-declaration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateMessage(tt.message, nil)

			if tt.wantErrors {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)

				if tt.wantFuncs != nil {
					assert.ElementsMatch(t, tt.wantFuncs, result.Functions, "Functions mismatch")
				}

				if tt.wantVars != nil {
					assert.ElementsMatch(t, tt.wantVars, result.Variables, "Variables mismatch")
				}
			}
		})
	}
}

func TestValidateMessageWithCustomErrorHandler(t *testing.T) {
	errorTypes := []string{}
	errorHandler := func(errType string, node interface{}) {
		errorTypes = append(errorTypes, errType)
	}

	msg := NewPatternMessage(
		[]Declaration{
			NewLocalDeclaration("x", NewExpression(NewLiteral("1"), nil, nil)),
			NewLocalDeclaration("x", NewExpression(NewLiteral("2"), nil, nil)),
		},
		NewPattern([]PatternElement{NewTextElement("test")}),
		"",
	)

	_, err := ValidateMessage(msg, errorHandler)

	assert.Error(t, err)
	assert.Contains(t, errorTypes, "duplicate-declaration")
}

func TestValidateComplexSelectMessage(t *testing.T) {
	tests := []struct {
		name       string
		message    Message
		wantErrors bool
		errorType  string
	}{
		{
			name: "multi-selector with proper keys",
			message: NewSelectMessage(
				nil,
				[]VariableRef{
					*NewVariableRef("count"),
					*NewVariableRef("gender"),
				},
				[]Variant{
					*NewVariant(
						[]VariantKey{NewLiteral("one"), NewLiteral("male")},
						NewPattern([]PatternElement{NewTextElement("He has one item")}),
					),
					*NewVariant(
						[]VariantKey{NewLiteral("one"), NewLiteral("female")},
						NewPattern([]PatternElement{NewTextElement("She has one item")}),
					),
					*NewVariant(
						[]VariantKey{NewCatchallKey(""), NewCatchallKey("")},
						NewPattern([]PatternElement{NewTextElement("They have items")}),
					),
				},
				"",
			),
			wantErrors: false,
		},
		{
			name: "multi-selector with key count mismatch",
			message: NewSelectMessage(
				nil,
				[]VariableRef{
					*NewVariableRef("count"),
					*NewVariableRef("gender"),
				},
				[]Variant{
					*NewVariant(
						[]VariantKey{NewLiteral("one")}, // Missing second key
						NewPattern([]PatternElement{NewTextElement("One")}),
					),
					*NewVariant(
						[]VariantKey{NewCatchallKey(""), NewCatchallKey("")},
						NewPattern([]PatternElement{NewTextElement("Other")}),
					),
				},
				"",
			),
			wantErrors: true,
			errorType:  "key-mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateMessage(tt.message, nil)

			if tt.wantErrors {
				assert.Error(t, err)
				if tt.errorType != "" {
					assert.Contains(t, err.Error(), tt.errorType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateMarkup(t *testing.T) {
	tests := []struct {
		name       string
		message    Message
		wantErrors bool
	}{
		{
			name: "valid markup without options",
			message: NewPatternMessage(
				nil,
				NewPattern([]PatternElement{
					NewMarkup("open", "b", nil, nil),
					NewTextElement("Bold text"),
					NewMarkup("close", "b", nil, nil),
				}),
				"",
			),
			wantErrors: false,
		},
		{
			name: "valid standalone markup",
			message: NewPatternMessage(
				nil,
				NewPattern([]PatternElement{
					NewTextElement("Line 1"),
					NewMarkup("standalone", "br", nil, nil),
					NewTextElement("Line 2"),
				}),
				"",
			),
			wantErrors: false,
		},
		{
			name: "markup with options",
			message: NewPatternMessage(
				nil,
				NewPattern([]PatternElement{
					NewMarkup("open", "link", ConvertMapToOptions(map[string]interface{}{
						"href": NewLiteral("https://example.com"),
					}), nil),
					NewTextElement("Click here"),
					NewMarkup("close", "link", nil, nil),
				}),
				"",
			),
			wantErrors: false,
		},
		{
			name: "markup with attributes",
			message: NewPatternMessage(
				nil,
				NewPattern([]PatternElement{
					NewMarkup("open", "img", nil, ConvertMapToAttributes(map[string]interface{}{
						"alt": NewLiteral("Image description"),
					})),
				}),
				"",
			),
			wantErrors: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateMessage(tt.message, nil)

			if tt.wantErrors {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAnnotatedVariables(t *testing.T) {
	tests := []struct {
		name       string
		message    Message
		wantErrors bool
		wantFuncs  []string
	}{
		{
			name: "input declaration with annotation",
			message: NewPatternMessage(
				[]Declaration{
					NewInputDeclaration(
						"count",
						NewVariableRefExpression(
							NewVariableRef("count"),
							NewFunctionRef("integer", nil),
							nil,
						),
					),
				},
				NewPattern([]PatternElement{NewTextElement("test")}),
				"",
			),
			wantErrors: false,
			wantFuncs:  []string{"integer"},
		},
		{
			name: "local declaration referencing annotated variable",
			message: NewPatternMessage(
				[]Declaration{
					NewInputDeclaration(
						"count",
						NewVariableRefExpression(
							NewVariableRef("count"),
							NewFunctionRef("integer", nil),
							nil,
						),
					),
					NewLocalDeclaration(
						"formatted",
						NewExpression(NewVariableRef("count"), nil, nil),
					),
				},
				NewPattern([]PatternElement{NewTextElement("test")}),
				"",
			),
			wantErrors: false,
			wantFuncs:  []string{"integer"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateMessage(tt.message, nil)

			if tt.wantErrors {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.wantFuncs != nil {
					assert.ElementsMatch(t, tt.wantFuncs, result.Functions)
				}
			}
		})
	}
}

func TestValidationResultStructure(t *testing.T) {
	msg := NewPatternMessage(
		[]Declaration{
			NewInputDeclaration(
				"count",
				NewVariableRefExpression(
					NewVariableRef("count"),
					NewFunctionRef("integer", nil),
					nil,
				),
			),
			NewLocalDeclaration(
				"date",
				NewExpression(
					NewVariableRef("timestamp"),
					NewFunctionRef("datetime", nil),
					nil,
				),
			),
		},
		NewPattern([]PatternElement{
			NewExpression(NewVariableRef("count"), nil, nil),
			NewTextElement(" items on "),
			NewExpression(NewVariableRef("date"), nil, nil),
			NewTextElement(" by "),
			NewExpression(NewVariableRef("user"), nil, nil),
		}),
		"",
	)

	result, err := ValidateMessage(msg, nil)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Check that functions are collected
	assert.Contains(t, result.Functions, "integer")
	assert.Contains(t, result.Functions, "datetime")

	// Check that variables are collected (excluding local variables)
	assert.Contains(t, result.Variables, "timestamp")
	assert.Contains(t, result.Variables, "user")
	assert.NotContains(t, result.Variables, "date") // Local variable should not be in Variables
}

func TestEdgeCasesInValidation(t *testing.T) {
	tests := []struct {
		name       string
		message    Message
		wantErrors bool
	}{
		{
			name: "empty pattern message",
			message: NewPatternMessage(
				nil,
				NewPattern([]PatternElement{}),
				"",
			),
			wantErrors: false,
		},
		{
			name: "pattern with only text",
			message: NewPatternMessage(
				nil,
				NewPattern([]PatternElement{
					NewTextElement("Just text, no expressions"),
				}),
				"",
			),
			wantErrors: false,
		},
		{
			name: "select message with single variant",
			message: NewSelectMessage(
				nil,
				[]VariableRef{*NewVariableRef("count")},
				[]Variant{
					*NewVariant(
						[]VariantKey{NewCatchallKey("")},
						NewPattern([]PatternElement{NewTextElement("Any count")}),
					),
				},
				"",
			),
			wantErrors: false,
		},
		{
			name: "expression with only function (no arg)",
			message: NewPatternMessage(
				nil,
				NewPattern([]PatternElement{
					NewExpression(nil, NewFunctionRef("randomValue", nil), nil),
				}),
				"",
			),
			wantErrors: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateMessage(tt.message, nil)

			if tt.wantErrors {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
