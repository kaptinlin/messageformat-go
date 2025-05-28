# MessageFormat 2.0 Go Implementation

[![Go Reference](https://pkg.go.dev/badge/github.com/kaptinlin/messageformat-go.svg)](https://pkg.go.dev/github.com/kaptinlin/messageformat-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/kaptinlin/messageformat-go)](https://goreportcard.com/report/github.com/kaptinlin/messageformat-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A **production-ready** Go implementation of the [Unicode MessageFormat 2.0 specification](https://unicode.org/reports/tr35/tr35-messageFormat.html), providing comprehensive internationalization (i18n) capabilities with advanced features like pluralization, gender selection, bidirectional text support, and custom formatting functions.

## 🏆 Specification Compliance

This implementation passes the official MessageFormat 2.0 test suite from the Unicode Consortium.

## 🚀 Quick Start

### Installation

```bash
go get github.com/kaptinlin/messageformat-go
```

**Requirements**: Go 1.21 or later

### Basic Example

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

## ✨ Key Features

### 🌍 MessageFormat 2.0 Support
- **Pattern Matching**: Advanced `.match` statements with exact number and plural category matching
- **Variable Declarations**: `.input` and `.local` declarations with function annotations
- **Standard Functions**: Built-in `:number`, `:integer`, `:string`, and `:datetime` formatting
- **Custom Functions**: Extensible function system with locale awareness
- **Markup Support**: `{#tag}`, `{/tag}`, `{#tag /}` syntax support
- **Unicode Compliance**: Unicode normalization and bidirectional text handling

### 🌐 International Features
- **Multi-Locale Support**: Intelligent locale fallback and negotiation
- **Automatic Direction Detection**: RTL/LTR detection for 25+ languages
- **Bidirectional Text Isolation**: Configurable Unicode bidi isolation
- **Locale-Aware Formatting**: Currency, numbers, dates, and percentages adapt to locale conventions
- **Mixed Content Handling**: Proper LTR/RTL text mixing in complex layouts

### 🛡️ Production Ready
- **Thread-Safe**: Safe for concurrent use after construction
- **Graceful Error Handling**: Fallback representations for missing variables
- **Performance Optimized**: Efficient parsing and formatting algorithms
- **TypeScript Compatible**: API designed to match the TypeScript implementation
- **Testing**: 100+ test cases covering specification compliance

## 📖 Documentation

| Guide | Description |
|-------|-------------|
| **[Getting Started](docs/getting-started.md)** | Installation, basic concepts, and first steps |
| **[Message Syntax](docs/message-syntax.md)** | MessageFormat 2.0 syntax reference |
| **[API Reference](docs/api-reference.md)** | API documentation with examples |
| **[Formatting Functions](docs/formatting-functions.md)** | Built-in and custom function development |
| **[Custom Functions](docs/custom-functions.md)** | Advanced function development guide |
| **[Error Handling](docs/error-handling.md)** | Error handling strategies |

## 🎯 Usage Examples

### Number Formatting with Localization
```go
mf, err := messageformat.New("de-DE", 
    "Preis: {$amount :number style=currency currency=EUR}")

result, err := mf.Format(map[string]interface{}{
    "amount": 1234.56,
})
// Output: "Preis: 1.234,56 €"
```

### Advanced Pluralization
```go
mf, err := messageformat.New("en", `
.input {$count :number}
.match $count
0   {{No items in your cart}}
1   {{One item in your cart}}
*   {{{$count} items in your cart}}
`)

result, err := mf.Format(map[string]interface{}{
    "count": 5,
})
// Output: "5 items in your cart"
```
### Multi-Selector Pattern Matching
```go
mf, err := messageformat.New("en", `
.input {$photoCount :number}
.input {$userGender :string}
.match $photoCount $userGender
0   *     {{{$userName} has no photos}}
1   male  {{{$userName} has one photo in his album}}
1   *     {{{$userName} has one photo in her album}}
*   male  {{{$userName} has {$photoCount} photos in his album}}
*   *     {{{$userName} has {$photoCount} photos in her album}}
`)
```

### Custom Functions with Locale Support
```go
import (
    "strings"
    "github.com/kaptinlin/messageformat-go/pkg/functions"
    "github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

func customUppercase(ctx functions.MessageFunctionContext, options map[string]interface{}, input interface{}) messagevalue.MessageValue {
    locales := ctx.Locales()
    locale := "en"
    if len(locales) > 0 {
        locale = locales[0]
    }
    
    str := fmt.Sprintf("%v", input)
    return messagevalue.NewStringValue(strings.ToUpper(str), locale, ctx.Source())
}

mf, err := messageformat.New("en", "Hello, {$name :uppercase}!",
    messageformat.WithFunction("uppercase", customUppercase),
)
```

### Structured Output for Rich Text
```go
parts, err := mf.FormatToParts(map[string]interface{}{
    "name": "World",
    "count": 42,
})

for _, part := range parts {
    switch p := part.(type) {
    case *messageformat.MessageTextPart:
        fmt.Printf("Text: %s\n", p.Value())
    case *messageformat.MessageNumberPart:
        fmt.Printf("Number: %s (locale: %s)\n", p.Value(), p.Locale())
    case *messageformat.MessageStringPart:
        fmt.Printf("Variable: %s\n", p.Value())
    }
}
```

## 🎛️ Configuration Options

### Functional Options (Recommended)
```go
mf, err := messageformat.New("ar", "مرحبا {$name}!",
    messageformat.WithBidiIsolation("default"),
    messageformat.WithDir("rtl"),
    messageformat.WithFunction("custom", myCustomFunction),
)
```

### Traditional Options Structure
```go
mf, err := messageformat.New("en", "Hello, {$name}!", &messageformat.MessageFormatOptions{
    BidiIsolation: messageformat.BidiNone,
    Dir:          messageformat.DirLTR,
    Functions:    map[string]messageformat.MessageFunction{
        "custom": myCustomFunction,
    },
})
```

### TypeScript Mapping Guide
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

## 🧪 Testing & Verification

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

# Run official MessageFormat 2.0 test suite only
make test-official

# Run tests with coverage report
make test-coverage

# Run benchmarks
make bench
```

### Development Workflow
```bash
# Show all available commands
make help

# Format code and run all checks
make verify

# Run examples to verify functionality
make examples
```

📋 **For detailed testing instructions, see [TESTING.md](TESTING.md)**

## 🌐 Features

### Unicode Features
- **Bidirectional Text**: Unicode Bidirectional Algorithm support
- **Text Isolation**: Configurable bidi isolation (`auto`, `none`, `always`)
- **Normalization**: Unicode normalization for consistent text handling
- **Mixed Scripts**: Proper handling of mixed LTR/RTL content

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines on:
- Development setup and workflow
- Code standards and testing requirements
- Commit message conventions (Conventional Commits)
- Pull request process

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Related Projects

- **[MessageFormat 2.0 Specification](https://github.com/unicode-org/message-format-wg)** - Official Unicode specification
- **[JavaScript/TypeScript Implementation](https://github.com/messageformat/messageformat)** - Reference implementation
- **[MessageFormat Working Group](https://github.com/unicode-org/message-format-wg)** - Unicode working group
- **[ICU MessageFormat](https://unicode-org.github.io/icu/userguide/format_parse/messages/)** - ICU implementation

## 🙏 Acknowledgments

This Go implementation is inspired by the [MessageFormat JavaScript/TypeScript library](https://github.com/messageformat/messageformat) and follows the official [Unicode MessageFormat 2.0 specification](https://unicode.org/reports/tr35/tr35-messageFormat.html). 

Special thanks to:
- The [Unicode MessageFormat Working Group](https://github.com/unicode-org/message-format-wg) for their work on internationalization standards
- The Unicode Consortium for maintaining the specification
- The open-source community for their contributions and feedback

---

**Ready to internationalize your Go applications?** Start with our [Getting Started Guide](docs/getting-started.md) or explore the [API Reference](docs/api-reference.md) for advanced usage patterns.
