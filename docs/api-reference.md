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
	"Price: {$amount :number style=currency currency=USD}",
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
	options ...FormatOption,
) (string, error)
```

Accepted format options:

- `messageformat.WithErrorHandler(...)`

Example:

```go
out, err := mf.Format(map[string]any{
	"name": "Alice",
	"count": 3,
})
```

With a format-time error handler:

```go
out, err := mf.Format(
	map[string]any{},
	messageformat.WithErrorHandler(func(err error) {
		log.Printf("format warning: %v", err)
	}),
)
```

### `(*MessageFormat).FormatToParts`

```go
func (mf *MessageFormat) FormatToParts(
	values map[string]any,
	options ...FormatOption,
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
	Logger        *slog.Logger
}
```

### Functional options

Available constructor options include:

- `WithBidiIsolation(...)`
- `WithDir(...)`
- `WithLocaleMatcher(...)`
- `WithFunction(...)`
- `WithFunctions(...)`
- `WithLogger(...)`

Format-time option:

- `WithErrorHandler(...)`

## Exported Helpers

The root package also re-exports several helpers:

- `ParseMessage`
- `StringifyMessage`
- `Validate`
- `Visit`
- data-model type guards such as `IsMessage`, `IsVariableRef`, and `IsFunctionRef`

These are useful when you want to work with the parsed message model directly rather than only formatting strings.

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
- `messagevalue.NewNumberValue(...)`
- `messagevalue.NewFallbackValue(...)`

## Defaults

Important runtime defaults:

- `BidiIsolation` defaults to `messageformat.BidiNone`
- `LocaleMatcher` defaults to `messageformat.LocaleBestFit`
- locale input is defensively copied during construction
- `MessageFormat` instances are safe for concurrent use after construction

## Errors

Construction returns syntax and validation errors immediately.

Formatting uses graceful degradation for recoverable runtime issues and can report them through `WithErrorHandler(...)`.

Syntax errors preserve specific categories such as:

- `missing-syntax`
- `bad-selector`
- `extra-content`
- `bad-input-expression`

See [Error Handling](error-handling.md) for the error model in more detail.
