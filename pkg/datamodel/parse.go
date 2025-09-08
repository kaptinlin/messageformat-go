// Package datamodel provides message parsing functionality for MessageFormat 2.0
// TypeScript original code: data-model/parse.ts module
package datamodel

import (
	"github.com/kaptinlin/messageformat-go/internal/cst"
	"github.com/kaptinlin/messageformat-go/pkg/errors"
)

// ParseMessage parses a MessageFormat 2.0 source string into a Message
// TypeScript original code:
// export function parseMessage(
//
//	src: string,
//	onError?: ErrorHandler
//
//	): Message {
//	  const cst = parseResource(src, onError);
//	  return cstToMessage(cst, onError);
//	}
func ParseMessage(source string) (Message, error) {
	// Parse the string using CST parser
	cstMessage := cst.ParseCST(source, false)

	// Check for CST parsing errors
	if len(cstMessage.Errors()) > 0 {
		// Return the first error
		firstError := cstMessage.Errors()[0]
		end := firstError.End
		return nil, errors.NewMessageSyntaxError(errors.ErrorTypeParseError, firstError.Start, &end, nil)
	}

	// Convert CST to datamodel
	message, err := FromCST(cstMessage)
	if err != nil {
		return nil, err
	}

	return message, nil
}
