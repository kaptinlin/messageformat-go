package functions

import (
	"testing"

	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
	"github.com/stretchr/testify/assert"
)

func TestNewFunctionRegistry(t *testing.T) {
	registry := NewFunctionRegistry()

	// Check that default functions are registered
	assert.True(t, len(registry.List()) >= 3)

	// Check specific functions
	_, exists := registry.Get("number")
	assert.True(t, exists)

	_, exists = registry.Get("integer")
	assert.True(t, exists)

	_, exists = registry.Get("string")
	assert.True(t, exists)
}

func TestNewFunctionRegistryWithDraft(t *testing.T) {
	registry := NewFunctionRegistryWithDraft()

	// Check that both default and draft functions are registered
	assert.True(t, len(registry.List()) >= 9)

	// Check default functions
	_, exists := registry.Get("number")
	assert.True(t, exists)

	// Check draft functions
	_, exists = registry.Get("currency")
	assert.True(t, exists)

	_, exists = registry.Get("date")
	assert.True(t, exists)
}

func TestFunctionRegistryRegister(t *testing.T) {
	registry := NewFunctionRegistry()

	// Register a custom function
	customFunc := func(ctx MessageFunctionContext, options map[string]interface{}, operand interface{}) messagevalue.MessageValue {
		return messagevalue.NewStringValue("custom", "en", "test")
	}

	registry.Register("custom", customFunc)

	// Check that it's registered
	fn, exists := registry.Get("custom")
	assert.True(t, exists)
	assert.NotNil(t, fn)

	// Check that it's in the list
	names := registry.List()
	assert.Contains(t, names, "custom")
}

func TestFunctionRegistryGet(t *testing.T) {
	registry := NewFunctionRegistry()

	// Test existing function
	fn, exists := registry.Get("string")
	assert.True(t, exists)
	assert.NotNil(t, fn)

	// Test non-existing function
	_, exists = registry.Get("nonexistent")
	assert.False(t, exists)
}

func TestFunctionRegistryClone(t *testing.T) {
	registry := NewFunctionRegistry()

	// Add a custom function
	customFunc := func(ctx MessageFunctionContext, options map[string]interface{}, operand interface{}) messagevalue.MessageValue {
		return messagevalue.NewStringValue("custom", "en", "test")
	}
	registry.Register("custom", customFunc)

	// Clone the registry
	cloned := registry.Clone()

	// Check that cloned registry has the same functions
	assert.Equal(t, len(registry.List()), len(cloned.List()))

	_, exists := cloned.Get("custom")
	assert.True(t, exists)

	// Modify original registry
	registry.Register("new", customFunc)

	// Check that cloned registry is not affected
	_, exists = cloned.Get("new")
	assert.False(t, exists)
}

func TestFunctionRegistryMerge(t *testing.T) {
	registry1 := NewFunctionRegistry()
	registry2 := NewFunctionRegistry()

	// Add different functions to each registry
	customFunc1 := func(ctx MessageFunctionContext, options map[string]interface{}, operand interface{}) messagevalue.MessageValue {
		return messagevalue.NewStringValue("custom1", "en", "test")
	}
	customFunc2 := func(ctx MessageFunctionContext, options map[string]interface{}, operand interface{}) messagevalue.MessageValue {
		return messagevalue.NewStringValue("custom2", "en", "test")
	}

	registry1.Register("custom1", customFunc1)
	registry2.Register("custom2", customFunc2)

	// Merge registry2 into registry1
	registry1.Merge(registry2)

	// Check that registry1 has both functions
	_, exists := registry1.Get("custom1")
	assert.True(t, exists)

	_, exists = registry1.Get("custom2")
	assert.True(t, exists)
}

func TestDefaultFunctions(t *testing.T) {
	// Test that default functions map contains expected functions
	assert.Contains(t, DefaultFunctions, "number")
	assert.Contains(t, DefaultFunctions, "integer")
	assert.Contains(t, DefaultFunctions, "string")
	assert.Contains(t, DefaultFunctions, "offset")
	assert.Equal(t, 4, len(DefaultFunctions))
}

func TestDraftFunctions(t *testing.T) {
	// Test that draft functions map contains expected functions
	assert.Contains(t, DraftFunctions, "currency")
	assert.Contains(t, DraftFunctions, "date")
	assert.Contains(t, DraftFunctions, "datetime")
	assert.Contains(t, DraftFunctions, "math")
	assert.Contains(t, DraftFunctions, "percent")
	assert.Contains(t, DraftFunctions, "time")
	assert.Contains(t, DraftFunctions, "unit")
	assert.Equal(t, 7, len(DraftFunctions))
}
