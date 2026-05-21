package datamodel

// Visitor represents a visitor for traversing message data model
// TypeScript original code:
// export function visit(
//
//	msg: Message,
//	visitors: {
//	  attributes?: (attributes: Attributes, context: 'declaration' | 'placeholder') => (() => void) | void;
//	  declaration?: (declaration: Declaration) => (() => void) | void;
//	  expression?: (expression: Expression, context: 'declaration' | 'placeholder') => (() => void) | void;
//	  functionRef?: (functionRef: FunctionRef, context: 'declaration' | 'placeholder', argument: Literal | VariableRef | undefined) => (() => void) | void;
//	  key?: (key: Literal | CatchallKey, index: number, keys: (Literal | CatchallKey)[]) => void;
//	  markup?: (markup: Markup, context: 'declaration' | 'placeholder') => (() => void) | void;
//	  node?: (node: Node, ...rest: unknown[]) => void;
//	  options?: (options: Options, context: 'declaration' | 'placeholder') => (() => void) | void;
//	  pattern?: (pattern: Pattern) => (() => void) | void;
//	  value?: (value: Literal | VariableRef, context: 'declaration' | 'selector' | 'placeholder', position: 'arg' | 'option' | 'attribute') => void;
//	  variant?: (variant: Variant) => (() => void) | void;
//	}
//
// )
type Visitor struct {
	Attributes  func(attributes Attributes, context VisitContext) func()
	Declaration func(declaration Declaration) func()
	Expression  func(expression *Expression, context VisitContext) func()
	FunctionRef func(functionRef *FunctionRef, context VisitContext, argument ExpressionArg) func()
	Key         func(key VariantKey, index int, keys []VariantKey)
	Markup      func(markup *Markup, context VisitContext) func()
	Node        func(node any, rest ...any)
	Options     func(options Options, context VisitContext) func()
	Pattern     func(pattern Pattern) func()
	Value       func(value ExpressionArg, context VisitContext, position ValuePosition)
	Variant     func(variant *Variant) func()
}

// Visit applies visitor functions to message nodes
func Visit(msg Message, visitor *Visitor) {
	if visitor == nil {
		return
	}

	for _, decl := range msg.Declarations() {
		var end func()
		if visitor.Declaration != nil {
			end = visitor.Declaration(decl)
		} else if visitor.Node != nil {
			visitor.Node(decl)
		}

		handleDeclarationValue(decl, visitor)

		if end != nil {
			end()
		}
	}

	switch m := msg.(type) {
	case *PatternMessage:
		handlePattern(m.Pattern(), visitor)
	case *SelectMessage:
		if visitor.Value != nil {
			for _, selector := range m.Selectors() {
				visitor.Value(&selector, VisitSelector, ValueArg)
			}
		}

		for _, variant := range m.Variants() {
			var end func()
			if visitor.Variant != nil {
				end = visitor.Variant(&variant)
			} else if visitor.Node != nil {
				visitor.Node(&variant)
			}

			if visitor.Key != nil {
				for i, key := range variant.Keys() {
					visitor.Key(key, i, variant.Keys())
				}
			}

			handlePattern(variant.Value(), visitor)

			if end != nil {
				end()
			}
		}
	}
}

func handleDeclarationValue(decl Declaration, visitor *Visitor) {
	switch d := decl.(type) {
	case *InputDeclaration:
		if d.Value() != nil {
			handleElement(d.Value(), VisitDeclaration, visitor)
		}
	case *LocalDeclaration:
		if d.Value() != nil {
			handleElement(d.Value(), VisitDeclaration, visitor)
		}
	}
}

func handleElement(elem any, context VisitContext, visitor *Visitor) {
	switch e := elem.(type) {
	case *Expression:
		var end func()
		if visitor.Expression != nil {
			end = visitor.Expression(e, context)
		} else if visitor.Node != nil {
			visitor.Node(e, context)
		}

		if e.Arg() != nil && visitor.Value != nil {
			visitor.Value(e.Arg(), context, ValueArg)
		}

		if e.FunctionRef() != nil {
			var endFunc func()
			if visitor.FunctionRef != nil {
				endFunc = visitor.FunctionRef(e.FunctionRef(), context, e.Arg())
			} else if visitor.Node != nil {
				visitor.Node(e.FunctionRef(), context, e.Arg())
			}

			handleOptions(e.FunctionRef().Options(), context, visitor)

			if endFunc != nil {
				endFunc()
			}
		}

		handleAttributes(e.Attributes(), context, visitor)

		if end != nil {
			end()
		}

	case *Markup:
		var end func()
		if visitor.Markup != nil {
			end = visitor.Markup(e, context)
		} else if visitor.Node != nil {
			visitor.Node(e, context)
		}

		handleOptions(e.Options(), context, visitor)

		handleAttributes(e.Attributes(), context, visitor)

		if end != nil {
			end()
		}
	}
}

func handlePattern(pattern Pattern, visitor *Visitor) {
	var end func()
	if visitor.Pattern != nil {
		end = visitor.Pattern(pattern)
	} else if visitor.Node != nil {
		visitor.Node(pattern)
	}

	for _, elem := range pattern.Elements() {
		switch e := elem.(type) {
		case *TextElement:
			if visitor.Node != nil {
				visitor.Node(e)
			}
		case *Expression:
			handleElement(e, VisitPlaceholder, visitor)
		case *Markup:
			handleElement(e, VisitPlaceholder, visitor)
		}
	}

	if end != nil {
		end()
	}
}

func handleOptions(options Options, context VisitContext, visitor *Visitor) {
	if options == nil {
		return
	}

	var end func()
	if visitor.Options != nil {
		end = visitor.Options(options, context)
	}

	if visitor.Value != nil {
		for _, value := range options {
			if value, ok := value.(ExpressionArg); ok {
				visitor.Value(value, context, ValueOption)
			}
		}
	}

	if end != nil {
		end()
	}
}

func handleAttributes(attributes Attributes, context VisitContext, visitor *Visitor) {
	if attributes == nil {
		return
	}

	var end func()
	if visitor.Attributes != nil {
		end = visitor.Attributes(attributes, context)
	}

	if visitor.Value != nil {
		for _, value := range attributes {
			if value, ok := value.(ExpressionArg); ok {
				visitor.Value(value, context, ValueAttribute)
			}
		}
	}

	if end != nil {
		end()
	}
}
