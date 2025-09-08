package functions

import (
	"sync"
)

// DefaultFunctions provides the built-in REQUIRED functions
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

// DraftFunctions provides the DRAFT functions (beta)
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
