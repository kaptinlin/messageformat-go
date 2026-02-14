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
func UnknownFunction(source string, input any, locale string) messagevalue.MessageValue {
	return messagevalue.NewUnknownValue(source, input, locale)
}
