package functions

import (
	"maps"
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
