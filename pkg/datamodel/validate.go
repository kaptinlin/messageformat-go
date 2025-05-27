package datamodel

import (
	"encoding/json"

	"github.com/kaptinlin/messageformat-go/pkg/errors"
	"golang.org/x/text/unicode/norm"
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
		end := 1
		switch errType {
		case "key-mismatch":
			err = errors.NewMessageSyntaxError(errors.ErrorTypeKeyMismatch, 0, &end, nil)
		case "missing-fallback":
			// Try to use convenience constructor if node implements Node interface
			if nodeImpl, ok := node.(errors.Node); ok {
				err = errors.NewMissingFallbackError(nodeImpl)
			} else {
				err = errors.NewMessageSyntaxError(errors.ErrorTypeMissingFallback, 0, &end, nil)
			}
		case "missing-selector-annotation":
			err = errors.NewMessageSyntaxError(errors.ErrorTypeMissingSelectorAnnotation, 0, &end, nil)
		case "duplicate-declaration":
			// Try to use convenience constructor if node implements Node interface
			if nodeImpl, ok := node.(errors.Node); ok {
				err = errors.NewDuplicateDeclarationError(nodeImpl)
			} else {
				err = errors.NewMessageSyntaxError(errors.ErrorTypeDuplicateDeclaration, 0, &end, nil)
			}
		case "duplicate-variant":
			err = errors.NewMessageSyntaxError(errors.ErrorTypeDuplicateVariant, 0, &end, nil)
		default:
			err = errors.NewMessageSyntaxError(errors.ErrorTypeParseError, 0, &end, nil)
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

// validateMessage implements the core validation logic matching TypeScript implementation
// TypeScript original code:
//
//	export function validate(msg: Message, onError: ...) {
//	  let selectorCount = 0;
//	  let missingFallback: VariableRef | Variant | null = null;
//	  const annotated = new Set<string>();
//	  const declared = new Set<string>();
//	  const functions = new Set<string>();
//	  const localVars = new Set<string>();
//	  const variables = new Set<string>();
//	  const variants = new Set<string>();
//	  let setArgAsDeclared = true;
//	  visit(msg, { ... });
//	}
func validateMessage(msg Message, onError func(string, interface{})) *ValidationResult {
	selectorCount := 0
	var missingFallback interface{}

	// Tracks directly & indirectly annotated variables for missing-selector-annotation
	annotated := make(map[string]bool)

	// Tracks declared variables for duplicate-declaration
	declared := make(map[string]bool)

	functions := make(map[string]bool)
	localVars := make(map[string]bool)
	variables := make(map[string]bool)
	variants := make(map[string]bool)

	// Visit all declarations first
	// TypeScript: visit(msg, { declaration(decl) { ... } })
	for _, decl := range msg.Declarations() {
		// Skip all ReservedStatement
		if decl.Name() == "" {
			continue
		}

		// TypeScript: if (decl.value.functionRef || (decl.type === 'local' && ...))
		if (decl.Type() == "input" && hasFunction(decl)) ||
			(decl.Type() == "local" && (hasFunction(decl) || referencesAnnotatedVariable(decl, annotated))) {
			annotated[decl.Name()] = true
		}

		// TypeScript: if (decl.type === 'local') localVars.add(decl.name);
		if decl.Type() == "local" {
			localVars[decl.Name()] = true
		}

		// TypeScript: setArgAsDeclared = decl.type === 'local';
		setArgAsDeclared := decl.Type() == "local"

		// Check for duplicate declaration before adding to declared set
		// This includes checking if the declaration references itself
		if declared[decl.Name()] || checkSelfReference(decl) {
			onError("duplicate-declaration", decl)
		} else {
			// Check if declaration references undeclared variables in function options
			if checkUndeclaredReferences(decl, declared) {
				onError("duplicate-declaration", decl)
			} else {
				declared[decl.Name()] = true
			}
		}

		// Visit expression in declaration
		visitExpression(decl, functions, variables, declared, setArgAsDeclared)
	}

	// Visit message pattern or selectors/variants
	switch m := msg.(type) {
	case *PatternMessage:
		visitPattern(m.Pattern(), functions, variables, onError)
	case *SelectMessage:
		// Visit selectors
		// TypeScript: case 'selector': selectorCount += 1; missingFallback = value; ...
		for _, selector := range m.Selectors() {
			selectorCount++
			missingFallback = selector
			variables[selector.Name()] = true

			if !annotated[selector.Name()] {
				onError("missing-selector-annotation", selector)
			}
		}

		// Visit variants
		// TypeScript: variant(variant) { ... }
		hasFallback := false
		for _, variant := range m.Variants() {
			keys := variant.Keys()

			// Check key count matches selector count
			if len(keys) != selectorCount {
				onError("key-mismatch", variant)
			}

			// Check for duplicate variants
			// TypeScript: const strKeys = JSON.stringify(keys.map(key => (key.type === 'literal' ? key.value : 0)));
			keyStrs := make([]interface{}, len(keys))
			allCatchall := true
			for i, key := range keys {
				if IsCatchallKey(key) {
					keyStrs[i] = 0
				} else if IsLiteral(key) {
					// Apply Unicode NFC normalization to literal keys for comparison
					// This matches the MessageFormat 2.0 specification requirement
					normalizedValue := norm.NFC.String(key.(*Literal).Value())
					keyStrs[i] = normalizedValue
					allCatchall = false
				} else {
					keyStrs[i] = 0
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

			// TypeScript: missingFallback &&= keys.every(key => key.type === '*') ? null : variant;
			if !allCatchall {
				missingFallback = variant
			} else {
				missingFallback = nil
			}

			// Visit variant pattern
			visitPattern(variant.Value(), functions, variables, onError)
		}

		// Check for missing fallback
		// TypeScript: if (missingFallback) onError('missing-fallback', missingFallback);
		if !hasFallback && missingFallback != nil {
			onError("missing-fallback", missingFallback)
		}
	}

	// TypeScript: for (const lv of localVars) variables.delete(lv);
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

// hasFunction checks if a declaration has a function reference
func hasFunction(decl Declaration) bool {
	switch d := decl.(type) {
	case *InputDeclaration:
		return d.value != nil && d.value.FunctionRef() != nil
	case *LocalDeclaration:
		return d.value != nil && d.value.FunctionRef() != nil
	}
	return false
}

// referencesAnnotatedVariable checks if a local declaration references an annotated variable
func referencesAnnotatedVariable(decl Declaration, annotated map[string]bool) bool {
	if decl.Type() != "local" {
		return false
	}

	if localDecl, ok := decl.(*LocalDeclaration); ok && localDecl.value != nil {
		if localDecl.value.Arg() != nil {
			if varRef, ok := localDecl.value.Arg().(*VariableRef); ok {
				return annotated[varRef.Name()]
			}
		}
	}
	return false
}

// visitExpression visits an expression in a declaration
// TypeScript: expression({ functionRef }) { if (functionRef) functions.add(functionRef.name); }
// TypeScript: value(value, context, position) { ... }
func visitExpression(decl Declaration, functions, variables, declared map[string]bool, setArgAsDeclared bool) {
	switch d := decl.(type) {
	case *InputDeclaration:
		if d.value != nil {
			// Add function to functions set
			if d.value.FunctionRef() != nil {
				functions[d.value.FunctionRef().Name()] = true
				// Check function options for variable references
				visitFunctionOptions(d.value.FunctionRef(), variables)
			}

			// Handle argument variable
			// TypeScript: case 'declaration': if (position !== 'arg' || setArgAsDeclared) { declared.add(value.name); }
			if d.value.Arg() != nil {
				// For VariableRefExpression, Arg() returns *VariableRef directly
				varRef := d.value.Arg()
				variables[varRef.Name()] = true
				// For input declarations, setArgAsDeclared is false, so we don't add to declared
				// This matches TypeScript: position === 'arg' && !setArgAsDeclared
			}
		}
	case *LocalDeclaration:
		if d.value != nil {
			// Add function to functions set
			if d.value.FunctionRef() != nil {
				functions[d.value.FunctionRef().Name()] = true
				// Check function options for variable references
				visitFunctionOptions(d.value.FunctionRef(), variables)
			}

			// Handle argument variable
			if d.value.Arg() != nil {
				if varRef, ok := d.value.Arg().(*VariableRef); ok {
					variables[varRef.Name()] = true
					// For local declarations, setArgAsDeclared is true, so we add to declared
					if setArgAsDeclared {
						declared[varRef.Name()] = true
					}
				}
			}
		}
	}
}

// visitFunctionOptions visits function options for variable references
func visitFunctionOptions(funcRef *FunctionRef, variables map[string]bool) {
	if funcRef.Options() != nil {
		for _, optValue := range funcRef.Options() {
			if varRef, ok := optValue.(*VariableRef); ok {
				variables[varRef.Name()] = true
			}
		}
	}
}

// visitPattern visits a pattern for expressions
func visitPattern(pattern Pattern, functions, variables map[string]bool, onError func(string, interface{})) {
	for _, elem := range pattern.Elements() {
		switch e := elem.(type) {
		case *Expression:
			// TypeScript: expression({ functionRef }) { if (functionRef) functions.add(functionRef.name); }
			if e.FunctionRef() != nil {
				functions[e.FunctionRef().Name()] = true
				// Check for duplicate options
				checkDuplicateOptions(e.FunctionRef(), onError)
				// Visit function options
				visitFunctionOptions(e.FunctionRef(), variables)
			}

			// TypeScript: value(value, context, position) { if (value.type !== 'variable') return; variables.add(value.name); }
			if e.Arg() != nil {
				if varRef, ok := e.Arg().(*VariableRef); ok {
					variables[varRef.Name()] = true
				}
			}
		case *Markup:
			// Check for duplicate options in markup
			if e.Options() != nil {
				seen := make(map[string]bool)
				for optName := range e.Options() {
					if seen[optName] {
						onError("duplicate-option-name", e)
						break
					}
					seen[optName] = true
				}
			}
		}
	}
}

// checkSelfReference checks if a declaration references itself
func checkSelfReference(decl Declaration) bool {
	declName := decl.Name()

	switch d := decl.(type) {
	case *LocalDeclaration:
		if d.value != nil {
			// Check if the argument references itself
			if d.value.Arg() != nil {
				if varRef, ok := d.value.Arg().(*VariableRef); ok {
					if varRef.Name() == declName {
						return true
					}
				}
			}

			// Check if function options reference the declaration
			if d.value.FunctionRef() != nil && d.value.FunctionRef().Options() != nil {
				for _, optValue := range d.value.FunctionRef().Options() {
					if varRef, ok := optValue.(*VariableRef); ok {
						if varRef.Name() == declName {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

// checkUndeclaredReferences checks if a declaration references undeclared variables
func checkUndeclaredReferences(decl Declaration, declared map[string]bool) bool {
	switch d := decl.(type) {
	case *LocalDeclaration:
		if d.value != nil {
			// Check if function options reference undeclared variables
			if d.value.FunctionRef() != nil && d.value.FunctionRef().Options() != nil {
				for _, optValue := range d.value.FunctionRef().Options() {
					if varRef, ok := optValue.(*VariableRef); ok {
						// If the variable is not declared yet, this is an error
						if !declared[varRef.Name()] {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

// checkDuplicateOptions checks for duplicate option names in a function reference
func checkDuplicateOptions(funcRef *FunctionRef, onError func(string, interface{})) {
	if funcRef.Options() == nil {
		return
	}

	seen := make(map[string]bool)
	for optName := range funcRef.Options() {
		if seen[optName] {
			onError("duplicate-option-name", funcRef)
			return
		}
		seen[optName] = true
	}
}
