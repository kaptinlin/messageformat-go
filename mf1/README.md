# MessageFormat Go MF1

Supported ICU MessageFormat v1 compatibility module for Go.

## Installation

```bash
go get github.com/kaptinlin/messageformat-go/mf1@latest
```

Import the package as:

```go
import mf "github.com/kaptinlin/messageformat-go/mf1"
```

## Status

- Supported compatibility surface for ICU MessageFormat v1
- Kept as product code and covered by repository lint and test workflows
- Intended for consumers that need the ICU MessageFormat v1 API shape

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	mf "github.com/kaptinlin/messageformat-go/mf1"
)

func main() {
	messageFormat, err := mf.New("en", nil)
	if err != nil {
		log.Fatal(err)
	}

	msg, err := messageFormat.Compile("Hello, {name}!")
	if err != nil {
		log.Fatal(err)
	}

	result, err := msg.Format(map[string]any{"name": "World"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)
}
```

`Compile` returns an immutable `*CompiledMessage`. Use `Format` for text and
`FormatValues` for the ordered value projection. Both methods accept only
`map[string]any`; nil means an empty map. Missing arguments render as empty
values unless `RequireAllArguments` is enabled, in which case formatting
returns `ErrMissingArgument`.

## Examples

- [Basic](./examples/basic/main.go)
- [E-commerce](./examples/ecommerce/main.go)
- [Multilingual](./examples/multilingual/main.go)
- [Performance](./examples/performance/main.go)

Run examples from the repository root:

```bash
go -C mf1 run ./examples/basic
go -C mf1 run ./examples/ecommerce
go -C mf1 run ./examples/multilingual
go -C mf1 run ./examples/performance
```

## Documentation

- [API reference](./docs/api-reference.md)
- [Performance guide](./docs/performance.md)
- [Examples guide](./examples/README.md)

## Notes

- `mf1` is not deprecated inside this repository.
- `mf1` must not be pruned during cleanup or refactoring.
- `mf1/go.mod` owns the module path `github.com/kaptinlin/messageformat-go/mf1`.
