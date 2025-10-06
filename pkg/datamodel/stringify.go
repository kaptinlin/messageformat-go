// Package datamodel provides message data model stringification for MessageFormat 2.0
// TypeScript original code: data-model/stringify.ts module
package datamodel

import (
	"fmt"
	"regexp"
	"strings"
)

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
func StringifyMessage(msg Message) string {
	var res strings.Builder

	// Intelligent size estimation based on message complexity
	estimatedSize := len(msg.Declarations()) * 50
	if IsSelectMessage(msg) {
		sm := msg.(*SelectMessage)
		estimatedSize += len(sm.Selectors()) * 30
		estimatedSize += len(sm.Variants()) * 100
	}
	res.Grow(estimatedSize)

	// Stringify declarations
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

		// Add selectors
		for _, selector := range sm.Selectors() {
			res.WriteString(" ")
			res.WriteString(stringifyVariableRef(&selector))
		}

		// Add variants
		for _, variant := range sm.Variants() {
			res.WriteString("\n")

			// Add keys
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

	return res.String()
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
				// Convert VariableRefExpression to Expression for stringification
				generalExpr := NewExpression(expr.Arg(), expr.FunctionRef(), expr.Attributes())
				return fmt.Sprintf(".input %s\n", stringifyExpression(generalExpr))
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
	if expr.Attributes() != nil {
		for name, value := range expr.Attributes() {
			parts = append(parts, stringifyAttribute(name, value))
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
	result := ":" + fr.Name()

	if fr.Options() != nil {
		for name, value := range fr.Options() {
			result += " " + stringifyOption(name, value)
		}
	}

	return result
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
	if markup.Options() != nil {
		for name, value := range markup.Options() {
			result.WriteString(" ")
			result.WriteString(stringifyOption(name, value))
		}
	}

	// Add attributes
	if markup.Attributes() != nil {
		for name, value := range markup.Attributes() {
			result.WriteString(" ")
			result.WriteString(stringifyAttribute(name, value))
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
func stringifyOption(name string, value interface{}) string {
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
func stringifyAttribute(name string, value interface{}) string {
	switch {
	case value == true:
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

	// Check if first element starts with dot (needs quoting)
	if !quoted && len(pattern.Elements()) > 0 {
		if textElem, ok := pattern.Elements()[0].(*TextElement); ok {
			if matched, _ := regexp.MatchString(`^\s*\.`, textElem.Value()); matched {
				quoted = true
			}
		}
	}

	// Process elements
	for _, elem := range pattern.Elements() {
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
	if value == "" {
		return false
	}

	// Check for special characters that require quoting
	specialChars := []string{" ", "\t", "\n", "\r", "{", "}", "|", "\\", "=", "@", "$", ":", "#", "/"}
	for _, char := range specialChars {
		if strings.Contains(value, char) {
			return false
		}
	}

	// Check if starts with dot
	if strings.HasPrefix(value, ".") {
		return false
	}

	return true
}
