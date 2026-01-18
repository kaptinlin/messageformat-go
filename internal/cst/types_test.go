// Package cst provides comprehensive tests for types.go accessor methods
package cst

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleMessage_Declarations(t *testing.T) {
	msg := NewSimpleMessage(Pattern{}, nil)

	// SimpleMessage should have empty declarations
	decls := msg.Declarations()
	assert.Empty(t, decls)
}

func TestInputDeclaration_Start(t *testing.T) {
	keyword := NewSyntax(0, 6, ".input")
	value := NewLiteral(7, 12, false, nil, "test", nil)
	decl := NewInputDeclaration(0, 12, keyword, value)

	assert.Equal(t, 0, decl.Start())
}

func TestLocalDeclaration_Start(t *testing.T) {
	keyword := NewSyntax(0, 6, ".local")
	target := NewVariableRef(7, 11, NewSyntax(7, 8, "$"), "var")
	equals := NewSyntax(12, 13, "=")
	value := NewLiteral(14, 18, false, nil, "test", nil)
	decl := NewLocalDeclaration(0, 18, keyword, target, &equals, value)

	assert.Equal(t, 0, decl.Start())
}

func TestVariant_Start(t *testing.T) {
	key := NewLiteral(0, 3, false, nil, "one", nil)
	pattern := NewPattern(4, 10, []Node{}, nil)
	variant := NewVariant(0, 10, []Key{key}, *pattern)

	assert.Equal(t, 0, variant.Start())
}

func TestCatchallKey_Type(t *testing.T) {
	key := NewCatchallKey(0, 1)
	assert.Equal(t, "*", key.Type())
}

func TestIdentifier_Name_EdgeCases(t *testing.T) {
	t.Run("empty identifier", func(t *testing.T) {
		id := Identifier{}
		name := id.Name()
		assert.Nil(t, name)
	})

	t.Run("two-part identifier", func(t *testing.T) {
		id := Identifier{
			NewSyntax(0, 2, "ns"),
			NewSyntax(2, 3, ":"),
		}
		name := id.Name()
		assert.Nil(t, name)
	})

	t.Run("four-part identifier", func(t *testing.T) {
		id := Identifier{
			NewSyntax(0, 2, "a"),
			NewSyntax(2, 3, ":"),
			NewSyntax(3, 4, "b"),
			NewSyntax(4, 5, ":"),
		}
		name := id.Name()
		assert.Nil(t, name)
	})
}
