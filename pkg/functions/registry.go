package functions

import (
	"maps"
	"slices"
	"sync"
)

// defaultFunctions stores the stable built-in functions as defined in
// LDML 48 MessageFormat.
// Reference: https://www.unicode.org/reports/tr35/tr35-76/tr35-messageFormat.html#contents-of-part-9-messageformat
//
// These functions are stable and covered by stability guarantees.
// They include :currency, :integer, :number, :offset, :percent, and :string.
//
// TypeScript original code:
//
//	export let DefaultFunctions = {
//	  currency,
//	  integer,
//	  number,
//	  offset,
//	  percent,
//	  string
//	};
//
// DefaultFunctions = Object.freeze(
//
//	Object.assign(Object.create(null), DefaultFunctions)
//
// );
var defaultFunctions = map[string]MessageFunction{
	"currency": CurrencyFunction,
	"integer":  IntegerFunction,
	"number":   NumberFunction,
	"offset":   OffsetFunction,
	"percent":  PercentFunction,
	"string":   StringFunction,
}

// DefaultFunctionMap returns a snapshot of the stable built-in functions.
//
// TypeScript original code:
// DefaultFunctions = Object.freeze(
//
//	Object.assign(Object.create(null), DefaultFunctions)
//
// );
func DefaultFunctionMap() map[string]MessageFunction {
	return maps.Clone(defaultFunctions)
}

// draftFunctions stores functions classified as DRAFT by LDML 48 MessageFormat.
// Reference: https://www.unicode.org/reports/tr35/tr35-76/tr35-messageFormat.html#contents-of-part-9-messageformat
//
// These functions are liable to change and are NOT covered by stability guarantees.
//
// TypeScript original code:
//
//	export let DraftFunctions = {
//	  date,
//	  datetime,
//	  time,
//	  unit
//	};
//
// DraftFunctions = Object.freeze(
//
//	Object.assign(Object.create(null), DraftFunctions)
//
// );
var draftFunctions = map[string]MessageFunction{
	"date":     DateFunction,
	"datetime": DatetimeFunction,
	"time":     TimeFunction,
	"unit":     UnitFunction,
}

// DraftFunctionMap returns a snapshot of the draft function set.
//
// TypeScript original code:
// DraftFunctions = Object.freeze(
//
//	Object.assign(Object.create(null), DraftFunctions)
//
// );
func DraftFunctionMap() map[string]MessageFunction {
	return maps.Clone(draftFunctions)
}

// FunctionRegistry manages function registration and lookup
type FunctionRegistry struct {
	functions map[string]MessageFunction
	mu        sync.RWMutex
}

// NewFunctionRegistry creates a new function registry
func NewFunctionRegistry() *FunctionRegistry {
	return &FunctionRegistry{
		functions: maps.Clone(defaultFunctions),
	}
}

// NewFunctionRegistryWithDraft creates a new function registry including draft functions
func NewFunctionRegistryWithDraft() *FunctionRegistry {
	registry := &FunctionRegistry{
		functions: maps.Clone(defaultFunctions),
	}

	maps.Copy(registry.functions, draftFunctions)

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

	return slices.AppendSeq(make([]string, 0, len(fr.functions)), maps.Keys(fr.functions))
}

// Clone creates a copy of the registry
func (fr *FunctionRegistry) Clone() *FunctionRegistry {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	return &FunctionRegistry{
		functions: maps.Clone(fr.functions),
	}
}

// Merge merges another registry into this one
func (fr *FunctionRegistry) Merge(other *FunctionRegistry) {
	fr.mu.Lock()
	defer fr.mu.Unlock()

	other.mu.RLock()
	defer other.mu.RUnlock()

	maps.Copy(fr.functions, other.functions)
}
