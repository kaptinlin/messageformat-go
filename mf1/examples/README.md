# MF1 Examples

Examples for the supported `github.com/kaptinlin/messageformat-go/mf1` module.

## Run

From the repository root:

```bash
go -C mf1 run ./examples/basic
go -C mf1 run ./examples/ecommerce
go -C mf1 run ./examples/multilingual
go -C mf1 run ./examples/performance
```

## Example Set

- `basic`: minimal interpolation, select, and plural usage
- `ecommerce`: message patterns for transactional flows
- `multilingual`: locale-specific behavior and plural categories
- `performance`: caching and repeated formatting patterns

## Import Path

All examples import the package path below:

```go
import mf "github.com/kaptinlin/messageformat-go/mf1"
```

Install and version the independent MF1 module:

```bash
go get github.com/kaptinlin/messageformat-go/mf1@latest
```
