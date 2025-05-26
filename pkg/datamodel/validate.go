package datamodel

import (
	"encoding/json"
	"fmt"

	"github.com/kaptinlin/messageformat-go/pkg/errors"
)

// ValidationResult contains the result of message validation
// TypeScript original code: validation return value
type ValidationResult struct {
	Functions []string // Set of function names used
	Variables []string // Set of variable names used
}

// ValidateMessage validates a message data model
// TypeScript original code:
// export function validate(
//
//	msg: Message,
//	onError: (type: MessageDataModelError['type'], node: Node) => void = (
//	  type,
//	  node
//	) => {
//	  throw new MessageDataModelError(type, node);
//	}
//
// )
func ValidateMessage(msg Message, onError func(string, interface{})) (*ValidationResult, error) {
	if onError == nil {
		onError = func(errType string, node interface{}) {
			// Default error handler - we'll collect errors instead of throwing
		}
	}

	var validationErrors []error
	errorHandler := func(errType string, node interface{}) {
		var err error
		switch errType {
		case "key-mismatch":
			err = errors.NewSyntaxError("variant key mismatch", 0, 0)
		case "missing-fallback":
			err = errors.NewSyntaxError("missing fallback variant", 0, 0)
		case "missing-selector-annotation":
			err = errors.NewSyntaxError("missing selector annotation", 0, 0)
		case "duplicate-declaration":
			err = errors.NewSyntaxError("duplicate declaration", 0, 0)
		case "duplicate-variant":
			err = errors.NewSyntaxError("duplicate variant", 0, 0)
		default:
			err = errors.NewSyntaxError(fmt.Sprintf("validation error: %s", errType), 0, 0)
		}
		validationErrors = append(validationErrors, err)
		onError(errType, node)
	}

	result := validateMessage(msg, errorHandler)

	if len(validationErrors) > 0 {
		return result, validationErrors[0] // Return first error
	}

	return result, nil
}

func validateMessage(msg Message, onError func(string, interface{})) *ValidationResult {
	selectorCount := 0
	var missingFallback interface{}

	// Track annotated variables for missing-selector-annotation
	annotated := make(map[string]bool)

	// Track declared variables for duplicate-declaration
	declared := make(map[string]bool)

	functions := make(map[string]bool)
	localVars := make(map[string]bool)
	variables := make(map[string]bool)
	variants := make(map[string]bool)

	// Validate declarations
	for _, decl := range msg.Declarations() {
		if decl.Name() == "" {
			continue // Skip reserved statements
		}

		// Check if declaration has function or references annotated variable
		var hasFunction bool
		var referencesAnnotated bool

		switch d := decl.(type) {
		case *InputDeclaration:
			if d.value != nil {
				hasFunction = d.value.FunctionRef() != nil
				if d.value.Arg() != nil && annotated[d.value.Arg().Name()] {
					referencesAnnotated = true
				}
			}
		case *LocalDeclaration:
			if d.value != nil {
				hasFunction = d.value.FunctionRef() != nil
				if d.value.Arg() != nil {
					if varRef, ok := d.value.Arg().(*VariableRef); ok && annotated[varRef.Name()] {
						referencesAnnotated = true
					}
				}
			}
		}

		if hasFunction || (IsLocalDeclaration(decl) && referencesAnnotated) {
			annotated[decl.Name()] = true
		}

		if IsLocalDeclaration(decl) {
			localVars[decl.Name()] = true
		}

		// Check for duplicate declarations
		if declared[decl.Name()] {
			onError("duplicate-declaration", decl)
		} else {
			declared[decl.Name()] = true
		}

		// Process expression
		switch d := decl.(type) {
		case *InputDeclaration:
			if d.value != nil {
				// Convert VariableRefExpression to Expression for validation
				generalExpr := NewExpression(d.value.Arg(), d.value.FunctionRef(), d.value.Attributes())
				validateExpression(generalExpr, functions, variables)
			}
		case *LocalDeclaration:
			if d.value != nil {
				validateExpression(d.value, functions, variables)
			}
		}
	}

	// Validate message-specific content
	switch m := msg.(type) {
	case *PatternMessage:
		validatePattern(m.Pattern(), functions, variables)
	case *SelectMessage:
		// Validate selectors
		for _, selector := range m.Selectors() {
			selectorCount++
			missingFallback = selector
			variables[selector.Name()] = true

			if !annotated[selector.Name()] {
				onError("missing-selector-annotation", selector)
			}
		}

		// Validate variants
		hasFallback := false
		for _, variant := range m.Variants() {
			// Check key count matches selector count
			if len(variant.Keys()) != selectorCount {
				onError("key-mismatch", variant)
			}

			// Check for duplicate variants
			keyStrs := make([]interface{}, len(variant.Keys()))
			allCatchall := true
			for i, key := range variant.Keys() {
				if IsCatchallKey(key) {
					keyStrs[i] = 0
				} else if IsLiteral(key) {
					keyStrs[i] = key.(*Literal).Value()
					allCatchall = false
				} else {
					keyStrs[i] = key.String()
					allCatchall = false
				}
			}

			if allCatchall {
				hasFallback = true
			}

			keyJSON, _ := json.Marshal(keyStrs)
			keyStr := string(keyJSON)
			if variants[keyStr] {
				onError("duplicate-variant", variant)
			} else {
				variants[keyStr] = true
			}

			// Validate variant pattern
			validatePattern(variant.Value(), functions, variables)
		}

		// Check for missing fallback
		if !hasFallback && missingFallback != nil {
			onError("missing-fallback", missingFallback)
		}
	}

	// Remove local variables from variables set
	for localVar := range localVars {
		delete(variables, localVar)
	}

	// Convert sets to slices
	functionList := make([]string, 0, len(functions))
	for fn := range functions {
		functionList = append(functionList, fn)
	}

	variableList := make([]string, 0, len(variables))
	for variable := range variables {
		variableList = append(variableList, variable)
	}

	return &ValidationResult{
		Functions: functionList,
		Variables: variableList,
	}
}

func validateExpression(expr *Expression, functions, variables map[string]bool) {
	if expr.FunctionRef() != nil {
		functions[expr.FunctionRef().Name()] = true
	}

	if expr.Arg() != nil {
		if IsVariableRef(expr.Arg()) {
			variables[expr.Arg().(*VariableRef).Name()] = true
		}
	}
}

func validatePattern(pattern Pattern, functions, variables map[string]bool) {
	for _, elem := range pattern.Elements() {
		switch e := elem.(type) {
		case *Expression:
			validateExpression(e, functions, variables)
		case *Markup:
			// Markup validation if needed
		}
	}
}
