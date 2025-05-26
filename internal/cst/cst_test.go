// Package cst provides tests for CST parsing
package cst

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCST_SimpleMessage(t *testing.T) {
	source := "Hello, world!"
	msg := ParseCST(source, false)

	assert.Equal(t, "simple", msg.Type())
	assert.Empty(t, msg.Errors())

	simple, ok := msg.(*SimpleMessage)
	assert.True(t, ok)
	assert.NotNil(t, simple.Pattern())
}

func TestParseCST_ComplexMessage(t *testing.T) {
	source := "{{Hello, world!}}"
	msg := ParseCST(source, false)

	assert.Equal(t, "complex", msg.Type())
	assert.Empty(t, msg.Errors())

	complex, ok := msg.(*ComplexMessage)
	assert.True(t, ok)
	assert.NotNil(t, complex.Pattern())
	assert.Empty(t, complex.Declarations())
}

func TestParseCST_SelectMessage(t *testing.T) {
	source := ".match $count one {{one item}} * {{many items}}"
	msg := ParseCST(source, false)

	assert.Equal(t, "select", msg.Type())

	select_, ok := msg.(*SelectMessage)
	assert.True(t, ok)
	assert.Len(t, select_.Selectors(), 1)
	assert.Len(t, select_.Variants(), 2)
}

func TestParseCST_WithDeclarations(t *testing.T) {
	source := ".input {$count :number} .local $formatted = {$count :number} {{You have {$formatted} items}}"
	msg := ParseCST(source, false)

	assert.Equal(t, "complex", msg.Type())

	complex, ok := msg.(*ComplexMessage)
	assert.True(t, ok)
	assert.Len(t, complex.Declarations(), 2)
}

func TestParseText(t *testing.T) {
	ctx := NewParseContext("Hello, world!", false)
	text := ParseText(ctx, 0)

	assert.Equal(t, "text", text.Type())
	assert.Equal(t, "Hello, world!", text.Value())
	assert.Equal(t, 0, text.Start())
	assert.Equal(t, 13, text.End())
}

func TestParseLiteral_Unquoted(t *testing.T) {
	ctx := NewParseContext("hello", false)
	literal := ParseLiteral(ctx, 0, true)

	assert.NotNil(t, literal)
	assert.Equal(t, "literal", literal.Type())
	assert.Equal(t, "hello", literal.Value())
	assert.False(t, literal.Quoted())
}

func TestParseLiteral_Quoted(t *testing.T) {
	ctx := NewParseContext("|hello world|", false)
	literal := ParseLiteral(ctx, 0, true)

	assert.NotNil(t, literal)
	assert.Equal(t, "literal", literal.Type())
	assert.Equal(t, "hello world", literal.Value())
	assert.True(t, literal.Quoted())
}

func TestParseVariable(t *testing.T) {
	ctx := NewParseContext("$count", false)
	variable := ParseVariable(ctx, 0)

	assert.Equal(t, "variable", variable.Type())
	assert.Equal(t, "count", variable.Name())
	assert.Equal(t, 0, variable.Start())
	assert.Equal(t, 6, variable.End())
}

func TestWhitespaces(t *testing.T) {
	ws := Whitespaces("  hello", 0)
	assert.True(t, ws.HasWS)
	assert.Equal(t, 2, ws.End)

	ws = Whitespaces("hello", 0)
	assert.False(t, ws.HasWS)
	assert.Equal(t, 0, ws.End)
}

func TestParseNameValue(t *testing.T) {
	name := ParseNameValue("hello", 0)
	assert.NotNil(t, name)
	assert.Equal(t, "hello", name.Value)
	assert.Equal(t, 5, name.End)

	name = ParseNameValue("123invalid", 0)
	assert.Nil(t, name)
}

func TestParseUnquotedLiteralValue(t *testing.T) {
	value := ParseUnquotedLiteralValue("hello", 0)
	assert.Equal(t, "hello", value)

	value = ParseUnquotedLiteralValue("hello world", 0)
	assert.Equal(t, "hello", value) // stops at whitespace

	value = ParseUnquotedLiteralValue("", 0)
	assert.Equal(t, "", value)
}
