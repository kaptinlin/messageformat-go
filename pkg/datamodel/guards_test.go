package datamodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCatchallKey(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"valid catchall key", NewCatchallKey(""), true},
		{"catchall key with value", NewCatchallKey("default"), true},
		{"literal", NewLiteral("test"), false},
		{"nil", nil, false},
		{"string", "test", false},
		{"number", 42, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCatchallKey(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"valid expression", NewExpression(nil, nil, nil), true},
		{"expression with literal", NewExpression(NewLiteral("test"), nil, nil), true},
		{"literal", NewLiteral("test"), false},
		{"nil", nil, false},
		{"string", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsExpression(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsFunctionRef(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"valid function ref", NewFunctionRef("number", nil), true},
		{"function ref with options", NewFunctionRef("number", ConvertMapToOptions(map[string]interface{}{"style": "decimal"})), true},
		{"literal", NewLiteral("test"), false},
		{"nil", nil, false},
		{"string", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsFunctionRef(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsLiteral(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"valid literal", NewLiteral("test"), true},
		{"empty literal", NewLiteral(""), true},
		{"expression", NewExpression(nil, nil, nil), false},
		{"nil", nil, false},
		{"string", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsLiteral(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsMarkup(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"valid markup", NewMarkup("open", "b", nil, nil), true},
		{"standalone markup", NewMarkup("standalone", "br", nil, nil), true},
		{"close markup", NewMarkup("close", "b", nil, nil), true},
		{"literal", NewLiteral("test"), false},
		{"nil", nil, false},
		{"string", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMarkup(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsMessage(t *testing.T) {
	pattern := NewPattern([]PatternElement{NewTextElement("Hello")})
	patternMsg := NewPatternMessage(nil, pattern, "")
	selectMsg := NewSelectMessage(nil, nil, nil, "")

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"pattern message", patternMsg, true},
		{"select message", selectMsg, true},
		{"literal", NewLiteral("test"), false},
		{"nil", nil, false},
		{"string", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMessage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsPatternMessage(t *testing.T) {
	pattern := NewPattern([]PatternElement{NewTextElement("Hello")})
	patternMsg := NewPatternMessage(nil, pattern, "")
	selectMsg := NewSelectMessage(nil, nil, nil, "")

	tests := []struct {
		name     string
		input    Message
		expected bool
	}{
		{"pattern message", patternMsg, true},
		{"select message", selectMsg, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPatternMessage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsSelectMessage(t *testing.T) {
	pattern := NewPattern([]PatternElement{NewTextElement("Hello")})
	patternMsg := NewPatternMessage(nil, pattern, "")
	selectMsg := NewSelectMessage(nil, nil, nil, "")

	tests := []struct {
		name     string
		input    Message
		expected bool
	}{
		{"pattern message", patternMsg, false},
		{"select message", selectMsg, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSelectMessage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsVariableRef(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"valid variable ref", NewVariableRef("test"), true},
		{"empty name variable ref", NewVariableRef(""), true},
		{"literal", NewLiteral("test"), false},
		{"nil", nil, false},
		{"string", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsVariableRef(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsInputDeclaration(t *testing.T) {
	expr := NewExpression(NewLiteral("test"), nil, nil)
	inputDecl := NewInputDeclaration("input1", ConvertExpressionToVariableRefExpression(expr))
	localDecl := NewLocalDeclaration("local1", expr)

	tests := []struct {
		name     string
		input    Declaration
		expected bool
	}{
		{"input declaration", inputDecl, true},
		{"local declaration", localDecl, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsInputDeclaration(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsLocalDeclaration(t *testing.T) {
	expr := NewExpression(NewLiteral("test"), nil, nil)
	inputDecl := NewInputDeclaration("input1", ConvertExpressionToVariableRefExpression(expr))
	localDecl := NewLocalDeclaration("local1", expr)

	tests := []struct {
		name     string
		input    Declaration
		expected bool
	}{
		{"input declaration", inputDecl, false},
		{"local declaration", localDecl, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsLocalDeclaration(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsTextElement(t *testing.T) {
	textElem := NewTextElement("Hello")
	expr := NewExpression(NewLiteral("test"), nil, nil)

	tests := []struct {
		name     string
		input    PatternElement
		expected bool
	}{
		{"text element", textElem, true},
		{"expression element", expr, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTextElement(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNilInputs(t *testing.T) {
	t.Run("IsPatternMessage with nil", func(t *testing.T) {
		result := IsPatternMessage(nil)
		assert.False(t, result)
	})

	t.Run("IsSelectMessage with nil", func(t *testing.T) {
		result := IsSelectMessage(nil)
		assert.False(t, result)
	})

	t.Run("IsInputDeclaration with nil", func(t *testing.T) {
		result := IsInputDeclaration(nil)
		assert.False(t, result)
	})

	t.Run("IsLocalDeclaration with nil", func(t *testing.T) {
		result := IsLocalDeclaration(nil)
		assert.False(t, result)
	})

	t.Run("IsTextElement with nil", func(t *testing.T) {
		result := IsTextElement(nil)
		assert.False(t, result)
	})
}

func TestIsVariantKey(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"literal key", NewLiteral("one"), true},
		{"catchall key", NewCatchallKey(""), true},
		{"catchall key with value", NewCatchallKey("default"), true},
		{"variable ref", NewVariableRef("test"), false},
		{"nil", nil, false},
		{"string", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsVariantKey(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsPatternElement(t *testing.T) {
	textElem := NewTextElement("Hello")
	expr := NewExpression(NewLiteral("test"), nil, nil)
	markup := NewMarkup("open", "b", nil, nil)

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"text element", textElem, true},
		{"expression element", expr, true},
		{"markup element", markup, true},
		{"literal (not pattern element)", NewLiteral("test"), false},
		{"nil", nil, false},
		{"string", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPatternElement(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsNode(t *testing.T) {
	literal := NewLiteral("test")
	varRef := NewVariableRef("test")
	funcRef := NewFunctionRef("number", nil)
	expr := NewExpression(literal, nil, nil)
	markup := NewMarkup("open", "b", nil, nil)
	catchall := NewCatchallKey("")
	inputDecl := NewInputDeclaration("input1", ConvertExpressionToVariableRefExpression(NewExpression(varRef, nil, nil)))
	localDecl := NewLocalDeclaration("local1", expr)

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"literal", literal, true},
		{"variable ref", varRef, true},
		{"function ref", funcRef, true},
		{"expression", expr, true},
		{"markup", markup, true},
		{"catchall key", catchall, true},
		{"input declaration", inputDecl, true},
		{"local declaration", localDecl, true},
		{"nil", nil, false},
		{"string", "test", false},
		{"number", 42, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNode(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsBooleanAttribute(t *testing.T) {
	boolAttr := NewBooleanAttribute()
	literal := NewLiteral("test")

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"boolean attribute", boolAttr, true},
		{"literal", literal, false},
		{"nil", nil, false},
		{"string", "test", false},
		{"boolean value", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBooleanAttribute(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsVariableRefExpression(t *testing.T) {
	varRef := NewVariableRef("test")
	varRefExpr := NewVariableRefExpression(varRef, nil, nil)
	regularExpr := NewExpression(NewLiteral("test"), nil, nil)

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"variable ref expression", varRefExpr, true},
		{"regular expression", regularExpr, false},
		{"literal", NewLiteral("test"), false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsVariableRefExpression(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
