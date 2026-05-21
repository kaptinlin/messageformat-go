# Architecture

## Processing Pipeline

MessageFormat Go uses a staged pipeline:

```text
source text -> internal/cst -> pkg/datamodel -> internal/resolve -> internal/selector -> pkg/messagevalue
```

Each stage owns one responsibility:

- `internal/cst` parses source syntax and records parser-level errors.
- `pkg/datamodel` converts CST nodes into the public message model and validates model invariants.
- `internal/resolve` resolves variables, literals, functions, markup, and bidi context.
- `internal/selector` chooses the pattern for `.match` messages.
- `pkg/messagevalue` represents resolved values and structured output parts.

> **Why**: Keeping CST, data model, resolution, selection, and parts separate makes spec behavior testable without turning parser details into public API.
>
> **Rejected**: A single parser/resolver package. It would make official test failures harder to classify and would leak implementation detail into exported types.

## Package Boundaries

Public packages:

- `pkg/datamodel`: public MessageFormat 2.0 data model and validation helpers.
- `pkg/functions`: built-in functions, custom function contracts, and registries.
- `pkg/messagevalue`: resolved values and formatted parts.
- `pkg/parts`: compatibility aliases for part constructors and interfaces.
- `pkg/errors`, `pkg/bidi`, and `pkg/logger`: supporting public utilities.

Internal packages:

- `internal/cst`: parser implementation details.
- `internal/resolve`: expression and function resolution.
- `internal/selector`: pattern selection for `.match`.
- `internal/intlbridge`: translation layer for `github.com/agentable/go-intl`.

Internal packages may depend on public packages. Public packages must not expose internal package types in exported signatures.

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

## Acceptance Criteria

- Public `go doc` output for root, `pkg/datamodel`, `pkg/functions`, and `pkg/messagevalue` does not expose `internal/cst`.
- Official tests continue to pass through `task test-v2`.
- New docs link to SPECS for contracts and to README/docs for usage.
