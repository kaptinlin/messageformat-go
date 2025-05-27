// Package messageformat provides the main MessageFormat 2.0 API
package messageformat

import (
	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	"github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
)

// Re-export essential constructor and utility functions for convenient access

// Main constructor functions
var (
	// MessageFormat constructors
	NewMessageFormat = New

	// Message validation
	ValidateMessage = datamodel.ValidateMessage

	// Error constructors
	NewMessageSyntaxError     = errors.NewMessageSyntaxError
	NewMessageResolutionError = errors.NewMessageResolutionError
	NewMessageSelectionError  = errors.NewMessageSelectionError

	// Type guards - essential utilities for runtime type checking
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
