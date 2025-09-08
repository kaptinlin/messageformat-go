# MessageFormat Go - Developer Guide

This comprehensive guide provides detailed instructions for working with the MessageFormat Go library using Claude Code (claude.ai/code).

## 🚀 Quick Start

**For New Contributors:**
1. **Choose v2** - Always work with v2 (root directory) unless explicitly maintaining v1
2. **Run Tests**: `make test` - Includes official Unicode MessageFormat 2.0 test suite
3. **Check Quality**: `make lint` - Ensures code meets project standards
4. **Follow Patterns**: Use TypeScript-compatible comments (see [Code Style](#code-style-requirements))

**Essential Commands:**
```bash
# v2 Development (Recommended)
make test && make lint    # Test and lint v2 code
make verify              # Complete verification pipeline

# v1 Maintenance (Limited)
cd v1 && make test && make lint    # v1 bug fixes only
```

## 📊 Project Status

### ✅ Production Ready
- **Performance**: 80%+ improvements in v1, optimized v2 implementation
- **Specification**: 100% MessageFormat 2.0 compatibility in v2
- **Quality**: Fixed 86+ lint issues, >80% test coverage
- **Testing**: Comprehensive test suites with official Unicode test validation
- **CI/CD**: Multi-platform automated testing and releases

## Development Commands

### Build & Test Commands

#### All Tests
```bash
# Run all tests (V1 + V2) with race detection
make test

# Run with coverage report
make test-coverage

# Run with verbose output
make test-verbose
```

#### Version-Specific Testing
```bash
# V1 Tests (ICU MessageFormat)
make test-v1

# V2 Tests (MessageFormat 2.0, includes official test suite)
make test-v2

# Official MessageFormat 2.0 test suite only
make test-official
```

#### Examples and Benchmarks
```bash
# Run all examples (V1 + V2)
make examples

# Run benchmarks
make bench
```

### Code Quality

```bash
# Format, vet, lint, and test
make verify

# Individual checks  
make fmt          # Format code
make vet          # Static analysis
make lint         # Comprehensive linting
```

### Coverage and Benchmarks

```bash
# Coverage report
make test-coverage

# Benchmarks
make bench
```

### Prerequisites
- Go 1.24+ required 
- Initialize git submodules for official tests: `git submodule update --init --recursive`
- The official test suite is located in `tests/messageformat-wg/` (git submodule)

## Architecture Overview

This repository contains **dual implementations** of MessageFormat:

### **v2 (Root Directory)** - MessageFormat 2.0 ⭐ **RECOMMENDED**
- **Location**: Root directory (`/`)
- **Specification**: Unicode MessageFormat 2.0 Tech Preview - **100% compliant**
- **Module**: `github.com/kaptinlin/messageformat-go`
- **Purpose**: Modern, specification-compliant implementation
- **Status**: ✅ **Production Ready** with complete TypeScript API compatibility
- **Recent Improvements**: Added `:offset`, `:percent` functions, enhanced selector support

### **v1 (Subdirectory)** - ICU MessageFormat 
- **Location**: `v1/` subdirectory
- **Specification**: Traditional ICU MessageFormat  
- **Module**: `github.com/kaptinlin/messageformat-go/v1`
- **Purpose**: Legacy compatibility and maintenance
- **Status**: ✅ **Optimized** - 80%+ performance improvements using golang.org/x/text
- **Scope**: **Maintenance-only** - bug fixes and security updates

## Import Paths

```go
// v2 (Recommended) - MessageFormat 2.0
import "github.com/kaptinlin/messageformat-go"
import "github.com/kaptinlin/messageformat-go/pkg/datamodel"

// v1 (Legacy) - ICU MessageFormat
import "github.com/kaptinlin/messageformat-go/v1"
```

### v2 Package Structure (MessageFormat 2.0)

#### Main API (`/` root)
- `messageformat.go` - Main MessageFormat 2.0 type and constructor
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

### v1 Package Structure (ICU MessageFormat)

#### Legacy Implementation (`v1/`)
- `messageformat.go` - ICU MessageFormat implementation
- `plurals.go` - Plural rules using golang.org/x/text
- `runtime.go` - Performance-optimized runtime with object pooling
- `formatters.go` - Number, date, and currency formatters
- `errors.go` - Static error definitions for lint compliance
- `parser.go` - ICU message parsing logic
- `perf_test.go` - Performance regression testing

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

## 📚 API Usage Examples

### v2 MessageFormat 2.0 Usage

#### Basic Message Formatting
```go
import "github.com/kaptinlin/messageformat-go"

// Simple message
mf, err := messageformat.New("en", "Hello {name}!")
if err != nil {
    log.Fatal(err)
}

result, err := mf.Format(map[string]interface{}{
    "name": "World",
})
// Output: "Hello World!"
```

#### Advanced Features (v2 Only)
```go
// Plural selection with exact matching
mf, err := messageformat.New("en", `
{count :integer select=cardinal} items remaining:
{one}   You have 1 item left
{other} You have {count} items left
`)

// Number formatting with offset
mf, err := messageformat.New("en", `
You and {count :offset subtract=1} others liked this post
`)

// Percentage formatting  
mf, err := messageformat.New("en", `
Progress: {progress :percent}
`)
```

#### Format to Parts (Rich Formatting)
```go
parts, err := mf.FormatToParts(map[string]interface{}{
    "count": 42,
})
// Returns structured formatting parts for complex UI rendering
```

### v1 ICU MessageFormat Usage (Legacy)
```go
import "github.com/kaptinlin/messageformat-go/v1"

// Basic ICU-style message
mf, err := v1.New("en", "You have {itemCount, plural, =0 {no items} one {one item} other {# items}}.")
result := mf.FormatSimple(map[string]interface{}{
    "itemCount": 5,
})
```

## 🛠️ Development Best Practices

### Code Quality Standards
1. **Always use v2** for new development
2. **TypeScript Comments**: Every function must include original TypeScript code
3. **Error Handling**: Use static error variables, never dynamic error creation
4. **Testing**: >80% coverage required, use testify framework
5. **Performance**: Run benchmarks for critical paths

### Security Guidelines
- Never commit sensitive data or API keys
- Use static error definitions to prevent information leakage
- Validate all user inputs in function implementations
- Follow principle of least privilege in API design

## Common Development Tasks

### Working with v2 (MessageFormat 2.0) - PREFERRED

#### Adding New Functions
1. Implement in `pkg/functions/` with proper TypeScript compatibility comments
2. Add to `DefaultFunctions` or `DraftFunctions` map
3. Write comprehensive table-driven tests
4. Follow the `MessageFunction` type signature

#### Working with Message Parsing
1. CST parsing handles raw syntax in `internal/cst/`
2. DataModel conversion in `pkg/datamodel/fromcst.go`
3. Validation in `pkg/datamodel/validate.go`
4. Resolution in `internal/resolve/`

#### Testing Against Official Suite
The repository includes the official Unicode MessageFormat Working Group test suite as a git submodule. Always run `make test` to ensure specification compliance.

### Working with v1 (ICU MessageFormat) - MAINTENANCE ONLY

#### v1 Development Guidelines
- **Bug fixes and security updates only** - no new features
- Maintain API compatibility with existing v1 users
- Performance optimizations are acceptable
- All changes must pass existing v1 test suite

#### v1 Testing
```bash
cd v1
make test          # Run basic tests
make lint          # Check code quality
```

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

## CI/CD Pipeline

### GitHub Actions Workflows

#### `main.yml` - Primary CI/CD
- **Triggers**: Push to main, PRs, v2.* tags
- **Jobs**: 
  - `test-v2`: v2 unit tests + official test suite
  - `lint-v2`: v2 code quality checks
  - `test-v1`: v1 compatibility tests
  - `lint-v1`: v1 maintenance checks
  - `security`: Security scanning

#### `v1-maintenance.yml` - v1 Legacy Support
- **Triggers**: Changes to v1/ directory, v1.* tags
- **Jobs**: 
  - Cross-platform testing (Ubuntu, Windows, macOS)
  - Performance regression detection
  - API compatibility validation
  - Security vulnerability scanning

#### `release.yml` - Automated Releases
- **Triggers**: v1.* and v2.* tags
- **Features**:
  - Pre-release testing
  - Automatic changelog generation
  - GitHub release creation
  - Documentation deployment
  - Slack notifications (if configured)

### Development Workflow

1. **For v2 development** (preferred):
   ```bash
   git checkout -b feature/new-v2-feature
   # Work in root directory
   make test && make lint
   git commit -m "feat: add new v2 feature"
   ```

2. **For v1 maintenance** (bug fixes only):
   ```bash
   git checkout -b fix/v1-critical-bug
   cd v1
   # Make minimal changes
   make test && make lint
   git commit -m "fix(v1): resolve critical bug"
   ```

3. **Release process**:
   ```bash
   # For v2 releases
   git tag v2.1.0
   git push origin v2.1.0
   
   # For v1 maintenance releases  
   git tag v1.3.1
   git push origin v1.3.1
   ```

## 🔧 Troubleshooting & FAQ

### Common Issues

#### "Tests failing in official test suite"
```bash
# Ensure git submodules are initialized
git submodule update --init --recursive

# Run specific test suite
make test-official
```

#### "Lint errors with golangci-lint"
```bash
# Check specific linter configuration
cat .golangci.yml

# Run with verbose output
golangci-lint run --verbose

# Auto-fix formatting issues
make fmt
```

#### "Performance regression in v1"
```bash
cd v1
make test         # Run V1 tests
make bench        # Run benchmarks
```

### Development FAQ

**Q: Should I use v1 or v2?**
A: **Always use v2** unless you're specifically maintaining v1 legacy code. v2 is production-ready with 100% MessageFormat 2.0 compliance.

**Q: How do I add a new MessageFormat function?**
A: 
1. Implement in `pkg/functions/` with TypeScript compatibility comments
2. Add to `DefaultFunctions` or `DraftFunctions` map in `registry.go`
3. Write comprehensive tests following existing patterns
4. Ensure the function follows the `MessageFunction` signature

**Q: Why are comments required in TypeScript format?**
A: This ensures 100% API compatibility with the official TypeScript reference implementation and helps maintain consistency across implementations.

**Q: How do I run only fast tests during development?**
A: Use `make test-unit` to skip the official test suite, or `go test ./pkg/...` for package-specific tests.

**Q: What's the difference between DefaultFunctions and DraftFunctions?**
A: 
- `DefaultFunctions`: Required by MessageFormat 2.0 spec (`:integer`, `:number`, `:string`, `:offset`)
- `DraftFunctions`: Optional/draft functions (`:currency`, `:date`, `:percent`, `:unit`, etc.)

### Performance Guidelines

#### v2 Optimization Tips
- Use `sync.Pool` for frequently allocated objects
- Minimize allocations in hot paths (formatting functions)
- Benchmark changes: `make bench`
- Profile memory usage: `go test -memprofile=mem.prof -bench=.`

#### v1 Maintenance Notes
- Performance optimizations are acceptable
- Use existing `sync.Pool` patterns in `runtime.go`
- Run `make test` to validate changes

### CI/CD Troubleshooting

#### "GitHub Actions failing"
- Check workflow logs in `.github/workflows/`
- Ensure cross-platform compatibility (Windows, macOS, Linux)
- Verify Go version compatibility (Go 1.24+)

#### "Security scan failures"
- Review security policy in `.github/workflows/`
- Update dependencies: `go mod tidy && go mod vendor`
- Check for known vulnerabilities: `go list -json -m all | nancy sleuth`

## 📖 Reference Links

### Specifications
- [Unicode MessageFormat 2.0](https://github.com/unicode-org/message-format-wg) - Official specification
- [ICU MessageFormat](https://unicode-org.github.io/icu/userguide/format_parse/messages/) - Legacy v1 reference

### Go Resources
- [golang.org/x/text](https://pkg.go.dev/golang.org/x/text) - Unicode text processing
- [Testify](https://github.com/stretchr/testify) - Required testing framework
- [golangci-lint](https://golangci-lint.run/) - Code quality tool

---

**Remember**: v2 is the future 🚀, v1 is maintenance-only 🔧