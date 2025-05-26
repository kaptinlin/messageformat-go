package resolve

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
)

func TestGetValue(t *testing.T) {
	scope := map[string]interface{}{
		"name":    "Alice",
		"age":     30,
		"user.id": 123,
		"settings": map[string]interface{}{
			"theme": "dark",
			"lang":  "en",
		},
	}

	// Direct lookup
	assert.Equal(t, "Alice", getValue(scope, "name"))
	assert.Equal(t, 30, getValue(scope, "age"))

	// Dotted property access
	assert.Equal(t, "dark", getValue(scope, "settings.theme"))
	assert.Equal(t, "en", getValue(scope, "settings.lang"))

	// Non-existent key
	assert.Nil(t, getValue(scope, "nonexistent"))

	// Non-scope value
	assert.Nil(t, getValue("not a scope", "name"))
	assert.Nil(t, getValue(nil, "name"))
}

func TestLookupVariableRef(t *testing.T) {
	ctx := NewContext(
		[]string{"en"},
		map[string]functions.MessageFunction{
			"string": functions.StringFunction,
		},
		map[string]interface{}{
			"name": "Alice",
			"age":  30,
		},
		nil,
	)

	// Existing variable
	ref := datamodel.NewVariableRef("name")
	value := lookupVariableRef(ctx, ref)
	assert.Equal(t, "Alice", value)

	// Non-existing variable
	ref2 := datamodel.NewVariableRef("nonexistent")
	value2 := lookupVariableRef(ctx, ref2)
	assert.Nil(t, value2)
}

func TestResolveVariableRef(t *testing.T) {
	ctx := NewContext(
		[]string{"en"},
		map[string]functions.MessageFunction{
			"string": functions.StringFunction,
			"number": functions.NumberFunction,
		},
		map[string]interface{}{
			"name": "Alice",
			"age":  30,
		},
		nil,
	)

	// String variable
	ref := datamodel.NewVariableRef("name")
	result := ResolveVariableRef(ctx, ref)
	assert.Equal(t, "string", result.Type())

	// Number variable
	ref2 := datamodel.NewVariableRef("age")
	result2 := ResolveVariableRef(ctx, ref2)
	assert.Equal(t, "number", result2.Type())

	// Non-existing variable (should return fallback)
	ref3 := datamodel.NewVariableRef("nonexistent")
	result3 := ResolveVariableRef(ctx, ref3)
	assert.Equal(t, "fallback", result3.Type())
}

func TestUnresolvedExpression(t *testing.T) {
	expr := datamodel.NewExpression(
		datamodel.NewLiteral("test"),
		nil,
		nil,
	)
	scope := map[string]interface{}{"key": "value"}

	unresolved := NewUnresolvedExpression(expr, scope)

	assert.Equal(t, expr, unresolved.Expression)
	assert.Equal(t, scope, unresolved.Scope)
}

func TestIsScope(t *testing.T) {
	// Valid scopes
	assert.True(t, isScope(map[string]interface{}{}))
	assert.True(t, isScope(map[interface{}]interface{}{}))
	assert.True(t, isScope(struct{}{}))

	// Invalid scopes
	assert.False(t, isScope(nil))
	assert.False(t, isScope("string"))
	assert.False(t, isScope(123))
	assert.False(t, isScope([]string{}))
}

func TestGetFirstLocale(t *testing.T) {
	// With locales
	assert.Equal(t, "en", getFirstLocale([]string{"en", "fr"}))
	assert.Equal(t, "fr", getFirstLocale([]string{"fr"}))

	// Without locales
	assert.Equal(t, "en", getFirstLocale([]string{}))
	assert.Equal(t, "en", getFirstLocale(nil))
}
