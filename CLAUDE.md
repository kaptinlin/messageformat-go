# MessageFormat Go

Go implementation of Unicode MessageFormat 2.0 specification with 100% spec compliance. Dual implementation: v2 (MessageFormat 2.0, root directory) is production-ready and recommended; v1 (ICU MessageFormat, v1/ subdirectory) is maintenance-only for legacy compatibility.

**Reference implementation:** TypeScript messageformat library — API compatibility target with identical method signatures and behavior.

## Commands

```bash
# Primary workflow
task test              # Run all tests (v1 + v2) with race detection
task lint              # Run golangci-lint + go mod tidy check
task verify            # Full verification: deps, fmt, vet, lint, test

# Version-specific testing
task test-v2           # Run v2 tests + official MessageFormat 2.0 test suite
task test-v1           # Run v1 tests only
task test-official     # Run official Unicode test suite only
task test-coverage     # Generate coverage report (coverage.html)

# Development
task fmt               # Format code
task vet               # Run go vet
task bench             # Run benchmarks
task examples          # Run all examples (v1 + v2)
task deps              # Download and tidy dependencies
task clean             # Clean build artifacts

# Prerequisites
task submodules        # Initialize git submodules (required for official tests)
```

## Architecture

```
messageformat-go/
├── messageformat.go       # Main MessageFormat 2.0 API
├── options.go            # Functional options pattern
├── exports.go            # Re-exports for TypeScript API compatibility
├── pkg/                  # Public packages
│   ├── datamodel/        # Message data model (Message, Expression, Pattern)
│   ├── functions/        # Built-in and custom function implementations
│   ├── messagevalue/     # Message value types and formatting parts
│   ├── errors/           # Custom error types (Syntax, Resolution, Selection)
│   ├── bidi/             # Bidirectional text support
│   ├── parts/            # Message part representations
│   └── logger/           # Logging utilities
├── internal/             # Private packages
│   ├── cst/              # Concrete Syntax Tree parser
│   ├── resolve/          # Expression resolution and context handling
│   └── selector/         # Pattern selection for .match statements
├── tests/                # Official MessageFormat 2.0 test suite
└── v1/                   # Legacy ICU MessageFormat (maintenance-only)
```

### Key Types and Interfaces

```go
// Main API (root package)
type MessageFormat struct { ... }
func New(locales string, source string, opts ...Option) (*MessageFormat, error)
func (mf *MessageFormat) Format(values map[string]interface{}) (string, error)
func (mf *MessageFormat) FormatToParts(values map[string]interface{}) ([]messagevalue.MessagePart, error)

// Core data model (pkg/datamodel)
type Message struct { ... }        // Root message node
type Expression struct { ... }     // Variable/function expression
type Pattern struct { ... }        // Text pattern with placeholders

// Function system (pkg/functions)
type MessageFunction func(
    ctx MessageFunctionContext,
    options map[string]any,
    operand any,
) messagevalue.MessageValue

// Built-in function registries
var DefaultFunctions map[string]MessageFunction  // :integer, :number, :string, :offset
var DraftFunctions map[string]MessageFunction    // :currency, :date, :datetime, :percent, :unit
```

### Processing Flow

```
Source String → CST (internal/cst) → DataModel (pkg/datamodel) → Resolution (internal/resolve) → MessageParts (pkg/messagevalue)
```

## Design Philosophy

- **TypeScript API Compatibility** — Maintains identical method signatures and behavior with TypeScript reference implementation. Every function includes original TypeScript code in comments for traceability.
- **Specification Compliance** — Strict adherence to Unicode MessageFormat 2.0 specification. Official test suite (git submodule) validates 100% compliance.
- **Two-Phase Processing** — Clean separation: CST parsing → DataModel conversion → Resolution/Formatting. Each phase has clear boundaries and responsibilities.
- **KISS** — Simple, focused implementations. No premature abstractions. Three similar lines are better than a helper used once.
- **DRY** — Shared function registry, unified error types, reusable resolution context across all message types.
- **YAGNI** — Only implement what's currently needed. v1 is maintenance-only; v2 focuses on spec compliance, not feature creep.

## Coding Rules

### Must Follow

- Go 1.26 — use modern language features (generics, slices/maps packages, clear(), for range N)
- **TypeScript comment format** — Every function must include original TypeScript code:
  ```go
  // FunctionName describes what this function does
  // TypeScript original code:
  // export function functionName(param: Type): ReturnType {
  //   // implementation
  // }
  func FunctionName(param Type) ReturnType { ... }
  ```
- **All comments in English only** — No other languages in code comments
- **testify for all tests** — Use `github.com/stretchr/testify/assert` and `testify/require`
- **Table-driven tests** — Use subtests with `t.Run()` for multiple test cases
- **Static error definitions** — Define errors as package-level variables, never create dynamically
- **Thread safety** — MessageFormat instances are immutable after construction, safe for concurrent use
- **Error returns** — All errors returned via `error`, never panic in production code
- Follow Google Go Best Practices: https://google.github.io/go-style/best-practices
- Follow Google Go Style Decisions: https://google.github.io/go-style/decisions

### Go 1.26 Features Used

| Feature | Where Used |
|---------|-----------|
| `maps.Clone()` | Function registry cloning |
| `maps` package | Options and attributes handling |
| `slices` package | Pattern and variant operations |
| Generics | messagevalue types, datamodel nodes |
| `for range N` | Iteration patterns throughout |

### Forbidden

- No `panic` in production code — all errors returned via `error`
- No dynamic error creation — use static error variables for lint compliance
- No premature abstraction — implement only what's currently needed
- No v1 feature additions — v1 is maintenance-only (bug fixes and security updates)
- No breaking changes to TypeScript-compatible API — maintain method signature compatibility

## Testing

- **Framework:** `github.com/stretchr/testify` required for all tests
- **Patterns:** Table-driven tests with `t.Run()`, `t.Parallel()` where safe
- **Coverage target:** >80% for all packages
- **Official test suite:** Git submodule at `tests/messageformat-wg/` validates spec compliance

```bash
# Run specific package tests
go test -race ./pkg/functions/
go test -race ./internal/resolve/

# Run specific test function
go test -race -run TestIntegerFunction ./pkg/functions/

# Run benchmarks
go test -bench=. -benchmem ./pkg/functions/
go test -bench=. -benchmem ./internal/resolve/

# Coverage for specific package
go test -race -coverprofile=coverage.out ./pkg/datamodel/
go tool cover -html=coverage.out
```

## Dependencies

| Dependency | Purpose |
|------------|---------|
| `golang.org/x/text` | Unicode text processing, plural rules, language tags |
| `github.com/dromara/carbon/v2` | Date/time formatting for :date, :datetime, :time functions |
| `github.com/Rhymond/go-money` | Currency formatting for :currency function |
| `github.com/go-json-experiment/json` | JSON v2 experimental API (for future use) |
| `github.com/stretchr/testify` | Testing framework (test-only) |

## Error Handling

Custom error types for different failure modes:

```go
// pkg/errors package
type MessageSyntaxError struct { ... }      // Parsing errors
type MessageResolutionError struct { ... }  // Variable/function resolution errors
type MessageSelectionError struct { ... }   // Pattern selection errors
```

All errors are static package-level variables to prevent information leakage and satisfy lint rules.

## Performance

### v2 Optimization Guidelines

- **Pre-compile regex** — Package-level regex compilation for validation patterns
- **Minimize allocations** — Use value types where possible, avoid unnecessary copying
- **Benchmark critical paths** — Run `task bench` for formatting functions and resolution logic

### v1 Performance (Maintenance-only)

- Uses `sync.Pool` for frequently allocated objects in hot paths
- 80%+ performance improvements over original implementation
- Performance optimizations acceptable; no new features

## Linting

golangci-lint v2.9.0. Config in `.golangci.yml`.

- Strict configuration with standard linters enabled
- Test files excluded from certain rules (funlen, gosec)
- Examples and internal packages have relaxed rules

## CI

GitHub Actions workflows:

- **main.yml** — Primary CI/CD (test-v2, lint-v2, test-v1, lint-v1, security)
- **v1-maintenance.yml** — v1 legacy support (cross-platform, performance regression)
- **release.yml** — Automated releases for v1.* and v2.* tags

Triggers: Push to main, PRs, version tags (v1.*, v2.*)

## Development Workflow

### For v2 Development (Recommended)

```bash
git checkout -b feature/new-v2-feature
# Work in root directory
task test && task lint
git commit -m "feat: add new v2 feature"
```

### For v1 Maintenance (Bug Fixes Only)

```bash
git checkout -b fix/v1-critical-bug
cd v1
# Make minimal changes
task test && task lint
git commit -m "fix(v1): resolve critical bug"
```

### Release Process

```bash
# For v2 releases
git tag v2.1.0
git push origin v2.1.0

# For v1 maintenance releases
git tag v1.3.1
git push origin v1.3.1
```

## Agent Skills

Specialized skills available in `.agents/skills/`:

| Skill | When to Use |
|-------|------------|
| [testing](.agents/skills/testing/) | Writing or reviewing Go tests — testify patterns, table-driven tests, mocking, concurrency testing, benchmarks |
| [linting](.agents/skills/linting/) | Setting up or running golangci-lint v2, fixing lint errors, configuring linters |
| [modernizing](.agents/skills/modernizing/) | Adopting Go 1.20-1.26 new features — generics, iterators, error handling, stdlib collections |
| [committing](.agents/skills/committing/) | Creating conventional commit messages for Go packages |
| [releasing](.agents/skills/releasing/) | Releasing a Go package — semantic versioning, tagging, dependency upgrades |
| [code-simplifying](.agents/skills/code-simplifying/) | Refining recently written Go code for clarity and consistency without changing functionality |
| [go-best-practices](.agents/skills/go-best-practices/) | Applying Google Go style guide — naming, error handling, interfaces, testing, concurrency |
| [agent-md-creating](.agents/skills/agent-md-creating/) | Generate CLAUDE.md and AGENTS.md for Go projects |
| [dependency-selecting](.agents/skills/dependency-selecting/) | Selecting Go dependencies from kaptinlin/agentable ecosystem and vetted external libraries |
| [readme-creating](.agents/skills/readme-creating/) | Generate README.md for Go libraries in kaptinlin and agentable ecosystem |
| [ralphy-initializing](.agents/skills/ralphy-initializing/) | Initialize Ralphy AI coding loop configuration for Go projects |
| [ralphy-todo-creating](.agents/skills/ralphy-todo-creating/) | Create Ralphy TODO.yaml task files from PRDs, plans, or issue trackers |
