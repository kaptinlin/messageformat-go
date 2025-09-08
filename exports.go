// Package messageformat provides the main MessageFormat 2.0 API
package messageformat

import (
	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
)

// Export convenience aliases for the main API functions
// These match TypeScript's exported functions while following Go conventions

// Data model operations - matches TypeScript exports
var (
	ParseMessage     = datamodel.ParseMessage
	StringifyMessage = datamodel.StringifyMessage
	Validate         = datamodel.ValidateMessage
	Visit            = datamodel.Visit
)

// Type guards - matches TypeScript exports
var (
	IsExpression     = datamodel.IsExpression
	IsFunctionRef    = datamodel.IsFunctionRef
	IsLiteral        = datamodel.IsLiteral
	IsMarkup         = datamodel.IsMarkup
	IsMessage        = datamodel.IsMessage
	IsPatternMessage = datamodel.IsPatternMessage
	IsSelectMessage  = datamodel.IsSelectMessage
	IsVariableRef    = datamodel.IsVariableRef
	IsCatchallKey    = datamodel.IsCatchallKey
)

// DefaultFunctions provides access to built-in functions
var DefaultFunctions = functions.DefaultFunctions

// DraftFunctions provides access to draft functions (beta)
var DraftFunctions = functions.DraftFunctions
