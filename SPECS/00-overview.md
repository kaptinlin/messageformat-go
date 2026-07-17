# MessageFormat Go Overview

## Scope

MessageFormat Go implements Unicode MessageFormat 2.0 for Go and maintains ICU MessageFormat v1 as the independent module under `mf1/`.

The root module owns the MessageFormat 2.0 parser, public data model, resolver, selector, built-in functions, structured parts, and examples. The `mf1/` module owns the supported ICU MessageFormat v1 compiler and has its own `go.mod`, tests, documentation, and module tags. Repository-wide verification covers both modules without creating a runtime dependency between them.

> **Why**: Applications often need both a current MessageFormat 2.0 implementation and a migration path for existing ICU MessageFormat v1 messages. Keeping both products in one repository makes that relationship visible while preserving truthful Go module boundaries.
>
> **Rejected**: Treating `mf1/` as cleanup residue or a package shipped by the root module. Both descriptions contradict its supported surface and manifest.

## Compatibility Targets

- The `tests/messageformat-wg` gitlink pins the official MessageFormat Working Group corpus used to verify MessageFormat 2.0 behavior.
- The public API must preserve the established MessageFormat 2 constructor, option, formatting, parts, and error semantics documented in [`20-api-contracts.md`](20-api-contracts.md).
- The TypeScript implementation is evidence for vocabulary and observable behavior, not a requirement to copy JavaScript unions, mutable objects, exceptions, or compatibility helpers into Go.
- Go APIs use typed inputs and error returns where they make invalid states harder to express without changing required formatting behavior.
- Dynamic `map[string]any` boundaries remain only where caller data or formatter extension values are inherently dynamic.

> **Why**: MessageFormat inputs are dynamic by design, but Go callers still need immutable formatters, explicit errors, and typed helper APIs where those helpers do not alter the cross-language contract.

## Runtime Guarantees

- Root and MF1 formatter instances are immutable after construction and safe for concurrent use.
- Constructor inputs that could mutate compiled behavior must be copied before storage.
- Built-in function maps returned by public helpers are snapshots. Mutating one returned map must not affect future formatters.
- Public collection accessors, including resolved options and message-value options, must return detached snapshots.
- Public data model variants are package-defined closed unions; construction and validation must not expose CST types or retain caller-owned mutable storage.
- Runtime data model nodes may preserve source positions, but must not retain parser objects.
- Public package documentation must not expose import paths from `internal/`.

## Conformance

Initialize the pinned official fixture explicitly, then run the read-only repository gate:

```bash
task submodules
task verify
```

`task verify` checks vet, tidy state, lint, race-enabled tests, and vulnerabilities for both Go modules. It does not format source, update module files, or change submodule pins. A passing official-suite result applies to the current `tests/messageformat-wg` gitlink; a different pin requires a new run.

To run only the MessageFormat 2.0 package and official-corpus checks:

```bash
task test-v2
```

## Forbidden

- Do not weaken MessageFormat 2.0 conformance to simplify an implementation.
- Do not delete, quarantine, or rename `mf1/` as "legacy" without an explicit product deprecation plan.
- Do not describe `mf1/` as part of the root Go module or add a root-to-MF1 runtime dependency.
- Do not expose `internal/cst` or other internal packages from public API signatures.
- Do not add mutable package-level public maps for defaults or extension configuration.
- Do not preserve obsolete API shapes with compatibility aliases when typed construction and one validation owner express the caller job directly.

## Acceptance Criteria

- `go doc github.com/kaptinlin/messageformat-go/pkg/datamodel` shows no public `internal/cst` signatures.
- `go list -m` from the repository root and from `mf1/` reports two independent module paths.
- README and examples show the documented default path and describe opt-out behavior only as configuration.
- `task verify` passes without changing tracked files or submodule gitlinks.
- SPECS remains the source of truth for design contracts; README and `docs/` remain usage guides.
