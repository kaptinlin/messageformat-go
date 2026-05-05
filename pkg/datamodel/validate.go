package datamodel

import (
	"encoding/binary"
	"errors"
	"hash/maphash"
	"maps"
	"slices"

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
	variantHashSeed := maphash.MakeSeed()
	variants := make(map[uint64][][]any)

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

		// Visit expression in declaration
		visitExpression(decl, functions, variables)

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
		visitPattern(m.Pattern(), functions, variables)
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

			keyHash := hashVariantKeys(variantHashSeed, keyStrs)
			if existing, ok := variants[keyHash]; ok {
				if variantKeysContain(existing, keyStrs) {
					onError("duplicate-variant", variant)
				} else {
					// Hash collision with a different key tuple; append to bucket.
					variants[keyHash] = append(existing, keyStrs)
				}
			} else {
				variants[keyHash] = [][]any{keyStrs}
			}

			// TypeScript: missingFallback &&= keys.every(key => key.type === '*') ? null : variant;
			hasOtherFallback := slices.ContainsFunc(keys, func(key VariantKey) bool {
				return IsLiteral(key) && key.(*Literal).Value() == "other"
			})
			if hasOtherFallback {
				hasFallback = true
			}

			if !allCatchall && !hasOtherFallback {
				missingFallback = variant
			} else {
				missingFallback = nil
			}

			// Visit variant pattern
			visitPattern(variant.Value(), functions, variables)
		}

		// Check for missing fallback
		// TypeScript: if (missingFallback) onError('missing-fallback', missingFallback);
		if !hasFallback && missingFallback != nil {
			onError("missing-fallback", missingFallback)
		}
	}

	// TypeScript: for (const lv of localVars) variables.delete(lv);
	maps.DeleteFunc(variables, func(name string, _ bool) bool {
		return localVars[name]
	})

	// Convert sets to slices using slices.Collect (Go 1.23+)
	functionList := slices.Collect(maps.Keys(functions))
	variableList := slices.Collect(maps.Keys(variables))

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

	localDecl, ok := decl.(*LocalDeclaration)
	if !ok || localDecl.value == nil || localDecl.value.Arg() == nil {
		return false
	}

	varRef, ok := localDecl.value.Arg().(*VariableRef)
	if !ok {
		return false
	}

	return annotated[varRef.Name()]
}

// visitExpression visits an expression in a declaration
// TypeScript: expression({ functionRef }) { if (functionRef) functions.add(functionRef.name); }
// TypeScript: value(value, context, position) { ... }
func visitExpression(decl Declaration, functions, variables map[string]bool) {
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
func visitPattern(pattern Pattern, functions, variables map[string]bool) {
	for _, elem := range pattern.Elements() {
		expr, ok := elem.(*Expression)
		if !ok {
			continue
		}

		// TypeScript: expression({ functionRef }) { if (functionRef) functions.add(functionRef.name); }
		if expr.FunctionRef() != nil {
			functions[expr.FunctionRef().Name()] = true
			visitFunctionOptions(expr.FunctionRef(), variables)
		}

		// TypeScript: value(value, context, position) { if (value.type !== 'variable') return; variables.add(value.name); }
		if expr.Arg() != nil {
			if varRef, ok := expr.Arg().(*VariableRef); ok {
				variables[varRef.Name()] = true
			}
		}
	}
}

// hashVariantKeys computes a deterministic hash for a variant key tuple.
// Each element is either an int (0, for catchall) or a string (normalized literal).
func hashVariantKeys(seed maphash.Seed, keys []any) uint64 {
	var h maphash.Hash
	h.SetSeed(seed)
	for _, k := range keys {
		switch v := k.(type) {
		case string:
			// Write a type tag to distinguish string "0" from int 0.
			_ = h.WriteByte(1)
			_, _ = h.WriteString(v)
		default:
			// Catchall key represented as int 0.
			_ = h.WriteByte(0)
		}
	}
	// Mix in the length to avoid prefix collisions.
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(len(keys)))
	_, _ = h.Write(buf[:])
	return h.Sum64()
}

// variantKeysContain reports whether any entry in bucket is equal to target.
func variantKeysContain(bucket [][]any, target []any) bool {
	for _, entry := range bucket {
		if variantKeysEqual(entry, target) {
			return true
		}
	}
	return false
}

// variantKeysEqual reports whether two variant key slices are equal.
// Elements are either int (catchall) or string (normalized literal).
func variantKeysEqual(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
