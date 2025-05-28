# Formatting Functions

MessageFormat 2.0 built-in functions for formatting different types of data. Documentation for all functions, options, and usage examples.

## 📖 Table of Contents

1. [Overview](#overview)
2. [Built-in Functions](#built-in-functions)
   - [`:string` Function](#string-function)
   - [`:number` Function](#number-function)
3. [Function Options](#function-options)
4. [Locale-Specific Behavior](#locale-specific-behavior)
5. [Advanced Usage](#advanced-usage)
6. [Custom Functions](#custom-functions)
7. [Error Handling](#error-handling)
8. [Best Practices](#best-practices)

## 🎯 Overview

MessageFormat Go provides a set of formatting functions that handle various data types and localization requirements. These functions are designed to work seamlessly with the MessageFormat 2.0 syntax and provide locale-aware formatting.

### Function Categories

| Category | Functions | Purpose |
|----------|-----------|---------|
| **Text** | `:string` | Text formatting and selection | 
| **Numeric** | `:number` | Number formatting and pluralization |

### Function Syntax

Functions in MessageFormat 2.0 use the following syntax:

```
{$variable :function option1=value1 option2=value2}
```

**Components:**
- `$variable`: The input value to format
- `:function`: The function name (prefixed with `:`)
- `option=value`: Function-specific options

## 🔧 Built-in Functions

### `:string` Function

The `:string` function handles text formatting and string selection operations.

**Purpose**: Text formatting and selection
**Input Types**: Any (converted to string)
**Output**: Formatted string

**Basic Usage:**

```go
mf := messageformat.MustNew("en", "Message: {$text :string}")
result, _ := mf.Format(map[string]interface{}{
    "text": "Hello, World!",
})
// Output: Message: ⁨Hello, World!⁩
```

**String Selection:**

```go
source := `
.input {$status :string}
.match $status
online  {{🟢 User is online}}
offline {{🔴 User is offline}}
*       {{❓ Unknown status}}
`

mf := messageformat.MustNew("en", source)
result, _ := mf.Format(map[string]interface{}{
    "status": "online",
})
// Output: 🟢 User is online
```

### `:number` Function

The `:number` function provides number formatting and pluralization capabilities.

**Purpose**: Number formatting and pluralization
**Input Types**: number, string (parseable as number)
**Output**: Formatted number with locale-specific formatting

**Basic Usage:**

```go
mf := messageformat.MustNew("en", "Count: {$count :number}")
result, _ := mf.Format(map[string]interface{}{
    "count": 1234.56,
})
// Output: Count: ⁨1,234.56⁩
```

**Pluralization:**

```go
source := `
.input {$count :number}
.match $count
0   {{No items}}
one {{One item}}
*   {{{$count} items}}
`

mf := messageformat.MustNew("en", source)
result, _ := mf.Format(map[string]interface{}{
    "count": 5,
})
// Output: ⁨5⁩ items
```

## 🔤 String Function

The `:string` function handles text formatting and provides selection capabilities.

### Basic Usage

```go
mf := messageformat.MustNew("en", "Message: {$text :string}")
result, _ := mf.Format(map[string]interface{}{
    "text": "Hello, World!",
})
// Output: Message: ⁨Hello, World!⁩
```

### String Selection

The string function can be used for selection based on string values:

```go
mf := messageformat.MustNew("en", `
.input {$status :string}
.match $status
active   {{User is currently active}}
inactive {{User is inactive}}
pending  {{User registration is pending}}
*        {{Unknown user status}}
`)

result, _ := mf.Format(map[string]interface{}{
    "status": "active",
})
// Output: User is currently active
```

### String Options

| Option | Type | Description | Example |
|--------|------|-------------|---------|
| `locale` | string | Override locale for formatting | `locale=en-US` |

**Examples:**

```go
// Locale override
mf := messageformat.MustNew("de", "Text: {$text :string locale=en}")
result, _ := mf.Format(map[string]interface{}{
    "text": "Hello",
})
// Uses English formatting rules despite German locale
```

## 🔢 Number Function

The `:number` function provides comprehensive number formatting and pluralization.

### Basic Number Formatting

```go
mf := messageformat.MustNew("en", "Count: {$num :number style=decimal}")
result, _ := mf.Format(map[string]interface{}{
    "num": 1234.56,
})
// Output: Count: ⁨1,234.56⁩
```

### Currency Formatting

```go
mf := messageformat.MustNew("en-US", "Price: {$amount :number style=currency currency=USD}")
result, _ := mf.Format(map[string]interface{}{
    "amount": 42.50,
})
// Output: Price: ⁨$42.50⁩

// Different locales
mf = messageformat.MustNew("de-DE", "Preis: {$amount :number style=currency currency=EUR}")
result, _ = mf.Format(map[string]interface{}{
    "amount": 42.50,
})
// Output: Preis: ⁨42,50 €⁩
```

### Percentage Formatting

```go
mf := messageformat.MustNew("en", "Progress: {$rate :number style=percent}")
result, _ := mf.Format(map[string]interface{}{
    "rate": 0.75,
})
// Output: Progress: ⁨75%⁩
```

### Number Pluralization

```go
mf := messageformat.MustNew("en", `
.input {$count :number}
.match $count
0   {{No items}}
one {{One item}}
*   {{{$count} items}}
`)

// Test different values
for _, count := range []int{0, 1, 5, 23} {
    result, _ := mf.Format(map[string]interface{}{"count": count})
    fmt.Printf("Count %d: %s\n", count, result)
}
// Output:
// Count 0: No items
// Count 1: One item
// Count 5: 5 items
// Count 23: 23 items
```

### Number Options

| Option | Type | Description | Default | Example |
|--------|------|-------------|---------|---------|
| `style` | string | Formatting style | `decimal` | `style=currency` |
| `currency` | string | Currency code (ISO 4217) | - | `currency=USD` |
| `currencyDisplay` | string | Currency display mode | `symbol` | `currencyDisplay=code` |
| `useGrouping` | boolean | Use grouping separators | `true` | `useGrouping=false` |
| `minimumIntegerDigits` | number | Minimum integer digits | `1` | `minimumIntegerDigits=3` |
| `minimumFractionDigits` | number | Minimum fraction digits | `0` | `minimumFractionDigits=2` |
| `maximumFractionDigits` | number | Maximum fraction digits | `3` | `maximumFractionDigits=2` |
| `minimumSignificantDigits` | number | Minimum significant digits | - | `minimumSignificantDigits=3` |
| `maximumSignificantDigits` | number | Maximum significant digits | - | `maximumSignificantDigits=5` |

**Style Values:**
- `decimal` - Standard number formatting
- `currency` - Currency formatting (requires `currency` option)
- `percent` - Percentage formatting
- `unit` - Unit formatting (future extension)

**Currency Display Values:**
- `symbol` - Currency symbol ($, €, ¥)
- `code` - Currency code (USD, EUR, JPY)
- `name` - Currency name (US Dollar, Euro, Japanese Yen)

**Examples:**

```go
// Detailed currency formatting
mf := messageformat.MustNew("en", `
Price: {$amount :number 
    style=currency 
    currency=USD 
    currencyDisplay=symbol 
    minimumFractionDigits=2 
    maximumFractionDigits=2}
`)

// No grouping
mf := messageformat.MustNew("en", "ID: {$id :number useGrouping=false}")

// Significant digits
mf := messageformat.MustNew("en", `
Value: {$num :number 
    minimumSignificantDigits=3 
    maximumSignificantDigits=5}
`)
```

## 📋 Function Options Reference

### Common Option Types

**Boolean Options:**
```go
// Boolean values
useGrouping=true
useGrouping=false
hour12=true
hour12=false
```

**String Options:**
```go
// String values (no quotes needed)
style=currency
currency=USD
timeZone=America/New_York
```

**Number Options:**
```go
// Numeric values
minimumFractionDigits=2
maximumFractionDigits=4
minimumIntegerDigits=3
```

### Option Validation

Invalid options are handled gracefully:

```go
// Invalid option - falls back to default
mf := messageformat.MustNew("en", "Price: {$amount :number style=invalid}")
result, _ := mf.Format(map[string]interface{}{
    "amount": 42.50,
})
// Uses default decimal style
```

## 🌍 Locale-Specific Behavior

Functions adapt their behavior based on the locale:

### Number Formatting by Locale

```go
locales := []string{"en-US", "de-DE", "fr-FR", "ja-JP", "ar-SA"}
amount := 1234567.89

for _, locale := range locales {
    mf := messageformat.MustNew(locale, "Amount: {$amount :number}")
    result, _ := mf.Format(map[string]interface{}{
        "amount": amount,
    })
    fmt.Printf("%s: %s\n", locale, result)
}
// Output:
// en-US: Amount: ⁨1,234,567.89⁩
// de-DE: Amount: ⁨1.234.567,89⁩
// fr-FR: Amount: ⁨1 234 567,89⁩
// ja-JP: Amount: ⁨1,234,567.89⁩
// ar-SA: Amount: ⁨١٬٢٣٤٬٥٦٧٫٨٩⁩
```

### Currency Formatting by Locale

```go
locales := []string{"en-US", "de-DE", "ja-JP"}
amount := 42.50

for _, locale := range locales {
    mf := messageformat.MustNew(locale, "Price: {$amount :number style=currency currency=USD}")
    result, _ := mf.Format(map[string]interface{}{
        "amount": amount,
    })
    fmt.Printf("%s: %s\n", locale, result)
}
// Output:
// en-US: Price: ⁨$42.50⁩
// de-DE: Price: ⁨42,50 $⁩
// ja-JP: Price: ⁨$43⁩ (rounded)
```

### Date Formatting by Locale

```go
locales := []string{"en-US", "de-DE", "ja-JP", "ar-SA"}
date := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)

for _, locale := range locales {
    mf := messageformat.MustNew(locale, "Date: {$date :date style=medium}")
    result, _ := mf.Format(map[string]interface{}{
        "date": date,
    })
    fmt.Printf("%s: %s\n", locale, result)
}
// Output:
// en-US: Date: ⁨Jun 15, 2024⁩
// de-DE: Date: ⁨15.06.2024⁩
// ja-JP: Date: ⁨2024/06/15⁩
// ar-SA: Date: ⁨١٥‏/٠٦‏/٢٠٢٤⁩
```

### Plural Rules by Locale

Different languages have different plural rules:

```go
// English: 0, 1, other
mfEn := messageformat.MustNew("en", `
.input {$count :number}
.match $count
0   {{No items}}
one {{One item}}
*   {{{$count} items}}
`)

// Polish: 0, 1, few (2-4), many (5+), other
mfPl := messageformat.MustNew("pl", `
.input {$count :number}
.match $count
0   {{Brak elementów}}
one {{Jeden element}}
few {{{$count} elementy}}
many {{{$count} elementów}}
*   {{{$count} elementów}}
`)

// Arabic: 0, 1, 2, few (3-10), many (11-99), other
mfAr := messageformat.MustNew("ar", `
.input {$count :number}
.match $count
0   {{لا توجد عناصر}}
one {{عنصر واحد}}
two {{عنصران}}
few {{{$count} عناصر}}
many {{{$count} عنصراً}}
*   {{{$count} عنصر}}
`)
```

## 🎨 Advanced Examples

### E-commerce Price Display

```go
mf := messageformat.MustNew("en", `
.input {$originalPrice :number style=currency currency=USD}
.input {$salePrice :number style=currency currency=USD}
.input {$discount :number style=percent}
.match $salePrice
0 {{Free! (was {$originalPrice})}}
* {{
    Sale Price: {$salePrice} 
    (was {$originalPrice}, save {$discount})
}}
`)

result, _ := mf.Format(map[string]interface{}{
    "originalPrice": 99.99,
    "salePrice":     79.99,
    "discount":      0.20,
})
// Output: Sale Price: ⁨$79.99⁩ (was ⁨$99.99⁩, save ⁨20%⁩)
```

### Multi-timezone Event Display

```go
mf := messageformat.MustNew("en", `
Conference Call:
• New York: {$date :datetime dateStyle=short timeStyle=short timeZone=America/New_York}
• London: {$date :datetime dateStyle=short timeStyle=short timeZone=Europe/London}
• Tokyo: {$date :datetime dateStyle=short timeStyle=short timeZone=Asia/Tokyo}
`)

result, _ := mf.Format(map[string]interface{}{
    "date": time.Date(2024, 6, 15, 18, 0, 0, 0, time.UTC),
})
// Output: Conference Call:
//         • New York: ⁨6/15/24, 2:00 PM⁩
//         • London: ⁨15/06/2024, 19:00⁩
//         • Tokyo: ⁨2024/06/16, 3:00⁩
```

### Financial Dashboard

```go
mf := messageformat.MustNew("en", `
.input {$balance :number style=currency currency=USD}
.input {$change :number style=currency currency=USD}
.input {$changePercent :number style=percent minimumFractionDigits=2}
.input {$trend :string}
.match $trend
positive {{
    Account Balance: {$balance}
    ↗ Up {$change} ({$changePercent}) today
}}
negative {{
    Account Balance: {$balance}
    ↘ Down {$change} ({$changePercent}) today
}}
* {{
    Account Balance: {$balance}
    → No change today
}}
`)

result, _ := mf.Format(map[string]interface{}{
    "balance":       15420.75,
    "change":        234.50,
    "changePercent": 0.0154,
    "trend":         "positive",
})
// Output: Account Balance: ⁨$15,420.75⁩
//         ↗ Up ⁨$234.50⁩ (⁨1.54%⁩) today
```

### Multilingual File Manager

```go
// English
enMf := messageformat.MustNew("en", `
.input {$fileCount :number}
.input {$totalSize :number}
.match $fileCount
0 {{No files selected}}
one {{Selected {$fileCount} file ({$totalSize} bytes)}}
* {{Selected {$fileCount} files ({$totalSize} bytes total)}}
`)

// German
deMf := messageformat.MustNew("de", `
.input {$fileCount :number}
.input {$totalSize :number useGrouping=true}
.match $fileCount
0 {{Keine Dateien ausgewählt}}
one {{Eine Datei ausgewählt ({$totalSize} Bytes)}}
* {{{$fileCount} Dateien ausgewählt ({$totalSize} Bytes insgesamt)}}
`)

// Japanese
jaMf := messageformat.MustNew("ja", `
.input {$fileCount :number}
.input {$totalSize :number}
.match $fileCount
0 {{ファイルが選択されていません}}
* {{選択されたファイル: {$fileCount}個 (合計 {$totalSize} バイト)}}
`)

variables := map[string]interface{}{
    "fileCount": 3,
    "totalSize": 1048576,
}

for name, mf := range map[string]*messageformat.MessageFormat{
    "English": enMf,
    "German":  deMf,
    "Japanese": jaMf,
} {
    result, _ := mf.Format(variables)
    fmt.Printf("%s: %s\n", name, result)
}
// Output:
// English: Selected ⁨3⁩ files (⁨1,048,576⁩ bytes total)
// German: ⁨3⁩ Dateien ausgewählt (⁨1.048.576⁩ Bytes insgesamt)
// Japanese: 選択されたファイル: ⁨3⁩個 (合計 ⁨1,048,576⁩ バイト)
```

## 💡 Best Practices

### Function Selection

```go
// ✅ Good: Use appropriate function for data type
mf := messageformat.MustNew("en", `
User ID: {$id :number}
Balance: {$amount :number style=currency currency=USD}
Status: {$status :string}
`)

// ❌ Avoid: Using wrong function for data type
mf := messageformat.MustNew("en", `
User ID: {$id :string}  // Should use :number for numeric IDs
Balance: {$amount :string}  // Should use :number
`)
```

### Option Usage

```go
// ✅ Good: Specify relevant options
mf := messageformat.MustNew("en", `
Price: {$amount :number style=currency currency=USD minimumFractionDigits=2}
`)

// ✅ Good: Use locale-appropriate defaults
mf := messageformat.MustNew("de", `
Preis: {$amount :number style=currency currency=EUR}
`)

// ❌ Avoid: Unnecessary options
mf := messageformat.MustNew("en", `
Count: {$num :number style=decimal useGrouping=true}  // style=decimal is default
`)
```

## Summary

This guide covers all built-in formatting functions in MessageFormat Go. For information about creating custom functions, see the [Custom Functions](custom-functions.md) guide. 
