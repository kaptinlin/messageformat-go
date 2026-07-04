# Public API Contracts

## Root Package

The root package is the primary user-facing API:

- `Parse(locales, source, options...)` parses source text and validates the resulting data model.
- `Compile(locales, message, options...)` accepts a public data model and stores a detached snapshot.
- `Format(values, options...)` returns the string projection of the selected pattern.
- `FormatToParts(values, options...)` returns the structured-parts projection of the selected pattern.

`Format(values map[string]any)` remains dynamic. Message parameters come from application data, not from Go compile-time schemas.

> **Why**: The formatter must accept arbitrary application values, but the compiled formatter itself can still be immutable and typed.
>
> **Rejected**: A generic `Format[T any]` API. It would give false precision because message templates choose variable names and functions at runtime.

## Rendering Contract

`Format` and `FormatToParts` are two public projections over the same selected pattern. They must agree on selection, fallback source, bidi isolation, locale, and error reporting, but they must not use structured parts as the intermediate representation for string output.

Rules:

- `Format` resolves each expression and calls the resolved message value's string conversion.
- `FormatToParts` resolves each expression and returns the resolved message value's parts.
- Markup affects structured parts and error reporting, but does not emit text in `Format`.
- Fallback string output uses the source expression in braces.
- Unknown `nil` and typed-nil values format as `null` in string output while preserving the original Go value for parts and `ValueOf`.

> **Rejected**: Building one public render result and deriving both string and parts from it. That makes structural part values decide string semantics and hides differences that callers rely on.

## Constructor Options

Functional options are the preferred configuration surface. `MessageFormatOptions` exists for callers that already have a configuration struct.

Required defaults:

- `BidiIsolation` defaults to `BidiDefault`.
- `LocaleMatcher` defaults to `LocaleBestFit`.
- `Dir` defaults to locale-derived direction when possible.
- Custom functions extend built-ins rather than replacing them.

Constructor input slices and maps must be cloned before storage.

String option helpers are part of the public configuration surface:

- `WithBidiIsolationString("none")` disables isolation.
- Any other bidi-isolation string resolves to default isolation.
- `WithDirString` preserves the supplied runtime string in resolved options.
- `WithLocaleMatcherString` preserves the supplied runtime string in resolved options.

> **Why**: Configuration often arrives from files, flags, or cross-language fixtures. The typed helpers are the preferred Go spelling, but string helpers keep the documented option vocabulary loadable without forcing callers to pre-parse every value.

## Data Model

The public data model represents MessageFormat 2.0 messages, declarations, selectors, variants, expressions, markup, options, and attributes.

Rules:

- Public interfaces must use Go unions where possible: `Declaration`, `PatternElement`, `VariantKey`, `ExpressionArg`, `OptionValue`, and `AttributeValue`.
- Runtime nodes store source spans, not parser objects.
- Parser adapters may consume CST values internally, but CST values must not be retained by public runtime nodes.
- Constructors must not retain caller-owned mutable slices or maps.
- Accessors for slices and maps must return detached snapshots.
- `Type()` may remain for debugging and JSON-like compatibility, but internal control flow should prefer type switches and small interfaces.

> **Why**: Go already has concrete types and type switches, so exported APIs should make invalid states harder to express while still supporting data-model inspection and serialization.

## Variable Resolution

Variable lookup has four distinct states:

- missing: no matching value exists in the active scope.
- nil: the scope contains the key with a nil value.
- typed nil: the scope contains a typed nil value.
- unknown: the scope contains a value that is neither a built-in scalar nor a local message value.

Rules:

- Missing values produce fallback values and report through the configured error handler.
- Nil and typed-nil values are found values, not missing variables.
- Unknown values produce `UnknownValue` and preserve their original Go value.
- Normalized key matching may find a variable when the stored key's normalized form matches the requested name.

> **Rejected**: Using `nil` as the only lookup result. Go needs an explicit found bit because nil can be a valid caller value.

## Pattern Selection

Selection for `.match` messages must be deterministic and side-effect aware.

Rules:

- Selector candidate keys are collected in variant order.
- Custom selectors are called only for real selection candidates.
- Selection must not call custom selectors with fabricated probe keys.
- Backtracking must preserve message variant order when multiple candidates survive.

> **Rejected**: Capability probing with fake keys. It changes observable custom selector behavior and makes selection depend on implementation mechanics rather than the message.

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
- `DefaultFunctionMap()` contains stable defaults: `:currency`, `:integer`, `:number`, `:offset`, `:percent`, and `:string`.
- `DraftFunctionMap()` contains draft functions: `:date`, `:datetime`, `:time`, and `:unit`; callers opt in with `WithFunctions`.
- `FunctionRegistry` is the mutable extension point.
- `:math` is an extension function, not an MF2 spec function; callers opt in with `WithFunction`.

## Message Values and Parts

`MessageValue` is the common formatting contract. Optional behavior belongs in smaller interfaces:

- `PartsFormatter` emits structured parts.
- `Valuer` exposes an underlying Go value.
- `Selector` participates in pattern selection.
- `OptionedValue` carries formatting options.

Concrete part types must expose typed accessors where the value type is known, while keeping `Value() any` for compatibility.

> **Why**: Go should not force every value to implement no-op selector or options methods just to satisfy one broad interface.

## Error Handling

Construction errors return `error`. Format-time recoverable errors flow through the configured error handler and degrade to fallback values when possible.

Production code must not panic for user-controlled message syntax, data model shape, or runtime values.

Rules:

- `MessageError.Type` remains the stable string category for callers that consume serialized or fixture-like errors.
- `MessageError.Kind()` returns the typed Go identity for `errors.Is` and `errors.As`.
- Error tests should assert kind, source/span, cause, and behavior before asserting full prose.

## Forbidden

- Do not expose mutable default function maps.
- Do not require all message values to implement selection.
- Do not require callers to import `internal/` packages.
- Do not use string `Type()` checks where a type switch expresses the same rule.
- Do not return dynamically created sentinel-like errors when a static package-level error is appropriate.
- Do not disable bidi isolation by default.
- Do not collapse `Format` and `FormatToParts` into one public intermediate representation.
- Do not stringify unknown Go objects before `UnknownValue` can preserve them.

## Acceptance Criteria

- Mutating caller-supplied slices or maps after construction does not change the constructed value.
- Mutating a data model after `Compile` does not change the compiled formatter.
- Mutating a map returned by `DefaultFunctionMap` or `DraftFunctionMap` does not affect new registries or formatters.
- Tests cover default bidi isolation, string option helpers, resolved options, custom function registration, `Format`, and `FormatToParts`.
- Tests cover selector no-probe behavior and deterministic candidate order.
- Tests cover missing, nil, typed nil, and unknown variable states.
- Tests cover accessor snapshots and compile-time data model snapshots.
- Tests cover `errors.Is`, `errors.As`, and `Kind()`.
- `task lint`, `task test-v2`, and `task test` pass after API changes.
