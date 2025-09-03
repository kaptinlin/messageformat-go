# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Build & Test Commands
```bash
# Run all tests (includes official Unicode MessageFormat 2.0 test suite)
make test

# Run only unit tests (excludes official test suite)
make test-unit

# Run only official MessageFormat 2.0 test suite  
make test-official

# Run tests with coverage report
make test-coverage

# Run benchmarks
make bench

# Format code, run linters, and tests
make verify

# Run CI pipeline locally
make ci
```

### Linting & Quality
```bash
# Run golangci-lint
make lint

# Format code
make fmt

# Run go vet
make vet

# Clean build artifacts
make clean
```

### Prerequisites
- Go 1.23+ required 
- Initialize git submodules for official tests: `git submodule update --init --recursive`
- The official test suite is located in `tests/messageformat-wg/` (git submodule)

## Architecture Overview

This is a Go implementation of the Unicode MessageFormat 2.0 specification, designed for API compatibility with the TypeScript reference implementation.

### Core Package Structure

#### Main API (`/` root)
- `messageformat.go` - Main MessageFormat type and constructor
- `options.go` - Functional options pattern for configuration
- `exports.go` - Re-exports for convenient access

#### Public Packages (`pkg/`)
- `datamodel/` - Message data model types (Message, Expression, Pattern, etc.)
- `functions/` - Built-in and custom function implementations
- `messagevalue/` - Message value types and formatting parts
- `errors/` - Custom error types for MessageFormat operations
- `bidi/` - Bidirectional text support
- `parts/` - Message part representations
- `logger/` - Logging utilities

#### Internal Packages (`internal/`)
- `cst/` - Concrete Syntax Tree parser and types
- `resolve/` - Expression resolution and context handling  
- `selector/` - Pattern selection logic for .match statements

### Key Architecture Principles

1. **TypeScript API Compatibility**: Maintains identical method signatures and behavior
2. **Specification Compliance**: Strict adherence to MessageFormat 2.0 Unicode specification
3. **Two-Phase Processing**: CST parsing → DataModel conversion → Resolution/Formatting
4. **Thread Safety**: MessageFormat instances are safe for concurrent use after construction

### Core Types Flow
```
Source String → CST (internal/cst) → DataModel (pkg/datamodel) → Resolution (internal/resolve) → MessageParts (pkg/messagevalue)
```

## Important Development Guidelines

### Code Style Requirements

All code must follow the established patterns from `.cursor/rules.mdc`:

1. **Mandatory Comment Format**:
```go
// FunctionName describes what this function does
// TypeScript original code:
// export function functionName(param: Type): ReturnType {
//   // implementation  
// }
func FunctionName(param Type) ReturnType {
    // implementation
}
```

2. **Language Requirements**:
- **ALL COMMENTS MUST BE IN ENGLISH ONLY**
- Include complete TypeScript original code unmodified
- Follow Go naming conventions while preserving API compatibility

3. **Testing Requirements**:
- MUST use `github.com/stretchr/testify` for all testing
- Use table-driven tests with `testify/assert` and `testify/require`
- Maintain test coverage > 80%

### Key Type Mappings (TypeScript → Go)
- `Record<string, unknown>` → `map[string]interface{}`
- `string | string[]` → `interface{}` (handled with type switching)
- `MessageFunction` → `func(ctx MessageFunctionContext, options map[string]interface{}, operand interface{}) MessageValue`

### Error Handling Patterns
- Return errors as last return value
- Use custom error types: `MessageSyntaxError`, `MessageResolutionError`, `MessageSelectionError`
- Collect multiple errors in slices when needed

## Common Development Tasks

### Adding New Functions
1. Implement in `pkg/functions/` with proper TypeScript compatibility comments
2. Add to `DefaultFunctions` or `DraftFunctions` map
3. Write comprehensive table-driven tests
4. Follow the `MessageFunction` type signature

### Working with Message Parsing
1. CST parsing handles raw syntax in `internal/cst/`
2. DataModel conversion in `pkg/datamodel/fromcst.go`
3. Validation in `pkg/datamodel/validate.go`
4. Resolution in `internal/resolve/`

### Testing Against Official Suite
The repository includes the official Unicode MessageFormat Working Group test suite as a git submodule. Always run `make test` to ensure specification compliance.

### Linting Configuration
- Uses golangci-lint v2.1.6 with strict configuration in `.golangci.yml`
- Excludes test files from certain rules (funlen, gosec, etc.)
- Enables comprehensive linters: gocritic, gosec, revive, errorlint, etc.

## Package Dependencies

### Core Dependencies
- `golang.org/x/text` - Unicode text processing
- `github.com/dromara/carbon/v2` - Date/time handling  
- `github.com/Rhymond/go-money` - Currency formatting
- `github.com/stretchr/testify` - Testing framework

### No External Runtime Dependencies
The library is designed to have minimal external dependencies for production use.