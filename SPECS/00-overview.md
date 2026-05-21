# MessageFormat Go Overview

## Scope

MessageFormat Go implements Unicode MessageFormat 2.0 for Go while keeping the supported ICU MessageFormat v1 package under `v1/`.

The root module owns the MessageFormat 2.0 parser, public data model, resolver, selector, built-in functions, structured parts, and examples. The `v1/` package remains product code and must continue to build and test with the root module.

> **Why**: Applications often need both a current MessageFormat 2.0 implementation and a migration path for existing ICU MessageFormat v1 messages. Keeping both in one module makes migration explicit without pretending the old API is dead.
>
> **Rejected**: Splitting `v1/` into a separate module during cleanup. That would move compatibility risk to users without improving the v2 implementation.

## Compatibility Targets

- The MessageFormat 2.0 behavior must follow the Unicode specification and official MessageFormat Working Group tests.
- The TypeScript `messageformat` project is the API compatibility reference for behavior and trace comments, not a requirement to copy JavaScript object shapes.
- Go APIs should expose typed constructors, typed option helpers, immutable snapshots, and explicit error returns where the language makes that practical.
- Dynamic `map[string]any` boundaries are allowed for caller-supplied format values, official JSON-like fixtures, custom function operands, and `v1/`.

> **Why**: MessageFormat inputs are dynamic by design, but Go callers should not have to work through JavaScript-style prototype, discriminator, or optional-method patterns when a smaller typed interface is clearer.

## Runtime Guarantees

- `MessageFormat` instances are immutable after construction and safe for concurrent formatting.
- Constructor inputs that could mutate compiled behavior must be copied before storage.
- Built-in function maps returned by public helpers are snapshots. Mutating one returned map must not affect future formatters or registries.
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
- Do not delete, quarantine, or rename `v1/` as "legacy" without an explicit product deprecation plan.
- Do not expose `internal/cst` or other internal packages from public API signatures.
- Do not add mutable package-level public maps for defaults or registries.
- Do not add JavaScript-specific safety rules, such as prototype-pollution key bans, at Go map boundaries unless the rule belongs to an import/export adapter.

## Acceptance Criteria

- `go doc github.com/kaptinlin/messageformat-go/pkg/datamodel` shows no public `internal/cst` signatures.
- `task lint`, `task test`, and `task test-v2` pass.
- README links users to usage docs while SPECS remains the source of truth for design contracts.
