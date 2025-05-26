package resolve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

func TestNewContext(t *testing.T) {
	locales := []string{"en", "fr"}
	funcs := map[string]functions.MessageFunction{
		"test": func(ctx functions.MessageFunctionContext, options map[string]interface{}, operand interface{}) messagevalue.MessageValue {
			return messagevalue.NewStringValue("test", "en", "test")
		},
	}
	scope := map[string]interface{}{
		"name": "Alice",
	}

	ctx := NewContext(locales, funcs, scope, nil)

	assert.NotNil(t, ctx)
	assert.Equal(t, locales, ctx.Locales)
	assert.Equal(t, "best fit", ctx.LocaleMatcher)
	assert.Equal(t, "Alice", ctx.Scope["name"])
	assert.NotNil(t, ctx.Functions["test"])
	assert.NotNil(t, ctx.LocalVars)
}

func TestNewContextWithNils(t *testing.T) {
	ctx := NewContext(nil, nil, nil, nil)

	assert.NotNil(t, ctx)
	assert.NotNil(t, ctx.Functions)
	assert.NotNil(t, ctx.Scope)
	assert.NotNil(t, ctx.LocalVars)
}

func TestContextClone(t *testing.T) {
	original := NewContext(
		[]string{"en"},
		map[string]functions.MessageFunction{"test": func(ctx functions.MessageFunctionContext, options map[string]interface{}, operand interface{}) messagevalue.MessageValue {
			return messagevalue.NewStringValue("test", "en", "test")
		}},
		map[string]interface{}{"name": "Alice"},
		nil,
	)

	// Add a local var
	mv := messagevalue.NewStringValue("test", "en", "test")
	original.LocalVars[mv] = true

	cloned := original.Clone()

	// Should have separate scope and local vars
	assert.Equal(t, original.Scope, cloned.Scope)
	assert.Equal(t, original.LocalVars, cloned.LocalVars)

	// Modify original - should not affect clone
	original.Scope["age"] = 30
	assert.NotContains(t, cloned.Scope, "age")

	// Should share immutable references
	assert.Equal(t, original.Functions, cloned.Functions)
	assert.Equal(t, original.Locales, cloned.Locales)
}

func TestContextCloneWithScope(t *testing.T) {
	original := NewContext(
		[]string{"en"},
		nil,
		map[string]interface{}{"name": "Alice"},
		nil,
	)

	newScope := map[string]interface{}{
		"age":  30,
		"city": "Paris",
	}

	cloned := original.CloneWithScope(newScope)

	// Should have original scope plus new scope
	assert.Equal(t, "Alice", cloned.Scope["name"])
	assert.Equal(t, 30, cloned.Scope["age"])
	assert.Equal(t, "Paris", cloned.Scope["city"])

	// Original should be unchanged
	assert.NotContains(t, original.Scope, "age")
	assert.NotContains(t, original.Scope, "city")
}
