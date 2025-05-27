# MessageFormat 2.0 Go Implementation

A complete implementation of [MessageFormat 2.0](https://github.com/unicode-org/message-format-wg) in Go, providing internationalization (i18n) support with advanced features like pluralization, gender selection, and custom formatting functions.

## ✅ Official Test Suite Compliance

**This implementation passes the complete official MessageFormat 2.0 test suite**, ensuring full compatibility with the specification and interoperability with other MessageFormat 2.0 implementations.

## 🚀 Quick Start

### Installation

```bash
go get github.com/kaptinlin/messageformat-go
```

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/kaptinlin/messageformat-go"
)

func main() {
    mf, err := messageformat.New("en", "Hello, {$name}!")
    if err != nil {
        panic(err)
    }

    result, err := mf.Format(map[string]interface{}{
        "name": "World",
    })
    if err != nil {
        panic(err)
    }

    fmt.Println(result) // Output: Hello, ⁨World⁩!
}
```

## ✨ Features

- **Complete MF2 Support**: Full MessageFormat 2.0 specification implementation
- **Official Test Suite Compliant**: Passes all official tests
- **TypeScript Compatible**: API designed to match the TypeScript implementation
- **Built-in Functions**: Number, integer, string, and datetime formatting with locale support
- **Custom Functions**: Extensible function system for custom formatters
- **Select Messages**: Pattern matching with pluralization and gender selection
- **Unicode Compliant**: Full Unicode normalization and bidirectional text support
- **Thread Safe**: Concurrent-safe after construction

## 📖 Usage Examples

### Number Formatting
```go
mf, err := messageformat.New("en", "You have {$count :number} messages")
result, err := mf.Format(map[string]interface{}{
    "count": 42,
})
// Output: "You have 42 messages"
```

### Select Messages (Pluralization)
```go
mf, err := messageformat.New("en", `
.input {$count :number}
.match $count
0   {{No items}}
one {{One item}}
*   {{{$count} items}}
`)

result, err := mf.Format(map[string]interface{}{
    "count": 5,
})
// Output: "5 items"
```

### Custom Functions
```go
import (
    "strings"
    "github.com/kaptinlin/messageformat-go"
    "github.com/kaptinlin/messageformat-go/pkg/functions"
    "github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

func customUppercase(ctx functions.MessageFunctionContext, options map[string]interface{}, input interface{}) messagevalue.MessageValue {
    locales := ctx.Locales()
    locale := "en"
    if len(locales) > 0 {
        locale = locales[0]
    }
    if str, ok := input.(string); ok {
        return messagevalue.NewStringValue(strings.ToUpper(str), locale, ctx.Source())
    }
    return messagevalue.NewStringValue(fmt.Sprintf("%v", input), locale, ctx.Source())
}

mf, err := messageformat.New("en", "Hello, {$name :custom}!",
    messageformat.WithFunction("custom", customUppercase),
)
```

### Structured Output
```go
parts, err := mf.FormatToParts(map[string]interface{}{
    "name": "World",
})

for _, part := range parts {
    switch p := part.(type) {
    case *messageformat.MessageTextPart:
        fmt.Printf("Text: %s\n", p.Value())
    case *messageformat.MessageStringPart:
        fmt.Printf("Variable: %s\n", p.Value())
    }
}
```

## 🎯 API Options

### Functional Options (Recommended)
```go
mf, err := messageformat.New("en", "Hello, {$name}!",
    messageformat.WithBidiIsolation("none"),
    messageformat.WithDir("ltr"),
    messageformat.WithFunction("custom", myCustomFunction),
)
```

### Traditional Options
```go
mf, err := messageformat.New("en", "Hello, {$name}!", &messageformat.MessageFormatOptions{
    BidiIsolation: stringPtr("none"),
    Dir:          stringPtr("ltr"),
    Functions:    map[string]messageformat.MessageFunction{
        "custom": myCustomFunction,
    },
})
```

### TypeScript Migration
```typescript
// TypeScript
const mf = new MessageFormat('en', 'Hello, {name}!', {
  bidiIsolation: 'none'
});
const result = mf.format({ name: 'World' });
```

```go
// Go equivalent
mf, err := messageformat.New("en", "Hello, {$name}!", 
    messageformat.WithBidiIsolation("none"),
)
result, err := mf.Format(map[string]interface{}{
    "name": "World",
})
```

## 🧪 Testing

### Prerequisites
Initialize git submodules to fetch the official test suite:

```bash
# Clone with submodules
git clone --recurse-submodules https://github.com/kaptinlin/messageformat-go.git

# Or initialize submodules after cloning
git submodule update --init --recursive
```

### Running Tests
```bash
# Run all tests including official test suite
make test

# Run unit tests only (excluding official test suite)
make test-unit

# Run official test suite only
make test-official

# Run tests with coverage report
make test-coverage

# Run tests with verbose output
make test-verbose

# Or use go commands directly
go test ./...                    # All tests
go test ./pkg/... ./internal/... # Unit tests only
go test ./tests/                 # Official test suite only
```

### Development Commands
```bash
# Show all available commands
make help

# Initialize git submodules (required for official tests)
make submodules

# Format code and run all checks
make verify



# Run examples
make examples

# Run benchmarks
make bench
```

📋 **For detailed testing instructions, see [TESTING.md](TESTING.md)**

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines on:
- Development setup and workflow
- Code standards and testing requirements
- Commit message conventions
- Pull request process

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Related Projects

- [MessageFormat 2.0 Specification](https://github.com/unicode-org/message-format-wg)
- [JavaScript/TypeScript Implementation](https://github.com/messageformat/messageformat)
- [MessageFormat Working Group](https://github.com/unicode-org/message-format-wg)

## 🙏 Credits

This Go implementation is based on the [MessageFormat JavaScript/TypeScript library](https://github.com/messageformat/messageformat). Special thanks to the [MessageFormat Working Group](https://github.com/unicode-org/message-format-wg) and the Unicode Consortium for their work on internationalization standards.
