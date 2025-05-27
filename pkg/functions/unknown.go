package functions

import (
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// UnknownFunction creates an unknown value for unrecognized input
//
// TypeScript original code:
// export const unknown = (
//
//	source: string,
//	input: unknown
//
//	): MessageUnknownValue => ({
//	  type: 'unknown',
//	  source,
//	  dir: 'auto',
//	  toParts: () => [{ type: 'unknown', value: input }],
//	  toString: () => String(input),
//	  valueOf: () => input
//	});
func UnknownFunction(source string, input interface{}, locale string) messagevalue.MessageValue {
	// For now, create a string value with the string representation of the input
	// This is a simplified implementation
	return messagevalue.NewStringValue(toString(input), source, locale)
}
