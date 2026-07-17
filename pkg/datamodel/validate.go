package datamodel

import (
	"encoding/binary"
	"errors"
	"hash/maphash"
	"maps"
	"slices"

	pkgerrors "github.com/kaptinlin/messageformat-go/pkg/errors"
	"golang.org/x/text/unicode/norm"
)

// ValidationResult contains the functions and external variables used by a message.
// Ordered model nodes contribute names in first-encounter order. Values owned by
// Options maps are visited in lexicographic key order.
// TypeScript original code: validation return value
type ValidationResult struct {
	Functions []string
	Variables []string
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
		switch errType {
		case pkgerrors.ErrorTypeKeyMismatch,
			pkgerrors.ErrorTypeMissingFallback,
			pkgerrors.ErrorTypeMissingSelectorAnnotation,
			pkgerrors.ErrorTypeDuplicateDeclaration,
			pkgerrors.ErrorTypeDuplicateVariant,
			pkgerrors.ErrorTypeInvalidMessage:
			node, _ := node.(pkgerrors.Node)
			err = pkgerrors.NewMessageDataModelError(pkgerrors.ErrorKind(errType), node)
		default:
			start := 0
			end := 1
			if node, ok := node.(pkgerrors.Node); ok {
				start, end = node.GetPosition()
			}
			err = pkgerrors.NewMessageSyntaxError(pkgerrors.ErrorTypeParseError, start, &end, nil)
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
	invalid := msg == nil
	switch message := msg.(type) {
	case *PatternMessage:
		invalid = message == nil
	case *SelectMessage:
		invalid = message == nil
	}
	if invalid {
		onError(pkgerrors.ErrorTypeInvalidMessage, nil)
		return &ValidationResult{
			Functions: []string{},
			Variables: []string{},
		}
	}

	state := newValidationState(onError)
	state.scanDeclarations(msg.Declarations())

	switch m := msg.(type) {
	case *PatternMessage:
		state.visitPattern(m.Pattern())
	case *SelectMessage:
		state.validateSelectMessage(m)
	}

	return state.result()
}

type validationState struct {
	onError func(string, any)

	annotated     map[string]bool
	declared      map[string]bool
	functions     map[string]bool
	functionOrder []string
	localVars     map[string]bool
	variables     map[string]bool
	variableOrder []string

	variantHashSeed maphash.Seed
	variants        map[uint64][][]any
}

func newValidationState(onError func(string, any)) *validationState {
	return &validationState{
		onError:         onError,
		annotated:       make(map[string]bool),
		declared:        make(map[string]bool),
		functions:       make(map[string]bool),
		localVars:       make(map[string]bool),
		variables:       make(map[string]bool),
		variantHashSeed: maphash.MakeSeed(),
		variants:        make(map[uint64][][]any),
	}
}

func (state *validationState) scanDeclarations(declarations []Declaration) {
	for i, decl := range declarations {
		if decl.Name() == "" {
			continue
		}
		if state.hasInvalidLocalArgumentReference(declarations, i, decl) {
			continue
		}
		state.reportInvalidLocalOptionReferences(declarations, i, decl)
		state.recordAnnotation(decl)
		state.recordLocalVariable(decl)
		visitDeclarationExpression(decl, state)
		state.recordDeclaration(decl)
	}
}

func (state *validationState) hasInvalidLocalArgumentReference(declarations []Declaration, index int, decl Declaration) bool {
	localDecl, ok := decl.(*LocalDeclaration)
	if !ok || localDecl.value == nil || localDecl.value.Arg() == nil {
		return false
	}

	varRef, ok := localDecl.value.Arg().(*VariableRef)
	if !ok {
		return false
	}
	if varRef.Name() == localDecl.Name() || referencesLaterLocal(declarations, index, varRef.Name()) {
		state.onError("duplicate-declaration", decl)
		return true
	}
	return false
}

func (state *validationState) reportInvalidLocalOptionReferences(declarations []Declaration, index int, decl Declaration) {
	localDecl, ok := decl.(*LocalDeclaration)
	if !ok || localDecl.value == nil || localDecl.value.FunctionRef() == nil {
		return
	}

	options := localDecl.value.FunctionRef().Options()
	for _, name := range slices.Sorted(maps.Keys(options)) {
		optValue := options[name]
		varRef, ok := optValue.(*VariableRef)
		if !ok {
			continue
		}
		if varRef.Name() == localDecl.Name() || referencesLaterLocal(declarations, index, varRef.Name()) {
			state.onError("duplicate-declaration", decl)
			return
		}
	}
}

func referencesLaterLocal(declarations []Declaration, index int, name string) bool {
	for _, laterDecl := range declarations[index+1:] {
		if laterDecl.Name() == name && IsLocalDeclaration(laterDecl) {
			return true
		}
	}
	return false
}

func (state *validationState) recordAnnotation(decl Declaration) {
	if (IsInputDeclaration(decl) && hasFunction(decl)) ||
		(IsLocalDeclaration(decl) && (hasFunction(decl) || referencesAnnotatedVariable(decl, state.annotated))) {
		state.annotated[decl.Name()] = true
	}
}

func (state *validationState) recordLocalVariable(decl Declaration) {
	if IsLocalDeclaration(decl) {
		state.localVars[decl.Name()] = true
	}
}

func (state *validationState) recordDeclaration(decl Declaration) {
	if state.declared[decl.Name()] {
		state.onError("duplicate-declaration", decl)
		return
	}
	state.declared[decl.Name()] = true
}

func (state *validationState) validateSelectMessage(message *SelectMessage) {
	selectorCount, missingFallback := state.validateSelectors(message.Selectors())
	missingFallback = state.validateVariants(message.Variants(), selectorCount, missingFallback)
	if missingFallback != nil {
		state.onError("missing-fallback", missingFallback)
	}
}

func (state *validationState) validateSelectors(selectors []VariableRef) (int, any) {
	var missingFallback any
	for _, selector := range selectors {
		missingFallback = selector
		state.recordVariable(selector.Name())
		if !state.annotated[selector.Name()] {
			state.onError("missing-selector-annotation", selector)
		}
	}
	return len(selectors), missingFallback
}

func (state *validationState) validateVariants(variants []Variant, selectorCount int, missingFallback any) any {
	for _, variant := range variants {
		keys := variant.Keys()
		if len(keys) != selectorCount {
			state.onError("key-mismatch", variant)
		}

		keyStrs, allCatchall := normalizedVariantKeys(keys)
		state.recordVariantKeys(variant, keyStrs)

		if allCatchall {
			missingFallback = nil
		} else if missingFallback != nil {
			missingFallback = variant
		}
		state.visitPattern(variant.Value())
	}
	return missingFallback
}

func normalizedVariantKeys(keys []VariantKey) ([]any, bool) {
	keyStrs := make([]any, len(keys))
	allCatchall := true
	for i, key := range keys {
		switch {
		case IsCatchallKey(key):
			keyStrs[i] = 0
		case IsLiteral(key):
			keyStrs[i] = norm.NFC.String(key.(*Literal).Value())
			allCatchall = false
		default:
			keyStrs[i] = 0
			allCatchall = false
		}
	}
	return keyStrs, allCatchall
}

func (state *validationState) recordVariantKeys(variant Variant, keyStrs []any) {
	keyHash := hashVariantKeys(state.variantHashSeed, keyStrs)
	if existing, ok := state.variants[keyHash]; ok {
		if variantKeysContain(existing, keyStrs) {
			state.onError("duplicate-variant", variant)
			return
		}
		state.variants[keyHash] = append(existing, keyStrs)
		return
	}
	state.variants[keyHash] = [][]any{keyStrs}
}

func (state *validationState) visitPattern(pattern Pattern) {
	visitPattern(pattern, state)
}

func (state *validationState) result() *ValidationResult {
	variables := make([]string, 0, len(state.variableOrder))
	for _, name := range state.variableOrder {
		if !state.localVars[name] {
			variables = append(variables, name)
		}
	}

	return &ValidationResult{
		Functions: slices.Clone(state.functionOrder),
		Variables: variables,
	}
}

func (state *validationState) recordFunction(name string) {
	if !state.functions[name] {
		state.functions[name] = true
		state.functionOrder = append(state.functionOrder, name)
	}
}

func (state *validationState) recordVariable(name string) {
	if !state.variables[name] {
		state.variables[name] = true
		state.variableOrder = append(state.variableOrder, name)
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

// visitDeclarationExpression visits one declaration's expression.
func visitDeclarationExpression(decl Declaration, state *validationState) {
	switch declaration := decl.(type) {
	case *InputDeclaration:
		visitExpression(declaration.value, state)
	case *LocalDeclaration:
		visitExpression(declaration.value, state)
	}
}

// visitExpression records one expression's function, argument, and options.
func visitExpression(expression *Expression, state *validationState) {
	if expression == nil {
		return
	}
	if function := expression.FunctionRef(); function != nil {
		state.recordFunction(function.Name())
	}
	if variable, ok := expression.Arg().(*VariableRef); ok {
		state.recordVariable(variable.Name())
	}
	if function := expression.FunctionRef(); function != nil {
		visitOptions(function.Options(), state)
	}
}

// visitOptions records variable values in lexicographic option-key order.
func visitOptions(options Options, state *validationState) {
	for _, name := range slices.Sorted(maps.Keys(options)) {
		if variable, ok := options[name].(*VariableRef); ok {
			state.recordVariable(variable.Name())
		}
	}
}

// visitPattern visits expressions and markup in source order.
func visitPattern(pattern Pattern, state *validationState) {
	for _, element := range pattern.Elements() {
		switch element := element.(type) {
		case *Expression:
			visitExpression(element, state)
		case *Markup:
			visitOptions(element.Options(), state)
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
