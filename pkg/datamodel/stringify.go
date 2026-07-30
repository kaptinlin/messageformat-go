// Package datamodel provides message data model stringification for MessageFormat 2.0
// TypeScript original code: data-model/stringify.ts module
package datamodel

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/kaptinlin/messageformat-go/internal/cst"
	pkgerrors "github.com/kaptinlin/messageformat-go/pkg/errors"
)

// startsWithDotPattern is pre-compiled regex for checking if text starts with a dot
var startsWithDotPattern = regexp.MustCompile(`^\s*\.`)

// StringifyMessage converts a message back to its syntax representation
// TypeScript original code:
//
//	export function stringifyMessage(msg: Message) {
//	  let res = '';
//	  for (const decl of msg.declarations) res += stringifyDeclaration(decl);
//	  if (isPatternMessage(msg)) {
//	    res += stringifyPattern(msg.pattern, !!res);
//	  } else if (isSelectMessage(msg)) {
//	    res += '.match';
//	    for (const sel of msg.selectors) res += ' ' + stringifyVariableRef(sel);
//	    for (const { keys, value } of msg.variants) {
//	      res += '\n';
//	      for (const key of keys) {
//	        res += (isLiteral(key) ? stringifyLiteral(key) : '*') + ' ';
//	      }
//	      res += stringifyPattern(value, true);
//	    }
//	  }
//	  return res;
//	}
func StringifyMessage(msg Message) (string, error) {
	invalid := msg == nil
	switch message := msg.(type) {
	case *PatternMessage:
		invalid = message == nil
	case *SelectMessage:
		invalid = message == nil
	}
	if invalid {
		return "", pkgerrors.NewMessageDataModelError(pkgerrors.ErrorTypeInvalidMessage, nil)
	}
	if err := validateStringifiableMessage(msg); err != nil {
		return "", err
	}
	if _, err := ValidateMessage(msg, nil); err != nil {
		return "", err
	}

	var res strings.Builder

	// Intelligent size estimation based on message complexity
	estimatedSize := len(msg.Declarations()) * 50
	if IsSelectMessage(msg) {
		sm := msg.(*SelectMessage)
		estimatedSize += len(sm.Selectors()) * 30
		estimatedSize += len(sm.Variants()) * 100
	}
	res.Grow(estimatedSize)

	for _, decl := range msg.Declarations() {
		res.WriteString(stringifyDeclaration(decl))
	}

	hasDeclarations := res.Len() > 0

	if IsPatternMessage(msg) {
		pm := msg.(*PatternMessage)
		res.WriteString(stringifyPattern(pm.Pattern(), hasDeclarations))
	} else if IsSelectMessage(msg) {
		sm := msg.(*SelectMessage)
		res.WriteString(".match")

		for _, selector := range sm.Selectors() {
			res.WriteString(" ")
			res.WriteString(stringifyVariableRef(&selector))
		}

		for _, variant := range sm.Variants() {
			res.WriteString("\n")

			for _, key := range variant.Keys() {
				if IsLiteral(key) {
					res.WriteString(stringifyLiteral(key.(*Literal)))
				} else {
					res.WriteString("*")
				}
				res.WriteString(" ")
			}

			res.WriteString(stringifyPattern(variant.Value(), true))
		}
	}

	return res.String(), nil
}

func validateStringifiableMessage(msg Message) error {
	var patterns []Pattern
	var declarations []Declaration
	switch message := msg.(type) {
	case *PatternMessage:
		declarations = message.declarations
		patterns = append(patterns, message.pattern)
	case *SelectMessage:
		declarations = message.declarations
		for i := range message.selectors {
			if !stringifiableName(message.selectors[i].name) {
				return pkgerrors.NewMessageDataModelError(pkgerrors.ErrorTypeInvalidMessage, &message.selectors[i])
			}
		}
		for _, variant := range message.variants {
			patterns = append(patterns, variant.value)
		}
	}

	for _, declaration := range declarations {
		switch declaration := declaration.(type) {
		case *InputDeclaration:
			if declaration == nil || !stringifiableExpression(declaration.value) {
				return pkgerrors.NewMessageDataModelError(pkgerrors.ErrorTypeInvalidMessage, declaration)
			}
		case *LocalDeclaration:
			if declaration == nil || !stringifiableName(declaration.name) || !stringifiableExpression(declaration.value) {
				return pkgerrors.NewMessageDataModelError(pkgerrors.ErrorTypeInvalidMessage, declaration)
			}
		}
	}

	for _, pattern := range patterns {
		for _, element := range pattern {
			switch element := element.(type) {
			case *Expression:
				if !stringifiableExpression(element) {
					return pkgerrors.NewMessageDataModelError(pkgerrors.ErrorTypeInvalidMessage, element)
				}
			case *Markup:
				if !stringifiableMarkup(element) {
					return pkgerrors.NewMessageDataModelError(pkgerrors.ErrorTypeInvalidMessage, element)
				}
			}
		}
	}
	return nil
}

func stringifiableExpression(expression *Expression) bool {
	if expression == nil || (expression.arg == nil && expression.functionRef == nil) {
		return false
	}
	if variable, ok := expression.arg.(*VariableRef); ok && !stringifiableName(variable.name) {
		return false
	}
	if expression.functionRef != nil && !stringifiableIdentifier(expression.functionRef.name) {
		return false
	}
	return (expression.functionRef == nil || stringifiableOptions(expression.functionRef.options)) &&
		stringifiableAttributes(expression.attributes)
}

func stringifiableMarkup(markup *Markup) bool {
	if markup == nil || !validMarkupKind(markup.kind) || !stringifiableIdentifier(markup.name) {
		return false
	}
	return stringifiableOptions(markup.options) && stringifiableAttributes(markup.attributes)
}

func stringifiableOptions(options Options) bool {
	for name, value := range options {
		if !stringifiableIdentifier(name) {
			return false
		}
		if variable, ok := value.(*VariableRef); ok && !stringifiableName(variable.name) {
			return false
		}
	}
	return true
}

func stringifiableAttributes(attributes Attributes) bool {
	for name := range attributes {
		if !stringifiableIdentifier(name) {
			return false
		}
	}
	return true
}

func stringifiableName(name string) bool {
	parsed := cst.ParseNameValue(name, 0)
	return parsed != nil && parsed.End == len(name) && parsed.Value == name
}

func stringifiableIdentifier(identifier string) bool {
	first := cst.ParseNameValue(identifier, 0)
	if first == nil || first.Value != identifier[:first.End] {
		return false
	}
	if first.End == len(identifier) {
		return true
	}
	if identifier[first.End] != ':' {
		return false
	}
	second := cst.ParseNameValue(identifier, first.End+1)
	return second != nil && second.End == len(identifier) && second.Value == identifier[first.End+1:]
}

// stringifyDeclaration converts a declaration to its string representation
// TypeScript original code:
//
//	function stringifyDeclaration(decl: Declaration) {
//	  switch (decl.type) {
//	    case 'input':
//	      return `.input ${stringifyExpression(decl.value)}\n`;
//	    case 'local':
//	      return `.local $${decl.name} = ${stringifyExpression(decl.value)}\n`;
//	  }
//	  // @ts-expect-error Guard against non-TS users with bad data
//	  throw new Error(`Unsupported ${decl.type} declaration`);
//	}
func stringifyDeclaration(decl Declaration) string {
	switch decl.Type() {
	case "input":
		if inputDecl, ok := decl.(*InputDeclaration); ok {
			if expr := inputDecl.value; expr != nil {
				return fmt.Sprintf(".input %s\n", stringifyExpression(expr))
			}
		}
		return ".input\n"
	case "local":
		if localDecl, ok := decl.(*LocalDeclaration); ok {
			if expr := localDecl.value; expr != nil {
				return fmt.Sprintf(".local $%s = %s\n", decl.Name(), stringifyExpression(expr))
			}
		}
		return fmt.Sprintf(".local $%s\n", decl.Name())
	default:
		return ""
	}
}

// stringifyExpression converts an expression to its string representation
// TypeScript original code:
//
//	function stringifyExpression({ arg, attributes, functionRef }: Expression) {
//	  let res: string;
//	  switch (arg?.type) {
//	    case 'literal':
//	      res = stringifyLiteral(arg);
//	      break;
//	    case 'variable':
//	      res = stringifyVariableRef(arg);
//	      break;
//	    default:
//	      res = '';
//	  }
//	  if (functionRef) {
//	    if (res) res += ' ';
//	    res += stringifyFunctionRef(functionRef);
//	  }
//	  if (attributes) {
//	    for (const [name, value] of attributes) {
//	      res += ' ' + stringifyAttribute(name, value);
//	    }
//	  }
//	  return `{${res}}`;
//	}
func stringifyExpression(expr *Expression) string {
	var parts []string

	// Add argument
	if expr.Arg() != nil {
		switch arg := expr.Arg().(type) {
		case *Literal:
			parts = append(parts, stringifyLiteral(arg))
		case *VariableRef:
			parts = append(parts, stringifyVariableRef(arg))
		}
	}

	// Add function reference
	if expr.FunctionRef() != nil {
		parts = append(parts, stringifyFunctionRef(expr.FunctionRef()))
	}

	// Add attributes
	if attributes := expr.Attributes(); attributes != nil {
		for _, name := range slices.Sorted(maps.Keys(attributes)) {
			parts = append(parts, stringifyAttribute(name, attributes[name]))
		}
	}

	return fmt.Sprintf("{%s}", strings.Join(parts, " "))
}

// stringifyFunctionRef converts a function reference to its string representation
// TypeScript original code:
//
//	function stringifyFunctionRef({ name, options }: FunctionRef) {
//	  let res = `:${name}`;
//	  if (options) {
//	    for (const [key, value] of options) {
//	      res += ' ' + stringifyOption(key, value);
//	    }
//	  }
//	  return res;
//	}
func stringifyFunctionRef(fr *FunctionRef) string {
	var result strings.Builder
	result.WriteString(":" + fr.Name())

	if options := fr.Options(); options != nil {
		for _, name := range slices.Sorted(maps.Keys(options)) {
			result.WriteString(" " + stringifyOption(name, options[name]))
		}
	}

	return result.String()
}

// stringifyMarkup converts a markup element to its string representation
// TypeScript original code:
//
//	function stringifyMarkup({ kind, name, options, attributes }: Markup) {
//	  let res = kind === 'close' ? '{/' : '{#';
//	  res += name;
//	  if (options) {
//	    for (const [name, value] of options) {
//	      res += ' ' + stringifyOption(name, value);
//	    }
//	  }
//	  if (attributes) {
//	    for (const [name, value] of attributes) {
//	      res += ' ' + stringifyAttribute(name, value);
//	    }
//	  }
//	  res += kind === 'standalone' ? ' /}' : '}';
//	  return res;
//	}
func stringifyMarkup(markup *Markup) string {
	var result strings.Builder

	if markup.Kind() == "close" {
		result.WriteString("{/")
	} else {
		result.WriteString("{#")
	}

	result.WriteString(markup.Name())

	// Add options
	if options := markup.Options(); options != nil {
		for _, name := range slices.Sorted(maps.Keys(options)) {
			result.WriteString(" ")
			result.WriteString(stringifyOption(name, options[name]))
		}
	}

	// Add attributes
	if attributes := markup.Attributes(); attributes != nil {
		for _, name := range slices.Sorted(maps.Keys(attributes)) {
			result.WriteString(" ")
			result.WriteString(stringifyAttribute(name, attributes[name]))
		}
	}

	if markup.Kind() == "standalone" {
		result.WriteString(" /}")
	} else {
		result.WriteString("}")
	}

	return result.String()
}

// stringifyLiteral converts a literal to its string representation
// TypeScript original code:
//
//	function stringifyLiteral({ value }: Literal) {
//	  if (isValidUnquotedLiteral(value)) return value;
//	  const esc = value.replace(/\\/g, '\\\\').replace(/\|/g, '\\|');
//	  return `|${esc}|`;
//	}
func stringifyLiteral(lit *Literal) string {
	value := lit.Value()

	// Check if value needs quoting (simplified version)
	if isValidUnquotedLiteral(value) {
		return value
	}

	// Escape backslashes and pipes
	escaped := strings.ReplaceAll(value, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "|", "\\|")

	return fmt.Sprintf("|%s|", escaped)
}

// stringifyVariableRef converts a variable reference to its string representation
// TypeScript original code:
//
//	function stringifyVariableRef(ref: VariableRef) {
//	  return '$' + ref.name;
//	}
func stringifyVariableRef(vr *VariableRef) string {
	return "$" + vr.Name()
}

// stringifyOption converts an option to its string representation
// TypeScript original code:
//
//	function stringifyOption(name: string, value: Literal | VariableRef) {
//	  const valueStr = isVariableRef(value)
//	    ? stringifyVariableRef(value)
//	    : stringifyLiteral(value);
//	  return `${name}=${valueStr}`;
//	}
func stringifyOption(name string, value any) string {
	var valueStr string

	switch {
	case IsVariableRef(value):
		valueStr = stringifyVariableRef(value.(*VariableRef))
	case IsLiteral(value):
		valueStr = stringifyLiteral(value.(*Literal))
	default:
		valueStr = fmt.Sprintf("%v", value)
	}

	return fmt.Sprintf("%s=%s", name, valueStr)
}

// stringifyAttribute converts an attribute to its string representation
// TypeScript original code:
//
//	function stringifyAttribute(name: string, value: true | Literal) {
//	  return value === true ? `@${name}` : `@${name}=${stringifyLiteral(value)}`;
//	}
func stringifyAttribute(name string, value any) string {
	switch {
	case value == true:
		return "@" + name
	case IsBooleanAttribute(value):
		return "@" + name
	case IsLiteral(value):
		return fmt.Sprintf("@%s=%s", name, stringifyLiteral(value.(*Literal)))
	default:
		return fmt.Sprintf("@%s=%v", name, value)
	}
}

// stringifyPattern converts a pattern to its string representation
// TypeScript original code:
//
//	function stringifyPattern(pattern: Pattern, quoted: boolean) {
//	  let res = '';
//	  if (!quoted && typeof pattern[0] === 'string' && /^\s*\./.test(pattern[0])) {
//	    quoted = true;
//	  }
//	  for (const el of pattern) {
//	    if (typeof el === 'string') res += el.replace(/[\\{}]/g, '\\$&');
//	    else if (el.type === 'markup') res += stringifyMarkup(el);
//	    else res += stringifyExpression(el);
//	  }
//	  return quoted ? `{{${res}}}` : res;
//	}
func stringifyPattern(pattern Pattern, quoted bool) string {
	var result strings.Builder
	elements := pattern.Elements()

	// Check if first element starts with dot (needs quoting)
	if !quoted && len(elements) > 0 {
		if textElem, ok := elements[0].(*TextElement); ok {
			if startsWithDotPattern.MatchString(textElem.Value()) {
				quoted = true
			}
		}
	}

	// Process elements
	for _, elem := range elements {
		switch e := elem.(type) {
		case *TextElement:
			// Escape special characters
			text := e.Value()
			text = strings.ReplaceAll(text, "\\", "\\\\")
			text = strings.ReplaceAll(text, "{", "\\{")
			text = strings.ReplaceAll(text, "}", "\\}")
			result.WriteString(text)
		case *Expression:
			result.WriteString(stringifyExpression(e))
		case *Markup:
			result.WriteString(stringifyMarkup(e))
		}
	}

	if quoted {
		return fmt.Sprintf("{{%s}}", result.String())
	}

	return result.String()
}

// isValidUnquotedLiteral checks if a literal value can be used without quotes
// Simplified version of the TypeScript isValidUnquotedLiteral function
func isValidUnquotedLiteral(value string) bool {
	return cst.IsValidUnquotedLiteral(value)
}
