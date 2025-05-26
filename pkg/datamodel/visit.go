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
	Attributes  func(attributes map[string]interface{}, context string) func()
	Declaration func(declaration Declaration) func()
	Expression  func(expression *Expression, context string) func()
	FunctionRef func(functionRef *FunctionRef, context string, argument interface{}) func()
	Key         func(key VariantKey, index int, keys []VariantKey)
	Markup      func(markup *Markup, context string) func()
	Node        func(node interface{}, rest ...interface{})
	Options     func(options map[string]interface{}, context string) func()
	Pattern     func(pattern Pattern) func()
	Value       func(value interface{}, context string, position string)
	Variant     func(variant *Variant) func()
}

// Visit applies visitor functions to message nodes
func Visit(msg Message, visitor *Visitor) {
	if visitor == nil {
		return
	}

	// Visit declarations
	for _, decl := range msg.Declarations() {
		var end func()
		if visitor.Declaration != nil {
			end = visitor.Declaration(decl)
		} else if visitor.Node != nil {
			visitor.Node(decl)
		}

		if decl.Value() != nil {
			handleElement(decl.Value(), "declaration", visitor)
		}

		if end != nil {
			end()
		}
	}

	// Visit message-specific content
	switch m := msg.(type) {
	case *PatternMessage:
		handlePattern(m.Pattern(), visitor)
	case *SelectMessage:
		// Visit selectors
		if visitor.Value != nil {
			for _, selector := range m.Selectors() {
				visitor.Value(selector, "selector", "arg")
			}
		}

		// Visit variants
		for _, variant := range m.Variants() {
			var end func()
			if visitor.Variant != nil {
				end = visitor.Variant(&variant)
			} else if visitor.Node != nil {
				visitor.Node(&variant)
			}

			// Visit keys
			if visitor.Key != nil {
				for i, key := range variant.Keys() {
					visitor.Key(key, i, variant.Keys())
				}
			}

			// Visit pattern
			handlePattern(variant.Value(), visitor)

			if end != nil {
				end()
			}
		}
	}
}

func handleElement(elem interface{}, context string, visitor *Visitor) {
	switch e := elem.(type) {
	case *Expression:
		var end func()
		if visitor.Expression != nil {
			end = visitor.Expression(e, context)
		} else if visitor.Node != nil {
			visitor.Node(e, context)
		}

		// Visit argument
		if e.Arg() != nil && visitor.Value != nil {
			visitor.Value(e.Arg(), context, "arg")
		}

		// Visit function reference
		if e.FunctionRef() != nil {
			var endFunc func()
			if visitor.FunctionRef != nil {
				endFunc = visitor.FunctionRef(e.FunctionRef(), context, e.Arg())
			} else if visitor.Node != nil {
				visitor.Node(e.FunctionRef(), context, e.Arg())
			}

			// Visit function options
			handleOptions(convertOptionsToMap(e.FunctionRef().Options()), context, visitor)

			if endFunc != nil {
				endFunc()
			}
		}

		// Visit attributes
		handleAttributes(convertAttributesToMap(e.Attributes()), context, visitor)

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

		// Visit markup options
		handleOptions(convertOptionsToMap(e.Options()), context, visitor)

		// Visit markup attributes
		handleAttributes(convertAttributesToMap(e.Attributes()), context, visitor)

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
			handleElement(e, "placeholder", visitor)
		case *Markup:
			handleElement(e, "placeholder", visitor)
		}
	}

	if end != nil {
		end()
	}
}

func handleOptions(options map[string]interface{}, context string, visitor *Visitor) {
	if options == nil {
		return
	}

	var end func()
	if visitor.Options != nil {
		end = visitor.Options(options, context)
	}

	// Visit option values
	if visitor.Value != nil {
		for _, value := range options {
			if IsLiteral(value) || IsVariableRef(value) {
				visitor.Value(value, context, "option")
			}
		}
	}

	if end != nil {
		end()
	}
}

func handleAttributes(attributes map[string]interface{}, context string, visitor *Visitor) {
	if attributes == nil {
		return
	}

	var end func()
	if visitor.Attributes != nil {
		end = visitor.Attributes(attributes, context)
	}

	// Visit attribute values
	if visitor.Value != nil {
		for _, value := range attributes {
			if value != true && (IsLiteral(value) || IsVariableRef(value)) {
				visitor.Value(value, context, "attribute")
			}
		}
	}

	if end != nil {
		end()
	}
}

// Helper functions to convert typed maps to interface{} maps
func convertOptionsToMap(options Options) map[string]interface{} {
	if options == nil {
		return nil
	}

	result := make(map[string]interface{})
	for k, v := range options {
		result[k] = v
	}
	return result
}

func convertAttributesToMap(attributes Attributes) map[string]interface{} {
	if attributes == nil {
		return nil
	}

	result := make(map[string]interface{})
	for k, v := range attributes {
		if _, ok := v.(*BooleanAttribute); ok {
			result[k] = true
		} else {
			result[k] = v
		}
	}
	return result
}
