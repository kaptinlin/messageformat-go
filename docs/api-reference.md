# API Reference

Reference for the main public API in `github.com/kaptinlin/messageformat-go`.

For runnable examples, see the root [README](../README.md) and the [`examples`](../examples) directory.

## Constructors

### `messageformat.Parse`

```go
func Parse(
	locales []string,
	source string,
	options ...Option,
) (*MessageFormat, error)
```

Typical usage:

```go
mf, err := messageformat.Parse([]string{"en"}, "Hello, {$name}!")
```

With multiple locales:

```go
mf, err := messageformat.Parse(
	[]string{"zh-CN", "en"},
	"Price: {$amount :currency currency=USD}",
)
```

With functional options:

```go
mf, err := messageformat.Parse(
	[]string{"ar"},
	"مرحبا {$name}!",
	messageformat.WithBidiIsolation(messageformat.BidiDefault),
	messageformat.WithDir(messageformat.DirRTL),
)
```

### `messageformat.Compile`

```go
func Compile(
	locales []string,
	message datamodel.Message,
	options ...Option,
) (*MessageFormat, error)
```

Use `Compile(...)` when you already have a parsed public data model and want to skip reparsing source text.

## Formatting Methods

### `(*MessageFormat).Format`

```go
func (mf *MessageFormat) Format(
	values map[string]any,
) (string, error)
```

Example:

```go
out, err := mf.Format(map[string]any{
	"name": "Alice",
	"count": 3,
})
```

Recoverable diagnostics are returned with usable fallback output:

```go
out, err := mf.Format(map[string]any{})
if err != nil {
	log.Printf("format diagnostics: %v", err)
}
fmt.Println(out)
```

### `(*MessageFormat).FormatToParts`

```go
func (mf *MessageFormat) FormatToParts(
	values map[string]any,
) ([]messagevalue.MessagePart, error)
```

Use `FormatToParts` when you need structured output for rich text rendering or post-processing.

Example:

```go
parts, err := mf.FormatToParts(map[string]any{"amount": 29.99})
if err != nil {
	log.Fatal(err)
}

for _, part := range parts {
	fmt.Printf("%s: %v\n", part.Type(), part.Value())
}
```

## Configuration

### `MessageFormatOptions`

Use `messageformat.Options(messageformat.MessageFormatOptions{...})` to convert a config struct into an `Option`.


```go
type MessageFormatOptions struct {
	BidiIsolation BidiIsolation
	Dir           Direction
	LocaleMatcher LocaleMatcher
	Functions     map[string]functions.MessageFunction
}
```

### Functional options

Available constructor options include:

- `WithBidiIsolation(...)`
- `WithDir(...)`
- `WithLocaleMatcher(...)`
- `WithFunction(...)`
- `WithFunctions(...)`

## Data Model Package

Import `github.com/kaptinlin/messageformat-go/pkg/datamodel` when working with
the parsed model directly:

- `datamodel.ParseMessage`
- `datamodel.StringifyMessage`
- `datamodel.ValidateMessage`
- `datamodel.Visit`
- type guards such as `datamodel.IsMessage`, `datamodel.IsVariableRef`, and
  `datamodel.IsFunctionRef`

The root package owns formatter construction and rendering; `pkg/datamodel`
owns model construction and inspection.

## Parts and Values

The root package re-exports several part aliases:

- `messageformat.Part`
- `messageformat.StringPart`
- `messageformat.NumberPart`
- `messageformat.DateTimePart`
- `messageformat.FallbackPart`
- `messageformat.MarkupPart`

Custom functions return concrete values from `pkg/messagevalue`, such as:

- `messagevalue.NewStringValue(...)`
- `messagevalue.NewNumberValue(...) (*NumberValue, error)`
- `messagevalue.NewDateTimeValue(...) (*DateTimeValue, error)`
- `messagevalue.NewFallbackValue(...)`

## Defaults

Important runtime defaults:

- `BidiIsolation` defaults to `messageformat.BidiDefault`
- `LocaleMatcher` defaults to `messageformat.LocaleBestFit`
- locale input is defensively copied during construction
- `MessageFormat` instances are safe for concurrent use after construction

Number values validate their Intl plan during construction. Their locale and
number parts report the dependency-resolved locale, and rendering and plural
selection apply the same resolved digit options.

Date/time values also validate one Intl plan during construction. Their value
and `DateTimePart` expose dependency-resolved locale, calendar, and time zone;
implicit zones preserve the input `time.Time` wall clock.

## Errors

Construction returns syntax and validation errors immediately.

Formatting uses graceful degradation for recoverable runtime issues: fallback
string or parts remain usable, and the returned error joins diagnostics in
encounter order.

Syntax errors preserve specific categories such as:

- `missing-syntax`
- `bad-selector`
- `extra-content`
- `bad-input-expression`

See [Error Handling](error-handling.md) for the error model in more detail.
