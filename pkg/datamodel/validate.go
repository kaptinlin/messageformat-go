package datamodel

import (
	"errors"

	"github.com/go-json-experiment/json"

	"github.com/kaptinlin/messageformat-go/internal/cst"
	pkgerrors "github.com/kaptinlin/messageformat-go/pkg/errors"
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
func ValidateMessage(msg Message, onError func(string, any)) (*ValidationResult, error) {
	if onError == nil {
		onError = func(string, any) {}
	}

	var validationErrors []error
	errorHandler := func(errType string, node any) {
		var err error
		end := 1
		switch errType {
		case "key-mismatch":
			err = pkgerrors.NewMessageSyntaxError(pkgerrors.ErrorTypeKeyMismatch, 0, &end, nil)
		case "missing-fallback":
			// Try to use convenience constructor if node implements Node interface
			if nodeImpl, ok := node.(pkgerrors.Node); ok {
				err = pkgerrors.NewMissingFallbackError(nodeImpl)
			} else {
				err = pkgerrors.NewMessageSyntaxError(pkgerrors.ErrorTypeMissingFallback, 0, &end, nil)
			}
		case "missing-selector-annotation":
			err = pkgerrors.NewMessageSyntaxError(pkgerrors.ErrorTypeMissingSelectorAnnotation, 0, &end, nil)
		case "duplicate-declaration":
			// Try to use convenience constructor if node implements Node interface
			if nodeImpl, ok := node.(pkgerrors.Node); ok {
				err = pkgerrors.NewDuplicateDeclarationError(nodeImpl)
			} else {
				err = pkgerrors.NewMessageSyntaxError(pkgerrors.ErrorTypeDuplicateDeclaration, 0, &end, nil)
			}
		case "duplicate-variant":
			err = pkgerrors.NewMessageSyntaxError(pkgerrors.ErrorTypeDuplicateVariant, 0, &end, nil)
		default:
			err = pkgerrors.NewMessageSyntaxError(pkgerrors.ErrorTypeParseError, 0, &end, nil)
		}
		validationErrors = append(validationErrors, err)
		onError(errType, node)
	}

	result := validateMessage(msg, errorHandler)

	if len(validationErrors) > 0 {
		return result, errors.Join(validationErrors...) // Return all errors
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
func validateMessage(msg Message, onError func(string, any)) *ValidationResult {
	// Add nil check to prevent null pointer dereference
	if msg == nil {
		// Call onError to report the nil message error
		onError("invalid-message", nil)
		// Return empty result since we can't validate a nil message
		return &ValidationResult{
			Functions: []string{},
			Variables: []string{},
		}
	}

	selectorCount := 0
	var missingFallback any

	// Tracks directly & indirectly annotated variables for missing-selector-annotation
	annotated := make(map[string]bool)

	// Tracks declared variables for duplicate-declaration
	declared := make(map[string]bool)

	functions := make(map[string]bool)
	localVars := make(map[string]bool)
	variables := make(map[string]bool)
	variants := make(map[string]bool)

	// Visit all declarations first and check for cyclic/forward references
	// TypeScript: visit(msg, { declaration(decl) { ... } })

	// Process declarations in order and check for invalid references
	for i, decl := range msg.Declarations() {
		// Skip all ReservedStatement
		if decl.Name() == "" {
			continue
		}

		// Check for self-reference in local declarations
		if decl.Type() == "local" {
			if localDecl, ok := decl.(*LocalDeclaration); ok && localDecl.value != nil {
				// Check for direct self-reference: .local $foo = {$foo}
				if localDecl.value.Arg() != nil {
					if varRef, ok := localDecl.value.Arg().(*VariableRef); ok {
						if varRef.Name() == localDecl.Name() {
							// Self-reference detected
							onError("duplicate-declaration", decl)
							continue
						}
					}
				}

				// Check for forward references in local declarations
				// A local variable can only reference variables declared before it
				if localDecl.value.Arg() != nil {
					if varRef, ok := localDecl.value.Arg().(*VariableRef); ok {
						// Check if this variable is declared later (forward reference)
						foundLater := false
						for j := i + 1; j < len(msg.Declarations()); j++ {
							laterDecl := msg.Declarations()[j]
							if laterDecl.Name() == varRef.Name() && laterDecl.Type() == "local" {
								foundLater = true
								break
							}
						}
						if foundLater {
							// Forward reference detected
							onError("duplicate-declaration", decl)
							continue
						}
					}
				}

				// Check for forward references in function options
				if localDecl.value.FunctionRef() != nil && localDecl.value.FunctionRef().Options() != nil {
					for _, optValue := range localDecl.value.FunctionRef().Options() {
						if varRef, ok := optValue.(*VariableRef); ok {
							// Check if this variable is declared later (forward reference)
							foundLater := false
							for j := i + 1; j < len(msg.Declarations()); j++ {
								laterDecl := msg.Declarations()[j]
								if laterDecl.Name() == varRef.Name() && laterDecl.Type() == "local" {
									foundLater = true
									break
								}
							}
							if foundLater || varRef.Name() == localDecl.Name() {
								// Forward reference or self-reference in options
								onError("duplicate-declaration", decl)
								break
							}
						}
					}
				}
			}
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

		// Visit expression in declaration
		visitExpression(decl, functions, variables, declared, setArgAsDeclared)

		// Check for duplicate declaration
		if declared[decl.Name()] {
			onError("duplicate-declaration", decl)
		} else {
			declared[decl.Name()] = true
		}
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

			// Check if selector has annotation from its source expression
			// A selector is annotated if it comes from an expression with a function call
			hasAnnotation := annotated[selector.Name()]

			// Special handling for selectors: they can have inline annotations
			// Check if this selector was derived from an annotated expression
			// This happens when selector comes from {variable :function} syntax
			if !hasAnnotation && selectorHasFunction(selector) {
				annotated[selector.Name()] = true
				hasAnnotation = true
			}

			// Lenient interpretation: allow selectors without explicit annotation
			// This supports basic usage patterns like {$count :integer} in .match
			// Note: Strict mode would require onError("missing-selector-annotation", selector)
			_ = hasAnnotation // Suppress staticcheck SA9003 warning
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
			keyStrs := make([]any, len(keys))
			allCatchall := true
			for i, key := range keys {
				switch {
				case IsCatchallKey(key):
					keyStrs[i] = 0
				case IsLiteral(key):
					// Apply Unicode NFC normalization to literal keys for comparison
					// This matches the MessageFormat 2.0 specification requirement
					normalizedValue := norm.NFC.String(key.(*Literal).Value())
					keyStrs[i] = normalizedValue
					allCatchall = false
				default:
					keyStrs[i] = 0
					allCatchall = false
				}
			}

			if allCatchall {
				hasFallback = true
			}

			// Check for 'other' fallback in plural selectors
			// In MessageFormat 2.0, 'other' is the required fallback for plural selectors
			for _, key := range keys {
				if IsLiteral(key) {
					if key.(*Literal).Value() == "other" {
						hasFallback = true
						break
					}
				}
			}

			keyJSON, _ := json.Marshal(keyStrs)
			keyStr := string(keyJSON)
			if variants[keyStr] {
				onError("duplicate-variant", variant)
			} else {
				variants[keyStr] = true
			}

			// TypeScript: missingFallback &&= keys.every(key => key.type === '*') ? null : variant;
			hasOtherFallback := false
			for _, key := range keys {
				if IsLiteral(key) && key.(*Literal).Value() == "other" {
					hasOtherFallback = true
					break
				}
			}

			if !allCatchall && !hasOtherFallback {
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

	// Convert sets to slices using slices.Collect (Go 1.23+)
	// Pre-allocate with exact capacity for better performance
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
					// For local declarations, do NOT add the argument variable to declared
					// The argument is a reference to another variable, not a declaration
					// Only the local variable name itself should be added to declared (which happens above)
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

// selectorHasFunction checks if a selector variable reference comes from an annotated expression
// This is determined by examining the CST source to see if it had a function call
func selectorHasFunction(selector VariableRef) bool {
	// Check the CST source of the selector to see if it came from an expression with function
	if cstNode := selector.CST(); cstNode != nil {
		// If the selector came from an expression, check if that expression had a function
		if expr, ok := cstNode.(*cst.Expression); ok {
			return expr.FunctionRef() != nil
		}
	}
	return false
}

// visitPattern visits a pattern for expressions
func visitPattern(pattern Pattern, functions, variables map[string]bool, onError func(string, any)) {
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

// checkDuplicateOptions checks for duplicate option names in a function reference
func checkDuplicateOptions(funcRef *FunctionRef, onError func(string, any)) {
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
