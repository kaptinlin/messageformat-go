# MessageFormat Go

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.26-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/kaptinlin/messageformat-go.svg)](https://pkg.go.dev/github.com/kaptinlin/messageformat-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/kaptinlin/messageformat-go)](https://goreportcard.com/report/github.com/kaptinlin/messageformat-go)

A Go implementation of Unicode MessageFormat 2.0 for parsing, validating, and formatting localized messages

## Features

- **MessageFormat 2.0**: Parse, validate, select, and format messages with the Unicode MessageFormat 2.0 model.
- **Established API surface**: Use stable constructor concepts, option names, formatting methods, and runtime defaults across MessageFormat 2 implementations.
- **Locale-aware functions**: Format numbers, dates, currencies, percentages, offsets, strings, and units through [`github.com/agentable/go-intl`](https://github.com/agentable/go-intl).
- **Custom formatters**: Register application functions with `WithFunction` or `WithFunctions`.
- **Structured rendering**: Use `FormatToParts` for rich text, markup-aware rendering, and post-processing.
- **Migration path**: Keep existing ICU MessageFormat v1 code on the supported `github.com/kaptinlin/messageformat-go/mf1` package.
- **Official conformance**: Run the Unicode MessageFormat Working Group test suite with `task test-official` or `task test-v2`.

## Installation

```bash
go get github.com/kaptinlin/messageformat-go
```

Requires **Go 1.26+**.

For ICU MessageFormat v1 compatibility, import:

```go
import mf1 "github.com/kaptinlin/messageformat-go/mf1"
```

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
| `Parse(locales, source, options...)` | Parse source text into a formatter |
| `Compile(locales, message, options...)` | Build a formatter from a parsed data model |
| `(*MessageFormat).Format(values, options...)` | Format to a string |
| `(*MessageFormat).FormatToParts(values, options...)` | Format to structured parts |
| `ParseMessage(source)` | Parse source into the public data model |
| `StringifyMessage(message)` | Convert the data model back to source |
| `Validate(message, onError)` | Validate a parsed message |

Full API details are available on [pkg.go.dev](https://pkg.go.dev/github.com/kaptinlin/messageformat-go) and in [`docs/api-reference.md`](docs/api-reference.md).

## Common Usage

### Select Messages

Declare selectors before `.match` messages:

```go
mf, err := messageformat.Parse([]string{"en"}, `
.input {$count :number}
.match $count
0   {{No items}}
one {{One item}}
*   {{{$count} items}}
`)
if err != nil {
	log.Fatal(err)
}

out, err := mf.Format(map[string]any{"count": 3})
if err != nil {
	log.Fatal(err)
}

fmt.Println(out)
```

### Built-In Formatting

Use MessageFormat functions directly in source text:

```go
mf, err := messageformat.Parse(
	[]string{"en"},
	"Total: {$amount :currency currency=USD}",
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

Built-in functions include `:number`, `:integer`, `:string`, `:offset`, `:currency`, `:date`, `:datetime`, `:percent`, `:time`, and `:unit`.

### Structured Parts

Use `FormatToParts` when a UI needs structured output instead of one string. `Format` follows the documented string conversion path; `FormatToParts` keeps the resolved part values for rich rendering.

```go
parts, err := mf.FormatToParts(map[string]any{"amount": 29.99})
if err != nil {
	log.Fatal(err)
}

for _, part := range parts {
	fmt.Printf("%s: %v\n", part.Type(), part.Value())
}
```

### Custom Functions

Custom functions receive locale context, resolved options, and the operand value:

```go
func uppercase(
	ctx messageformat.MessageFunctionContext,
	options functions.Options,
	operand any,
) messagevalue.MessageValue {
	locale := "en"
	if locales := ctx.Locales(); len(locales) > 0 {
		locale = locales[0]
	}

	return messagevalue.NewStringValue(
		strings.ToUpper(fmt.Sprint(operand)),
		locale,
		ctx.Source(),
	)
}

mf, err := messageformat.Parse(
	[]string{"en"},
	"Hello, {$name :uppercase}!",
	messageformat.WithFunction("uppercase", uppercase),
)
```

See [`docs/custom-functions.md`](docs/custom-functions.md) and [`examples/custom-functions`](examples/custom-functions) for complete examples.

## Configuration

Use functional options for focused constructor changes:

| Option | Purpose | Default |
|--------|---------|---------|
| `WithBidiIsolation(strategy)` | Control Unicode bidi isolation markers | `BidiDefault` |
| `WithDir(direction)` | Set message base direction | Locale-derived |
| `WithLocaleMatcher(matcher)` | Select locale matching behavior | `LocaleBestFit` |
| `WithFunction(name, fn)` | Register one custom function | Built-ins only |
| `WithFunctions(funcs)` | Register multiple custom functions | Built-ins only |
| `WithLogger(logger)` | Attach an instance logger | Package logger |

Example:

```go
mf, err := messageformat.Parse(
	[]string{"ar"},
	"مرحبا {$name}!",
	messageformat.WithDir(messageformat.DirRTL),
)
```

Use `messageformat.Options(...)` when a struct is more convenient:

```go
mf, err := messageformat.Parse(
	[]string{"en"},
		"Hello, {$name}!",
		messageformat.Options(messageformat.MessageFormatOptions{
			BidiIsolation: messageformat.BidiDefault,
			LocaleMatcher: messageformat.LocaleBestFit,
		}),
	)
```

## Conformance

This repository verifies behavior against the official Unicode MessageFormat Working Group test suite in `tests/messageformat-wg/`.

```bash
task test-official  # Official MessageFormat 2.0 suite only
task test-v2        # Package tests plus official suite, with race detection
```

Project design contracts live in [`SPECS/`](SPECS/).

## Documentation

| Guide | Use it for |
|-------|------------|
| [`docs/getting-started.md`](docs/getting-started.md) | Installation and first steps |
| [`docs/message-syntax.md`](docs/message-syntax.md) | MessageFormat 2.0 syntax |
| [`docs/formatting-functions.md`](docs/formatting-functions.md) | Built-in formatter behavior |
| [`docs/custom-functions.md`](docs/custom-functions.md) | Custom formatter authoring |
| [`docs/error-handling.md`](docs/error-handling.md) | Syntax, resolution, and selection errors |
| [`docs/api-reference.md`](docs/api-reference.md) | Public API reference |
| [`SPECS/`](SPECS/) | Design contracts and architecture boundaries |

Runnable examples live in [`examples/`](examples/).

## Development

```bash
task submodules      # Initialize official test suite submodule
task test            # Run all tests with race detection
task test-v2         # Run package tests and the official suite
task lint            # Run golangci-lint and go mod tidy checks
task verify          # Run deps, fmt, vet, lint, test, and vuln
task examples        # Run example programs
```

## Contributing

Contributions are welcome. Start with [`CONTRIBUTING.md`](CONTRIBUTING.md), then run `task verify` before opening a pull request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
