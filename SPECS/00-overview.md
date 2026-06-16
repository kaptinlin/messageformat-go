# MessageFormat Go Overview

## Scope

MessageFormat Go implements Unicode MessageFormat 2.0 for Go while keeping the supported ICU MessageFormat v1 package under `mf1/`.

The root module owns the MessageFormat 2.0 parser, public data model, resolver, selector, built-in functions, structured parts, and examples. The `mf1/` package remains product code and must continue to build and test with the root module.

> **Why**: Applications often need both a current MessageFormat 2.0 implementation and a migration path for existing ICU MessageFormat v1 messages. Keeping both in one module makes migration explicit without pretending the old API is dead.
>
> **Rejected**: Treating `mf1/` as cleanup residue. That would move compatibility risk to users without improving the v2 implementation.

## Compatibility Targets

- The MessageFormat 2.0 behavior must follow the Unicode specification and official MessageFormat Working Group tests.
- The public API must preserve the established MessageFormat 2 constructor, option, formatting, parts, and error semantics documented in [`20-api-contracts.md`](20-api-contracts.md).
- Go conveniences are allowed only when they do not change public defaults, option meanings, error identity, or formatting behavior.
- Dynamic `map[string]any` boundaries are allowed for caller-supplied format values, official JSON-like fixtures, custom function operands, and `mf1/`.

> **Why**: MessageFormat inputs are dynamic by design, but Go callers still need immutable formatters, explicit errors, and typed helper APIs where those helpers do not alter the cross-language contract.

## Runtime Guarantees

- `MessageFormat` instances are immutable after construction and safe for concurrent formatting.
- Constructor inputs that could mutate compiled behavior must be copied before storage.
- Built-in function maps returned by public helpers are snapshots. Mutating one returned map must not affect future formatters or registries.
- Public data model accessors must not expose internal mutable slice or map storage.
- Runtime data model nodes may preserve source positions, but must not retain parser objects.
- Public package documentation must not expose import paths from `internal/`.

## Conformance

The primary conformance gate is:

```bash
task test-v2
```

This runs package tests and the official MessageFormat 2.0 suite with race detection. The broader repository gate is:

```bash
task lint && task test
```

## Forbidden

- Do not weaken MessageFormat 2.0 conformance to simplify an implementation.
- Do not delete, quarantine, or rename `mf1/` as "legacy" without an explicit product deprecation plan.
- Do not expose `internal/cst` or other internal packages from public API signatures.
- Do not add mutable package-level public maps for defaults or registries.
- Do not replace the documented public surface with a language-native redesign when the established API already defines the caller contract.

## Acceptance Criteria

- `go doc github.com/kaptinlin/messageformat-go/pkg/datamodel` shows no public `internal/cst` signatures.
- README and examples show the documented default path and describe opt-out behavior only as configuration.
- `task lint`, `task test`, and `task test-v2` pass.
- SPECS remains the source of truth for design contracts; README and `docs/` remain usage guides.
