# MessageFormat Go

Go implementation of Unicode MessageFormat 2.0, verified against the repository's pinned official test corpus.

**Reference implementation:** TypeScript messageformat library — evidence for option vocabulary and observable behavior, not a requirement to copy JavaScript unions, mutability, or exceptions into Go.

## Commands

```bash
# Primary workflow
task test              # Run all tests with race detection
task lint              # Check tidy state and lint both modules
task verify            # Read-only vet, lint, test, and vuln checks for both modules

# Version-specific testing
task test-v2           # Run package tests + official MessageFormat 2.0 test suite
task test-official     # Run official Unicode test suite only
task test-coverage     # Generate coverage report (coverage.html)

# Development
task fmt               # Format code
task vet               # Run go vet
task bench             # Run benchmarks
task examples          # Run all examples
task deps              # Download and tidy dependencies
task clean             # Clean build artifacts

# Prerequisites
task submodules        # Initialize the official test fixture
```

## Agent Operating Rules

- **Think before coding** — State assumptions, surface ambiguity, and choose the simplest interpretation that preserves correctness.
- **KISS/DRY/YAGNI** — Prefer the standard library, platform behavior, and existing utilities; share duplicated knowledge, not merely similar shape.
- **Development-cost realism** — Do not weaken architecture or maintainability solely to reduce an estimate.
- **User-path bug fixes** — Reproduce through the closest public path, then verify with focused and repository gates.
- **Vertical slice** — Establish the thinnest end-to-end behavior before expanding edge cases and polish.
- **Surgical changes** — Touch only the task-owned surface and preserve unrelated worktree changes.
- **Goal-driven execution** — Define observable completion criteria and continue through verification when implementation is requested.
- **Respect context budgets** — Keep evidence focused and checkpoint long work instead of silently dropping scope.
- **Resolve conflicts explicitly** — Prefer the newer, tested, local convention and identify stale alternatives for removal.
- **Read before writing** — Inspect contracts, callers, tests, and shared utilities before changing a surface.
- **Test intent** — Prove public behavior and regressions; do not mirror prose after stronger behavior coverage exists.
- **Fail loud** — Report skipped checks, uncertainty, and partial work; never call unverified work complete.

## Architecture

```text
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
│   └── parts/            # Message part representations
├── internal/             # Private packages
│   ├── cst/              # Concrete Syntax Tree parser
│   ├── resolve/          # Expression resolution and context handling
│   └── selector/         # Pattern selection for .match statements
├── mf1/                  # Independent ICU MessageFormat v1 module, kept as product code
├── tests/                # Official MessageFormat 2.0 test suite
└── examples/             # Example programs
```

### Processing Flow

```text
Source String → CST (internal/cst) → DataModel (pkg/datamodel) → Resolution (internal/resolve) → MessageParts (pkg/messagevalue)
```

## Agent Workflow

1. Read the relevant `SPECS/` owner before designing or modifying code.
2. Use CodeGraph first when `.codegraph/` exists and code relationships need to be located or understood.
3. Consult `.reference/messageformat` only for vocabulary and observable behavior that matches the local problem; do not copy dynamic JavaScript surfaces mechanically.
4. Run the narrowest relevant tests while iterating, then run `task verify` before closeout.

## SPECS Index

| Spec | Owner |
|------|-------|
| [`SPECS/00-overview.md`](SPECS/00-overview.md) | Product scope, module identities, runtime guarantees, verification boundary |
| [`SPECS/20-api-contracts.md`](SPECS/20-api-contracts.md) | Root and MF1 public APIs, ownership, errors, options, values, and parts |
| [`SPECS/40-architecture.md`](SPECS/40-architecture.md) | Package/module boundaries, pipeline, tests, CI, and documentation ownership |

## Design Philosophy

- **Behavioral Reference, Go-Native Boundary** — Use the TypeScript implementation to verify vocabulary and behavior, while expressing caller jobs with typed Go inputs, explicit errors, and detached ownership.
- **Specification Alignment** — The `tests/messageformat-wg` gitlink pins the official corpus used by `task test-official`. A passing run describes that pin; a changed pin requires a new verification run.
- **Two-Phase Processing** — Clean separation: CST parsing → DataModel conversion → Resolution/Formatting. Each phase has clear boundaries and responsibilities.
- **Keep `mf1` Intact** — The independent `mf1/` module is a supported compatibility surface, not legacy code. Do not prune, delete, sideline, or treat it as dead code during cleanup, modernization, or refactoring.
- **KISS** — Simple, focused implementations. No premature abstractions. Three similar lines are better than a helper used once.
- **DRY** — Shared built-in function catalogs, unified error types, reusable resolution context across all message types.
- **YAGNI** — Only implement what's currently needed. Focus on spec alignment, not feature creep.

## Coding Rules

### Must Follow

- Use the Go toolchain pinned in `go.mod` — assume its modern language features are available (generics, `slices` / `maps` packages, `clear()`, `for range N`, range-over-func iterators)
- **Code comments** — Write comments in English, add godoc for exported APIs, and give concise rationale only for non-obvious contracts, invariants, algorithms, or deliberate divergence
- **Reference provenance** — When upstream provenance matters, cite an exact, openable repository-relative path such as `.reference/messageformat/mf2/messageformat/src/messageformat.ts`; do not copy reference implementation bodies into Go comments
- **testify for all tests** — Use `github.com/stretchr/testify/assert` and `testify/require`
- **Table-driven tests** — Use subtests with `t.Run()` for multiple test cases
- **Stable error identity** — Define sentinel identities at package scope and wrap them when contextual detail is required
- **Thread safety** — MessageFormat instances are immutable after construction, safe for concurrent use
- **Closed options** — Root constructor enum values use exported typed constants; `Parse` and `Compile` reject unsupported values with `ErrInvalidOption`
- **Snapshot ownership** — Clone caller-owned containers before storage and return detached maps/slices from public inspection accessors
- **Typed MF1 jobs** — Use `mf1.New(string, ...)` for locale lookup and `mf1.NewWithPlural(mf1.PluralProfile, ...)` for custom plural behavior; malformed locales fail while valid unsupported locales use the stable fallback
- **Error returns** — All errors returned via `error`, never panic in production code
- **Preserve `mf1` support** — Changes must keep `mf1/` building and tested; do not remove or downgrade `mf1` because it supports ICU MessageFormat v1
- Follow Google Go Best Practices: <https://google.github.io/go-style/best-practices>
- Follow Google Go Style Decisions: <https://google.github.io/go-style/decisions>

### Modern Go Features Used

| Feature | Where Used |
|---------|-----------|
| `maps.Clone()` | Function catalog and option snapshots |
| `maps` package | Options and attributes handling |
| `slices` package | Pattern and variant operations |
| Generics | `messagevalue` types, `datamodel` nodes |
| `for range N` | Iteration patterns throughout |

### Forbidden

- No `panic` in production code — all errors returned via `error`
- No dynamic error creation — use static error variables for lint compliance
- No premature abstraction — implement only what's currently needed
- Do not reintroduce removed string option wrappers, MF1 dynamic overloads, partial locale catalogs, or package-wide mutable logger state
- No working around dependency bugs — report them in `reports/<dependency-name>.md` instead of reimplementing dependency behavior
- No documentation masquerading as code, policy-only gates that restate docs, or tests that only mirror SPECS
- Do not copy JavaScript dynamic unions when a typed Go API expresses the supported caller jobs directly
- Do not classify `mf1/` as legacy, deprecated, or prune-only code unless the user explicitly requests a product-level deprecation plan

## Testing

- **Framework:** `github.com/stretchr/testify` required for all tests
- **Patterns:** Table-driven tests with `t.Run()`, `t.Parallel()` where safe
- **Coverage:** Use coverage reports to find untested behavior; acceptance depends on behavior, race, lint, and pinned-corpus gates rather than a fixed percentage
- **Official test suite:** Git submodule at `tests/messageformat-wg/` pins the corpus checked by `task test-official`

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
| `github.com/agentable/go-intl` | ECMA-402 Intl runtime — backs `:number` / `:integer` / `:string` / `:currency` / `:date` / `:datetime` / `:time` / `:percent` / `:unit` via `numberformat`, `datetimeformat`, `pluralrules`. Constructors are variadic `New(loc, options ...Options)`; `Append*` and `CanonicalKey` helpers do not exist on the dependency surface and must not be relied on. Delegated CLDR/TZDB validation covers MF2 well-formed `timeZone` / `calendar` / `unit` values. |
| `golang.org/x/text` | BCP 47 / Unicode text utilities for `pkg/datamodel`, `internal/cst`, and `pkg/messagevalue/string.go` (no longer used for plural/number/message) |
| `github.com/go-json-experiment/json` | JSON v2 decoding for number options and official corpus fixtures |
| `github.com/stretchr/testify` | Testing framework (test-only) |

## Error Handling

Custom error types for different failure modes:

```go
// pkg/errors package
type MessageSyntaxError struct { ... }      // Parsing errors
type MessageDataModelError struct { ... }   // Parsed model invariant errors
type MessageResolutionError struct { ... }  // Variable/function resolution errors
type MessageSelectionError struct { ... }   // Pattern selection errors
```

Stable error identities are package-level variables or typed error kinds. Add context by wrapping those identities; do not create dynamic sentinel-like errors.

## Dependency Issue Reporting

When a dependency bug or limitation blocks work, do not reimplement the dependency or silently bypass it. Create `reports/<dependency-name>.md` with the dependency/version, trigger, expected and actual behavior, relevant errors, and a proposed upstream resolution; continue only with work that does not depend on the defect.

## Performance

- **Pre-compile regex** — Package-level regex compilation for validation patterns
- **Minimize allocations** — Use value types where possible, avoid unnecessary copying
- **Benchmark critical paths** — Run `task bench` for formatting functions and resolution logic

## Linting

golangci-lint pinned via `.golangci.version`. Config in `.golangci.yml`.

- Strict configuration with standard linters enabled
- Test files excluded from certain rules (funlen, gosec)
- Examples and internal packages have relaxed rules
- The `//go:fix inline` directive is incompatible with `govet inline` on generic functions and must not be applied to generic helpers (see `mf1/types.go::Ptr`).

## CI

GitHub Actions workflows:

- **ci.yml** — Primary CI checks (tests, lint, security)

CI initializes only the pinned official fixture, caches both module manifests, and checks root and MF1 independently. Triggers: pushes to main and pull requests.

## Development Workflow

### Development

```bash
task verify
git commit -m "feat: add new feature"
```

### Release Process

```bash
git tag -a vX.Y.Z -m vX.Y.Z
git tag -a mf1/vX.Y.Z -m mf1/vX.Y.Z
git push origin main vX.Y.Z mf1/vX.Y.Z
```

## Agent Skills

Specialized skills available in `.agents/skills/`:

| Skill | When to Use |
|-------|------------|
| [agent-md-writing](.agents/skills/agent-md-writing/) | Revising project-specific CLAUDE.md and AGENTS.md development guidance |
| [api-surface-designing](.agents/skills/api-surface-designing/) | Designing exported types, options, errors, defaults, and caller-facing contracts |
| [code-review](.agents/skills/code-review/) | Reviewing completed changes for contract fit, regressions, tests, and safety |
| [committing](.agents/skills/committing/) | Creating narrow conventional commits after staged-diff and gate review |
| [concurrency-hardening](.agents/skills/concurrency-hardening/) | Reviewing goroutines, shared state, immutability, and race-sensitive code |
| [github-actions-configuring](.agents/skills/github-actions-configuring/) | Configuring Go CI workflows and multi-module cache/check ownership |
| [go-best-practices](.agents/skills/go-best-practices/) | Applying Google Go style, errors, naming, interfaces, tests, and concurrency guidance |
| [golangci-linting](.agents/skills/golangci-linting/) | Running or configuring golangci-lint v2 and fixing reported issues |
| [improvement-proposing](.agents/skills/improvement-proposing/) | Producing evidence-backed, bounded improvement proposals for this repository |
| [modernizing](.agents/skills/modernizing/) | Adopting Go 1.20-1.26 new features — generics, iterators, error handling, stdlib collections |
| [releasing](.agents/skills/releasing/) | Releasing a Go package — semantic versioning, tagging, dependency upgrades |
| [readme-writing](.agents/skills/readme-writing/) | Maintaining usage-first installation, examples, configuration, and command documentation |
| [spec-writing](.agents/skills/spec-writing/) | Maintaining durable target-state contracts with verification paths |
| [tdd-implementing](.agents/skills/tdd-implementing/) | Implementing behavior changes through focused red-green-refactor cycles |
