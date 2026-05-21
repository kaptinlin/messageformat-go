# Public API Contracts

## Root Package

The root package is the primary user-facing API:

- `Parse(locales, source, options...)` parses source text and validates the resulting data model.
- `Compile(locales, message, options...)` accepts a public data model and stores a detached snapshot.
- `Format(values, options...)` returns plain text.
- `FormatToParts(values, options...)` returns structured parts for rich rendering.

`Format(values map[string]any)` remains dynamic. Message parameters come from application data, not from Go compile-time schemas.

> **Why**: The formatter must accept arbitrary application values, but the compiled formatter itself can still be immutable and typed.
>
> **Rejected**: A generic `Format[T any]` API. It would give false precision because message templates choose variable names and functions at runtime.

## Constructor Options

Functional options are the preferred configuration surface. `MessageFormatOptions` exists for callers that already have a configuration struct.

Required defaults:

- `BidiIsolation` defaults to `BidiNone` unless an RTL locale triggers the existing automatic defaulting behavior.
- `LocaleMatcher` defaults to `LocaleBestFit`.
- `Dir` defaults to locale-derived direction when possible.
- Custom functions extend built-ins rather than replacing them.

Constructor input slices and maps must be cloned before storage.

## Data Model

The public data model represents MessageFormat 2.0 messages, declarations, selectors, variants, expressions, markup, options, and attributes.

Rules:

- Public interfaces must use Go unions where possible: `Declaration`, `PatternElement`, `VariantKey`, `ExpressionArg`, `OptionValue`, and `AttributeValue`.
- CST references are parser implementation details and stay behind unexported methods and constructors.
- Constructors must not retain caller-owned mutable slices or maps.
- `Type()` may remain for debugging, JSON-like compatibility, and TypeScript traceability, but internal control flow should prefer type switches and small interfaces.

> **Why**: TypeScript discriminated unions need string tags. Go already has concrete types and type switches, so exported APIs should make invalid states harder to express.

## Functions

Custom functions implement:

```go
type MessageFunction func(
	ctx MessageFunctionContext,
	options Options,
	operand any,
) messagevalue.MessageValue
```

`functions.Options` is a read-oriented helper boundary. Use `String`, `Int`, `Bool`, `Value`, `Has`, and `Map` instead of requiring every custom function to repeat ad hoc map coercion.

Built-in function registry rules:

- `DefaultFunctionMap()` and `DraftFunctionMap()` return detached snapshots.
- `FunctionRegistry` is the mutable extension point.
- `:math` is a TypeScript-compatible extension, not an MF2 spec function.

## Message Values and Parts

`MessageValue` is the common formatting contract. Optional behavior belongs in smaller interfaces:

- `PartsFormatter` emits structured parts.
- `Valuer` exposes an underlying Go value.
- `Selector` participates in pattern selection.
- `OptionedValue` carries formatting options.

Concrete part types must expose typed accessors where the value type is known, while keeping `Value() any` for compatibility.

> **Why**: TypeScript can check optional methods at runtime. Go should not force every value to implement no-op selector or options methods just to satisfy one broad interface.

## Error Handling

Construction errors return `error`. Format-time recoverable errors flow through the configured error handler and degrade to fallback values when possible.

Production code must not panic for user-controlled message syntax, data model shape, or runtime values.

## Forbidden

- Do not expose mutable default function maps.
- Do not require all message values to implement selection.
- Do not require callers to import `internal/` packages.
- Do not use string `Type()` checks where a type switch expresses the same rule.
- Do not return dynamically created sentinel-like errors when a static package-level error is appropriate.

## Acceptance Criteria

- Mutating caller-supplied slices or maps after construction does not change the constructed value.
- Mutating a data model after `Compile` does not change the compiled formatter.
- Mutating a map returned by `DefaultFunctionMap` or `DraftFunctionMap` does not affect new registries or formatters.
- `task lint && task test` passes after API changes.
