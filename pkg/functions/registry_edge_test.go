package functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEdgeCases tests edge cases in function registry handling
// Reference: TypeScript commit 09b01970 - tests for prototype pollution prevention
func TestEdgeCases(t *testing.T) {
	t.Run("unknown function with reserved name", func(t *testing.T) {
		registry := NewFunctionRegistry()

		// Test that unknown function names don't interfere with Go built-ins
		_, exists := registry.Get("toString")
		assert.False(t, exists, "Should not have toString function")

		_, exists = registry.Get("valueOf")
		assert.False(t, exists, "Should not have valueOf function")
	})

	t.Run("empty function registry", func(t *testing.T) {
		registry := &FunctionRegistry{
			functions: make(map[string]MessageFunction),
		}

		// Should return not found, not nil pointer
		_, exists := registry.Get("number")
		assert.False(t, exists, "Empty registry should not have any functions")

		// List should return empty slice, not nil
		names := registry.List()
		assert.NotNil(t, names, "List should return empty slice, not nil")
		assert.Empty(t, names, "Empty registry should list no functions")
	})

	t.Run("function overriding", func(t *testing.T) {
		registry := NewFunctionRegistry()

		// Get original number function
		originalFn, exists := registry.Get("number")
		assert.True(t, exists)
		assert.NotNil(t, originalFn)

		// Override with different function
		registry.Register("number", StringFunction)

		// Should get the registered function
		newFn, exists := registry.Get("number")
		assert.True(t, exists)
		assert.NotNil(t, newFn)
	})

	t.Run("case sensitive function names", func(t *testing.T) {
		registry := NewFunctionRegistry()

		// Function names should be case-sensitive
		_, exists := registry.Get("number")
		assert.True(t, exists, "Should find 'number'")

		_, exists = registry.Get("Number")
		assert.False(t, exists, "Should not find 'Number' (wrong case)")

		_, exists = registry.Get("NUMBER")
		assert.False(t, exists, "Should not find 'NUMBER' (wrong case)")
	})

	t.Run("nil function handling", func(t *testing.T) {
		registry := NewFunctionRegistry()

		// Registering nil function should work (for unregistering)
		registry.Register("custom", nil)

		fn, exists := registry.Get("custom")
		assert.True(t, exists, "Should find registered key")
		assert.Nil(t, fn, "Function should be nil")
	})

	t.Run("registry clone independence", func(t *testing.T) {
		original := NewFunctionRegistry()
		clone := original.Clone()

		// Add function to clone
		clone.Register("custom", StringFunction)

		// Original should not have the custom function
		_, exists := original.Get("custom")
		assert.False(t, exists, "Original should not be affected by clone modifications")

		// Clone should have it
		_, exists = clone.Get("custom")
		assert.True(t, exists, "Clone should have the custom function")
	})

	t.Run("registry merge behavior", func(t *testing.T) {
		registry1 := NewFunctionRegistry()
		registry2 := &FunctionRegistry{
			functions: make(map[string]MessageFunction),
		}

		registry2.Register("custom", StringFunction)

		// Merge registry2 into registry1
		registry1.Merge(registry2)

		// registry1 should now have the custom function
		_, exists := registry1.Get("custom")
		assert.True(t, exists, "Merged registry should have custom function")

		// Should still have default functions
		_, exists = registry1.Get("number")
		assert.True(t, exists, "Should still have default functions after merge")
	})

	t.Run("special character in function names", func(t *testing.T) {
		registry := NewFunctionRegistry()

		// These should not exist and should not cause issues
		_, exists := registry.Get("")
		assert.False(t, exists, "Empty string should not be a valid function name")

		_, exists = registry.Get(":")
		assert.False(t, exists, "Colon should not be a valid function name")

		_, exists = registry.Get(":number")
		assert.False(t, exists, "Function names with colon prefix should not exist")
	})
}

// TestDefaultFunctionsImmutability ensures default functions can't be accidentally modified
func TestDefaultFunctionsImmutability(t *testing.T) {
	// Get count of default functions
	originalCount := len(DefaultFunctions)

	// Create new registry
	registry := NewFunctionRegistry()

	// Add custom function to registry
	registry.Register("custom", StringFunction)

	// DefaultFunctions should still have same count
	assert.Equal(t, originalCount, len(DefaultFunctions),
		"DefaultFunctions should not be modified by registry operations")

	// Should not contain the custom function
	_, exists := DefaultFunctions["custom"]
	assert.False(t, exists, "DefaultFunctions should not contain custom functions")
}

// TestDraftFunctionsImmutability ensures draft functions can't be accidentally modified
func TestDraftFunctionsImmutability(t *testing.T) {
	// Get count of draft functions
	originalCount := len(DraftFunctions)

	// Create new registry with draft functions
	registry := NewFunctionRegistryWithDraft()

	// Add custom function to registry
	registry.Register("custom", StringFunction)

	// DraftFunctions should still have same count
	assert.Equal(t, originalCount, len(DraftFunctions),
		"DraftFunctions should not be modified by registry operations")

	// Should not contain the custom function
	_, exists := DraftFunctions["custom"]
	assert.False(t, exists, "DraftFunctions should not contain custom functions")
}
