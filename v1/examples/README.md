# V1 Examples

Examples for the supported `github.com/kaptinlin/messageformat-go/v1` package.

## Run

From the repository root:

```bash
go run ./v1/examples/basic
go run ./v1/examples/ecommerce
go run ./v1/examples/multilingual
go run ./v1/examples/performance
```

## Example Set

- `basic`: minimal interpolation, select, and plural usage
- `ecommerce`: message patterns for transactional flows
- `multilingual`: locale-specific behavior and plural categories
- `performance`: caching and repeated formatting patterns

## Import Path

All examples import the package path below:

```go
import mf "github.com/kaptinlin/messageformat-go/v1"
```

That path is a package inside the root module. Consumers should version the root module:

```bash
go get github.com/kaptinlin/messageformat-go@latest
```
