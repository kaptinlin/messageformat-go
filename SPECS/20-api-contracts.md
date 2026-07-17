# Public API Contracts

## Root Package

The root package is the primary user-facing API:

- `Parse(locales, source, options...)` parses source text and validates the resulting data model.
- `Compile(locales, message, options...)` accepts the sealed, immutable public data model. Model constructors snapshot mutable inputs before `Compile` can retain the value.
- `Format(values)` returns the string projection of the selected pattern and any runtime diagnostics.
- `FormatToParts(values)` returns the structured-parts projection of the selected pattern and any runtime diagnostics.

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
validate the value. Formatting does not log or invoke package-owned callbacks.

## Data Model

The public data model represents MessageFormat 2.0 messages, declarations, selectors, variants, expressions, markup, options, and attributes.

Rules:

- Public variant interfaces use package-defined closed Go unions: `Message`, `Declaration`, `PatternElement`, `VariantKey`, `ExpressionArg`, `OptionValue`, and `AttributeValue`.
- Runtime nodes store source spans, not parser objects.
- `ParseMessage` owns CST conversion. CST adapters remain internal and CST values must not be retained by public runtime nodes.
- Constructors must not retain caller-owned mutable slices or maps.
- Accessors for slices and maps must return detached snapshots.
- `NewExpression` returns an error unless an argument or function reference is
  present; typed-nil arguments count as absent.
- `NewInputDeclaration` accepts one expression, requires a non-nil variable
  argument, and derives its name from that variable rather than storing a
  second caller-supplied name.
- Composite model constructors return errors for nil or typed-nil members in
  declaration, pattern-element, and variant-key unions. `ErrNilMember` is the
  stable error identity; nil and empty sequences remain valid empty values.
- `NewFunctionRef`, `NewExpression`, and `NewMarkup` reject nil or typed-nil
  option and attribute union values before snapshotting their maps. Nil and
  empty maps remain valid and mean no options or attributes.
- `ParseMessage` preserves exact byte spans on parsed message roots and select
  variants so model-validation errors can identify their owning source range.
  Programmatically constructed roots and variants use the unknown `(-1, -1)`
  span.
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

- Missing values produce fallback values and contribute a diagnostic to the returned error.
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

Function catalog and configuration rules:

- `DefaultFunctionMap()` and `DraftFunctionMap()` return detached snapshots.
- `DefaultFunctionMap()` contains stable defaults: `:currency`, `:integer`, `:number`, `:offset`, `:percent`, and `:string`.
- `DraftFunctionMap()` contains draft functions: `:date`, `:datetime`, `:time`, and `:unit`; callers opt in with `WithFunctions`.
- `WithFunction` and `WithFunctions` are the only custom-function configuration handoff. Constructors snapshot their input maps.
- `:math` is an extension function, not an MF2 spec function; callers opt in with `WithFunction`.

## Message Values and Parts

`MessageValue` is the common formatting contract. Optional behavior belongs in smaller interfaces:

- `PartsFormatter` emits structured parts.
- `Valuer` exposes an underlying Go value.
- `Selector` participates in pattern selection.
- `OptionedValue` carries formatting options.

Concrete part types must expose typed accessors where the value type is known, while keeping `Value() any` for compatibility.
`Parts()` accessors return detached shallow copies.

`Options()` on package-owned values and parts returns a detached shallow copy.
Resolver wrappers must preserve that ownership boundary even when a custom
`OptionedValue` exposes its own mutable map. Opaque option values and functions
retain identity; the package copies containers, not arbitrary object graphs.

Number-value construction is fallible and validates one immutable Intl plan.
The value and its parts expose the dependency-resolved locale; string, parts,
and plural selection apply the same resolved digit semantics. Invalid required
number options return a typed diagnostic and fallback rather than retrying
after deleting caller semantics.

Date/time-value construction is fallible and validates one immutable Intl
plan. The value and its top-level part expose the dependency-resolved locale,
calendar, and time zone. When no time-zone option is present, UTC, named-zone,
fixed-offset, and local `time.Time` inputs preserve their wall clock. Invalid
required options return a typed diagnostic and fallback; style/field conflicts
remain intact for dependency validation rather than being silently rewritten.

> **Why**: Go should not force every value to implement no-op selector or options methods just to satisfy one broad interface.

## MF1 Module

The independent module `github.com/kaptinlin/messageformat-go/mf1` exposes one
typed entry point per caller job:

```go
type Formatter func(value any, locale, style string) (string, error)

type PluralProfile struct {
    Locale    string
    Select    PluralFunction
    Cardinals []PluralCategory
    Ordinals  []PluralCategory
}

func New(locale string, options *MessageFormatOptions) (*MessageFormat, error)
func NewWithPlural(profile PluralProfile, options *MessageFormatOptions) (*MessageFormat, error)
func (*MessageFormat) Compile(source string) (*CompiledMessage, error)
func (*CompiledMessage) Format(values map[string]any) (string, error)
func (*CompiledMessage) FormatValues(values map[string]any) ([]any, error)
func SupportedLocalesOf(locales []string) ([]string, error)
func GetPlural(locale string) (PluralObject, error)
```

Rules:

- Empty, wildcard, and tags rejected by strict locale parsing return `ErrInvalidLocale`.
- A syntactically valid locale without plural data uses the package-private stable fallback locale.
- `SupportedLocalesOf` filters the caller-owned candidate list in input order; the module does not publish a partial global locale catalog.
- `NewWithPlural` validates the profile locale, selector, and complete cardinal
  and ordinal category sets. Invalid category sets return
  `ErrInvalidPluralCategories`; a nil selector returns
  `ErrInvalidPluralFunction`.
- Constructor option maps and profile category slices are snapshots.
  `ResolvedOptions` returns detached maps, slices, and plural-category slices.
- `MessageFormat` may compile and execute concurrently after construction.
- `CompiledMessage` is immutable and may execute both projections concurrently.
- Nil values maps are empty input. Missing plain arguments become empty strings
  unless `RequireAllArguments` is enabled, in which case both projections
  return `ErrMissingArgument`.
- Built-in `number`, `date`, and `time` arguments format through `go-intl` on
  the compiled-message path. Number formatting uses no style or `integer`,
  `percent`, and `currency[:CODE]`; date/time styles are empty/default,
  `short`, `long`, and `full`.
- Unknown built-in styles return `ErrInvalidFormatterStyle`. Invalid operands
  retain the formatter-specific error identity. Constructor `Currency` and
  `TimeZone` options reach the dependency unchanged.
- `MessageFormatOptions.CustomFormatters` is `map[string]Formatter`. Names must
  be lexer-reachable, non-reserved identifiers and handlers must be non-nil;
  invalid registrations return `ErrInvalidFormatter` at construction.
- Custom formatters receive the original value, effective locale, and trimmed
  style. Handler errors preserve identity through both projections. The
  formatter map is snapshotted; handlers must themselves support concurrent calls
  when the compiled message is used concurrently.

> **Rejected**: A JavaScript-shaped `New(any, ...)`, wildcard locale catalog,
> mutable exported default locale, or compatibility overloads. Those surfaces
> combine unrelated jobs and admit states that the runtime does not use.

## Error Handling

Construction errors return `error`. Format-time recoverable failures preserve
fallback output and are returned as `errors.Join(...)` diagnostics in encounter
order. Successful format calls return a nil error.

Production code must not panic for user-controlled message syntax, data model shape, or runtime values.

Rules:

- `MessageError.Type` remains the stable string category for callers that consume serialized or fixture-like errors.
- `MessageError.Kind()` returns the typed Go identity for `errors.Is` and `errors.As`.
- Error tests should assert kind, source/span, cause, and behavior before asserting full prose.

## Forbidden

- Do not expose mutable default function maps.
- Do not add string compatibility helpers for typed constructor options.
- Do not add package-wide mutable logger state.
- Do not add format-time logger or error-handler options; callers own error presentation.
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
- Mutating data-model constructor inputs or collection accessor results after
  `Compile` does not change the compiled formatter.
- Mutating a map returned by `DefaultFunctionMap` or `DraftFunctionMap` does not affect new formatters.
- `options_test.go` proves every constructor option vocabulary and `ErrInvalidOption` path through both `Parse` and `Compile`.
- `pkg/messagevalue/value_test.go` and `internal/resolve/function_ref_test.go` prove option and part accessor snapshot ownership.
- `pkg/datamodel/fromcst_test.go` and `pkg/datamodel/validate_test.go` prove syntax/data-model error ownership and closed construction paths.
- `mf1/constructor_external_test.go` and `mf1/messageformat_test.go` prove typed constructors, malformed-versus-unsupported locale behavior, detached resolved options, and concurrent use.
- Tests cover default bidi isolation, resolved options, custom function registration, `Format`, and `FormatToParts`.
- Tests cover selector no-probe behavior and deterministic candidate order.
- Tests cover missing, nil, typed nil, and unknown variable states.
- Tests prove data-model constructor and accessor ownership through compiled formatting.
- Tests cover `errors.Is`, `errors.As`, and `Kind()`.
- `task verify` passes after API changes.
