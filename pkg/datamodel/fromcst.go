// Package datamodel provides CST to data model conversion
// TypeScript original code: data-model/from-cst.ts module
package datamodel

import (
	"github.com/kaptinlin/messageformat-go/internal/cst"
	"github.com/kaptinlin/messageformat-go/pkg/errors"
)

// FromCST converts a CST message structure into its data model representation
// TypeScript original code:
//
//	export function messageFromCST(msg: CST.Message): Model.Message {
//	  for (const error of msg.errors) throw error;
//	  const declarations: Model.Declaration[] = msg.declarations
//	    ? msg.declarations.map(asDeclaration)
//	    : [];
//	  if (msg.type === 'select') {
//	    return {
//	      type: 'select',
//	      declarations,
//	      selectors: msg.selectors.map(sel => asValue(sel)),
//	      variants: msg.variants.map(variant => ({
//	        keys: variant.keys.map(key =>
//	          key.type === '*' ? { type: '*', [cstKey]: key } : asValue(key)
//	        ),
//	        value: asPattern(variant.value),
//	        [cstKey]: variant
//	      })),
//	      [cstKey]: msg
//	    };
//	  } else {
//	    return {
//	      type: 'message',
//	      declarations,
//	      pattern: asPattern(msg.pattern),
//	      [cstKey]: msg
//	    };
//	  }
//	}
func FromCST(msg cst.Message) (Message, error) {
	// Check for CST errors first
	if len(msg.Errors()) > 0 {
		// Return the first error
		firstError := msg.Errors()[0]
		end := firstError.End
		return nil, errors.NewMessageSyntaxError(errors.ErrorTypeParseError, firstError.Start, &end, nil)
	}

	// Convert declarations
	var declarations []Declaration
	switch m := msg.(type) {
	case *cst.SimpleMessage:
		// Simple messages have no declarations
		declarations = []Declaration{}
	case *cst.ComplexMessage:
		declarations = make([]Declaration, len(m.Declarations()))
		for i, decl := range m.Declarations() {
			converted, err := asDeclaration(decl)
			if err != nil {
				return nil, err
			}
			declarations[i] = converted
		}
	case *cst.SelectMessage:
		declarations = make([]Declaration, len(m.Declarations()))
		for i, decl := range m.Declarations() {
			converted, err := asDeclaration(decl)
			if err != nil {
				return nil, err
			}
			declarations[i] = converted
		}
	}

	// Handle different message types
	switch m := msg.(type) {
	case *cst.SelectMessage:
		// Convert selectors
		selectors := make([]VariableRef, len(m.Selectors()))
		for i, sel := range m.Selectors() {
			converted, err := asVariableRef(&sel)
			if err != nil {
				return nil, err
			}
			selectors[i] = *converted
		}

		// Convert variants
		variants := make([]Variant, len(m.Variants()))
		for i, variant := range m.Variants() {
			converted, err := asVariant(variant)
			if err != nil {
				return nil, err
			}
			variants[i] = *converted
		}

		return NewSelectMessage(declarations, selectors, variants, ""), nil

	case *cst.SimpleMessage, *cst.ComplexMessage:
		// Get pattern from either simple or complex message
		var pattern cst.Pattern
		if simple, ok := m.(*cst.SimpleMessage); ok {
			pattern = simple.Pattern()
		} else if complex, ok := m.(*cst.ComplexMessage); ok {
			pattern = complex.Pattern()
		}

		// Convert pattern
		convertedPattern, err := asPattern(pattern)
		if err != nil {
			return nil, err
		}

		return NewPatternMessage(declarations, *convertedPattern, ""), nil

	default:
		end := 1
		return nil, errors.NewMessageSyntaxError(errors.ErrorTypeParseError, 0, &end, nil)
	}
}

// asDeclaration converts a CST declaration to a data model declaration
func asDeclaration(decl cst.Declaration) (Declaration, error) {
	switch d := decl.(type) {
	case *cst.InputDeclaration:
		// Convert the value expression
		expr, err := asExpression(d.Value(), false)
		if err != nil {
			return nil, err
		}

		// Type assert to Expression
		expression, ok := expr.(*Expression)
		if !ok {
			end := d.End()
			return nil, errors.NewMessageSyntaxError(errors.ErrorTypeParseError, d.Start(), &end, nil)
		}

		// Validate that the expression has a variable argument
		if expression.Arg() == nil {
			end := d.End()
			return nil, errors.NewMessageSyntaxError(errors.ErrorTypeParseError, d.Start(), &end, nil)
		}

		// Get the variable name from the argument
		var varName string
		if varRef, ok := expression.Arg().(*VariableRef); ok {
			varName = varRef.Name()
		} else {
			end := d.End()
			return nil, errors.NewMessageSyntaxError(errors.ErrorTypeParseError, d.Start(), &end, nil)
		}

		return NewInputDeclaration(varName, ConvertExpressionToVariableRefExpression(expression)), nil

	case *cst.LocalDeclaration:
		// Convert target variable
		target, err := asVariableRef(d.Target())
		if err != nil {
			return nil, err
		}

		// Convert value expression
		expr, err := asExpression(d.Value(), false)
		if err != nil {
			return nil, err
		}

		// Type assert to Expression
		expression, ok := expr.(*Expression)
		if !ok {
			end := d.End()
			return nil, errors.NewMessageSyntaxError(errors.ErrorTypeParseError, d.Start(), &end, nil)
		}

		return NewLocalDeclaration(target.Name(), expression), nil

	default:
		end := decl.End()
		return nil, errors.NewMessageSyntaxError(errors.ErrorTypeParseError, decl.Start(), &end, nil)
	}
}

// asPattern converts a CST pattern to a data model pattern
func asPattern(pattern cst.Pattern) (*Pattern, error) {
	elements := make([]PatternElement, len(pattern.Body()))

	for i, elem := range pattern.Body() {
		switch e := elem.(type) {
		case *cst.Text:
			elements[i] = NewTextElement(e.Value())
		case *cst.Expression:
			// Convert expression, allowing markup
			converted, err := asExpression(e, true)
			if err != nil {
				return nil, err
			}
			elements[i] = converted
		default:
			end := elem.End()
			return nil, errors.NewMessageSyntaxError(errors.ErrorTypeParseError, elem.Start(), &end, nil)
		}
	}

	result := NewPattern(elements)
	return &result, nil
}

// asExpression converts a CST expression to a data model expression or markup
func asExpression(exp cst.Node, allowMarkup bool) (PatternElement, error) {
	switch e := exp.(type) {
	case *cst.Expression:
		// Check for markup
		if allowMarkup && e.Markup() != nil {
			return asMarkup(e)
		}

		// Convert argument
		var arg any
		if e.Arg() != nil {
			converted, err := asValue(e.Arg())
			if err != nil {
				return nil, err
			}
			arg = converted
		}

		// Convert function reference
		var functionRef *FunctionRef
		if e.FunctionRef() != nil {
			if funcRef, ok := e.FunctionRef().(*cst.FunctionRef); ok {
				converted, err := asFunctionRef(funcRef)
				if err != nil {
					return nil, err
				}
				functionRef = converted
			} else {
				end := e.End()
				return nil, errors.NewMessageSyntaxError(errors.ErrorTypeParseError, e.Start(), &end, nil)
			}
		}

		// Convert attributes
		var attributes map[string]any
		if len(e.Attributes()) > 0 {
			attributes = make(map[string]any)
			for _, attr := range e.Attributes() {
				name := asName(attr.Name())
				if attr.Value() != nil {
					value, err := asValue(attr.Value())
					if err != nil {
						return nil, err
					}
					attributes[name] = value
				} else {
					attributes[name] = true
				}
			}
		}

		return NewExpression(arg, functionRef, ConvertMapToAttributes(attributes)), nil

	case *cst.Junk:
		end := e.End()
		return nil, errors.NewMessageSyntaxError(errors.ErrorTypeParseError, e.Start(), &end, nil)

	default:
		end := exp.End()
		return nil, errors.NewMessageSyntaxError(errors.ErrorTypeParseError, exp.Start(), &end, nil)
	}
}

// asMarkup converts a CST expression with markup to a data model markup
func asMarkup(exp *cst.Expression) (*Markup, error) {
	markup := exp.Markup()
	if markup == nil {
		end := exp.End()
		return nil, errors.NewMessageSyntaxError(errors.ErrorTypeParseError, exp.Start(), &end, nil)
	}

	name := asName(markup.Name())

	// Determine markup kind
	var kind string
	openSyntax := markup.Open()
	openValue := openSyntax.Value()
	switch openValue {
	case "/":
		kind = "close"
	case "#":
		if markup.Close() != nil {
			kind = "standalone"
		} else {
			kind = "open"
		}
	default:
		kind = "open"
	}

	// Convert options
	var options map[string]any
	if len(markup.Options()) > 0 {
		options = make(map[string]any)
		for _, opt := range markup.Options() {
			optName := asName(opt.Name())
			value, err := asValue(opt.Value())
			if err != nil {
				return nil, err
			}
			options[optName] = value
		}
	}

	// Convert attributes
	var attributes map[string]any
	if len(exp.Attributes()) > 0 {
		attributes = make(map[string]any)
		for _, attr := range exp.Attributes() {
			attrName := asName(attr.Name())
			if attr.Value() != nil {
				value, err := asValue(attr.Value())
				if err != nil {
					return nil, err
				}
				attributes[attrName] = value
			} else {
				attributes[attrName] = true
			}
		}
	}

	return NewMarkup(kind, name, ConvertMapToOptions(options), ConvertMapToAttributes(attributes)), nil
}

// asFunctionRef converts a CST function reference to a data model function reference
func asFunctionRef(funcRef *cst.FunctionRef) (*FunctionRef, error) {
	name := asName(funcRef.Name())

	// Convert options
	var options map[string]any
	if len(funcRef.Options()) > 0 {
		options = make(map[string]any)
		for _, opt := range funcRef.Options() {
			optName := asName(opt.Name())
			value, err := asValue(opt.Value())
			if err != nil {
				return nil, err
			}
			options[optName] = value
		}
	}

	return NewFunctionRef(name, ConvertMapToOptions(options)), nil
}

// asVariant converts a CST variant to a data model variant
func asVariant(variant cst.Variant) (*Variant, error) {
	// Convert keys
	keys := make([]VariantKey, len(variant.Keys()))
	for i, key := range variant.Keys() {
		switch k := key.(type) {
		case *cst.CatchallKey:
			keys[i] = NewCatchallKey("*")
		case *cst.Literal:
			literal, err := asLiteral(k)
			if err != nil {
				return nil, err
			}
			keys[i] = literal
		default:
			end := key.End()
			return nil, errors.NewMessageSyntaxError(errors.ErrorTypeParseError, key.Start(), &end, nil)
		}
	}

	// Convert pattern
	pattern, err := asPattern(variant.Value())
	if err != nil {
		return nil, err
	}

	return NewVariant(keys, *pattern), nil
}

// asValue converts a CST value to a data model value
func asValue(value cst.Node) (any, error) {
	switch v := value.(type) {
	case *cst.Literal:
		return asLiteral(v)
	case *cst.VariableRef:
		return asVariableRef(v)
	default:
		end := value.End()
		return nil, errors.NewMessageSyntaxError(errors.ErrorTypeParseError, value.Start(), &end, nil)
	}
}

// asLiteral converts a CST literal to a data model literal
func asLiteral(literal *cst.Literal) (*Literal, error) {
	return NewLiteral(literal.Value()), nil
}

// asVariableRef converts a CST variable reference to a data model variable reference
func asVariableRef(varRef cst.Node) (*VariableRef, error) {
	switch v := varRef.(type) {
	case *cst.VariableRef:
		return NewVariableRef(v.Name()), nil
	default:
		end := varRef.End()
		return nil, errors.NewMessageSyntaxError(errors.ErrorTypeParseError, varRef.Start(), &end, nil)
	}
}

// asName converts a CST identifier to a string name
func asName(id cst.Identifier) string {
	switch len(id) {
	case 1:
		return id[0].Value()
	case 3:
		return id[0].Value() + ":" + id[2].Value()
	default:
		// Return empty string for invalid identifiers
		return ""
	}
}
