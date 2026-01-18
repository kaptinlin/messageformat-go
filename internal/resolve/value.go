// Package resolve provides value resolution functions for MessageFormat 2.0
// TypeScript original code: resolve/resolve-value.ts module
package resolve

import (
	"fmt"
	"strings"

	"github.com/kaptinlin/messageformat-go/pkg/datamodel"
	"github.com/kaptinlin/messageformat-go/pkg/logger"
)

// resolveValue resolves a literal or variable reference to its value
// TypeScript original code:
// export function resolveValue(
//
//	ctx: Context,
//	value: Literal | VariableRef
//
//	): unknown {
//	  switch (value.type) {
//	    case 'literal':
//	      return value.value;
//	    case 'variable':
//	      return lookupVariableRef(ctx, value);
//	    default:
//	      // @ts-expect-error - should never happen
//	      throw new Error(`Unsupported value: ${value.type}`);
//	  }
//	}
func resolveValue(ctx *Context, value datamodel.Node) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case *datamodel.Literal:
		return v.Value(), nil
	case *datamodel.VariableRef:
		return lookupVariableRef(ctx, v), nil
	default:
		// Should never happen - matches TypeScript @ts-expect-error
		logger.Error("unsupported value type", "type", v.Type())
		return nil, fmt.Errorf("unsupported value: %s", v.Type())
	}
}

// getValueSource returns the source representation of a value
// TypeScript original code:
// export function getValueSource(value: Literal | VariableRef): string;
// export function getValueSource(
//
//	value: Literal | VariableRef | undefined
//
// ): string | undefined;
//
//	export function getValueSource(value: Literal | VariableRef | undefined) {
//	  switch (value?.type) {
//	    case 'literal':
//	      return (
//	        '|' + value.value.replaceAll('\\', '\\\\').replaceAll('|', '\\|') + '|'
//	      );
//	    case 'variable':
//	      return '$' + value.name;
//	    default:
//	      return undefined;
//	  }
//	}
func getValueSource(value datamodel.Node) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case *datamodel.Literal:
		// Escape backslashes and pipes - matches TypeScript replaceAll logic
		escaped := strings.ReplaceAll(v.Value(), "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "|", "\\|")
		return "|" + escaped + "|"
	case *datamodel.VariableRef:
		return "$" + v.Name()
	default:
		return ""
	}
}
