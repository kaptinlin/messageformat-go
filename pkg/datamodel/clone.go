package datamodel

// CloneMessage returns a detached copy of message.
//
// TypeScript original code:
// // No direct equivalent; TypeScript keeps object references.
func CloneMessage(message Message) Message {
	switch m := message.(type) {
	case nil:
		return nil
	case *PatternMessage:
		return newPatternMessageWithCST(
			cloneDeclarationsDeep(m.declarations),
			clonePatternDeep(m.pattern),
			m.comment,
			m.cst,
		)
	case *SelectMessage:
		return newSelectMessageWithCST(
			cloneDeclarationsDeep(m.declarations),
			cloneVariableRefs(m.selectors),
			cloneVariantsDeep(m.variants),
			m.comment,
			m.cst,
		)
	default:
		return message
	}
}

// cloneDeclarationsDeep copies declaration nodes for MessageFormat snapshots.
//
// TypeScript original code:
// // No direct equivalent; TypeScript keeps object references.
func cloneDeclarationsDeep(declarations []Declaration) []Declaration {
	if declarations == nil {
		return []Declaration{}
	}
	cloned := make([]Declaration, len(declarations))
	for i, declaration := range declarations {
		cloned[i] = cloneDeclaration(declaration)
	}
	return cloned
}

// cloneDeclaration copies a declaration node.
//
// TypeScript original code:
// // No direct equivalent; TypeScript keeps object references.
func cloneDeclaration(declaration Declaration) Declaration {
	switch d := declaration.(type) {
	case nil:
		return nil
	case *InputDeclaration:
		return newInputDeclarationWithCST(d.name, cloneVariableRefExpression(d.value), d.cst)
	case *LocalDeclaration:
		return newLocalDeclarationWithCST(d.name, cloneExpression(d.value), d.cst)
	default:
		return declaration
	}
}

// cloneVariableRefExpression copies a variable-ref expression node.
//
// TypeScript original code:
// // No direct equivalent; TypeScript keeps object references.
func cloneVariableRefExpression(expression *VariableRefExpression) *VariableRefExpression {
	if expression == nil {
		return nil
	}
	return newVariableRefExpressionWithCST(
		cloneVariableRef(expression.arg),
		cloneFunctionRef(expression.functionRef),
		cloneAttributesDeep(expression.attributes),
		expression.cst,
	)
}

// cloneVariantsDeep copies select variants.
//
// TypeScript original code:
// // No direct equivalent; TypeScript keeps object references.
func cloneVariantsDeep(variants []Variant) []Variant {
	if variants == nil {
		return []Variant{}
	}
	cloned := make([]Variant, len(variants))
	for i := range variants {
		cloned[i] = Variant{
			keys:  cloneVariantKeysDeep(variants[i].keys),
			value: clonePatternDeep(variants[i].value),
			cst:   variants[i].cst,
		}
	}
	return cloned
}

// cloneVariantKeysDeep copies variant keys.
//
// TypeScript original code:
// // No direct equivalent; TypeScript keeps object references.
func cloneVariantKeysDeep(keys []VariantKey) []VariantKey {
	if keys == nil {
		return []VariantKey{}
	}
	cloned := make([]VariantKey, len(keys))
	for i, key := range keys {
		cloned[i] = cloneVariantKey(key)
	}
	return cloned
}

// cloneVariantKey copies a select key node.
//
// TypeScript original code:
// // No direct equivalent; TypeScript keeps object references.
func cloneVariantKey(key VariantKey) VariantKey {
	switch k := key.(type) {
	case nil:
		return nil
	case *Literal:
		return cloneLiteral(k)
	case *CatchallKey:
		return newCatchallKeyWithCST(k.value, k.cst)
	default:
		return key
	}
}

// clonePatternDeep copies a message pattern and its elements.
//
// TypeScript original code:
// // No direct equivalent; TypeScript keeps object references.
func clonePatternDeep(pattern Pattern) Pattern {
	if pattern == nil {
		return Pattern{}
	}
	elements := make([]PatternElement, len(pattern))
	for i, element := range pattern {
		elements[i] = clonePatternElement(element)
	}
	return NewPattern(elements)
}

// clonePatternElement copies a pattern element node.
//
// TypeScript original code:
// // No direct equivalent; TypeScript keeps object references.
func clonePatternElement(element PatternElement) PatternElement {
	switch e := element.(type) {
	case nil:
		return nil
	case *TextElement:
		return newTextElementWithCST(e.value, e.cst)
	case *Expression:
		return cloneExpression(e)
	case *Markup:
		return cloneMarkup(e)
	default:
		return element
	}
}

// cloneExpression copies an expression node.
//
// TypeScript original code:
// // No direct equivalent; TypeScript keeps object references.
func cloneExpression(expression *Expression) *Expression {
	if expression == nil {
		return nil
	}
	return newExpressionWithCST(
		cloneExpressionArg(expression.arg),
		cloneFunctionRef(expression.functionRef),
		cloneAttributesDeep(expression.attributes),
		expression.cst,
	)
}

// cloneExpressionArg copies an expression argument.
//
// TypeScript original code:
// // No direct equivalent; TypeScript keeps object references.
func cloneExpressionArg(arg ExpressionArg) ExpressionArg {
	switch a := arg.(type) {
	case nil:
		return nil
	case *Literal:
		return cloneLiteral(a)
	case *VariableRef:
		return cloneVariableRef(a)
	default:
		return arg
	}
}

// cloneFunctionRef copies a function reference node.
//
// TypeScript original code:
// // No direct equivalent; TypeScript keeps object references.
func cloneFunctionRef(ref *FunctionRef) *FunctionRef {
	if ref == nil {
		return nil
	}
	return newFunctionRefWithCST(ref.name, cloneOptionsDeep(ref.options), ref.cst)
}

// cloneMarkup copies a markup node.
//
// TypeScript original code:
// // No direct equivalent; TypeScript keeps object references.
func cloneMarkup(markup *Markup) *Markup {
	if markup == nil {
		return nil
	}
	cloned, _ := newMarkupWithCST(
		markup.kind,
		markup.name,
		cloneOptionsDeep(markup.options),
		cloneAttributesDeep(markup.attributes),
		markup.cst,
	)
	return cloned
}

// cloneOptionsDeep copies option map values.
//
// TypeScript original code:
// // No direct equivalent; TypeScript keeps object references.
func cloneOptionsDeep(options Options) Options {
	if options == nil {
		return nil
	}
	cloned := make(Options, len(options))
	for name, value := range options {
		cloned[name] = cloneOptionValue(value)
	}
	return cloned
}

// cloneOptionValue copies a function option value.
//
// TypeScript original code:
// // No direct equivalent; TypeScript keeps object references.
func cloneOptionValue(value OptionValue) OptionValue {
	switch v := value.(type) {
	case nil:
		return nil
	case *Literal:
		return cloneLiteral(v)
	case *VariableRef:
		return cloneVariableRef(v)
	default:
		return value
	}
}

// cloneAttributesDeep copies attribute map values.
//
// TypeScript original code:
// // No direct equivalent; TypeScript keeps object references.
func cloneAttributesDeep(attributes Attributes) Attributes {
	if attributes == nil {
		return nil
	}
	cloned := make(Attributes, len(attributes))
	for name, value := range attributes {
		cloned[name] = cloneAttributeValue(value)
	}
	return cloned
}

// cloneAttributeValue copies a markup or expression attribute value.
//
// TypeScript original code:
// // No direct equivalent; TypeScript keeps object references.
func cloneAttributeValue(value AttributeValue) AttributeValue {
	switch v := value.(type) {
	case nil:
		return nil
	case *Literal:
		return cloneLiteral(v)
	case *BooleanAttribute:
		return newBooleanAttributeWithCST(v.cst)
	default:
		return value
	}
}

// cloneLiteral copies a literal node.
//
// TypeScript original code:
// // No direct equivalent; TypeScript keeps object references.
func cloneLiteral(literal *Literal) *Literal {
	if literal == nil {
		return nil
	}
	return newLiteralWithCST(literal.value, literal.cst)
}

// cloneVariableRef copies a variable reference node.
//
// TypeScript original code:
// // No direct equivalent; TypeScript keeps object references.
func cloneVariableRef(ref *VariableRef) *VariableRef {
	if ref == nil {
		return nil
	}
	return newVariableRefWithCST(ref.name, ref.cst)
}
