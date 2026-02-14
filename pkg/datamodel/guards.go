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
	m, ok := msg.(Message)
	return ok && (m.Type() == "message" || m.Type() == "select")
}

// IsPatternMessage checks if a message is a pattern message
// TypeScript original code:
// export const isPatternMessage = (msg: Message): msg is PatternMessage =>
//
//	msg.type === 'message';
func IsPatternMessage(msg Message) bool {
	return msg != nil && msg.Type() == "message"
}

// IsSelectMessage checks if a message is a select message
// TypeScript original code:
// export const isSelectMessage = (msg: Message): msg is SelectMessage =>
//
//	msg.type === 'select';
func IsSelectMessage(msg Message) bool {
	return msg != nil && msg.Type() == "select"
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
	return decl != nil && decl.Type() == "input"
}

// IsLocalDeclaration checks if a declaration is a local declaration
// TypeScript original code: Declaration type checking
func IsLocalDeclaration(decl Declaration) bool {
	return decl != nil && decl.Type() == "local"
}

// IsTextElement checks if a pattern element is text
// TypeScript original code: Pattern element type checking (string type)
func IsTextElement(elem PatternElement) bool {
	return elem != nil && elem.Type() == "text"
}

// IsVariantKey checks if an object is a valid variant key
// TypeScript original code: Array<Literal | CatchallKey> element checking
func IsVariantKey(key any) bool {
	return IsLiteral(key) || IsCatchallKey(key)
}

// IsPatternElement checks if an object is a valid pattern element
// TypeScript original code: Array<string | Expression | Markup> element checking
func IsPatternElement(elem any) bool {
	pe, ok := elem.(PatternElement)
	if !ok {
		return false
	}
	elemType := pe.Type()
	return elemType == "text" || elemType == "expression" || elemType == "markup"
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
	node, ok := obj.(Node)
	if !ok {
		return false
	}
	nodeType := node.Type()
	return nodeType == "input" || nodeType == "local" || // Declaration types
		nodeType == "*" || // CatchallKey
		nodeType == "expression" ||
		nodeType == "literal" ||
		nodeType == "variable" ||
		nodeType == "function" ||
		nodeType == "markup"
}

// IsBooleanAttribute checks if an attribute value is a boolean attribute
// TypeScript original code: true type in Attributes
func IsBooleanAttribute(attr any) bool {
	_, ok := attr.(*BooleanAttribute)
	return ok
}

// IsVariableRefExpression checks if an expression is a VariableRefExpression
// TypeScript original code: Expression<VariableRef> type checking
func IsVariableRefExpression(expr any) bool {
	_, ok := expr.(*VariableRefExpression)
	return ok
}
