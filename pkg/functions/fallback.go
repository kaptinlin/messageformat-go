package functions

import (
	"cmp"

	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// FallbackFunction creates a fallback value for runtime/formatting errors
//
// TypeScript original code:
//
//	export const fallback = (source: string = '�'): MessageFallback => ({
//	  type: 'fallback',
//	  source,
//	  toParts: () => [{ type: 'fallback', source }],
//	  toString: () => `{${source}}`
//	});
func FallbackFunction(source string, locale string) messagevalue.MessageValue {
	return messagevalue.NewFallbackValue(cmp.Or(source, "�"), locale)
}
