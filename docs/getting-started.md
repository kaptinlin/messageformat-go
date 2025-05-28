# Getting Started

Welcome to MessageFormat Go! This guide will help you get up and running with a production-ready implementation of the [Unicode MessageFormat 2.0 specification](https://unicode.org/reports/tr35/tr35-messageFormat.html).

## 📖 Table of Contents

1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [Basic Concepts](#basic-concepts)
4. [Your First Messages](#your-first-messages)
5. [International Features](#international-features)
6. [Pattern Matching](#pattern-matching)
7. [Number Formatting](#number-formatting)
8. [Error Handling](#error-handling)
9. [Next Steps](#next-steps)

## 📦 Installation

### Requirements

- **Go 1.21 or later**
- No external dependencies (pure Go implementation)

### Install the Package

```bash
go get github.com/kaptinlin/messageformat-go
```

### Verify Installation

Create a simple test file to verify the installation:

```go
// test.go
package main

import (
    "fmt"
    "log"
    "github.com/kaptinlin/messageformat-go"
)

func main() {
    mf, err := messageformat.New("en", "Hello, {$name}!")
    if err != nil {
        log.Fatal(err)
    }
    
    result, err := mf.Format(map[string]interface{}{
        "name": "World",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(result) // Output: Hello, ⁨World⁩!
}
```

Run the test:

```bash
go run test.go
```

If you see "Hello, ⁨World⁩!", you're ready to go! The `⁨` and `⁩` characters are Unicode bidirectional isolation characters that ensure proper text rendering in international contexts.

## 🚀 Quick Start

Let's start with the most common use cases:

### Simple Variable Substitution

```go
package main

import (
    "fmt"
    "log"
    "github.com/kaptinlin/messageformat-go"
)

func main() {
    // Create a message template
    mf, err := messageformat.New("en", "Welcome, {$username}!")
    if err != nil {
        log.Fatal(err)
    }
    
    // Format with variables
    result, err := mf.Format(map[string]interface{}{
        "username": "Alice",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(result) // Output: Welcome, ⁨Alice⁩!
}
```

### Multiple Variables

```go
mf, err := messageformat.New("en", 
    "Hello {$firstName} {$lastName}! You have {$count :number} messages.")
if err != nil {
    log.Fatal(err)
}

result, err := mf.Format(map[string]interface{}{
    "firstName": "John",
    "lastName":  "Doe", 
    "count":     5,
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(result) 
// Output: Hello ⁨John⁩ ⁨Doe⁩! You have 5 messages.
```

### Currency Formatting

```go
mf, err := messageformat.New("en", 
    "Total: {$amount :number style=currency currency=USD}")
if err != nil {
    log.Fatal(err)
}

result, err := mf.Format(map[string]interface{}{
    "amount": 42.50,
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(result) // Output: Total: $42.50
```

## 🧠 Basic Concepts

### MessageFormat 2.0 Syntax

MessageFormat 2.0 uses a powerful syntax for internationalization:

- **Variables**: `{$variableName}` - Insert variable values
- **Functions**: `{$variable :function}` - Apply formatting functions
- **Options**: `{$variable :function option=value}` - Configure function behavior
- **Declarations**: `.input {$var :function}` - Declare variable types
- **Pattern Matching**: `.match $var` - Conditional message variants

### Thread Safety

MessageFormat instances are **thread-safe** after construction. You can safely share them across goroutines:

```go
// Global message template (safe to share)
var welcomeMessage = messageformat.MustNew("en", "Welcome, {$name}!")

func handleUser(name string) {
    // Safe to call from multiple goroutines
    result, _ := welcomeMessage.Format(map[string]interface{}{
        "name": name,
    })
    fmt.Println(result)
}
```

### Graceful Error Handling

MessageFormat Go uses graceful degradation - missing variables show fallback representations instead of causing errors:

```go
mf := messageformat.MustNew("en", "Hello {$name}! You have {$count :number} items.")

// Missing variables - uses fallback representations
result, err := mf.Format(map[string]interface{}{
    "name": "Alice",
    // "count" is missing
})

fmt.Printf("Result: %s\n", result) // Hello ⁨Alice⁩! You have ⁨{$count}⁩ items.
fmt.Printf("Error: %v\n", err)     // <nil> (no error returned)
```

## 📝 Your First Messages

### Step 1: Simple Text with Variables

```go
package main

import (
    "fmt"
    "github.com/kaptinlin/messageformat-go"
)

func main() {
    // Create message templates
    greeting := messageformat.MustNew("en", "Good morning, {$name}!")
    farewell := messageformat.MustNew("en", "Goodbye, {$name}! See you {$when}.")
    
    // User data
    user := map[string]interface{}{
        "name": "Sarah",
        "when": "tomorrow",
    }
    
    // Format messages
    fmt.Println(greeting.Format(user))  // Good morning, ⁨Sarah⁩!
    fmt.Println(farewell.Format(user))  // Goodbye, ⁨Sarah⁩! See you ⁨tomorrow⁩.
}
```

### Step 2: Adding Number Formatting

```go
// Shopping cart message
cart := messageformat.MustNew("en", 
    "Cart total: {$total :number style=currency currency=USD} ({$items :number} items)")

result, _ := cart.Format(map[string]interface{}{
    "total": 127.50,
    "items": 3,
})

fmt.Println(result) // Cart total: $127.50 (3 items)
```

### Step 3: Percentage and Decimal Formatting

```go
// Progress and statistics
progress := messageformat.MustNew("en", 
    "Download progress: {$percent :number style=percent}")

stats := messageformat.MustNew("en", 
    "Average rating: {$rating :number minimumFractionDigits=1}")

fmt.Println(progress.Format(map[string]interface{}{
    "percent": 0.75,
})) // Download progress: 75%

fmt.Println(stats.Format(map[string]interface{}{
    "rating": 4.5,
})) // Average rating: 4.5
```

## 🌍 International Features

### Multi-Locale Support

MessageFormat Go supports multiple locales with intelligent fallback:

```go
// Locale array with fallback
mf, err := messageformat.New([]string{"zh-CN", "en", "fr"}, 
    "Price: {$amount :number style=currency currency=USD}", 
    &messageformat.MessageFormatOptions{
        LocaleMatcher: messageformat.LocaleBestFit,
    })

result, _ := mf.Format(map[string]interface{}{
    "amount": 99.99,
})

fmt.Println(result) // Price: $99.99 (formatted according to best available locale)

// Check resolved options
resolved := mf.ResolvedOptions()
fmt.Printf("Direction: %s\n", resolved.Dir)
fmt.Printf("LocaleMatcher: %s\n", resolved.LocaleMatcher)
```

### Automatic Direction Detection

MessageFormat Go automatically detects text direction for 25+ languages:

```go
// Arabic (RTL)
arabicMf := messageformat.MustNew("ar", "مرحبا {$name}!")
fmt.Println(arabicMf.ResolvedOptions().Dir) // Output: rtl

// English (LTR)  
englishMf := messageformat.MustNew("en", "Hello {$name}!")
fmt.Println(englishMf.ResolvedOptions().Dir) // Output: ltr

// Hebrew (RTL)
hebrewMf := messageformat.MustNew("he", "שלום {$name}!")
fmt.Println(hebrewMf.ResolvedOptions().Dir) // Output: rtl

// Chinese (LTR)
chineseMf := messageformat.MustNew("zh", "你好 {$name}!")
fmt.Println(chineseMf.ResolvedOptions().Dir) // Output: ltr
```

### Bidirectional Text Isolation

Handle mixed LTR/RTL content properly:

```go
// Mixed content with bidi isolation
mixed := messageformat.MustNew("ar", 
    "البريد الإلكتروني: {$email} - مرحبا {$name}!", 
    &messageformat.MessageFormatOptions{
        BidiIsolation: messageformat.BidiDefault,
    })

result, _ := mixed.Format(map[string]interface{}{
    "email": "user@example.com", // LTR content
    "name":  "أحمد",             // RTL content
})

fmt.Println(result)
// Output: البريد الإلكتروني: ⁨user@example.com⁩ - مرحبا ⁨أحمد⁩!
// Note: Proper bidi isolation characters are automatically added
```

### Locale-Specific Number Formatting

Different locales format numbers differently:

```go
// German locale (uses comma for decimal separator)
germanMf := messageformat.MustNew("de-DE", 
    "Preis: {$amount :number style=currency currency=EUR}")

// Japanese locale (no decimal places for JPY)
japaneseMf := messageformat.MustNew("ja-JP", 
    "価格: {$amount :number style=currency currency=JPY}")

// French locale (uses space as thousands separator)
frenchMf := messageformat.MustNew("fr-FR", 
    "Prix: {$amount :number style=currency currency=EUR}")

amount := 1234.56

fmt.Println(germanMf.Format(map[string]interface{}{"amount": amount}))
// Output: Preis: €1.234,56

fmt.Println(japaneseMf.Format(map[string]interface{}{"amount": 1234}))
// Output: 価格: ¥1,234

fmt.Println(frenchMf.Format(map[string]interface{}{"amount": amount}))
// Output: Prix: €1 234,56
```

## 🔀 Pattern Matching

Pattern matching allows conditional message variants based on variable values:

### Basic Pluralization

```go
// Simple plural handling
mf := messageformat.MustNew("en", `
.input {$count :number}
.match $count
0   {{No items in cart}}
one {{One item in cart}}
*   {{{$count} items in cart}}
`)

// Test different values
for _, count := range []int{0, 1, 5} {
    result, _ := mf.Format(map[string]interface{}{
        "count": count,
    })
    fmt.Printf("Count %d: %s\n", count, result)
}
// Output:
// Count 0: No items in cart
// Count 1: One item in cart
// Count 5: 5 items in cart
```

### Exact Number Matching

```go
// Exact numbers take priority over plural categories
mf := messageformat.MustNew("en", `
.input {$count :number}
.match $count
0   {{Your inbox is empty}}
1   {{You have one new message}}
2   {{You have a couple of messages}}
*   {{You have {$count} new messages}}
`)

for _, count := range []int{0, 1, 2, 5} {
    result, _ := mf.Format(map[string]interface{}{
        "count": count,
    })
    fmt.Printf("Count %d: %s\n", count, result)
}
// Output:
// Count 0: Your inbox is empty
// Count 1: You have one new message
// Count 2: You have a couple of messages
// Count 5: You have 5 new messages
```

### String Selection

```go
// Select based on string values
mf := messageformat.MustNew("en", `
.match {$status :string}
online  {{User is currently online}}
offline {{User is offline}}
away    {{User is away}}
*       {{User status unknown}}
`)

statuses := []string{"online", "offline", "away", "busy"}

for _, status := range statuses {
    result, _ := mf.Format(map[string]interface{}{
        "status": status,
    })
    fmt.Printf("Status %s: %s\n", status, result)
}
// Output:
// Status online: User is currently online
// Status offline: User is offline
// Status away: User is away
// Status busy: User status unknown
```

### Multi-Dimensional Matching

```go
// Match on multiple variables
mf := messageformat.MustNew("en", `
.input {$gender :string}
.input {$count :number}
.match $gender $count
male 0     {{{$name} has no items in his cart}}
male one   {{{$name} has one item in his cart}}
male *     {{{$name} has {$count} items in his cart}}
female 0   {{{$name} has no items in her cart}}
female one {{{$name} has one item in her cart}}
female *   {{{$name} has {$count} items in her cart}}
* 0        {{{$name} has no items in their cart}}
* one      {{{$name} has one item in their cart}}
* *        {{{$name} has {$count} items in their cart}}
`)

result, _ := mf.Format(map[string]interface{}{
    "name":   "Alex",
    "gender": "female",
    "count":  3,
})

fmt.Println(result) // Output: ⁨Alex⁩ has 3 items in her cart
```

## 🔢 Number Formatting

### Currency Formatting

```go
// Different currencies and locales
currencies := []struct {
    locale   string
    currency string
    amount   float64
}{
    {"en-US", "USD", 42.50},
    {"de-DE", "EUR", 42.50},
    {"ja-JP", "JPY", 1000},
    {"en-GB", "GBP", 42.50},
}

for _, c := range currencies {
    mf := messageformat.MustNew(c.locale, 
        "Price: {$amount :number style=currency currency="+c.currency+"}")
    
    result, _ := mf.Format(map[string]interface{}{
        "amount": c.amount,
    })
    
    fmt.Printf("%s: %s\n", c.locale, result)
}
// Output:
// en-US: Price: $42.50
// de-DE: Price: €42.50
// ja-JP: Price: ¥1,000
// en-GB: Price: £42.50
```

### Percentage Formatting

```go
mf := messageformat.MustNew("en", 
    "Progress: {$progress :number style=percent}")

values := []float64{0.0, 0.25, 0.5, 0.75, 1.0}

for _, value := range values {
    result, _ := mf.Format(map[string]interface{}{
        "progress": value,
    })
    fmt.Printf("%.2f: %s\n", value, result)
}
// Output:
// 0.00: Progress: 0%
// 0.25: Progress: 25%
// 0.50: Progress: 50%
// 0.75: Progress: 75%
// 1.00: Progress: 100%
```

### Decimal Formatting with Options

```go
// Control decimal places
precise := messageformat.MustNew("en", 
    "Value: {$value :number minimumFractionDigits=2 maximumFractionDigits=4}")

values := []float64{1, 1.5, 1.234, 1.23456789}

for _, value := range values {
    result, _ := precise.Format(map[string]interface{}{
        "value": value,
    })
    fmt.Printf("%.6f: %s\n", value, result)
}
// Output:
// 1.000000: Value: 1.00
// 1.500000: Value: 1.50
// 1.234000: Value: 1.234
// 1.234568: Value: 1.2346
```

## 🛡️ Error Handling

### Graceful Degradation

MessageFormat Go handles errors gracefully:

```go
mf := messageformat.MustNew("en", "Hello {$name}! You have {$count :number} items.")

// Missing variables - uses fallback representations
result, err := mf.Format(map[string]interface{}{
    "name": "Alice",
    // "count" is missing
})

fmt.Printf("Result: %s\n", result) // Hello ⁨Alice⁩! You have ⁨{$count}⁩ items.
fmt.Printf("Error: %v\n", err)     // <nil> (no error returned)
```

### Error Callbacks

Capture warnings without stopping execution:

```go
var warnings []error
onError := func(err error) {
    warnings = append(warnings, err)
}

mf := messageformat.MustNew("en", "Hello {$name} and {$missing}!")

result, err := mf.Format(map[string]interface{}{
    "name": "Alice",
    // "missing" not provided
}, onError)

fmt.Printf("Result: %s\n", result) // Hello ⁨Alice⁩ and ⁨{$missing}⁩!
fmt.Printf("Error: %v\n", err)     // <nil>
fmt.Printf("Warnings: %d\n", len(warnings)) // 1

for _, warning := range warnings {
    fmt.Printf("Warning: %v\n", warning) // undefined variable: missing
}
```

### Construction Errors

Handle syntax errors during message creation:

```go
// Invalid syntax
_, err := messageformat.New("en", "Invalid {$name")
if err != nil {
    fmt.Printf("Syntax error: %v\n", err)
    // Output: Syntax error: expected '}', found EOF
}

// Invalid configuration
_, err = messageformat.New("en", "Hello!", 
    &messageformat.MessageFormatOptions{
        Dir: "invalid-direction",
    })
if err != nil {
    fmt.Printf("Config error: %v\n", err)
    // Output: Config error: dir must be one of: ltr, rtl, auto
}
```

## 🎯 Next Steps

Continue your MessageFormat journey:

### 1. **[Message Syntax](message-syntax.md)**
Syntax reference for complex patterns and markup

### 2. **[API Reference](api-reference.md)**
Detailed methods, options, and configuration guide

### 3. **[Formatting Functions](formatting-functions.md)**
Built-in functions for numbers, dates, and text formatting

### 4. **[Custom Functions](custom-functions.md)**
Create your own formatting functions

### 5. **[Error Handling](error-handling.md)**
Error handling strategies

## 🏗️ Common Patterns

### Global Message Templates

```go
package main

import (
    "github.com/kaptinlin/messageformat-go"
)

// Define reusable message templates
var (
    WelcomeMsg = messageformat.MustNew("en", "Welcome, {$name}!")
    ErrorMsg   = messageformat.MustNew("en", "Error: {$message}")
    CountMsg   = messageformat.MustNew("en", `
.input {$count :number}
.match $count
0   {{No items}}
one {{One item}}
*   {{{$count} items}}
`)
)

func main() {
    // Use anywhere in your app
    result, _ := WelcomeMsg.Format(map[string]interface{}{
        "name": "Alice",
    })
    fmt.Println(result) // Welcome, ⁨Alice⁩!
}
```

### Simple Message Service

```go
type Messages struct {
    locale string
}

func NewMessages(locale string) *Messages {
    return &Messages{locale: locale}
}

func (m *Messages) Welcome(name string) string {
    mf := messageformat.MustNew(m.locale, "Welcome, {$name}!")
    result, _ := mf.Format(map[string]interface{}{"name": name})
    return result
}

func (m *Messages) ItemCount(count int) string {
    mf := messageformat.MustNew(m.locale, `
.input {$count :number}
.match $count
0   {{No items}}
one {{One item}}
*   {{{$count} items}}
`)
    result, _ := mf.Format(map[string]interface{}{"count": count})
    return result
}

// Usage
messages := NewMessages("en")
fmt.Println(messages.Welcome("Bob"))      // Welcome, ⁨Bob⁩!
fmt.Println(messages.ItemCount(5))        // 5 items
```

### Web Handler Example

```go
func greetHandler(w http.ResponseWriter, r *http.Request) {
    name := r.URL.Query().Get("name")
    if name == "" {
        name = "Guest"
    }
    
    result, _ := WelcomeMsg.Format(map[string]interface{}{
        "name": name,
    })
    
    fmt.Fprint(w, result)
}
```

## Summary

You're now ready to build internationalized applications with MessageFormat Go! The library's graceful error handling, thread safety, and Unicode support make it suitable for production use.
