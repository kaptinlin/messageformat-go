# MessageFormat Go

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.26.2-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/kaptinlin/messageformat-go.svg)](https://pkg.go.dev/github.com/kaptinlin/messageformat-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/kaptinlin/messageformat-go)](https://goreportcard.com/report/github.com/kaptinlin/messageformat-go)

A Go implementation of Unicode MessageFormat 2.0.

This repository tracks the Unicode MessageFormat 2.0 specification and verifies behavior against the official MessageFormat Working Group test suite bundled as a git submodule in `tests/messageformat-wg/`.

## Features

- **Unicode MessageFormat 2.0**: Parse, validate, and format messages using the Unicode MessageFormat 2.0 model.
- **Go-first public API**: Strongly typed constructors and format options.
- **Rich formatting**: Built-in support for numbers, integers, strings, dates, currencies, percentages, offsets, and units.
- **Custom functions**: Register locale-aware formatters with `WithFunction` or `WithFunctions`.
- **Structured output**: Render to strings with `Format` or rich parts with `FormatToParts`.
- **Predictable defaults**: Instances are safe for concurrent use after construction and default to clean output without bidi isolation markers.
- **Spec verification**: The repository includes the official MessageFormat Working Group test suite as a git submodule.

## Installation

```bash
go get github.com/kaptinlin/messageformat-go
```

Requires **Go 1.26.2+**.

This repository is a multi-module workspace:

- `github.com/kaptinlin/messageformat-go` for the MessageFormat 2.0 implementation
- `github.com/kaptinlin/messageformat-go/v1` for the supported ICU MessageFormat v1 compatibility module

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/messageformat-go"
)

func main() {
	mf, err := messageformat.Parse([]string{"en"}, "Hello, {$name}!")
	if err != nil {
		log.Fatal(err)
	}

	out, err := mf.Format(map[string]any{"name": "World"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(out)
}
```

## Core API

| API | Purpose |
|-----|---------|
| `Parse(locales, source, options...)` | Parse and validate a message from source text |
| `Compile(locales, message, options...)` | Reuse a prebuilt data model |
| `(*MessageFormat).Format(values, options...)` | Format to a string |
| `(*MessageFormat).FormatToParts(values, options...)` | Format to structured parts |
| `ParseMessage(source)` | Parse source into the public data model |
| `StringifyMessage(message)` | Convert the data model back to source |
| `Validate(message, scope)` | Validate a parsed message |

Full API details live on [pkg.go.dev](https://pkg.go.dev/github.com/kaptinlin/messageformat-go) and in [`docs/api-reference.md`](docs/api-reference.md).

## Spec And Reference Tracking

- Specification: [Unicode MessageFormat 2.0](https://unicode.org/reports/tr35/tr35-messageFormat.html)
- Official test suite submodule: `tests/messageformat-wg`
- Current pinned test-suite commit: `dd86e42e10d1d0c9c4401d0781cdd87ee7166366`
- TypeScript reference implementation submodule: `.reference/messageformat`
- Current pinned reference commit: `ea2b8b9c49a5caa5d6720948a9f4b5fc4a908b72`

The implementation aims to stay aligned with the spec and the official test suite first, while using the TypeScript `messageformat` project as the API-compatibility reference.

## MessageFormat 2.0 Basics

Variables use the MessageFormat 2.0 form with a `$` prefix:

```go
mf, err := messageformat.Parse([]string{"en"}, "Hello, {$name}!")
```

Select messages require declared selectors:

```go
mf, err := messageformat.Parse([]string{"en"}, `
.input {$count :number}
.match $count
0   {{No items}}
one {{One item}}
*   {{{$count} items}}
`)
```

The parser preserves syntax error types, so malformed selectors and missing syntax are reported with specific error categories instead of a generic parse error.

## Formatting Examples

### Numbers and currencies

```go
mf, err := messageformat.Parse(
	[]string{"en"},
	"Total: {$amount :number style=currency currency=USD}",
)
if err != nil {
	log.Fatal(err)
}

out, err := mf.Format(map[string]any{"amount": 29.99})
if err != nil {
	log.Fatal(err)
}

fmt.Println(out)
```

### Structured parts

```go
mf, err := messageformat.Parse([]string{"en"}, "Hello, {$name}!")
if err != nil {
	log.Fatal(err)
}

parts, err := mf.FormatToParts(map[string]any{"name": "World"})
if err != nil {
	log.Fatal(err)
}

for _, part := range parts {
	fmt.Printf("%s: %v\n", part.Type(), part.Value())
}
```

### Custom functions

```go
func uppercase(
	ctx messageformat.MessageFunctionContext,
	options map[string]any,
	operand any,
) messagevalue.MessageValue {
	return messagevalue.NewStringValue(
		strings.ToUpper(fmt.Sprint(operand)),
		ctx.Locales()[0],
		ctx.Source(),
	)
}

mf, err := messageformat.Parse(
	[]string{"en"},
	"Hello, {$name :uppercase}!",
	messageformat.WithFunction("uppercase", uppercase),
)
```

See [`examples/custom-functions/main.go`](examples/custom-functions/main.go) for a fuller example.

## Defaults and Configuration

This package deliberately chooses simple defaults:

- `BidiIsolation` defaults to `BidiNone`, so formatted output does not include Unicode isolation markers unless you opt in.
- `LocaleMatcher` defaults to `LocaleBestFit`.
- `MessageFormat` instances defensively copy locale input and are safe for concurrent use after construction.

Use functional options when you want focused overrides:

```go
mf, err := messageformat.Parse(
	[]string{"ar"},
	"مرحبا {$name}!",
	messageformat.WithBidiIsolation(messageformat.BidiDefault),
	messageformat.WithDir(messageformat.DirRTL),
)
```

Use `messageformat.Options(...)` when a struct is more convenient:

```go
mf, err := messageformat.Parse([]string{"en"}, "Hello, {$name}!", messageformat.Options(messageformat.MessageFormatOptions{
	BidiIsolation: messageformat.BidiNone,
	LocaleMatcher: messageformat.LocaleBestFit,
}))
```

## Documentation

| Guide | Description |
|-------|-------------|
| [`docs/getting-started.md`](docs/getting-started.md) | Installation, first steps, and basic concepts |
| [`docs/message-syntax.md`](docs/message-syntax.md) | MessageFormat 2.0 syntax reference |
| [`docs/api-reference.md`](docs/api-reference.md) | Public API overview |
| [`docs/formatting-functions.md`](docs/formatting-functions.md) | Built-in formatter behavior |
| [`docs/custom-functions.md`](docs/custom-functions.md) | Writing and registering custom functions |
| [`docs/error-handling.md`](docs/error-handling.md) | Error categories and handling patterns |
| [`TESTING.md`](TESTING.md) | Test layout and verification commands |

## Development

```bash
task test           # Run all tests with race detection
task test-v2        # Run the official MessageFormat 2.0 suite and package tests
task test-official  # Run the official MessageFormat 2.0 suite only
task lint           # Run golangci-lint and tidy checks
task verify         # Run deps, fmt, vet, lint, test, and vuln
task examples       # Run all example programs
```

The repository uses `go.work` to develop the root module and `v1/` together. Validation for `v1` is still executed with `GOWORK=off` so the sub-module stays independently consumable.

If this is a fresh clone, initialize the test suite submodule first:

```bash
task submodules
```

## Contributing

Contributions are welcome. Start with [`CONTRIBUTING.md`](CONTRIBUTING.md), then run `task verify` before opening a pull request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
