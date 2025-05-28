# API Reference

API documentation for MessageFormat Go implementation of the [Unicode MessageFormat 2.0 specification](https://unicode.org/reports/tr35/tr35-messageFormat.html).

## 📖 Table of Contents

1. [Package Import](#package-import)
2. [Constructor Functions](#constructor-functions)
3. [MessageFormat Type](#messageformat-type)
4. [Configuration Options](#configuration-options)
5. [Formatting Methods](#formatting-methods)
6. [International Features](#international-features)
7. [Error Handling](#error-handling)
8. [Value Types](#value-types)

## 📦 Package Import

```go
import "github.com/kaptinlin/messageformat-go"
```

All examples in this document assume this import statement.

## 🔧 Constructor Functions

### messageformat.New

Creates a new MessageFormat instance with support for single or multiple locales.

```go
func New(locales interface{}, source string, options *MessageFormatOptions) (*MessageFormat, error)
```

**Parameters:**
- `locales` (string | []string): Single locale string or array of locales for negotiation
- `source` (string): MessageFormat 2.0 source text
- `options` (*MessageFormatOptions): Optional configuration options

**Returns:**
- `*MessageFormat`: Configured MessageFormat instance
- `error`: Syntax or configuration error

**Examples:**

```go
// Simple message with single locale
mf, err := messageformat.New("en", "Hello, {$name}!")
if err != nil {
    log.Fatal(err)
}

// Multi-locale support with fallback
mf, err := messageformat.New([]string{"zh-CN", "en", "fr"}, 
    "Price: {$amount :number style=currency currency=USD}", 
    &messageformat.MessageFormatOptions{
        LocaleMatcher: messageformat.LocaleBestFit,
    })

// RTL language with bidirectional text isolation
mf, err := messageformat.New("ar", "مرحبا {$name}!", 
    &messageformat.MessageFormatOptions{
        BidiIsolation: messageformat.BidiDefault,
    })

// Complex message with pattern matching
mf, err := messageformat.New("en", `
.input {$count :number}
.match $count
0   {{No items}}
one {{One item}}
*   {{{$count} items}}
`)

// Message with explicit direction control
mf, err := messageformat.New("en", "Hello, {$name}!", 
    &messageformat.MessageFormatOptions{
        Dir:           messageformat.DirRTL, // Force RTL for English
        BidiIsolation: messageformat.BidiDefault,
    })
```

### messageformat.MustNew

Creates a new MessageFormat instance, panicking on error. Useful for static messages where syntax is guaranteed to be correct.

```go
func MustNew(locales interface{}, source string, options *MessageFormatOptions) *MessageFormat
```

**Parameters:**
- Same as `New()`

**Returns:**
- `*MessageFormat`: Configured MessageFormat instance

**Panics:**
- If construction fails due to syntax or configuration errors

**Examples:**

```go
// Use when you're certain the syntax is correct
mf := messageformat.MustNew("en", "Hello, {$name}!")

// Multi-locale with options
mf := messageformat.MustNew([]string{"de-DE", "en"}, 
    "Price: {$amount :number style=currency currency=EUR}", 
    &messageformat.MessageFormatOptions{
        LocaleMatcher: messageformat.LocaleLookup,
    })

// Useful for static messages in package initialization
var welcomeMessage = messageformat.MustNew("en", "Welcome to {$appName}!")
```

## 🎯 MessageFormat Type

The `MessageFormat` type represents a compiled message template with international support.

### Thread Safety

MessageFormat instances are **thread-safe** after construction. You can safely use the same instance across multiple goroutines for formatting operations.

```go
var globalMessage = messageformat.MustNew("en", "User {$name} logged in")

// Safe to use from multiple goroutines
func handleLogin(name string) {
    result, _ := globalMessage.Format(map[string]interface{}{
        "name": name,
    })
    log.Println(result)
}
```

## ⚙️ Configuration Options

### MessageFormatOptions

Configuration options for MessageFormat instances.

```go
type MessageFormatOptions struct {
    BidiIsolation BidiIsolation                           // Bidirectional text isolation
    Dir           Direction                               // Text direction
    LocaleMatcher LocaleMatcher                           // Locale matching strategy
    Functions     map[string]functions.MessageFunction    // Custom functions
}
```

### BidiIsolation

Controls bidirectional text isolation behavior for mixed LTR/RTL content.

```go
type BidiIsolation string

const (
    BidiDefault BidiIsolation = "default"  // Enable bidi isolation (recommended)
    BidiNone    BidiIsolation = "none"     // Disable bidi isolation
)
```

**Examples:**

```go
// Enable bidi isolation (default for RTL locales)
mf := messageformat.MustNew("ar", "Email: {$email}", 
    &messageformat.MessageFormatOptions{
        BidiIsolation: messageformat.BidiDefault,
    })

result, _ := mf.Format(map[string]interface{}{
    "email": "user@example.com",
})
// Output: Email: ⁨user@example.com⁩

// Disable bidi isolation
mf := messageformat.MustNew("ar", "Simple text", 
    &messageformat.MessageFormatOptions{
        BidiIsolation: messageformat.BidiNone,
    })
```

### Direction

Controls text direction behavior.

```go
type Direction string

const (
    DirLTR  Direction = "ltr"   // Left-to-right
    DirRTL  Direction = "rtl"   // Right-to-left
    DirAuto Direction = "auto"  // Automatic detection based on locale
)
```

**Examples:**

```go
// Automatic direction detection (default)
mf := messageformat.MustNew("ar", "مرحبا", 
    &messageformat.MessageFormatOptions{
        Dir: messageformat.DirAuto, // Will resolve to RTL for Arabic
    })

// Force RTL for English
mf := messageformat.MustNew("en", "Hello", 
    &messageformat.MessageFormatOptions{
        Dir: messageformat.DirRTL,
    })

// Force LTR for Arabic
mf := messageformat.MustNew("ar", "مرحبا", 
    &messageformat.MessageFormatOptions{
        Dir: messageformat.DirLTR,
    })
```

### LocaleMatcher

Controls locale negotiation strategy when multiple locales are provided.

```go
type LocaleMatcher string

const (
    LocaleBestFit LocaleMatcher = "best-fit"  // Best fit algorithm (default)
    LocaleLookup  LocaleMatcher = "lookup"    // Exact lookup algorithm
)
```

**Examples:**

```go
// Best fit matching (default)
mf := messageformat.MustNew([]string{"zh-CN", "en", "fr"}, "Hello", 
    &messageformat.MessageFormatOptions{
        LocaleMatcher: messageformat.LocaleBestFit,
    })

// Exact lookup matching
mf := messageformat.MustNew([]string{"en-US", "en-GB"}, "Hello", 
    &messageformat.MessageFormatOptions{
        LocaleMatcher: messageformat.LocaleLookup,
    })
```

## 🔧 Formatting Methods

### Format

Formats the message with provided variables, supporting optional error callbacks.

```go
func (mf *MessageFormat) Format(variables map[string]interface{}, onError ...func(error)) (string, error)
```

**Parameters:**
- `variables` (map[string]interface{}): Variable values for substitution
- `onError` (...func(error)): Optional error callback for capturing warnings

**Returns:**
- `string`: Formatted message text
- `error`: Runtime formatting error (usually nil due to graceful degradation)

**Examples:**

```go
mf := messageformat.MustNew("en", "Hello, {$name}!")

// Basic usage
result, err := mf.Format(map[string]interface{}{
    "name": "Alice",
})
// result: "Hello, ⁨Alice⁩!"

// With multiple variables
mf2 := messageformat.MustNew("en", "Welcome {$firstName} {$lastName}!")
result, err = mf2.Format(map[string]interface{}{
    "firstName": "John",
    "lastName":  "Doe",
})
// result: "Welcome ⁨John⁩ ⁨Doe⁩!"

// With missing variables (uses fallback)
var warnings []error
onError := func(err error) {
    warnings = append(warnings, err)
}

result, err = mf.Format(map[string]interface{}{
    // Missing "name" variable
}, onError)
// result: "Hello, ⁨{$name}⁩!" (fallback representation)
// err: nil (graceful degradation)
// warnings: contains warning about missing variable

// RTL text with automatic bidi isolation
rtlMf := messageformat.MustNew("ar", "مرحبا {$name}!", 
    &messageformat.MessageFormatOptions{
        BidiIsolation: messageformat.BidiDefault,
    })
result, _ = rtlMf.Format(map[string]interface{}{
    "name": "أحمد",
})
// result: "مرحبا ⁨أحمد⁩!"

// Number formatting
numberMf := messageformat.MustNew("en", 
    "Price: {$amount :number style=currency currency=USD}")
result, _ = numberMf.Format(map[string]interface{}{
    "amount": 42.50,
})
// result: "Price: $42.50"

// Pattern matching
patternMf := messageformat.MustNew("en", `
.input {$count :number}
.match $count
0   {{No items}}
one {{One item}}
*   {{{$count} items}}
`)
result, _ = patternMf.Format(map[string]interface{}{
    "count": 5,
})
// result: "5 items"
```

### FormatToParts

Formats the message and returns structured parts for advanced processing like rich text rendering.

```go
func (mf *MessageFormat) FormatToParts(variables map[string]interface{}) ([]Part, error)
```

**Parameters:**
- `variables` (map[string]interface{}): Variable values for substitution

**Returns:**
- `[]Part`: Array of message parts with type information
- `error`: Runtime formatting error

**Part Types:**
- `"text"`: Literal text content
- `"number"`: Formatted number (currency, percentage, decimal)
- `"markup"`: Markup placeholders (`{#tag}`, `{/tag}`)
- `"bidi"`: Bidirectional text isolation characters

**Examples:**

```go
mf := messageformat.MustNew("en", "Price: {$amount :number style=currency currency=USD}")

parts, err := mf.FormatToParts(map[string]interface{}{
    "amount": 42.50,
})

for _, part := range parts {
    fmt.Printf("Type: %s, Value: %s\n", part.Type(), part.Value())
}
// Output:
// Type: text, Value: Price: 
// Type: number, Value: $42.50

// With detailed number parts
if numberPart, ok := parts[1].(*messagevalue.NumberPart); ok {
    subParts := numberPart.Parts()
    for _, subPart := range subParts {
        fmt.Printf("  SubType: %s, Value: %s\n", subPart.Type(), subPart.Value())
    }
    // Output:
    // SubType: currency, Value: $
    // SubType: integer, Value: 42
    // SubType: decimal, Value: .
    // SubType: fraction, Value: 50
}

// With markup
markupMf := messageformat.MustNew("en", "Welcome {#b}bold text{/b} and normal text")
parts, _ = markupMf.FormatToParts(nil)

for _, part := range parts {
    fmt.Printf("Type: %s, Value: %s\n", part.Type(), part.Value())
    if part.Type() == "markup" {
        if mp, ok := part.(*messagevalue.MarkupPart); ok {
            fmt.Printf("  Kind: %s, Name: %s\n", mp.Kind(), mp.Name())
        }
    }
}
// Output:
// Type: text, Value: Welcome 
// Type: markup, Value: 
//   Kind: open, Name: b
// Type: text, Value: bold text
// Type: markup, Value: 
//   Kind: close, Name: b
// Type: text, Value:  and normal text
```

### ResolvedOptions

Returns the resolved configuration options after locale negotiation and direction detection.

```go
func (mf *MessageFormat) ResolvedOptions() ResolvedMessageFormatOptions
```

**Returns:**
- `ResolvedMessageFormatOptions`: Resolved configuration

**ResolvedMessageFormatOptions Structure:**

```go
type ResolvedMessageFormatOptions struct {
    BidiIsolation BidiIsolation                           // Resolved bidi isolation setting
    Dir           Direction                               // Resolved text direction
    LocaleMatcher LocaleMatcher                           // Locale matching strategy
    Functions     map[string]functions.MessageFunction    // Available functions
}
```

**Examples:**

```go
// Multi-locale with automatic direction detection
mf := messageformat.MustNew([]string{"ar", "en"}, "Hello {$name}!", 
    &messageformat.MessageFormatOptions{
        LocaleMatcher: messageformat.LocaleBestFit,
    })

resolved := mf.ResolvedOptions()
fmt.Printf("Direction: %s\n", resolved.Dir)           // Output: rtl (Arabic detected)
fmt.Printf("LocaleMatcher: %s\n", resolved.LocaleMatcher) // Output: best-fit
fmt.Printf("BidiIsolation: %s\n", resolved.BidiIsolation) // Output: default

// Check available functions
if _, hasNumber := resolved.Functions["number"]; hasNumber {
    fmt.Println("Number function is available")
}

// Direction detection examples
examples := []struct {
    locale      string
    expectedDir Direction
}{
    {"ar", DirRTL}, {"he", DirRTL}, {"fa", DirRTL}, {"ur", DirRTL},
    {"en", DirLTR}, {"zh", DirLTR}, {"fr", DirLTR}, {"de", DirLTR},
    {"ja", DirLTR}, {"ko", DirLTR}, {"es", DirLTR}, {"pt", DirLTR},
}

for _, example := range examples {
    mf := messageformat.MustNew(example.locale, "Test")
    resolved := mf.ResolvedOptions()
    fmt.Printf("%s: %s\n", example.locale, resolved.Dir)
}
```

## 🌍 International Features

### Automatic Language Direction Detection

MessageFormat Go automatically detects text direction for 25+ languages:

**RTL Languages:**
- Arabic (`ar`)
- Hebrew (`he`) 
- Persian/Farsi (`fa`)
- Urdu (`ur`)

**LTR Languages:**
- English (`en`)
- Chinese (`zh`)
- French (`fr`)
- German (`de`)
- Japanese (`ja`)
- Korean (`ko`)
- Spanish (`es`)
- Portuguese (`pt`)
- Italian (`it`)
- Dutch (`nl`)
- Russian (`ru`)
- Hindi (`hi`)
- Thai (`th`)
- Vietnamese (`vi`)
- And more...

**Examples:**

```go
// Automatic RTL detection
arabicMf := messageformat.MustNew("ar", "مرحبا {$name}!")
resolved := arabicMf.ResolvedOptions()
fmt.Println(resolved.Dir) // Output: rtl

// Automatic LTR detection
englishMf := messageformat.MustNew("en", "Hello {$name}!")
resolved = englishMf.ResolvedOptions()
fmt.Println(resolved.Dir) // Output: ltr

// Override automatic detection
forcedRTL := messageformat.MustNew("en", "Hello {$name}!", 
    &messageformat.MessageFormatOptions{
        Dir: messageformat.DirRTL, // Force RTL for English
    })
resolved = forcedRTL.ResolvedOptions()
fmt.Println(resolved.Dir) // Output: rtl
```

### Multi-Locale Support

Support for locale arrays with intelligent fallback:

```go
// Locale negotiation with fallback
mf := messageformat.MustNew([]string{"zh-Hans-CN", "zh", "en"}, 
    "Price: {$amount :number style=currency currency=CNY}", 
    &messageformat.MessageFormatOptions{
        LocaleMatcher: messageformat.LocaleBestFit,
    })

result, _ := mf.Format(map[string]interface{}{
    "amount": 100.50,
})
// Uses best available locale for formatting

// Compare different locale matchers
bestFitMf := messageformat.MustNew([]string{"en-US", "fr-FR"}, "Test", 
    &messageformat.MessageFormatOptions{
        LocaleMatcher: messageformat.LocaleBestFit,
    })

lookupMf := messageformat.MustNew([]string{"en-US", "fr-FR"}, "Test", 
    &messageformat.MessageFormatOptions{
        LocaleMatcher: messageformat.LocaleLookup,
    })

fmt.Println(bestFitMf.ResolvedOptions().LocaleMatcher) // Output: best-fit
fmt.Println(lookupMf.ResolvedOptions().LocaleMatcher)  // Output: lookup
```

### Mixed Content Handling

Proper handling of mixed LTR/RTL content with bidirectional text isolation:

```go
// Mixed English email in Arabic context
mixed := messageformat.MustNew("ar", 
    "البريد الإلكتروني: {$email} - مرحبا {$name}!", 
    &messageformat.MessageFormatOptions{
        BidiIsolation: messageformat.BidiDefault,
    })

result, _ := mixed.Format(map[string]interface{}{
    "email": "user@example.com", // LTR content
    "name":  "أحمد",             // RTL content
})
// Output: "البريد الإلكتروني: ⁨user@example.com⁩ - مرحبا ⁨أحمد⁩!"
// Proper bidi isolation applied automatically

// Compare with isolation disabled
noIsolation := messageformat.MustNew("ar", 
    "البريد الإلكتروني: {$email} - مرحبا {$name}!", 
    &messageformat.MessageFormatOptions{
        BidiIsolation: messageformat.BidiNone,
    })

result2, _ := noIsolation.Format(map[string]interface{}{
    "email": "user@example.com",
    "name":  "أحمد",
})
// Output: "البريد الإلكتروني: user@example.com - مرحبا أحمد"
// No isolation characters
```

### Locale-Aware Number Formatting

Numbers, currencies, and percentages adapt to locale conventions:

```go
// Different locales with same currency
examples := []struct {
    locale string
    amount float64
}{
    {"en-US", 1234.56},
    {"de-DE", 1234.56},
    {"fr-FR", 1234.56},
}

for _, example := range examples {
    mf := messageformat.MustNew(example.locale, 
        "Price: {$amount :number style=currency currency=EUR}")
    
    result, _ := mf.Format(map[string]interface{}{
        "amount": example.amount,
    })
    fmt.Printf("%s: %s\n", example.locale, result)
}
// Output varies by locale formatting conventions

// Japanese locale with JPY currency
japaneseMf := messageformat.MustNew("ja-JP", 
    "価格: {$amount :number style=currency currency=JPY}")
result, _ := japaneseMf.Format(map[string]interface{}{
    "amount": 1234,
})
// Output: "価格: ¥1,234"

// Percentage formatting
percentMf := messageformat.MustNew("en", 
    "Progress: {$rate :number style=percent}")
result, _ = percentMf.Format(map[string]interface{}{
    "rate": 0.75,
})
// Output: "Progress: 75%"
```

## 🛡️ Error Handling

### Graceful Degradation

MessageFormat Go provides graceful error handling with fallback representations:

```go
mf := messageformat.MustNew("en", "Hello {$name}!")

// Missing variable - uses fallback representation
result, err := mf.Format(map[string]interface{}{
    // No variables provided
})

fmt.Println(result) // Output: "Hello ⁨{$name}⁩!"
fmt.Println(err)    // Output: <nil> (graceful degradation)
```

### Error Callbacks

Capture warnings and errors during formatting:

```go
var warnings []error
onError := func(err error) {
    warnings = append(warnings, err)
}

mf := messageformat.MustNew("en", "Hello {$name} and {$missing}!")
result, err := mf.Format(map[string]interface{}{
    "name": "Alice",
    // "missing" variable not provided
}, onError)

fmt.Println(result) // "Hello ⁨Alice⁩ and ⁨{$missing}⁩!"
fmt.Printf("Warnings: %d\n", len(warnings)) // 1 warning about missing variable
```

### Construction Errors

Handle errors during MessageFormat creation:

```go
// Syntax error
_, err := messageformat.New("en", "Invalid syntax: {$missing")
if err != nil {
    fmt.Printf("Syntax error: %v\n", err)
}

// Invalid locale
_, err = messageformat.New("invalid-locale", "Hello!")
if err != nil {
    fmt.Printf("Locale error: %v\n", err)
}

// Invalid options
_, err = messageformat.New("en", "Hello!", 
    &messageformat.MessageFormatOptions{
        Dir: "invalid-direction", // Invalid direction
    })
if err != nil {
    fmt.Printf("Option error: %v\n", err)
}
```

## 🔧 Value Types

### Supported Input Types

MessageFormat Go accepts various Go types as variable values:

```go
mf := messageformat.MustNew("en", 
    "Data: {$str}, {$num :number}, {$bool}, {$nil}")

result, _ := mf.Format(map[string]interface{}{
    "str":  "text",           // string
    "num":  42.5,             // float64
    "bool": true,             // bool
    "nil":  nil,              // nil (uses fallback)
})
// Output: "Data: ⁨text⁩, 42.5, ⁨true⁩, ⁨{$nil}⁩"

// All numeric types are supported
numericMf := messageformat.MustNew("en", "Numbers: {$int :number}, {$float :number}")
result, _ = numericMf.Format(map[string]interface{}{
    "int":   42,      // int
    "float": 3.14159, // float64
})
// Output: "Numbers: 42, 3.14159"

// Complex types are converted to strings
complexMf := messageformat.MustNew("en", "Complex: {$slice}, {$map}")
result, _ = complexMf.Format(map[string]interface{}{
    "slice": []int{1, 2, 3},
    "map":   map[string]int{"a": 1},
})
// Output: "Complex: ⁨[1 2 3]⁩, ⁨map[a:1]⁩"
```

### Custom Functions

Define custom formatting functions with locale awareness:

```go
// Custom uppercase function
customUppercase := func(
    msgCtx functions.MessageFunctionContext,
    options map[string]interface{},
    input interface{},
) messagevalue.MessageValue {
    inputStr := fmt.Sprintf("%v", input)
    upperStr := strings.ToUpper(inputStr)
    
    // Get locale from context
    locales := msgCtx.Locales()
    locale := "en"
    if len(locales) > 0 {
        locale = locales[0]
    }
    
    return messagevalue.NewStringValue(upperStr, locale, msgCtx.Source())
}

// Custom reverse function
customReverse := func(
    msgCtx functions.MessageFunctionContext,
    options map[string]interface{},
    input interface{},
) messagevalue.MessageValue {
    inputStr := fmt.Sprintf("%v", input)
    runes := []rune(inputStr)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    reversedStr := string(runes)
    
    locales := msgCtx.Locales()
    locale := "en"
    if len(locales) > 0 {
        locale = locales[0]
    }
    
    return messagevalue.NewStringValue(reversedStr, locale, msgCtx.Source())
}

mf := messageformat.MustNew("en", "Name: {$first :uppercase} {$last :reverse}", 
    &messageformat.MessageFormatOptions{
        Functions: map[string]functions.MessageFunction{
            "uppercase": customUppercase,
            "reverse":   customReverse,
        },
    })

result, _ := mf.Format(map[string]interface{}{
    "first": "john",
    "last":  "doe",
})
// Output: "Name: ⁨JOHN⁩ ⁨eod⁩"
```

## 🚀 Advanced Usage

### Performance Optimization

```go
// Pre-compile messages for better performance
var (
    welcomeMsg = messageformat.MustNew("en", "Welcome {$name}!")
    errorMsg   = messageformat.MustNew("en", "Error: {$message}")
)

func handleRequest(name string) {
    // Fast formatting with pre-compiled message
    result, _ := welcomeMsg.Format(map[string]interface{}{
        "name": name,
    })
    fmt.Println(result)
}
```

### Concurrent Usage

```go
// MessageFormat instances are thread-safe
var sharedMessage = messageformat.MustNew("en", "User {$id}: {$action}")

func processUsers(users []User) {
    var wg sync.WaitGroup
    
    for _, user := range users {
        wg.Add(1)
        go func(u User) {
            defer wg.Done()
            
            // Safe concurrent access
            result, _ := sharedMessage.Format(map[string]interface{}{
                "id":     u.ID,
                "action": u.LastAction,
            })
            log.Println(result)
        }(user)
    }
    
    wg.Wait()
}
```

### Edge Case Handling

```go
// Comprehensive edge case handling
mf := messageformat.MustNew("en", "Complex: {$data}")

testCases := []interface{}{
    nil,                    // nil value
    "",                     // empty string
    0,                      // zero number
    false,                  // false boolean
    []int{1, 2, 3},        // slice (converted to string)
    map[string]int{"a": 1}, // map (converted to string)
}

for i, testCase := range testCases {
    result, _ := mf.Format(map[string]interface{}{
        "data": testCase,
    })
    fmt.Printf("Case %d: %s\n", i, result)
}
```

## Summary

This API reference provides information needed to effectively use MessageFormat Go in your applications. For syntax details, see the [Message Syntax](message-syntax.md). For error handling strategies, see the [Error Handling](error-handling.md).
