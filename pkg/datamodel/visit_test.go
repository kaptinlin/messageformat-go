package datamodel

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVisit_PatternMessageTraversal(t *testing.T) {
	t.Parallel()

	input := NewInputDeclaration(
		"name",
		NewVariableRefExpression(
			NewVariableRef("name"),
			NewFunctionRef("string", ConvertMapToOptions(map[string]any{"case": "title"})),
			ConvertMapToAttributes(map[string]any{"required": true, "label": "Name"}),
		),
	)
	expression := NewExpression(
		NewVariableRef("name"),
		NewFunctionRef("string", ConvertMapToOptions(map[string]any{"case": NewVariableRef("style")})),
		ConvertMapToAttributes(map[string]any{"id": "user-name", "hidden": true}),
	)
	message := NewPatternMessage(
		[]Declaration{input},
		NewPattern([]PatternElement{NewTextElement("Hello "), expression}),
		"",
	)

	var events []string
	Visit(message, &Visitor{
		Declaration: func(declaration Declaration) func() {
			events = append(events, fmt.Sprintf("declaration:%s", declaration.Name()))
			return func() { events = append(events, fmt.Sprintf("end-declaration:%s", declaration.Name())) }
		},
		Expression: func(expression *Expression, context string) func() {
			events = append(events, fmt.Sprintf("expression:%s", context))
			return func() { events = append(events, fmt.Sprintf("end-expression:%s", context)) }
		},
		FunctionRef: func(functionRef *FunctionRef, context string, argument any) func() {
			events = append(events, fmt.Sprintf("function:%s:%s:%T", context, functionRef.Name(), argument))
			return func() { events = append(events, fmt.Sprintf("end-function:%s:%s", context, functionRef.Name())) }
		},
		Options: func(options map[string]any, context string) func() {
			events = append(events, fmt.Sprintf("options:%s:%d", context, len(options)))
			return func() { events = append(events, fmt.Sprintf("end-options:%s", context)) }
		},
		Attributes: func(attributes map[string]any, context string) func() {
			events = append(events, fmt.Sprintf("attributes:%s:%d", context, len(attributes)))
			return func() { events = append(events, fmt.Sprintf("end-attributes:%s", context)) }
		},
		Pattern: func(pattern Pattern) func() {
			events = append(events, fmt.Sprintf("pattern:%d", pattern.Len()))
			return func() { events = append(events, "end-pattern") }
		},
		Value: func(value any, context string, position string) {
			events = append(events, fmt.Sprintf("value:%s:%s:%T", context, position, value))
		},
	})

	assert.Contains(t, events, "declaration:name")
	assert.Contains(t, events, "pattern:2")
	assert.Contains(t, events, "expression:placeholder")
	assert.Contains(t, events, "value:placeholder:option:*datamodel.VariableRef")
	assert.Contains(t, events, "value:placeholder:attribute:*datamodel.Literal")
	assert.NotContains(t, events, "value:placeholder:attribute:*datamodel.BooleanAttribute")
	assert.Contains(t, events, "end-pattern")
}

func TestVisit_SelectMessageTraversal(t *testing.T) {
	t.Parallel()

	selector := NewVariableRef("count")
	one := NewVariant(
		[]VariantKey{NewLiteral("one")},
		NewPattern([]PatternElement{NewExpression(NewVariableRef("count"), nil, nil)}),
	)
	fallback := NewVariant(
		[]VariantKey{NewCatchallKey("")},
		NewPattern([]PatternElement{NewMarkup("standalone", "br", ConvertMapToOptions(map[string]any{"id": "line"}), nil)}),
	)
	message := NewSelectMessage(nil, []VariableRef{*selector}, []Variant{*one, *fallback}, "items")

	var selectors []string
	var keys []string
	var variants int
	var markupContexts []string
	Visit(message, &Visitor{
		Value: func(value any, context string, position string) {
			if context == "selector" {
				selectors = append(selectors, fmt.Sprintf("%s:%T", position, value))
			}
		},
		Key: func(key VariantKey, index int, keysInVariant []VariantKey) {
			keys = append(keys, fmt.Sprintf("%d/%d:%s", index, len(keysInVariant), key.String()))
		},
		Variant: func(variant *Variant) func() {
			variants++
			return nil
		},
		Markup: func(markup *Markup, context string) func() {
			markupContexts = append(markupContexts, fmt.Sprintf("%s:%s", context, markup.Name()))
			return nil
		},
	})

	require.Len(t, selectors, 1)
	assert.Equal(t, "arg:datamodel.VariableRef", selectors[0])
	assert.Equal(t, []string{"0/1:one", "0/1:*"}, keys)
	assert.Equal(t, 2, variants)
	assert.Equal(t, []string{"placeholder:br"}, markupContexts)
}

func TestVisit_NilVisitor(t *testing.T) {
	t.Parallel()

	require.NotPanics(t, func() {
		Visit(NewPatternMessage(nil, NewPattern(nil), ""), nil)
	})
}
