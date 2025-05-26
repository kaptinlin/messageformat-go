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

func stringifyFunctionRef(fr *FunctionRef) string {
	result := ":" + fr.Name()

	if fr.Options() != nil {
		for name, value := range fr.Options() {
			result += " " + stringifyOption(name, value)
		}
	}

	return result
}

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

func stringifyVariableRef(vr *VariableRef) string {
	return "$" + vr.Name()
}

func stringifyOption(name string, value interface{}) string {
	var valueStr string

	if IsVariableRef(value) {
		valueStr = stringifyVariableRef(value.(*VariableRef))
	} else if IsLiteral(value) {
		valueStr = stringifyLiteral(value.(*Literal))
	} else {
		valueStr = fmt.Sprintf("%v", value)
	}

	return fmt.Sprintf("%s=%s", name, valueStr)
}

func stringifyAttribute(name string, value interface{}) string {
	if value == true {
		return "@" + name
	} else if IsLiteral(value) {
		return fmt.Sprintf("@%s=%s", name, stringifyLiteral(value.(*Literal)))
	} else {
		return fmt.Sprintf("@%s=%v", name, value)
	}
}

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
