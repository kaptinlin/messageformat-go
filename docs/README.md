# MessageFormat 2.0 Go Documentation

Documentation for MessageFormat 2.0 Go implementation with **100% Unicode specification compliance**.

## 📋 Overview

This Go implementation provides support for the [Unicode MessageFormat 2.0 specification](https://unicode.org/reports/tr35/tr35-messageFormat.html), offering internationalization capabilities for modern applications.

### Key Features
- **100% Specification Compliance**: Passes all official MessageFormat 2.0 tests
- **Unicode Support**: Bidirectional text and normalization support
- **Production Ready**: Thread-safe, performant, and reliable
- **Rich Formatting**: Numbers, currencies, dates, and custom functions
- **Pattern Matching**: Advanced pluralization and conditional logic

### Quick Start
```bash
go get github.com/kaptinlin/messageformat-go
```

```go
import "github.com/kaptinlin/messageformat-go"

mf, err := messageformat.New("en", "Hello, {$name}!")
result, err := mf.Format(map[string]interface{}{
    "name": "World",
})
// Output: Hello, ⁨World⁩!
```

## 📚 Documentation Guide

| Guide | Description |
|-------|-------------|
| **[Getting Started](getting-started.md)** | Installation, basic concepts, and first steps |
| **[Message Syntax](message-syntax.md)** | MessageFormat 2.0 syntax reference |
| **[API Reference](api-reference.md)** | API documentation with examples |
| **[Formatting Functions](formatting-functions.md)** | Built-in formatting functions |
| **[Custom Functions](custom-functions.md)** | Custom function development |
| **[Error Handling](error-handling.md)** | Error handling strategies |

## 🔢 Number Formatting

### Basic Numbers
```go
mf, _ := messageformat.New("en", "Count: {$count :number}")
result, _ := mf.Format(map[string]interface{}{"count": 1234})
// Output: Count: 1,234
```

### Currency
```go
mf, _ := messageformat.New("en-US", 
    "Price: {$amount :number style=currency currency=USD}")
result, _ := mf.Format(map[string]interface{}{"amount": 42.50})
// Output: Price: $42.50
```

### Percentage
```go
mf, _ := messageformat.New("en", 
    "Progress: {$progress :number style=percent}")
result, _ := mf.Format(map[string]interface{}{"progress": 0.75})
// Output: Progress: 75%
```

## 🔀 Pattern Matching

### Simple Pluralization
```go
mf, _ := messageformat.New("en", `
.input {$count :number}
.match $count
0   {{No items}}
1   {{One item}}
*   {{{$count} items}}
`)
```

### Multi-Selector
```go
mf, _ := messageformat.New("en", `
.input {$count :number}
.input {$gender :string}
.match $count $gender
0   *      {{No photos}}
1   male   {{One photo in his album}}
1   female {{One photo in her album}}
*   male   {{{$count} photos in his album}}
*   female {{{$count} photos in her album}}
`)
```

## 🌍 Internationalization

### Multi-Locale Support
```go
mf, _ := messageformat.New([]string{"de-DE", "en"}, 
    "Preis: {$amount :number style=currency currency=EUR}")
```

### RTL Languages
```go
mf, _ := messageformat.New("ar", "مرحبا {$name}!",
    messageformat.WithBidiIsolation("default"),
)
```

### Direction Detection
```go
mf, _ := messageformat.New("he", "שלום {$name}!")
options := mf.ResolvedOptions()
fmt.Println(options.Dir) // Output: rtl
```

## 📅 Date & Time Formatting

### Basic Date
```go
mf, _ := messageformat.New("en", 
    "Date: {$date :datetime style=short}")
result, _ := mf.Format(map[string]interface{}{
    "date": time.Now(),
})
```

### Custom Date Format
```go
mf, _ := messageformat.New("en", 
    "Date: {$date :datetime dateStyle=full timeStyle=short}")
```

## 🎨 Custom Functions

### Simple Custom Function
```go
import (
    "strings"
    "github.com/kaptinlin/messageformat-go/pkg/functions"
    "github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

func uppercase(ctx functions.MessageFunctionContext, options map[string]interface{}, input interface{}) messagevalue.MessageValue {
    str := fmt.Sprintf("%v", input)
    locale := "en"
    if len(ctx.Locales()) > 0 {
        locale = ctx.Locales()[0]
    }
    return messagevalue.NewStringValue(strings.ToUpper(str), locale, ctx.Source())
}

mf, _ := messageformat.New("en", "Hello, {$name :uppercase}!",
    messageformat.WithFunction("uppercase", uppercase),
)
```

## 🔧 Configuration Options

### Functional Options (Recommended)
```go
mf, _ := messageformat.New("ar", "مرحبا {$name}!",
    messageformat.WithBidiIsolation("default"),
    messageformat.WithDir("rtl"),
    messageformat.WithFunction("custom", myFunction),
)
```

### Options Struct
```go
mf, _ := messageformat.New("en", "Hello, {$name}!", 
    &messageformat.MessageFormatOptions{
        BidiIsolation: messageformat.BidiNone,
        Dir:          messageformat.DirLTR,
        Functions: map[string]messageformat.MessageFunction{
            "custom": myFunction,
        },
    })
```

## 🔧 Structured Output

### FormatToParts
```go
parts, _ := mf.FormatToParts(map[string]interface{}{
    "name": "World",
    "count": 42,
})

for _, part := range parts {
    switch p := part.(type) {
    case *messageformat.MessageTextPart:
        fmt.Printf("Text: %s\n", p.Value())
    case *messageformat.MessageNumberPart:
        fmt.Printf("Number: %s\n", p.Value())
    case *messageformat.MessageStringPart:
        fmt.Printf("Variable: %s\n", p.Value())
    }
}
```

## 🛡️ Error Handling

### Graceful Degradation
```go
mf, _ := messageformat.New("en", "Hello, {$name}!")

// Missing variable - graceful fallback
result, err := mf.Format(map[string]interface{}{})
fmt.Println(result) // Output: Hello, ⁨{$name}⁩!
fmt.Println(err)    // Output: <nil>
```

### Error Callbacks
```go
var warnings []error
onError := func(err error) {
    warnings = append(warnings, err)
}

result, err := mf.Format(map[string]interface{}{}, onError)
fmt.Printf("Warnings: %d\n", len(warnings))
```

## 🎯 Common Patterns

### Shopping Cart
```go
mf, _ := messageformat.New("en", `
.input {$count :number}
.match $count
0 {{Your cart is empty}}
1 {{You have one item in your cart}}
* {{You have {$count} items in your cart}}
`)
```

### User Notifications
```go
mf, _ := messageformat.New("en", `
.input {$photoCount :number}
.input {$userGender :string}
.match $photoCount $userGender
0   *     {{{$userName} has no photos}}
1   male  {{{$userName} has one photo}}
1   *     {{{$userName} has one photo}}
*   male  {{{$userName} has {$photoCount} photos}}
*   *     {{{$userName} has {$photoCount} photos}}
`)
```

### Multi-Currency Pricing
```go
currencies := map[string]string{
    "en-US": "USD",
    "de-DE": "EUR", 
    "ja-JP": "JPY",
}

for locale, currency := range currencies {
    mf, _ := messageformat.New(locale, 
        "Price: {$amount :number style=currency currency="+currency+"}")
    result, _ := mf.Format(map[string]interface{}{"amount": 42.50})
    fmt.Printf("%s: %s\n", locale, result)
}
```

## 🌐 Locale Support

### RTL Languages
- Arabic (`ar`), Hebrew (`he`), Persian (`fa`), Urdu (`ur`)

### LTR Languages  
- English (`en`), Chinese (`zh`), French (`fr`), German (`de`)
- Japanese (`ja`), Korean (`ko`), Spanish (`es`), Portuguese (`pt`)
- Italian (`it`), Dutch (`nl`), Russian (`ru`), Hindi (`hi`)
- And 15+ more with automatic direction detection

## 🔍 Debugging

### Inspect Resolved Options
```go
mf, _ := messageformat.New("ar", "مرحبا {$name}!")
options := mf.ResolvedOptions()
fmt.Printf("Locale: %s\n", options.Locale)
fmt.Printf("Direction: %s\n", options.Dir)
fmt.Printf("Bidi Isolation: %s\n", options.BidiIsolation)
```

### Verbose Error Information
```go
mf, err := messageformat.New("en", "Invalid {$syntax")
if err != nil {
    fmt.Printf("Parse error: %v\n", err)
}
```

## ⚡ Performance Tips

### Reuse MessageFormat Instances
```go
// Good - reuse instances
mf, _ := messageformat.New("en", "Hello, {$name}!")
for _, name := range names {
    result, _ := mf.Format(map[string]interface{}{"name": name})
    fmt.Println(result)
}
```

### Pre-compile Complex Messages
```go
// Pre-compile once, use many times
complexMf, _ := messageformat.New("en", `
.input {$count :number}
.input {$gender :string}
.match $count $gender
// ... complex pattern
`)
```

## 🧪 Testing

### Run Tests
```bash
# All tests including official test suite
task test

# Unit tests only (faster)
task test-unit

# Official MessageFormat 2.0 test suite
task test-official
```

### Benchmarks
```bash
# Performance benchmarks
make bench

# Specific benchmarks
go test -bench=BenchmarkSimpleMessage ./...
```

## 🔗 External Resources

- **[Unicode MessageFormat 2.0 Specification](https://unicode.org/reports/tr35/tr35-messageFormat.html)**
- **[Official Test Suite](https://github.com/unicode-org/message-format-wg/tree/main/test)**
- **[JavaScript Implementation](https://github.com/messageformat/messageformat)**

## Summary

MessageFormat Go provides a production-ready implementation of Unicode MessageFormat 2.0 with support for internationalization, pluralization, and custom formatting. The library offers graceful error handling, thread safety, and extensive locale support for building multilingual applications.
