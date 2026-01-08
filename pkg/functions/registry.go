package functions

import (
	"sync"
)

// DefaultFunctions provides the built-in REQUIRED functions as defined in
// LDML 48 MessageFormat specification.
// Reference: https://www.unicode.org/reports/tr35/tr35-76/tr35-messageFormat.html#contents-of-part-9-messageformat
//
// These functions are stable and covered by stability guarantees.
// They include: :integer, :number, :offset, and :string
//
// TypeScript original code:
//
//	export let DefaultFunctions = {
//	  integer,
//	  number,
//	  string
//	};
//
// DefaultFunctions = Object.freeze(
//
//	Object.assign(Object.create(null), DefaultFunctions)
//
// );
var DefaultFunctions = map[string]MessageFunction{
	"integer": IntegerFunction,
	"number":  NumberFunction,
	"string":  StringFunction,
	"offset":  OffsetFunction,
}

// DraftFunctions provides functions classified as DRAFT by the
// LDML 48 MessageFormat specification.
// Reference: https://www.unicode.org/reports/tr35/tr35-76/tr35-messageFormat.html#contents-of-part-9-messageformat
//
// These functions are liable to change and are NOT covered by stability guarantees.
//
// Note: As of LDML 48, :currency and :percent have been finalized and are now stable.
// However, they remain in this collection for backward compatibility.
// The :unit function is still in DRAFT status.
//
// TypeScript original code:
//
//	export let DraftFunctions = {
//	  currency,
//	  date,
//	  datetime,
//	  math,
//	  time,
//	  unit
//	};
//
// DraftFunctions = Object.freeze(
//
//	Object.assign(Object.create(null), DraftFunctions)
//
// );
var DraftFunctions = map[string]MessageFunction{
	"currency": CurrencyFunction,
	"date":     DateFunction,
	"datetime": DatetimeFunction,
	"math":     MathFunction,
	"percent":  PercentFunction,
	"time":     TimeFunction,
	"unit":     UnitFunction,
}

// FunctionRegistry manages function registration and lookup
type FunctionRegistry struct {
	functions map[string]MessageFunction
	mu        sync.RWMutex
}

// NewFunctionRegistry creates a new function registry
func NewFunctionRegistry() *FunctionRegistry {
	registry := &FunctionRegistry{
		functions: make(map[string]MessageFunction),
	}

	// Register default functions
	for name, fn := range DefaultFunctions {
		registry.functions[name] = fn
	}

	return registry
}

// NewFunctionRegistryWithDraft creates a new function registry including draft functions
func NewFunctionRegistryWithDraft() *FunctionRegistry {
	registry := &FunctionRegistry{
		functions: make(map[string]MessageFunction),
	}

	// Register default functions
	for name, fn := range DefaultFunctions {
		registry.functions[name] = fn
	}

	// Register draft functions
	for name, fn := range DraftFunctions {
		registry.functions[name] = fn
	}

	return registry
}

// Register adds a function to the registry
func (fr *FunctionRegistry) Register(name string, fn MessageFunction) {
	fr.mu.Lock()
	defer fr.mu.Unlock()
	fr.functions[name] = fn
}

// Get retrieves a function from the registry
func (fr *FunctionRegistry) Get(name string) (MessageFunction, bool) {
	fr.mu.RLock()
	defer fr.mu.RUnlock()
	fn, exists := fr.functions[name]
	return fn, exists
}

// List returns all registered function names
func (fr *FunctionRegistry) List() []string {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	names := make([]string, 0, len(fr.functions))
	for name := range fr.functions {
		names = append(names, name)
	}
	return names
}

// Clone creates a copy of the registry
func (fr *FunctionRegistry) Clone() *FunctionRegistry {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	newRegistry := &FunctionRegistry{
		functions: make(map[string]MessageFunction, len(fr.functions)),
	}

	for name, fn := range fr.functions {
		newRegistry.functions[name] = fn
	}

	return newRegistry
}

// Merge merges another registry into this one
func (fr *FunctionRegistry) Merge(other *FunctionRegistry) {
	fr.mu.Lock()
	defer fr.mu.Unlock()

	other.mu.RLock()
	defer other.mu.RUnlock()

	for name, fn := range other.functions {
		fr.functions[name] = fn
	}
}
