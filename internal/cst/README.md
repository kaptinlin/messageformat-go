# CST Package

This package provides Concrete Syntax Tree (CST) parsing for MessageFormat 2.0.

## Files

- `types.go` - CST type definitions
- `parser.go` - Main CST parsing logic
- `values.go` - Value parsing (text, literals, variables)
- `expression.go` - Expression parsing
- `names.go` - Name and identifier parsing
- `util.go` - Utility functions

## Tests

- `cst_test.go` - Basic CST parsing tests
- `resource_option_test.go` - Resource option parsing tests (migrated from TypeScript)

## Resource Option

The `resource` option affects how escape sequences and whitespace are handled:

### With `resource: true`
- Extended escape sequences are supported: `\n`, `\r`, `\t`, `\x01`, `\u0002`, `\U000003`
- Leading whitespace after newlines is trimmed in text and quoted literals

### With `resource: false` (default)
- Only basic escape sequences: `\\`, `\{`, `\|`, `\}`
- Extended escape sequences generate `bad-escape` errors
- Whitespace is preserved as-is

## TypeScript Compatibility

The resource option tests in `resource_option_test.go` are direct migrations from the TypeScript implementation's `resource-option.test.ts`, ensuring identical behavior across implementations. 