# MessageFormat 2.0 Go Implementation

A complete implementation of [MessageFormat 2.0](https://github.com/unicode-org/message-format-wg) in Go, providing internationalization (i18n) support with advanced features like pluralization, gender selection, and custom formatting functions.

This library is a Go port of the [Unicode MessageFormat 2 packages](https://github.com/messageformat/messageformat) from the JavaScript/TypeScript implementation, providing a formatter and other tools for [Unicode MessageFormat 2.0](https://unicode.org/reports/tr35/tr35-messageFormat.html) (MF2), the new standard for localization developed by the [MessageFormat Working Group](https://github.com/unicode-org/message-format-wg).

The API provided by this library is current as of the [LDML 47](https://www.unicode.org/reports/tr35/tr35-75/tr35-messageFormat.html) (March 2025) Final version of the MF2 specification.

## 🚀 Quick Start

```bash
go get github.com/kaptinlin/messageformat-go
```

```go
package main

import (
    "fmt"
    "github.com/kaptinlin/messageformat-go"
)

func main() {
    // Create a MessageFormat instance
    mf, err := messageformat.New("en", "Hello, {$name}!")
    if err != nil {
        panic(err)
    }

    // Format the message
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

- **Complete MF2 Support**: Full implementation of MessageFormat 2.0 specification
- **TypeScript Compatible**: API designed to match the TypeScript implementation
- **Modern Go API**: Functional options pattern with backward compatibility
- **Built-in Functions**: Number, integer, string formatting with locale support
- **Custom Functions**: Extensible function system for custom formatters
- **Select Messages**: Pattern matching with pluralization and gender selection
- **Bidirectional Text**: Full Unicode bidi algorithm support
- **Error Handling**: Comprehensive error reporting with position tracking
- **Thread Safe**: Concurrent-safe after construction

## 📖 Usage Examples

### Basic Formatting
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
func customUppercase(ctx messageformat.MessageFunctionContext, options map[string]interface{}, input interface{}) messageformat.MessageValue {
    if str, ok := input.(string); ok {
        return messageformat.NewStringValue(strings.ToUpper(str))
    }
    return messageformat.NewStringValue(fmt.Sprintf("%v", input))
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

## 🎯 API Design

### Modern Functional Options
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

## 🔄 Migration from TypeScript

This implementation maintains API compatibility with the TypeScript version:

```typescript
// TypeScript
const mf = new MessageFormat('en', 'Hello, {name}!', {
  bidiIsolation: 'none'
});
const result = mf.format({ name: 'World' });
```

```go
// Go equivalent
mf, err := messageformat.New("en", "Hello, {$name}!", &messageformat.MessageFormatOptions{
    BidiIsolation: stringPtr("none"),
})
result, err := mf.Format(map[string]interface{}{
    "name": "World",
})
```

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/datamodel
go test ./internal/cst
```

## 🤝 Contributing

1. Follow Go conventions and best practices
2. Maintain JavaScript/TypeScript API compatibility
3. Include comprehensive test coverage
4. Update documentation for new features

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Related Projects

- [MessageFormat 2.0 Specification](https://github.com/unicode-org/message-format-wg)
- [JavaScript/TypeScript Implementation](https://github.com/messageformat/messageformat)
- [MessageFormat Working Group](https://github.com/unicode-org/message-format-wg)

## 🙏 Credits

This Go implementation is based on the excellent work of the [MessageFormat JavaScript/TypeScript library](https://github.com/messageformat/messageformat) by the MessageFormat team. Special thanks to:

- The [MessageFormat Working Group](https://github.com/unicode-org/message-format-wg) for developing the MessageFormat 2.0 specification
- The maintainers and contributors of the original JavaScript/TypeScript implementation
- The Unicode Consortium for their work on internationalization standards

This project aims to bring the same level of functionality and API compatibility to the Go ecosystem while maintaining the high standards set by the original implementation.