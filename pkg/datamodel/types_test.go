package datamodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPatternMessage(t *testing.T) {
	// Create test data
	literal := NewLiteral("hello")
	expr := NewExpression(literal, nil, nil)
	decl := NewLocalDeclaration("test", expr)
	textElem := NewTextElement("Hello ")
	exprElem := expr
	pattern := NewPattern([]PatternElement{textElem, exprElem})

	msg := NewPatternMessage([]Declaration{decl}, pattern, "test comment")

	assert.Equal(t, "message", msg.Type())
	assert.Len(t, msg.Declarations(), 1)
	assert.Equal(t, "test comment", msg.Comment())
	assert.Len(t, msg.Pattern().Elements(), 2)
}

func TestSelectMessage(t *testing.T) {
	// Create test data
	selector := NewVariableRef("count")
	literal1 := NewLiteral("one")
	catchall := NewCatchallKey("")

	variant1 := NewVariant([]VariantKey{literal1}, NewPattern([]PatternElement{NewTextElement("One item")}))
	variant2 := NewVariant([]VariantKey{catchall}, NewPattern([]PatternElement{NewTextElement("Many items")}))

	msg := NewSelectMessage(nil, []VariableRef{*selector}, []Variant{*variant1, *variant2}, "")

	assert.Equal(t, "select", msg.Type())
	assert.Len(t, msg.Selectors(), 1)
	assert.Equal(t, "count", msg.Selectors()[0].Name())
	assert.Len(t, msg.Variants(), 2)
}

func TestDeclarations(t *testing.T) {
	literal := NewLiteral("test")
	varRef := NewVariableRef("testVar")
	expr := NewExpression(literal, nil, nil)
	varExpr := NewExpression(varRef, nil, nil)

	t.Run("InputDeclaration", func(t *testing.T) {
		decl := NewInputDeclaration("input1", ConvertExpressionToVariableRefExpression(varExpr))

		assert.Equal(t, "input", decl.Type())
		assert.Equal(t, "input1", decl.Name())
		assert.NotNil(t, decl.Value())
	})

	t.Run("LocalDeclaration", func(t *testing.T) {
		decl := NewLocalDeclaration("local1", expr)

		assert.Equal(t, "local", decl.Type())
		assert.Equal(t, "local1", decl.Name())
		assert.Equal(t, expr, decl.Value())
	})
}

func TestExpression(t *testing.T) {
	literal := NewLiteral("test")
	funcRef := NewFunctionRef("number", ConvertMapToOptions(map[string]interface{}{"style": "decimal"}))
	attrs := ConvertMapToAttributes(map[string]interface{}{"id": "test"})

	expr := NewExpression(literal, funcRef, attrs)

	assert.Equal(t, "expression", expr.Type())
	assert.Equal(t, literal, expr.Arg())
	assert.Equal(t, funcRef, expr.FunctionRef())
	assert.Equal(t, attrs, expr.Attributes())
}

func TestLiteral(t *testing.T) {
	lit := NewLiteral("hello world")

	assert.Equal(t, "literal", lit.Type())
	assert.Equal(t, "hello world", lit.Value())
	assert.Equal(t, "hello world", lit.String())
}

func TestVariableRef(t *testing.T) {
	vr := NewVariableRef("userName")

	assert.Equal(t, "variable", vr.Type())
	assert.Equal(t, "userName", vr.Name())
}

func TestFunctionRef(t *testing.T) {
	options := ConvertMapToOptions(map[string]interface{}{
		"style":    "currency",
		"currency": "USD",
	})
	fr := NewFunctionRef("number", options)

	assert.Equal(t, "function", fr.Type())
	assert.Equal(t, "number", fr.Name())
	assert.Equal(t, options, fr.Options())
}

func TestMarkup(t *testing.T) {
	options := ConvertMapToOptions(map[string]interface{}{"href": "https://example.com"})
	attrs := ConvertMapToAttributes(map[string]interface{}{"target": "_blank"})

	markup := NewMarkup("open", "link", options, attrs)

	assert.Equal(t, "markup", markup.Type())
	assert.Equal(t, "open", markup.Kind())
	assert.Equal(t, "link", markup.Name())
	assert.Equal(t, options, markup.Options())
	assert.Equal(t, attrs, markup.Attributes())
}

func TestCatchallKey(t *testing.T) {
	t.Run("with value", func(t *testing.T) {
		ck := NewCatchallKey("default")

		assert.Equal(t, "*", ck.Type())
		assert.Equal(t, "default", ck.Value())
		assert.Equal(t, "default", ck.String())
	})

	t.Run("without value", func(t *testing.T) {
		ck := NewCatchallKey("")

		assert.Equal(t, "*", ck.Type())
		assert.Equal(t, "", ck.Value())
		assert.Equal(t, "*", ck.String())
	})
}

func TestPattern(t *testing.T) {
	textElem := NewTextElement("Hello ")
	literal := NewLiteral("world")
	expr := NewExpression(literal, nil, nil)

	pattern := NewPattern([]PatternElement{textElem, expr})

	assert.Len(t, pattern.Elements(), 2)
	assert.Equal(t, "text", pattern.Elements()[0].Type())
	assert.Equal(t, "expression", pattern.Elements()[1].Type())
}

func TestTextElement(t *testing.T) {
	te := NewTextElement("Hello, world!")

	assert.Equal(t, "text", te.Type())
	assert.Equal(t, "Hello, world!", te.Value())
}

func TestVariant(t *testing.T) {
	literal := NewLiteral("one")
	catchall := NewCatchallKey("")
	keys := []VariantKey{literal, catchall}
	pattern := NewPattern([]PatternElement{NewTextElement("One item")})

	variant := NewVariant(keys, pattern)

	assert.Len(t, variant.Keys(), 2)
	assert.Equal(t, literal, variant.Keys()[0])
	assert.Equal(t, catchall, variant.Keys()[1])
	assert.Equal(t, pattern, variant.Value())
}

func TestNilHandling(t *testing.T) {
	t.Run("PatternMessage with nil declarations", func(t *testing.T) {
		pattern := NewPattern(nil)
		msg := NewPatternMessage(nil, pattern, "")

		assert.NotNil(t, msg.Declarations())
		assert.Len(t, msg.Declarations(), 0)
	})

	t.Run("SelectMessage with nil arrays", func(t *testing.T) {
		msg := NewSelectMessage(nil, nil, nil, "")

		assert.NotNil(t, msg.Declarations())
		assert.NotNil(t, msg.Selectors())
		assert.NotNil(t, msg.Variants())
		assert.Len(t, msg.Declarations(), 0)
		assert.Len(t, msg.Selectors(), 0)
		assert.Len(t, msg.Variants(), 0)
	})

	t.Run("Pattern with nil elements", func(t *testing.T) {
		pattern := NewPattern(nil)

		assert.NotNil(t, pattern.Elements())
		assert.Len(t, pattern.Elements(), 0)
	})

	t.Run("Variant with nil keys", func(t *testing.T) {
		pattern := NewPattern(nil)
		variant := NewVariant(nil, pattern)

		assert.NotNil(t, variant.Keys())
		assert.Len(t, variant.Keys(), 0)
	})
}
