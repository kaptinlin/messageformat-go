# V1 API Reference - ICU MessageFormat

This document provides comprehensive API documentation for MessageFormat Go v1 implementation.

## Core API

### MessageFormat Type

```go
type MessageFormat struct {
    // Internal fields (private)
}
```

The main MessageFormat type that handles message compilation and formatting.

### Constructor Functions

#### New(locale string, options *Options) (*MessageFormat, error)

Creates a new MessageFormat instance for the specified locale.

**Parameters:**
- `locale` (string): BCP 47 language tag (e.g., "en", "zh-CN", "fr-FR")
- `options` (*Options): Configuration options (can be nil for defaults)

**Returns:**
- `*MessageFormat`: New MessageFormat instance
- `error`: Error if locale is invalid or initialization fails

**Example:**
```go
mf, err := messageformat.New("en", nil)
if err != nil {
    log.Fatal(err)
}
```

### Instance Methods

#### Compile(pattern string) (*CompiledMessage, error)

Compiles a message pattern into an optimized, reusable message function.

**Parameters:**
- `pattern` (string): ICU MessageFormat pattern string

**Returns:**
- `*CompiledMessage`: Compiled message ready for formatting
- `error`: Compilation error if pattern is invalid

**Example:**
```go
compiled, err := mf.Compile("Hello, {name}!")
if err != nil {
    log.Fatal(err)
}
```

#### ResolvedOptions() ResolvedOptions

Returns the resolved configuration options for this MessageFormat instance.

**Returns:**
- `ResolvedOptions`: Current configuration settings

### CompiledMessage Methods

#### Format(args map[string]interface{}) (string, error)

Formats the compiled message with the provided arguments.

**Parameters:**
- `args` (map[string]interface{}): Variable values for message placeholders

**Returns:**
- `string`: Formatted message string
- `error`: Formatting error (e.g., missing required arguments)

**Example:**
```go
result, err := compiled.Format(map[string]interface{}{
    "name": "World",
})
// result: "Hello, World!"
```

#### FormatToString(args map[string]interface{}) string

Convenience method that formats the message and returns only the string result.
Panics if formatting fails.

**Parameters:**
- `args` (map[string]interface{}): Variable values for message placeholders

**Returns:**
- `string`: Formatted message string

**Example:**
```go
result := compiled.FormatToString(map[string]interface{}{
    "name": "World",
})
```

## Configuration Types

### Options

```go
type Options struct {
    Currency           string                 // Default currency (ISO 4217, e.g., "USD")
    BiDiSupport        bool                  // Enable bidirectional text support
    StrictPluralKeys   bool                  // Strict plural key validation
    CustomFormatters   map[string]Formatter  // Custom formatting functions
}
```

Configuration options for MessageFormat initialization.

### ResolvedOptions

```go
type ResolvedOptions struct {
    Locale   string // Resolved locale (may differ from requested)
    Currency string // Active currency setting
    // ... other resolved settings
}
```

Resolved configuration after MessageFormat initialization.

### Formatter

```go
type Formatter interface {
    Format(value interface{}, locale string, options map[string]string) (string, error)
}
```

Interface for custom formatting functions.

## Static Functions

### SupportedLocalesOf(locales []string) []string

Returns an array of supported locales from the provided list.

**Parameters:**
- `locales` ([]string): List of BCP 47 language tags to check

**Returns:**
- `[]string`: Array of supported locales from the input

**Example:**
```go
supported := messageformat.SupportedLocalesOf([]string{"en", "fr", "invalid-locale"})
// supported: ["en", "fr"]
```

### Escape(text string) string

Escapes special MessageFormat characters in plain text.

**Parameters:**
- `text` (string): Text to escape

**Returns:**
- `string`: Escaped text safe for use in MessageFormat patterns

**Example:**
```go
escaped := messageformat.Escape("Text with {braces} and 'quotes'")
// escaped: "Text with '{braces}' and ''quotes''"
```

## Error Types

### SyntaxError

```go
type SyntaxError struct {
    Message string
    Pattern string
    Offset  int
}
```

Represents a syntax error in a MessageFormat pattern.

### ResolutionError  

```go
type ResolutionError struct {
    Message string
    Key     string
}
```

Represents an error during message resolution (e.g., missing argument).

## Message Pattern Syntax

### Basic Placeholders

```
Hello, {name}!
```

Simple variable replacement.

### Plural Forms

```go
pattern := `{count, plural,
    =0 {no items}
    one {one item}  
    other {# items}
}`
```

### Select Statements

```go
pattern := `{gender, select,
    male {He}
    female {She}
    other {They}
} will arrive soon.`
```

### Number Formatting

```go
pattern := `Price: {amount, number, currency}`
```

### Date Formatting

```go
pattern := `Today is {date, date, short}`
```

## Performance Considerations

### Compilation vs Runtime

- **Compile once, use many times**: Always compile messages once and reuse
- **Thread safety**: CompiledMessage instances are safe for concurrent use
- **Memory efficiency**: Compiled messages use optimized internal representations

### Best Practices

```go
// ✅ Good: Compile once, reuse
compiled, _ := mf.Compile(pattern)
for i := 0; i < 1000; i++ {
    result, _ := compiled.Format(args)
}

// ❌ Bad: Recompiling every time  
for i := 0; i < 1000; i++ {
    compiled, _ := mf.Compile(pattern)
    result, _ := compiled.Format(args)
}
```

## Advanced Usage

### Custom Formatters

```go
type CurrencyFormatter struct{}

func (f *CurrencyFormatter) Format(value interface{}, locale string, options map[string]string) (string, error) {
    // Custom currency formatting logic
    return fmt.Sprintf("$%.2f", value), nil
}

opts := &messageformat.Options{
    CustomFormatters: map[string]messageformat.Formatter{
        "currency": &CurrencyFormatter{},
    },
}

mf, _ := messageformat.New("en", opts)
```

### Bidirectional Text Support

```go
opts := &messageformat.Options{
    BiDiSupport: true,
}

mf, _ := messageformat.New("ar", opts)
```

## Migration from JavaScript/TypeScript

V1 maintains 100% API compatibility with the official messageformat library:

| JavaScript/TypeScript | Go v1 |
|----------------------|-------|
| `new MessageFormat('en')` | `messageformat.New("en", nil)` |
| `mf.compile(pattern)` | `mf.Compile(pattern)` |
| `compiled(args)` | `compiled.Format(args)` |
| `MessageFormat.supportedLocalesOf()` | `messageformat.SupportedLocalesOf()` |

## Troubleshooting

### Common Issues

1. **Invalid Locale**: Ensure locale follows BCP 47 format
2. **Missing Arguments**: Check that all placeholder variables are provided
3. **Syntax Errors**: Validate MessageFormat pattern syntax
4. **Performance**: Compile messages once and reuse compiled instances

### Debug Mode

Enable verbose error messages during development:

```go
// Check error details
if err != nil {
    if syntaxErr, ok := err.(*messageformat.SyntaxError); ok {
        fmt.Printf("Syntax error at position %d: %s\n", 
            syntaxErr.Offset, syntaxErr.Message)
    }
}
```