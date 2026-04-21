# Documentation

Documentation index for the MessageFormat Go v2 package.

This directory is organized for two use cases:

- learn the Unicode MessageFormat 2.0 model and this package's defaults
- look up API, syntax, formatting, custom functions, and error handling

## Reading Order

If you are new to the package, read the guides in this order:

1. [Getting Started](getting-started.md)
2. [Message Syntax](message-syntax.md)
3. [Formatting Functions](formatting-functions.md)
4. [Custom Functions](custom-functions.md)
5. [Error Handling](error-handling.md)
6. [API Reference](api-reference.md)

## Guide Map

| Guide | Use it for |
|-------|------------|
| [Getting Started](getting-started.md) | Installation, first message, and core concepts |
| [Message Syntax](message-syntax.md) | `.input`, `.local`, `.match`, variables, markup, and syntax rules |
| [Formatting Functions](formatting-functions.md) | Built-in formatter behavior and option shapes |
| [Custom Functions](custom-functions.md) | Registering custom functions and returning message values |
| [Error Handling](error-handling.md) | Syntax, resolution, and selection error behavior |
| [API Reference](api-reference.md) | Constructor, formatting methods, options, and exported helpers |

## Defaults That Matter

These docs assume the current package defaults:

- `BidiIsolation` defaults to `BidiNone`
- `LocaleMatcher` defaults to `LocaleBestFit`
- `MessageFormat` instances are safe for concurrent use after construction
- syntax errors preserve their specific error type instead of being flattened into a generic parse error

If you need bidi isolation markers for RTL-safe embedding, opt in explicitly:

```go
mf, err := messageformat.Parse(
	[]string{"ar"},
	"مرحبا {$name}!",
	messageformat.WithBidiIsolation(messageformat.BidiDefault),
)
```

## Common Entry Points

For quick lookup, these are the APIs most users start with:

| API | Purpose |
|-----|---------|
| `messageformat.Parse(...)` | Parse and validate a message from source text |
| `messageformat.Compile(...)` | Create an instance from a parsed data model |
| `mf.Format(...)` | Format to a string |
| `mf.FormatToParts(...)` | Format to structured parts |
| `messageformat.ParseMessage(...)` | Parse to the public data model |
| `messageformat.Validate(...)` | Validate a parsed message |

## Examples

Runnable examples live outside this directory:

- [`examples/basic`](../examples/basic)
- [`examples/advanced`](../examples/advanced)
- [`examples/pluralization`](../examples/pluralization)
- [`examples/custom-functions`](../examples/custom-functions)

Run them with:

```bash
task examples
```

## Development and Verification

For repository-level commands, use the root documentation:

- [`README.md`](../README.md) for package overview and main workflow
- [`TESTING.md`](../TESTING.md) for test layout and verification commands
- [`CONTRIBUTING.md`](../CONTRIBUTING.md) for contribution workflow
