# Architecture

## Processing Pipeline

MessageFormat Go uses a staged pipeline:

```text
source text -> internal/cst -> pkg/datamodel -> selection/resolution -> pkg/messagevalue -> string or parts
```

Each stage owns one responsibility:

- `internal/cst` parses source syntax and records parser-level errors.
- `pkg/datamodel` converts parser output into the public message model and validates model invariants.
- `internal/selector` chooses the pattern for `.match` messages, using resolved selector values.
- `internal/resolve` resolves variables, literals, functions, markup, and bidi context for both selection and final rendering.
- `pkg/messagevalue` represents resolved values, string conversion, value preservation, and structured output parts.

> **Why**: Keeping CST, data model, resolution, selection, and parts separate makes spec behavior testable without turning parser details into public API.
>
> **Rejected**: A single parser/resolver package. It would make official test failures harder to classify and would leak implementation detail into exported types.

## Package Boundaries

Public packages:

- `pkg/datamodel`: public MessageFormat 2.0 data model and validation helpers.
- `pkg/functions`: built-in functions, custom function contracts, and registries.
- `pkg/messagevalue`: resolved values and formatted parts.
- `pkg/parts`: compatibility aliases for part constructors and interfaces.
- `pkg/errors` and `pkg/bidi`: supporting public utilities.

Internal packages:

- `internal/cst`: parser implementation details.
- `internal/resolve`: expression and function resolution.
- `internal/selector`: pattern selection for `.match`.
- `internal/intlbridge`: translation layer for `github.com/agentable/go-intl`.

Internal packages may depend on public packages. Public packages must not expose internal package types in exported signatures.

## Module Boundaries

The repository contains two published Go modules:

- `github.com/kaptinlin/messageformat-go` at the repository root owns MessageFormat 2.0.
- `github.com/kaptinlin/messageformat-go/mf1` under `mf1/` owns ICU MessageFormat v1 compatibility.

Each module has its own manifest, dependency graph, tests, lint invocation, vet
pass, and vulnerability scan. Neither module requires the other. Repository
automation may aggregate their checks, but it must not blur their import or
release identities.

## Data Model Boundary

`pkg/datamodel` owns the public message IR. Parser output is an input to construction, not part of runtime node identity.

Rules:

- Data model nodes may store source spans for diagnostics.
- Data model nodes must not retain CST or parser object references.
- Accessors must return detached slices or maps when returning collection data.
- `Compile` must clone the supplied data model before storing it.

> **Rejected**: Keeping parser nodes attached to public data model nodes for convenience. It couples diagnostics to parser internals and weakens formatter immutability.

## Selection Boundary

`internal/selector` owns pattern selection. Resolution may provide selector-capable values, but selection must not infer capability by causing fake selector calls.

Rules:

- Candidate keys are derived from real variants.
- Candidate key order follows variant order.
- Custom selectors are called only with real candidate keys.
- Selection errors flow through the formatting error path and degrade to fallback behavior where possible.

## Rendering Boundary

`messageformat.go` owns final rendering. The resolver returns `MessageValue`; final rendering decides whether the caller asked for string output or structured parts.

Rules:

- `Format` uses message value string conversion.
- `FormatToParts` uses message value parts.
- Shared facts are selection, fallback source, bidi isolation, locale, and error reporting.
- Public parts must not become the internal string-rendering representation.

## Intl Boundary

`github.com/agentable/go-intl` is the ECMA-402 runtime boundary for number, date/time, currency, percent, unit, string, and plural behavior.

Rules:

- Keep dependency-specific adaptation in a narrow package or function boundary.
- Do not rely on helper APIs that are not part of the dependency surface.
- Let the dependency own CLDR and timezone validation where it already provides that behavior.

> **Why**: ECMA-402 behavior changes with locale data and runtime semantics. A narrow bridge keeps those changes from spreading through parser and data model code.

## Tests

The test suite has three layers:

- Package tests for local invariants and edge cases.
- Root and examples tests for public usage.
- Official MessageFormat Working Group tests under `tests/`.

Use `testify/assert` and `testify/require` in tests. Prefer table-driven subtests for option matrices, function behavior, and parse/format cases.

Required invariant coverage:

- selector no-probe and deterministic candidate order
- missing, nil, typed-nil, and unknown variable states
- data model accessor snapshots and compile-time snapshots
- string and parts projection behavior
- typed error identity through `errors.Is`, `errors.As`, and `Kind()`
- documentation examples that show the default public path

## Verification Boundary

`Taskfile.yml` owns local check semantics. `task verify` aggregates root and MF1
vet, tidy checks, lint, race-enabled tests, and vulnerability scans. Verification
is observational: it must not run dependency updates, source formatters, or
submodule initialization, and it must leave tracked files and gitlinks unchanged.

`.github/workflows/ci.yml` owns CI environment and orchestration. It invokes the
Taskfile checks, caches both `go.sum` files, and lints both module directories.
Only the test job initializes `tests/messageformat-wg`; `.reference/messageformat`
is evidence for maintainers and is never a build or CI prerequisite.

The official test result is scoped to the corpus commit pinned by the
`tests/messageformat-wg` gitlink. Documentation must point to that pin and the
reproduction command instead of recording a case count or an absolute
compliance percentage.

## Documentation Boundaries

- README is a usage guide.
- `docs/` contains tutorials and API reference for users.
- `SPECS/` contains design contracts and architectural decisions.
- AGENTS/CLAUDE guidance contains agent workflow and coding conventions.

## Forbidden

- Do not import `internal/cst` from public API examples or user-facing signatures.
- Do not duplicate design contracts in README when a SPECS file owns them.
- Do not move formatting behavior into parser packages.
- Do not make official JSON fixture helpers the shape of the public Go API.
- Do not add new public packages for one-off helpers unless the boundary is stable and user-facing.
- Do not make `.reference/messageformat` or recursive submodule updates a verification prerequisite.
- Do not run mutating dependency, formatting, or submodule commands from `task verify`.
- Do not keep temporary planning files once their durable decisions have been accepted into SPECS.

## Acceptance Criteria

- Public `go doc` output for root, `pkg/datamodel`, `pkg/functions`, and `pkg/messagevalue` does not expose `internal/cst`.
- `task verify` covers both module manifests and passes without changing tracked files or submodule gitlinks.
- CI cache inputs include `go.sum` and `mf1/go.sum`, and only the test job initializes the official fixture.
- Official tests continue to pass through `task test-v2` against the pinned corpus.
- New docs link to SPECS for contracts and to README/docs for usage.
- Temporary roadmap files are absent after accepted decisions move into SPECS.
