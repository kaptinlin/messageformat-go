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

# Run basic examples - learn fundamentals
go run examples/basic/main.go

# Run pluralization examples - see .match patterns
go run examples/pluralization/main.go

# Run custom functions examples - build your own formatters
go run examples/custom-functions/main.go

# Run advanced features examples - production patterns
go run examples/advanced/main.go
```

### Expected Output Examples

**Basic Example:**
```
=== MessageFormat 2.0 Basic Usage Examples ===
1. Simple Variable Substitution:
   Input: "Hello, {$name}!"
   Variables: name = "World"
   Output: Hello, World!
```

**Pluralization Example:**
```
1. Basic Pluralization:
   count = 0: No messages
   count = 1: One message
   count = 5: 5 messages
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
Demonstrates the most fundamental MessageFormat 2.0 features:
- Simple variable substitution (`{$name}`)
- Number formatting with `:number` function
- Multiple variable handling
- Localization comparison (English vs Chinese)
- Built-in formatting functions (currency, integer, string)
- Error handling patterns
- Functional options pattern vs traditional struct options
- Bidirectional text isolation controls

### 2. Pluralization (`pluralization/`)
Shows how to use MessageFormat 2.0's `.match` statements and selection:
- Basic pluralization with `.input {$count :number}` and `.match $count`
- Localized pluralization comparison (English vs Chinese)
- Gender selection with `.match $gender`
- Complex multi-selector matching (count + gender)
- Status-based selection patterns
- Time-based conditional selection
- File type selection with fallback patterns

### 3. Custom Functions (`custom-functions/`)
Demonstrates how to create and use custom formatting functions:
- Custom function registration with `WithFunction()`
- Text transformation functions (uppercase, reverse)
- Functions with options (emoji with type parameter)
- Time-based formatting functions (timeago)
- String formatting with alignment
- Multiple functions in one message
- Comprehensive error handling in custom functions
- Type-safe function parameter handling

### 4. Advanced Features (`advanced/`)
Covers advanced MessageFormat 2.0 capabilities:
- Structured output with `FormatToParts()` for rich text rendering
- Bidirectional text support with configurable isolation
- Complex multi-selector pattern matching (.match statements)
- Custom functions with complex logic and styling
- Performance optimization techniques and caching
- Custom error handlers with `WithErrorHandler()`
- Nested patterns and local declarations
- Multi-locale support and fallback handling
- Working with structured data types

## 🌍 Languages Demonstrated

The examples showcase internationalization with:
- **English (en)** - Primary language for all examples
- **Chinese Simplified (zh-CN)** - Used in localization and pluralization comparison examples

All examples are designed to be easily adapted for additional locales.

## 🔧 Technical Features Demonstrated

- ✅ **MessageFormat 2.0 Specification**: Complete `.input`, `.match`, and pattern syntax
- ✅ **Variable Substitution**: `{$name}` syntax with type safety
- ✅ **Number Formatting**: `:number`, `:integer` functions with locale support
- ✅ **Pluralization**: Complex `.match` statements with exact numbers and categories
- ✅ **Custom Functions**: Registration with `WithFunction()` and parameter options
- ✅ **Bidirectional Text**: Unicode bidi isolation controls
- ✅ **Error Handling**: Graceful fallbacks and custom error handlers
- ✅ **Performance**: Optimization techniques and caching patterns
- ✅ **Structured Output**: `FormatToParts()` for rich text rendering
- ✅ **Multi-locale**: Language-specific formatting and fallback handling
- ✅ **Functional Options**: Modern Go API patterns vs traditional structs

## 📚 Learning Resources

- [MessageFormat 2.0 Specification](https://github.com/unicode-org/message-format-wg)
- [Go Library Documentation](https://github.com/kaptinlin/messageformat-go)
- [Internationalization Best Practices](https://unicode.org/reports/tr35/tr35-messageFormat.html)
