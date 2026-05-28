package datamodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type unknownCloneMessage struct{}

func (*unknownCloneMessage) Type() string                { return "custom" }
func (*unknownCloneMessage) Declarations() []Declaration { return nil }
func (*unknownCloneMessage) Comment() string             { return "custom" }

func TestCloneMessageHandlesNilAndUnknownMessages(t *testing.T) {
	t.Parallel()

	assert.Nil(t, CloneMessage(nil))

	message := &unknownCloneMessage{}
	assert.Same(t, message, CloneMessage(message))
}

func TestCloneMessageDeepCopiesPatternMessage(t *testing.T) {
	t.Parallel()

	inputDecl := NewInputDeclaration(
		"user",
		NewVariableRefExpression(
			NewVariableRef("user"),
			NewFunctionRef("string", ConvertMapToOptions(map[string]any{"fallback": "guest"})),
			ConvertMapToAttributes(map[string]any{"required": true}),
		),
	)
	localDecl := NewLocalDeclaration(
		"total",
		NewExpression(
			NewLiteral("42"),
			NewFunctionRef("number", ConvertMapToOptions(map[string]any{"minimumFractionDigits": "2"})),
			ConvertMapToAttributes(map[string]any{"title": "before"}),
		),
	)
	expr := NewExpression(
		NewVariableRef("total"),
		NewFunctionRef("number", ConvertMapToOptions(map[string]any{
			"currency": NewVariableRef("currencyCode"),
			"style":    "currency",
		})),
		ConvertMapToAttributes(map[string]any{"id": "price", "visible": true}),
	)
	markup := mustMarkup(
		t,
		MarkupOpen,
		"strong",
		ConvertMapToOptions(map[string]any{"class": "price"}),
		ConvertMapToAttributes(map[string]any{"data-value": "total", "hidden": true}),
	)
	message := NewPatternMessage(
		[]Declaration{inputDecl, localDecl},
		NewPattern([]PatternElement{NewTextElement("Total: "), expr, markup}),
		"invoice total",
	)

	cloned := requirePatternClone(t, CloneMessage(message))
	requirePatternMessageShape(t, cloned)

	message.Declarations()[0] = NewLocalDeclaration("changed", NewExpression(NewLiteral("changed"), nil, nil))
	message.Pattern()[0] = NewTextElement("Changed: ")
	inputDecl.Value().FunctionRef().Options()["fallback"] = NewLiteral("anonymous")
	inputDecl.Value().Attributes()["required"] = NewLiteral("false")
	localDecl.Value().FunctionRef().Options()["minimumFractionDigits"] = NewLiteral("0")
	localDecl.Value().Attributes()["title"] = NewLiteral("after")
	expr.FunctionRef().Options()["style"] = NewLiteral("decimal")
	expr.FunctionRef().Options()["currency"] = NewVariableRef("otherCurrency")
	expr.Attributes()["id"] = NewLiteral("changed")
	markup.Options()["class"] = NewLiteral("changed")
	markup.Attributes()["data-value"] = NewLiteral("changed")

	requirePatternMessageShape(t, cloned)
}

func TestCloneMessageDeepCopiesSelectMessage(t *testing.T) {
	t.Parallel()

	inputDecl := NewInputDeclaration("count", NewVariableRefExpression(NewVariableRef("count"), nil, nil))
	oneVariant := NewVariant(
		[]VariantKey{NewLiteral("one")},
		NewPattern([]PatternElement{NewTextElement("one item")}),
	)
	fallbackExpr := NewExpression(
		NewVariableRef("count"),
		NewFunctionRef("number", ConvertMapToOptions(map[string]any{"select": "ordinal"})),
		nil,
	)
	fallbackVariant := NewVariant(
		[]VariantKey{NewCatchallKey("")},
		NewPattern([]PatternElement{fallbackExpr}),
	)
	message := NewSelectMessage(
		[]Declaration{inputDecl},
		[]VariableRef{*NewVariableRef("count")},
		[]Variant{*oneVariant, *fallbackVariant},
		"count selector",
	)

	cloned := requireSelectClone(t, CloneMessage(message))
	requireSelectMessageShape(t, cloned)

	message.Selectors()[0] = *NewVariableRef("other")
	variants := message.Variants()
	variants[0].Keys()[0] = NewLiteral("changed")
	variants[0].Value()[0] = NewTextElement("changed item")
	variants[1].Keys()[0] = NewCatchallKey("changed")
	fallbackExpr.FunctionRef().Options()["select"] = NewLiteral("cardinal")

	requireSelectMessageShape(t, cloned)
}

func requirePatternClone(t *testing.T, message Message) *PatternMessage {
	t.Helper()

	cloned, ok := message.(*PatternMessage)
	require.True(t, ok)
	return cloned
}

func requireSelectClone(t *testing.T, message Message) *SelectMessage {
	t.Helper()

	cloned, ok := message.(*SelectMessage)
	require.True(t, ok)
	return cloned
}

func requirePatternMessageShape(t *testing.T, message *PatternMessage) {
	t.Helper()

	assert.Equal(t, "message", message.Type())
	assert.Equal(t, "invoice total", message.Comment())

	declarations := message.Declarations()
	require.Len(t, declarations, 2)
	inputDecl := requireInputDeclaration(t, declarations[0])
	assert.Equal(t, "user", inputDecl.Name())
	assert.Equal(t, "user", inputDecl.Value().Arg().Name())
	assert.Equal(t, "guest", inputDecl.Value().FunctionRef().Options()["fallback"].String())
	assert.Equal(t, "true", inputDecl.Value().Attributes()["required"].String())

	localDecl := requireLocalDeclaration(t, declarations[1])
	assert.Equal(t, "total", localDecl.Name())
	assert.Equal(t, "42", localDecl.Value().Arg().String())
	assert.Equal(t, "2", localDecl.Value().FunctionRef().Options()["minimumFractionDigits"].String())
	assert.Equal(t, "before", localDecl.Value().Attributes()["title"].String())

	pattern := message.Pattern()
	require.Len(t, pattern, 3)
	requireTextValue(t, "Total: ", pattern[0])

	expr := requireExpressionElement(t, pattern[1])
	variableRef, ok := expr.Arg().(*VariableRef)
	require.True(t, ok)
	assert.Equal(t, "total", variableRef.Name())
	assert.Equal(t, "currency", expr.FunctionRef().Options()["style"].String())
	assert.Equal(t, "currencyCode", expr.FunctionRef().Options()["currency"].String())
	assert.Equal(t, "price", expr.Attributes()["id"].String())
	assert.Equal(t, "true", expr.Attributes()["visible"].String())

	clonedMarkup, ok := pattern[2].(*Markup)
	require.True(t, ok)
	assert.Equal(t, MarkupOpen, clonedMarkup.Kind())
	assert.Equal(t, "strong", clonedMarkup.Name())
	assert.Equal(t, "price", clonedMarkup.Options()["class"].String())
	assert.Equal(t, "total", clonedMarkup.Attributes()["data-value"].String())
	assert.Equal(t, "true", clonedMarkup.Attributes()["hidden"].String())
}

func requireSelectMessageShape(t *testing.T, message *SelectMessage) {
	t.Helper()

	assert.Equal(t, "select", message.Type())
	assert.Equal(t, "count selector", message.Comment())

	declarations := message.Declarations()
	require.Len(t, declarations, 1)
	assert.Equal(t, "count", requireInputDeclaration(t, declarations[0]).Name())

	selectors := message.Selectors()
	require.Len(t, selectors, 1)
	assert.Equal(t, "count", selectors[0].Name())

	variants := message.Variants()
	require.Len(t, variants, 2)
	require.Len(t, variants[0].Keys(), 1)
	assert.Equal(t, "one", variants[0].Keys()[0].String())
	requireTextValue(t, "one item", variants[0].Value()[0])

	require.Len(t, variants[1].Keys(), 1)
	assert.Equal(t, "*", variants[1].Keys()[0].String())
	expr := requireExpressionElement(t, variants[1].Value()[0])
	assert.Equal(t, "ordinal", expr.FunctionRef().Options()["select"].String())
}

func requireInputDeclaration(t *testing.T, declaration Declaration) *InputDeclaration {
	t.Helper()

	input, ok := declaration.(*InputDeclaration)
	require.True(t, ok)
	return input
}

func requireLocalDeclaration(t *testing.T, declaration Declaration) *LocalDeclaration {
	t.Helper()

	local, ok := declaration.(*LocalDeclaration)
	require.True(t, ok)
	return local
}

func requireExpressionElement(t *testing.T, element PatternElement) *Expression {
	t.Helper()

	expression, ok := element.(*Expression)
	require.True(t, ok)
	return expression
}
