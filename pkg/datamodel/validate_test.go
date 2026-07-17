package datamodel

import (
	"errors"
	"testing"

	pkgerrors "github.com/kaptinlin/messageformat-go/pkg/errors"
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
			message: mustPatternMessage(t,
				nil,
				mustPattern(t, []PatternElement{NewTextElement("Hello World")}),
				""),

			wantErrors: false,
			wantFuncs:  []string{},
			wantVars:   []string{},
		},
		{
			name: "valid pattern message with expression",
			message: mustPatternMessage(t,
				nil,
				mustPattern(t, []PatternElement{
					NewTextElement("Hello "),
					mustExpression(t, NewVariableRef("name"), nil, nil),
				}),

				""),

			wantErrors: false,
			wantFuncs:  []string{},
			wantVars:   []string{"name"},
		},
		{
			name: "valid pattern message with function",
			message: mustPatternMessage(t,
				nil,
				mustPattern(t, []PatternElement{
					mustExpression(t,
						NewVariableRef("count"),
						mustFunctionRef(t, "number", nil),
						nil),
				}),

				""),

			wantErrors: false,
			wantFuncs:  []string{"number"},
			wantVars:   []string{"count"},
		},
		{
			name: "valid input declaration",
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

			wantErrors: false,
			wantFuncs:  []string{"integer"},
			wantVars:   []string{"count"},
		},
		{
			name: "valid local declaration",
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

			wantErrors: false,
			wantFuncs:  []string{"number"},
			wantVars:   []string{"count"},
		},
		{
			name: "duplicate declaration error",
			message: mustPatternMessage(t,
				[]Declaration{
					NewLocalDeclaration("x", mustExpression(t, NewLiteral("1"), nil, nil)),
					NewLocalDeclaration("x", mustExpression(t, NewLiteral("2"), nil, nil)),
				},
				mustPattern(t, []PatternElement{NewTextElement("test")}),
				""),

			wantErrors:    true,
			errorContains: "duplicate-declaration",
		},
		{
			name: "self-reference in local declaration",
			message: mustPatternMessage(t,
				[]Declaration{
					NewLocalDeclaration("x", mustExpression(t, NewVariableRef("x"), nil, nil)),
				},
				mustPattern(t, []PatternElement{NewTextElement("test")}),
				""),

			wantErrors:    true,
			errorContains: "duplicate-declaration",
		},
		{
			name: "valid select message with catchall",
			message: mustSelectMessage(t,
				[]Declaration{
					mustInputDeclaration(t,

						mustExpression(t,
							NewVariableRef("count"),
							mustFunctionRef(t, "number", nil),
							nil)),
				},
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

			wantErrors: false,
			wantFuncs:  []string{"number"},
			wantVars:   []string{"count"},
		},
		{
			name: "select message with 'other' literal as last variant requires catchall",
			message: mustSelectMessage(t,
				[]Declaration{
					mustInputDeclaration(t,

						mustExpression(t,
							NewVariableRef("count"),
							mustFunctionRef(t, "number", nil),
							nil)),
				},
				[]VariableRef{*NewVariableRef("count")},
				[]Variant{
					*mustVariant(t,
						[]VariantKey{NewLiteral("one")},
						mustPattern(t, []PatternElement{NewTextElement("One item")})),

					*mustVariant(t,
						[]VariantKey{NewLiteral("other")},
						mustPattern(t, []PatternElement{NewTextElement("Many items")})),
				},
				""),

			// Per MF2 spec only catchall (*) keys count as fallback; the
			// literal "other" key does not satisfy missing-fallback.
			wantErrors:    true,
			errorContains: "missing-fallback",
		},
		{
			name: "select message with key mismatch",
			message: mustSelectMessage(t,
				nil,
				[]VariableRef{
					*NewVariableRef("count"),
					*NewVariableRef("type"),
				},
				[]Variant{
					*mustVariant(t,
						[]VariantKey{NewLiteral("one")},
						mustPattern(t, []PatternElement{NewTextElement("One item")})),

					*mustVariant(t,
						[]VariantKey{NewLiteral("other"), NewCatchallKey("")},
						mustPattern(t, []PatternElement{NewTextElement("Many items")})),
				},
				""),

			wantErrors:    true,
			errorContains: "key-mismatch",
		},
		{
			name: "select message with duplicate variants",
			message: mustSelectMessage(t,
				nil,
				[]VariableRef{*NewVariableRef("count")},
				[]Variant{
					*mustVariant(t,
						[]VariantKey{NewLiteral("one")},
						mustPattern(t, []PatternElement{NewTextElement("First one")})),

					*mustVariant(t,
						[]VariantKey{NewLiteral("one")},
						mustPattern(t, []PatternElement{NewTextElement("Second one")})),

					*mustVariant(t,
						[]VariantKey{NewCatchallKey("")},
						mustPattern(t, []PatternElement{NewTextElement("Other")})),
				},
				""),

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
			errorContains: "invalid-message",
		},
		{
			name: "function in option value",
			message: mustPatternMessage(t,
				nil,
				mustPattern(t, []PatternElement{
					mustExpression(t,
						NewVariableRef("count"),
						mustFunctionRef(t, "number", ConvertMapToOptions(map[string]any{
							"minDigits": NewVariableRef("digits"),
						})),

						nil),
				}),

				""),

			wantErrors: false,
			wantFuncs:  []string{"number"},
			wantVars:   []string{"count", "digits"},
		},
		{
			name: "multiple functions and variables",
			message: mustPatternMessage(t,
				[]Declaration{
					mustInputDeclaration(t,

						mustExpression(t,
							NewVariableRef("price"),
							mustFunctionRef(t, "number", nil),
							nil)),

					NewLocalDeclaration(
						"date",
						mustExpression(t,
							NewVariableRef("timestamp"),
							mustFunctionRef(t, "datetime", nil),
							nil),
					),
				},
				mustPattern(t, []PatternElement{
					mustExpression(t, NewVariableRef("price"), nil, nil),
					NewTextElement(" on "),
					mustExpression(t, NewVariableRef("date"), nil, nil),
				}),

				""),

			wantErrors: false,
			wantFuncs:  []string{"number", "datetime"},
			wantVars:   []string{"price", "timestamp"},
		},
		{
			name: "forward reference in local declaration",
			message: mustPatternMessage(t,
				[]Declaration{
					NewLocalDeclaration("a", mustExpression(t, NewVariableRef("b"), nil, nil)),
					NewLocalDeclaration("b", mustExpression(t, NewLiteral("value"), nil, nil)),
				},
				mustPattern(t, []PatternElement{NewTextElement("test")}),
				""),

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
	errorHandler := func(errType string, node any) {
		errorTypes = append(errorTypes, errType)
	}

	msg := mustPatternMessage(t,
		[]Declaration{
			NewLocalDeclaration("x", mustExpression(t, NewLiteral("1"), nil, nil)),
			NewLocalDeclaration("x", mustExpression(t, NewLiteral("2"), nil, nil)),
		},
		mustPattern(t, []PatternElement{NewTextElement("test")}),
		"")

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
			message: mustSelectMessage(t,
				[]Declaration{
					mustInputDeclaration(t,

						mustExpression(t, NewVariableRef("count"), mustFunctionRef(t, "number", nil), nil)),

					mustInputDeclaration(t,

						mustExpression(t, NewVariableRef("gender"), mustFunctionRef(t, "string", nil), nil)),
				},
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

			wantErrors: false,
		},
		{
			name: "multi-selector with key count mismatch",
			message: mustSelectMessage(t,
				nil,
				[]VariableRef{
					*NewVariableRef("count"),
					*NewVariableRef("gender"),
				},
				[]Variant{
					*mustVariant(t,
						[]VariantKey{NewLiteral("one")},
						mustPattern(t, []PatternElement{NewTextElement("One")})),

					*mustVariant(t,
						[]VariantKey{NewCatchallKey(""), NewCatchallKey("")},
						mustPattern(t, []PatternElement{NewTextElement("Other")})),
				},
				""),

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
			message: mustPatternMessage(t,
				nil,
				mustPattern(t, []PatternElement{
					mustMarkup(t, "open", "b", nil, nil),
					NewTextElement("Bold text"),
					mustMarkup(t, "close", "b", nil, nil),
				}),

				""),

			wantErrors: false,
		},
		{
			name: "valid standalone markup",
			message: mustPatternMessage(t,
				nil,
				mustPattern(t, []PatternElement{
					NewTextElement("Line 1"),
					mustMarkup(t, "standalone", "br", nil, nil),
					NewTextElement("Line 2"),
				}),

				""),

			wantErrors: false,
		},
		{
			name: "markup with options",
			message: mustPatternMessage(t,
				nil,
				mustPattern(t, []PatternElement{
					mustMarkup(t, "open", "link", ConvertMapToOptions(map[string]any{
						"href": NewLiteral("https://example.com"),
					}), nil),
					NewTextElement("Click here"),
					mustMarkup(t, "close", "link", nil, nil),
				}),

				""),

			wantErrors: false,
		},
		{
			name: "markup with attributes",
			message: mustPatternMessage(t,
				nil,
				mustPattern(t, []PatternElement{
					mustMarkup(t, "open", "img", nil, ConvertMapToAttributes(map[string]any{
						"alt": NewLiteral("Image description"),
					})),
				}),

				""),

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
			message: mustPatternMessage(t,
				[]Declaration{
					mustInputDeclaration(t,

						mustExpression(t,
							NewVariableRef("count"),
							mustFunctionRef(t, "integer", nil),
							nil)),
				},
				mustPattern(t, []PatternElement{NewTextElement("test")}),
				""),

			wantErrors: false,
			wantFuncs:  []string{"integer"},
		},
		{
			name: "local declaration referencing annotated variable",
			message: mustPatternMessage(t,
				[]Declaration{
					mustInputDeclaration(t,

						mustExpression(t,
							NewVariableRef("count"),
							mustFunctionRef(t, "integer", nil),
							nil)),

					NewLocalDeclaration(
						"formatted",
						mustExpression(t, NewVariableRef("count"), nil, nil),
					),
				},
				mustPattern(t, []PatternElement{NewTextElement("test")}),
				""),

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

// TestValidateSelectMessageOtherLiteralFallback asserts that a literal "other"
// key does NOT satisfy the missing-fallback requirement — only a catchall (*)
// key counts as a fallback variant per MF2 spec.
func TestValidateSelectMessageOtherLiteralFallback(t *testing.T) {
	t.Parallel()

	message := mustSelectMessage(t,
		[]Declaration{
			mustInputDeclaration(t,

				mustExpression(t, NewVariableRef("count"), mustFunctionRef(t, "integer", nil), nil)),
		},
		[]VariableRef{*NewVariableRef("count")},
		[]Variant{
			*mustVariant(t,
				[]VariantKey{NewLiteral("one")},
				mustPattern(t, []PatternElement{NewTextElement("one")})),

			*mustVariant(t,
				[]VariantKey{NewLiteral("other")},
				mustPattern(t, []PatternElement{
					mustExpression(t, NewVariableRef("count"), mustFunctionRef(t, "integer", nil), nil),
				})),
		},
		"")

	_, err := ValidateMessage(message, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing-fallback")
}

func TestValidationResultStructure(t *testing.T) {
	msg := mustPatternMessage(t,
		[]Declaration{
			mustInputDeclaration(t,

				mustExpression(t,
					NewVariableRef("count"),
					mustFunctionRef(t, "integer", nil),
					nil)),

			NewLocalDeclaration(
				"date",
				mustExpression(t,
					NewVariableRef("timestamp"),
					mustFunctionRef(t, "datetime", nil),
					nil),
			),
		},
		mustPattern(t, []PatternElement{
			mustExpression(t, NewVariableRef("count"), nil, nil),
			NewTextElement(" items on "),
			mustExpression(t, NewVariableRef("date"), nil, nil),
			NewTextElement(" by "),
			mustExpression(t, NewVariableRef("user"), nil, nil),
		}),

		"")

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

func TestValidationResultExcludesAllLocalVariables(t *testing.T) {
	msg := mustPatternMessage(t,
		[]Declaration{
			NewLocalDeclaration("first", mustExpression(t, NewVariableRef("external"), nil, nil)),
			NewLocalDeclaration("second", mustExpression(t, NewVariableRef("first"), nil, nil)),
		},
		mustPattern(t, []PatternElement{
			mustExpression(t, NewVariableRef("first"), nil, nil),
			mustExpression(t, NewVariableRef("second"), nil, nil),
			mustExpression(t, NewVariableRef("external"), nil, nil),
		}),

		"")

	result, err := ValidateMessage(msg, nil)

	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"external"}, result.Variables)
}

// TestValidationResultOrderAndCoverage proves validation reports all dependencies deterministically.
// TypeScript original code:
// return { functions, variables };
func TestValidationResultOrderAndCoverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		message       Message
		wantFunctions []string
		wantVariables []string
	}{
		{
			name: "declarations expressions and markup options",
			message: mustPatternMessage(t,
				[]Declaration{
					mustInputDeclaration(t, mustExpression(t,
						NewVariableRef("inputArg"),
						mustFunctionRef(t, "declFn", Options{
							"z": NewVariableRef("declZ"),
							"a": NewVariableRef("declA"),
						}), nil)),
					NewLocalDeclaration("local", mustExpression(t,
						NewVariableRef("localArg"),
						mustFunctionRef(t, "localFn", Options{
							"z": NewVariableRef("localZ"),
							"a": NewVariableRef("localA"),
						}), nil)),
				},
				mustPattern(t, []PatternElement{
					mustExpression(t,
						NewVariableRef("exprArg"),
						mustFunctionRef(t, "exprFn", Options{
							"z": NewVariableRef("exprZ"),
							"a": NewVariableRef("exprA"),
						}), nil),
					mustMarkup(t, MarkupStandalone, "link", Options{
						"z": NewVariableRef("markupZ"),
						"a": NewVariableRef("markupA"),
					}, nil),
				}), ""),
			wantFunctions: []string{"declFn", "localFn", "exprFn"},
			wantVariables: []string{
				"inputArg", "declA", "declZ",
				"localArg", "localA", "localZ",
				"exprArg", "exprA", "exprZ",
				"markupA", "markupZ",
			},
		},
		{
			name: "selectors and variant patterns",
			message: mustSelectMessage(t,
				[]Declaration{mustInputDeclaration(t, mustExpression(t,
					NewVariableRef("selector"), mustFunctionRef(t, "selectorFn", nil), nil))},
				[]VariableRef{*NewVariableRef("selector")},
				[]Variant{
					*mustVariant(t, []VariantKey{NewLiteral("one")}, mustPattern(t, []PatternElement{
						mustExpression(t, NewVariableRef("variantArg"), mustFunctionRef(t, "variantFn", nil), nil),
					})),
					*mustVariant(t, []VariantKey{NewCatchallKey("")}, mustPattern(t, []PatternElement{
						mustExpression(t, NewVariableRef("fallbackArg"), mustFunctionRef(t, "fallbackFn", nil), nil),
					})),
				}, ""),
			wantFunctions: []string{"selectorFn", "variantFn", "fallbackFn"},
			wantVariables: []string{"selector", "variantArg", "fallbackArg"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for range 100 {
				result, err := ValidateMessage(tc.message, nil)
				require.NoError(t, err)
				assert.Equal(t, tc.wantFunctions, result.Functions)
				assert.Equal(t, tc.wantVariables, result.Variables)
			}
		})
	}
}

func TestEdgeCasesInValidation(t *testing.T) {
	tests := []struct {
		name       string
		message    Message
		wantErrors bool
	}{
		{
			name: "empty pattern message",
			message: mustPatternMessage(t,
				nil,
				mustPattern(t, []PatternElement{}),
				""),

			wantErrors: false,
		},
		{
			name: "pattern with only text",
			message: mustPatternMessage(t,
				nil,
				mustPattern(t, []PatternElement{
					NewTextElement("Just text, no expressions"),
				}),

				""),

			wantErrors: false,
		},
		{
			name: "select message with single variant",
			message: mustSelectMessage(t,
				[]Declaration{
					mustInputDeclaration(t,

						mustExpression(t, NewVariableRef("count"), mustFunctionRef(t, "number", nil), nil)),
				},
				[]VariableRef{*NewVariableRef("count")},
				[]Variant{
					*mustVariant(t,
						[]VariantKey{NewCatchallKey("")},
						mustPattern(t, []PatternElement{NewTextElement("Any count")})),
				},
				""),

			wantErrors: false,
		},
		{
			name: "expression with only function (no arg)",
			message: mustPatternMessage(t,
				nil,
				mustPattern(t, []PatternElement{
					mustExpression(t, nil, mustFunctionRef(t, "randomValue", nil), nil),
				}),

				""),

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

func TestValidateMessageErrorTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		message  Message
		wantType string
	}{
		{
			name: "local option forward reference",
			message: mustPatternMessage(t,
				[]Declaration{
					NewLocalDeclaration("formatted", mustExpression(t, NewLiteral("value"), mustFunctionRef(t, "string", ConvertMapToOptions(map[string]any{"case": NewVariableRef("style")})), nil)),
					NewLocalDeclaration("style", mustExpression(t, NewLiteral("title"), nil, nil)),
				},
				mustPattern(t, []PatternElement{NewTextElement("test")}),
				""),

			wantType: "duplicate-declaration",
		},
		{
			name: "local option self reference",
			message: mustPatternMessage(t,
				[]Declaration{
					NewLocalDeclaration("formatted", mustExpression(t, NewLiteral("value"), mustFunctionRef(t, "string", ConvertMapToOptions(map[string]any{"case": NewVariableRef("formatted")})), nil)),
				},
				mustPattern(t, []PatternElement{NewTextElement("test")}),
				""),

			wantType: "duplicate-declaration",
		},
		{
			name: "select without fallback",
			message: mustSelectMessage(t,
				[]Declaration{
					mustInputDeclaration(t,

						mustExpression(t, NewVariableRef("count"), mustFunctionRef(t, "number", nil), nil)),
				},
				[]VariableRef{*NewVariableRef("count")},
				[]Variant{
					*mustVariant(t, []VariantKey{NewLiteral("one")}, mustPattern(t, []PatternElement{NewTextElement("one")})),
				},
				""),

			wantType: "missing-fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var gotTypes []string
			_, err := ValidateMessage(tt.message, func(errType string, node any) {
				gotTypes = append(gotTypes, errType)
			})

			require.Error(t, err)
			require.Len(t, gotTypes, 1)
			assert.Equal(t, tt.wantType, gotTypes[0])
		})
	}
}

func TestValidateMessageReturnsDataModelErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		source   string
		wantType string
	}{
		{
			name:     "variant key mismatch",
			source:   ".input {$foo :x} .match $foo * * {{foo}}",
			wantType: "key-mismatch",
		},
		{
			name:     "missing fallback",
			source:   ".input {$foo :x} .match $foo 1 {{_}}",
			wantType: "missing-fallback",
		},
		{
			name:     "missing selector annotation",
			source:   ".input {$foo} .match $foo one {{one}} * {{other}}",
			wantType: "missing-selector-annotation",
		},
		{
			name:     "duplicate declaration",
			source:   ".input {$foo} .input {$foo} {{_}}",
			wantType: "duplicate-declaration",
		},
		{
			name:     "duplicate variant",
			source:   ".input {$var :string} .match $var * {{first}} * {{second}}",
			wantType: "duplicate-variant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			message, err := ParseMessage(tt.source)
			require.NoError(t, err)

			_, err = ValidateMessage(message, nil)
			require.Error(t, err)

			var modelErr *pkgerrors.MessageDataModelError
			require.True(t, errors.As(err, &modelErr), "got %T: %v", err, err)
			assert.Equal(t, tt.wantType, modelErr.ErrorType())
			assert.LessOrEqual(t, modelErr.Start, modelErr.End)
		})
	}
}

func TestValidateMessageRejectsNilMessages(t *testing.T) {
	t.Parallel()

	var pattern *PatternMessage
	var selectMessage *SelectMessage
	tests := []struct {
		name    string
		message Message
	}{
		{name: "nil interface"},
		{name: "nil pattern message", message: pattern},
		{name: "nil select message", message: selectMessage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ValidateMessage(tt.message, nil)
			require.Error(t, err)
			var modelErr *pkgerrors.MessageDataModelError
			require.ErrorAs(t, err, &modelErr)
			assert.Equal(t, "invalid-message", modelErr.ErrorType())
		})
	}
}
