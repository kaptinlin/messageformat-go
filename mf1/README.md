# MessageFormat Go `mf1`

Supported ICU MessageFormat v1 compatibility package for this repository.

`mf1` is the package for ICU MessageFormat v1 compatibility. Import it as:

```go
import mf "github.com/kaptinlin/messageformat-go/mf1"
```

Do not run `go get github.com/kaptinlin/messageformat-go/mf1` as if it were a separate module. Use the root module version instead:

```bash
go get github.com/kaptinlin/messageformat-go@latest
```

## Status

- Supported compatibility surface for ICU MessageFormat v1
- Kept as product code and covered by repository lint and test workflows
- Intended for consumers that still need the legacy MessageFormat v1 API shape

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

	result, err := msg(map[string]any{"name": "World"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)
}
```

## Examples

- [Basic](./examples/basic/main.go)
- [E-commerce](./examples/ecommerce/main.go)
- [Multilingual](./examples/multilingual/main.go)
- [Performance](./examples/performance/main.go)

Run examples from the repository root:

```bash
go run ./mf1/examples/basic
go run ./mf1/examples/basic
go run ./mf1/examples/ecommerce
go run ./mf1/examples/multilingual
go run ./mf1/examples/performance
```

## Documentation

- [API reference](./docs/api-reference.md)
- [Performance guide](./docs/performance.md)
- [Examples guide](./examples/README.md)

## Notes

- `mf1` is not deprecated inside this repository.
- `mf1` must not be pruned during cleanup or refactoring.
- Release tags apply to the root module; `mf1` ships as part of that module.
