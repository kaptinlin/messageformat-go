# MessageFormat 2.0 Go Examples

This directory contains comprehensive examples demonstrating how to use the [MessageFormat 2.0 Go library](https://github.com/kaptinlin/messageformat-go).

MessageFormat 2.0 is the next-generation internationalization standard developed by the Unicode Consortium, providing powerful message formatting capabilities including pluralization, gender selection, custom formatting functions, and more advanced features.

## 🚀 Quick Start

```bash
# Clone the repository
git clone https://github.com/kaptinlin/messageformat-go
cd messageformat-go

# Install dependencies
go mod tidy

# Run basic examples
go run examples/basic/main.go

# Run pluralization examples
go run examples/pluralization/main.go

# Run custom functions examples
go run examples/custom-functions/main.go

# Run advanced features examples
go run examples/advanced/main.go
```

## 📁 Project Structure

```
examples/
├── basic/              # Basic usage examples
├── pluralization/      # Pluralization and select messages
├── custom-functions/   # Custom formatting functions
├── advanced/          # Advanced features and patterns
└── README.md          # This file
```

## 📖 Example Categories

### 1. Basic Usage (`basic/`)
Demonstrates the most fundamental message formatting features:
- Variable substitution
- Simple formatting
- Multiple variables
- Error handling
- Built-in functions

### 2. Pluralization (`pluralization/`)
Shows how to use MessageFormat 2.0's select message functionality:
- Basic plural forms
- Gender selection
- Complex matching patterns
- Multiple selector combinations

### 3. Custom Functions (`custom-functions/`)
Demonstrates how to create and use custom formatting functions:
- Function registration
- Custom formatters
- Function options
- Error handling in functions

### 4. Advanced Features (`advanced/`)
Covers advanced MessageFormat 2.0 capabilities:
- Bidirectional text support
- Structured output (FormatToParts)
- Complex message patterns
- Performance optimization
- Best practices

## 🌍 Supported Languages

The examples demonstrate internationalization with:
- English (en)
- Chinese Simplified (zh-CN)

## 🔧 Technical Features Demonstrated

- ✅ Complete MessageFormat 2.0 support
- ✅ Multi-language internationalization
- ✅ Pluralization handling
- ✅ Custom formatting functions
- ✅ Error handling best practices
- ✅ Performance optimization examples
- ✅ Functional options pattern
- ✅ Structured output support

## 📚 Learning Resources

- [MessageFormat 2.0 Specification](https://github.com/unicode-org/message-format-wg)
- [Go Library Documentation](https://github.com/kaptinlin/messageformat-go)
- [Internationalization Best Practices](https://unicode.org/reports/tr35/tr35-messageFormat.html)
