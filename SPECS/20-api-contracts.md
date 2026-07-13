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

Constructor option values must use the exported typed constants. `Parse` and
`Compile` reject values outside those closed vocabularies with
`ErrInvalidOption`.

There is no parallel string-helper vocabulary. Callers that load strings from
configuration convert them to the exported types, then let `Parse` or `Compile`
validate the value. Logging is instance-scoped through `WithLogger`; a nil
logger resolves to `slog.Default()` without package-owned mutable logger state.

## Data Model

The public data model represents MessageFormat 2.0 messages, declarations, selectors, variants, expressions, markup, options, and attributes.

Rules:

- Public variant interfaces use package-defined closed Go unions: `Message`, `Declaration`, `PatternElement`, `VariantKey`, `ExpressionArg`, `OptionValue`, and `AttributeValue`.
- Runtime nodes store source spans, not parser objects.
- `ParseMessage` owns CST conversion. CST adapters remain internal and CST values must not be retained by public runtime nodes.
- Constructors must not retain caller-owned mutable slices or maps.
- Accessors for slices and maps must return detached snapshots.
- Parser failures return syntax errors; a well-formed parse result that violates data-model invariants returns a data-model error.
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

`Options()` on package-owned values and parts returns a detached shallow copy.
Resolver wrappers must preserve that ownership boundary even when a custom
`OptionedValue` exposes its own mutable map. Opaque option values and functions
retain identity; the package copies containers, not arbitrary object graphs.

> **Why**: Go should not force every value to implement no-op selector or options methods just to satisfy one broad interface.

## MF1 Module

The independent module `github.com/kaptinlin/messageformat-go/mf1` exposes one
typed entry point per caller job:

```go
func New(locale string, options *MessageFormatOptions) (*MessageFormat, error)
func NewWithPlural(plural PluralFunction, options *MessageFormatOptions) (*MessageFormat, error)
func SupportedLocalesOf(locales []string) ([]string, error)
func GetPlural(locale string) (PluralObject, error)
```

Rules:

- Empty, wildcard, and tags rejected by strict locale parsing return `ErrInvalidLocale`.
- A syntactically valid locale without plural data uses the package-private stable fallback locale.
- `SupportedLocalesOf` filters the caller-owned candidate list in input order; the module does not publish a partial global locale catalog.
- `NewWithPlural` rejects nil with `ErrInvalidPluralFunction`.
- Constructor option maps and plural containers are snapshots. `ResolvedOptions` returns detached maps, slices, and plural-category slices.
- `MessageFormat` may compile and execute concurrently after construction.

> **Rejected**: A JavaScript-shaped `New(any, ...)`, wildcard locale catalog,
> mutable exported default locale, or compatibility overloads. Those surfaces
> combine unrelated jobs and admit states that the runtime does not use.

## Error Handling

Construction errors return `error`. Format-time recoverable errors flow through the configured error handler and degrade to fallback values when possible.

Production code must not panic for user-controlled message syntax, data model shape, or runtime values.

Rules:

- `MessageError.Type` remains the stable string category for callers that consume serialized or fixture-like errors.
- `MessageError.Kind()` returns the typed Go identity for `errors.Is` and `errors.As`.
- Error tests should assert kind, source/span, cause, and behavior before asserting full prose.

## Forbidden

- Do not expose mutable default function maps.
- Do not add string compatibility helpers for typed constructor options.
- Do not add package-wide mutable logger state.
- Do not require all message values to implement selection.
- Do not require callers to import `internal/` packages.
- Do not reopen the sealed data model with public CST adapters or external message implementations.
- Do not add MF1 wildcard or partial global locale catalogs.
- Do not use string `Type()` checks where a type switch expresses the same rule.
- Do not return dynamically created sentinel-like errors when a static package-level error is appropriate.
- Do not disable bidi isolation by default.
- Do not collapse `Format` and `FormatToParts` into one public intermediate representation.
- Do not stringify unknown Go objects before `UnknownValue` can preserve them.

## Acceptance Criteria

- Mutating caller-supplied slices or maps after construction does not change the constructed value.
- Mutating a data model after `Compile` does not change the compiled formatter.
- Mutating a map returned by `DefaultFunctionMap` or `DraftFunctionMap` does not affect new registries or formatters.
- `options_test.go` proves every constructor option vocabulary and `ErrInvalidOption` path through both `Parse` and `Compile`.
- `pkg/messagevalue/value_test.go` and `internal/resolve/function_ref_test.go` prove option-accessor snapshot ownership.
- `pkg/datamodel/fromcst_test.go` and `pkg/datamodel/validate_test.go` prove syntax/data-model error ownership and closed construction paths.
- `mf1/constructor_external_test.go` and `mf1/messageformat_test.go` prove typed constructors, malformed-versus-unsupported locale behavior, detached resolved options, and concurrent use.
- Tests cover default bidi isolation, resolved options, custom function registration, `Format`, and `FormatToParts`.
- Tests cover selector no-probe behavior and deterministic candidate order.
- Tests cover missing, nil, typed nil, and unknown variable states.
- Tests cover accessor snapshots and compile-time data model snapshots.
- Tests cover `errors.Is`, `errors.As`, and `Kind()`.
- `task verify` passes after API changes.
