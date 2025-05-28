# Message Syntax

MessageFormat 2.0 syntax reference as defined in the [Unicode MessageFormat 2.0 specification](https://unicode.org/reports/tr35/tr35-messageFormat.html). Covers basic variable substitution to complex selection logic.

## 📖 Table of Contents

1. [Message Structure](#message-structure)
2. [Simple Messages](#simple-messages)
3. [Complex Messages](#complex-messages)
4. [Declarations](#declarations)
5. [Expressions](#expressions)
6. [Functions and Options](#functions-and-options)
7. [Pattern Matching](#pattern-matching)
8. [Escape Sequences](#escape-sequences)
9. [Bidirectional Text](#bidirectional-text)
10. [Advanced Examples](#advanced-examples)

## 🔧 Message Structure

MessageFormat 2.0 supports two fundamental message types:

### Simple Messages
Direct text with optional variable substitution and formatting:
```
Hello, {$name}!
```

### Complex Messages  
Messages with declarations and selection logic:
```
.input {$count :number}
.match $count
one {{You have one item}}
*   {{You have {$count} items}}
```

## 🔤 Simple Messages

Simple messages are the most straightforward form of MessageFormat. They consist of literal text with optional placeholders.

### Literal Text

Plain text is rendered as-is:

```go
mf, _ := messageformat.New("en", "Welcome to our application!")
result, _ := mf.Format(nil)
// Output: Welcome to our application!
```

### Variable Substitution

Variables are enclosed in curly braces and prefixed with `$`:

```go
mf, _ := messageformat.New("en", "Hello, {$name}!")
result, _ := mf.Format(map[string]interface{}{
    "name": "Alice",
})
// Output: Hello, ⁨Alice⁩!
```

### Multiple Variables

You can use multiple variables in a single message:

```go
mf, _ := messageformat.New("en", "Welcome {$firstName} {$lastName} to {$siteName}!")
result, _ := mf.Format(map[string]interface{}{
    "firstName": "John",
    "lastName":  "Doe", 
    "siteName":  "MessageFormat Go",
})
// Output: Welcome ⁨John⁩ ⁨Doe⁩ to ⁨MessageFormat Go⁩!
```

### Function Calls

Variables can be processed by functions:

```go
mf, _ := messageformat.New("en", "Today is {$date :datetime dateStyle=full}")
result, _ := mf.Format(map[string]interface{}{
    "date": time.Now(),
})
// Output: Today is ⁨Monday, January 15, 2024⁩
```

## 🔀 Complex Messages

Complex messages use declarations and selection logic to handle conditional formatting and pluralization.

### Basic Structure

Complex messages follow this pattern:
```
.declaration1
.declaration2
.match $selector1 $selector2
key1 key2 {{pattern1}}
key3 key4 {{pattern2}}
*    *    {{default pattern}}
```

### Simple Pluralization

```go
mf, _ := messageformat.New("en", `
.input {$count :number select=cardinal}
.match $count
0   {{No messages}}
one {{One message}}
*   {{{$count} messages}}
`)

// Test different counts
for _, count := range []int{0, 1, 5} {
    result, _ := mf.Format(map[string]interface{}{"count": count})
    fmt.Printf("Count %d: %s\n", count, result)
}
// Output:
// Count 0: No messages
// Count 1: One message
// Count 5: 5 messages
```

### Multi-dimensional Selection

```go
mf, _ := messageformat.New("en", `
.input {$gender :string}
.input {$count :number select=cardinal}
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
    "name":   "Sarah",
    "gender": "female",
    "count":  3,
})
// Output: Sarah has 3 items in her cart
```

## 📋 Declarations

Declarations preprocess variables before they're used in patterns or selectors.

### Input Declarations

Input declarations process external variables:

```go
// Basic input declaration
.input {$count :number}

// Input declaration with options
.input {$price :number style=currency currency=USD}

// Input declaration with multiple options
.input {$date :datetime dateStyle=full timeStyle=short}
```

Example usage:
```go
mf, _ := messageformat.New("en", `
.input {$amount :number style=currency currency=EUR}
Your balance is {$amount}
`)

result, _ := mf.Format(map[string]interface{}{
    "amount": 1234.56,
})
// Output: Your balance is €1,234.56
```

### Local Declarations

Local declarations create computed variables:

```go
mf, _ := messageformat.New("en", `
.input {$price :number}
.local $tax = {$price :number style=percent}
.local $total = {$price :number style=currency currency=USD}
Item: {$total} (includes {$tax} tax)
`)

result, _ := mf.Format(map[string]interface{}{
    "price": 100.00,
})
// Output: Item: $100.00 (includes 100% tax)
```

### Declaration Order

Declarations are processed in order, and later declarations can reference earlier ones:

```go
mf, _ := messageformat.New("en", `
.input {$basePrice :number}
.local $tax = {$basePrice :number}
.local $total = {$basePrice :number}
Base: {$basePrice}, Tax: {$tax}, Total: {$total}
`)
```

## 🔧 Expressions

Expressions are the building blocks of MessageFormat messages. They can appear in patterns, declarations, and as selectors.

### Variable References

Reference external or local variables:
```
{$variableName}
{$user}
{$count}
```

### Literal Values

Embed literal values directly:
```
{|literal text|}
{|42|}
{|true|}
```

### Function Calls

Apply functions to operands:
```
{$count :number}
{$date :datetime dateStyle=short}
{|Hello| :string}
```

### Function Calls with Options

Functions can accept multiple options:
```
{$price :number style=currency currency=USD minimumFractionDigits=2}
{$date :datetime dateStyle=full timeStyle=medium}
```

### Annotation-only Expressions

Functions without operands:
```
{:number}  // Uses default value
{:datetime}  // Uses current time
```

## ⚙️ Functions and Options

Functions transform and format values. MessageFormat 2.0 includes several built-in functions.

#### `:number` Function

The `:number` function formats numeric values and supports pluralization:

```go
mf, _ := messageformat.New("en", "Price: {$price :number style=currency currency=USD}")
mf, _ := messageformat.New("en", "Progress: {$rate :number style=percent}")
```

**Pattern Matching with Numbers:**
```go
source := `.match {$count :number}
0   {{No items}}
one {{One item}}
*   {{{$count} items}}`
```

#### `:string` Function

The `:string` function handles string values and selection:

```go
source := `.match {$status :string}
online  {{User is online}}
offline {{User is offline}}
*       {{Status unknown}}`
```

**Note**: According to the TC39 Intl.MessageFormat proposal, only `:number` and `:string` are standard functions. Other functions like `:datetime`, `:date`, `:time`, `:integer` are not part of the official specification.

## 🎯 Pattern Matching

Pattern matching allows messages to vary based on input values:

```go
source := `
.input {$count :number}
.match $count
0   {{You have no new notifications}}
one {{You have {$count} new notification}}
*   {{You have {$count} new notifications}}
`

mf, _ := messageformat.New("en", source)
result, _ := mf.Format(map[string]interface{}{
    "count": 1,
})
// Output: You have ⁨1⁩ new notification
```

### Complex Pattern Matching

```go
source := `
.input {$count :number}
.match $count
0   {{Your inbox is empty}}
one {{You have one unread message}}
*   {{You have {$count} unread messages}}
`

mf, _ := messageformat.New("en", source)
```

## 🔤 Escape Sequences

Special characters can be escaped when needed as literal text.

### Escaping Braces

Use `{{` and `}}` to include literal braces:
```go
mf, _ := messageformat.New("en", "Code: {{function() { return {$value}; }}")
// Output: Code: {function() { return ⁨42⁩; }
```

### Escaping in Patterns

Within pattern text, use backslash escaping:
```go
mf, _ := messageformat.New("en", `
.input {$count :number}
.match $count
one {{You have \{one\} item}}
*   {{You have \{{$count}\} items}}
`)
```

### Reserved Characters

These characters have special meaning and may need escaping:
- `{` and `}` - Expression delimiters
- `$` - Variable prefix
- `:` - Function prefix
- `|` - Literal delimiter
- `*` - Wildcard selector
- `.` - Declaration prefix

### Whitespace Handling

Leading and trailing whitespace in patterns is preserved:
```go
mf, _ := messageformat.New("en", `
.input {$count :number}
.match $count
one {{ You have one item }}
*   {{ You have {$count} items }}
`)
// Note the spaces around the text
```

## 🔄 Bidirectional Text

MessageFormat 2.0 provides support for bidirectional text.

### Automatic Isolation

By default, MessageFormat isolates variables to prevent text direction spillover:

```go
// Arabic text with English variables
mf, _ := messageformat.New("ar", "مرحبا {$name}!")
result, _ := mf.Format(map[string]interface{}{
    "name": "John",
})
// Output includes proper bidi isolation characters
```

### Bidi Isolation Options

Control bidirectional text handling:

```go
// Disable bidi isolation
mf, _ := messageformat.New("en", "Hello {$name}!", 
    messageformat.WithBidiIsolation("none"))

// Use compatibility mode (default)
mf, _ := messageformat.New("en", "Hello {$name}!", 
    messageformat.WithBidiIsolation("compatibility"))
```

### Text Direction

Specify text direction explicitly:

```go
// Left-to-right
mf, _ := messageformat.New("en", "Hello {$name}!", 
    messageformat.WithDir("ltr"))

// Right-to-left
mf, _ := messageformat.New("ar", "مرحبا {$name}!", 
    messageformat.WithDir("rtl"))

// Auto-detect
mf, _ := messageformat.New("en", "Hello {$name}!", 
    messageformat.WithDir("auto"))
```

## 🎨 Advanced Examples

### E-commerce Product Catalog

```go
mf, _ := messageformat.New("en", `
.input {$itemCount :number}
.input {$totalPrice :number style=currency currency=USD}
.input {$shippingCost :number style=currency currency=USD}
.match $itemCount
0 {{
    Your cart is empty. Start shopping to see items here!
}}
one {{
    You have {$itemCount} item in your cart.
    Subtotal: {$totalPrice}
    Shipping: {$shippingCost}
    Total: {$totalPrice :number style=currency currency=USD}
}}
* {{
    You have {$itemCount} items in your cart.
    Subtotal: {$totalPrice}
    Shipping: {$shippingCost}
    Total: {$totalPrice :number style=currency currency=USD}
}}
`)
```

### Social Media Notifications

```go
mf, _ := messageformat.New("en", `
.input {$actorGender :string}
.input {$actionType :string}
.input {$objectCount :number}
.match $actionType $actorGender $objectCount
like male 1 {{{$actor} liked your post}}
like female 1 {{{$actor} liked your post}}
like * 1 {{{$actor} liked your post}}
like male * {{{$actor} liked {$objectCount} of your posts}}
like female * {{{$actor} liked {$objectCount} of your posts}}
like * * {{{$actor} liked {$objectCount} of your posts}}
comment male 1 {{{$actor} commented on your post}}
comment female 1 {{{$actor} commented on your post}}
comment * 1 {{{$actor} commented on your post}}
comment male * {{{$actor} commented on {$objectCount} of your posts}}
comment female * {{{$actor} commented on {$objectCount} of your posts}}
comment * * {{{$actor} commented on {$objectCount} of your posts}}
* * * {{{$actor} performed an action}}
`)
```

### Financial Dashboard

```go
mf, _ := messageformat.New("en", `
.input {$accountType :string}
.input {$balance :number style=currency currency=USD}
.input {$changePercent :number style=percent}
.input {$trend :string}
.match $accountType $trend
checking positive {{
    Checking Account: {$balance}
    ↗ Up {$changePercent} this month
}}
checking negative {{
    Checking Account: {$balance}
    ↘ Down {$changePercent} this month
}}
savings positive {{
    Savings Account: {$balance}
    ↗ Gained {$changePercent} this month
}}
savings negative {{
    Savings Account: {$balance}
    ↘ Lost {$changePercent} this month
}}
investment positive {{
    Investment Portfolio: {$balance}
    ↗ Gained {$changePercent} this month
}}
investment negative {{
    Investment Portfolio: {$balance}
    ↘ Lost {$changePercent} this month
}}
* * {{
    Account Balance: {$balance}
    Change: {$changePercent}
}}
`)
```

### Multilingual Support

```go
// English
enMessage := `
.input {$fileCount :number}
.input {$totalSize :number}
.match $fileCount
0 {{No files selected}}
one {{Selected {$fileCount} file ({$totalSize} bytes)}}
* {{Selected {$fileCount} files ({$totalSize} bytes total)}}
`

// Spanish
esMessage := `
.input {$fileCount :number}
.input {$totalSize :number}
.match $fileCount
0 {{Ningún archivo seleccionado}}
one {{Seleccionado {$fileCount} archivo ({$totalSize} bytes)}}
* {{Seleccionados {$fileCount} archivos ({$totalSize} bytes en total)}}
`

// French
frMessage := `
.input {$fileCount :number}
.input {$totalSize :number}
.match $fileCount
0 {{Aucun fichier sélectionné}}
one {{Fichier sélectionné: {$fileCount} ({$totalSize} octets)}}
* {{Fichiers sélectionnés: {$fileCount} ({$totalSize} octets au total)}}
`

// Usage
messages := map[string]string{
    "en": enMessage,
    "es": esMessage,
    "fr": frMessage,
}

for locale, message := range messages {
    mf, _ := messageformat.New(locale, message)
    result, _ := mf.Format(map[string]interface{}{
        "fileCount": 3,
        "totalSize": 1024000,
    })
    fmt.Printf("%s: %s\n", locale, result)
}
```

## Summary

This syntax guide should help you master MessageFormat 2.0. For more specific information about functions and error handling, see the other documentation sections. 
