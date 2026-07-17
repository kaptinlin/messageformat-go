// Package datamodel provides type guard functions for message data model types
// TypeScript original code: data-model/type-guards.ts module
package datamodel

// IsCatchallKey checks if a variant key is a catchall key
// TypeScript original code:
// export const isCatchallKey = (key: any): key is CatchallKey =>
//
//	!!key && typeof key === 'object' && key.type === '*';
func IsCatchallKey(key any) bool {
	_, ok := key.(*CatchallKey)
	return ok
}

// IsExpression checks if a pattern element is an expression
// TypeScript original code:
// export const isExpression = (part: any): part is Expression =>
//
//	!!part && typeof part === 'object' && part.type === 'expression';
func IsExpression(part any) bool {
	_, ok := part.(*Expression)
	return ok
}

// IsFunctionRef checks if a part is a function reference
// TypeScript original code:
// export const isFunctionRef = (part: any): part is FunctionRef =>
//
//	!!part && typeof part === 'object' && part.type === 'function';
func IsFunctionRef(part any) bool {
	_, ok := part.(*FunctionRef)
	return ok
}

// IsLiteral checks if a part is a literal
// TypeScript original code:
// export const isLiteral = (part: any): part is Literal =>
//
//	!!part && typeof part === 'object' && part.type === 'literal';
func IsLiteral(part any) bool {
	_, ok := part.(*Literal)
	return ok
}

// IsMarkup checks if a pattern element is markup
// TypeScript original code:
// export const isMarkup = (part: any): part is Markup =>
//
//	!!part && typeof part === 'object' && part.type === 'markup';
func IsMarkup(part any) bool {
	_, ok := part.(*Markup)
	return ok
}

// IsMessage checks if an object is a message
// TypeScript original code:
// export const isMessage = (msg: any): msg is Message =>
//
//	!!msg &&
//	typeof msg === 'object' &&
//	(msg.type === 'message' || msg.type === 'select');
func IsMessage(msg any) bool {
	switch msg.(type) {
	case *PatternMessage, *SelectMessage:
		return true
	default:
		return false
	}
}

// IsPatternMessage checks if a message is a pattern message
// TypeScript original code:
// export const isPatternMessage = (msg: Message): msg is PatternMessage =>
//
//	msg.type === 'message';
func IsPatternMessage(msg Message) bool {
	_, ok := msg.(*PatternMessage)
	return ok
}

// IsSelectMessage checks if a message is a select message
// TypeScript original code:
// export const isSelectMessage = (msg: Message): msg is SelectMessage =>
//
//	msg.type === 'select';
func IsSelectMessage(msg Message) bool {
	_, ok := msg.(*SelectMessage)
	return ok
}

// IsVariableRef checks if a part is a variable reference
// TypeScript original code:
// export const isVariableRef = (part: any): part is VariableRef =>
//
//	!!part && typeof part === 'object' && part.type === 'variable';
func IsVariableRef(part any) bool {
	_, ok := part.(*VariableRef)
	return ok
}

// Additional type guards for Go-specific needs

// IsInputDeclaration checks if a declaration is an input declaration
// TypeScript original code: Declaration type checking
func IsInputDeclaration(decl Declaration) bool {
	_, ok := decl.(*InputDeclaration)
	return ok
}

// IsLocalDeclaration checks if a declaration is a local declaration
// TypeScript original code: Declaration type checking
func IsLocalDeclaration(decl Declaration) bool {
	_, ok := decl.(*LocalDeclaration)
	return ok
}

// IsTextElement checks if a pattern element is text
// TypeScript original code: Pattern element type checking (string type)
func IsTextElement(elem PatternElement) bool {
	_, ok := elem.(*TextElement)
	return ok
}

// IsVariantKey checks if an object is a valid variant key
// TypeScript original code: Array<Literal | CatchallKey> element checking
func IsVariantKey(key any) bool {
	return IsLiteral(key) || IsCatchallKey(key)
}

// IsPatternElement checks if an object is a valid pattern element
// TypeScript original code: Array<string | Expression | Markup> element checking
func IsPatternElement(elem any) bool {
	switch elem.(type) {
	case *TextElement, *Expression, *Markup:
		return true
	default:
		return false
	}
}

// IsNode checks if an object is a valid data model node
// TypeScript original code:
// export type Node =
//
//	| Declaration
//	| Variant
//	| CatchallKey
//	| Expression
//	| Literal
//	| VariableRef
//	| FunctionRef
//	| Markup;
func IsNode(obj any) bool {
	switch obj.(type) {
	case *InputDeclaration, *LocalDeclaration, *CatchallKey, *Expression, *Literal, *VariableRef, *FunctionRef, *Markup, *BooleanAttribute:
		return true
	default:
		return false
	}
}

// IsBooleanAttribute checks if an attribute value is a boolean attribute
// TypeScript original code: true type in Attributes
func IsBooleanAttribute(attr any) bool {
	_, ok := attr.(*BooleanAttribute)
	return ok
}
