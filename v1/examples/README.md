# MessageFormat v1 Go Examples

This directory contains comprehensive examples demonstrating how to use the [MessageFormat v1 Go library](https://github.com/kaptinlin/messageformat-go/v1).

MessageFormat v1 is the traditional ICU MessageFormat specification, providing reliable internationalization support with pluralization, gender selection, custom formatting functions, and proven production stability.

## 🚀 Quick Start

```bash
# Clone the repository
git clone https://github.com/kaptinlin/messageformat-go
cd messageformat-go/v1

# Install dependencies
go mod tidy

# Run basic examples - learn fundamentals
go run examples/basic/main.go

# Run e-commerce examples - see real-world patterns
go run examples/ecommerce/main.go

# Run multilingual examples - explore CLDR locale support
go run examples/multilingual/main.go

# Run performance examples - optimization techniques
go run examples/performance/main.go
```

### Expected Output Examples

**Basic Example:**
```
=== Basic MessageFormat Examples ===
1. Simple Variable Interpolation:
   Template: Hello, {name}!
   Data: {"name": "Alice"}
   Result: Hello, Alice!
```

**E-commerce Example:**
```
1. Shopping Cart Messages:
   Items: 0, Gender: male -> He has no items in his cart
   Items: 1, Gender: female -> She has 1 item in her cart
```

## 📁 Project Structure

```
v1/examples/
├── basic/              # Basic usage patterns
├── ecommerce/          # Real-world e-commerce scenarios
├── multilingual/       # Multi-locale and CLDR support
├── performance/        # Performance optimization techniques
└── README.md          # This file
```

## 📖 Example Categories

### 1. Basic Usage (`basic/`)
Demonstrates the most fundamental ICU MessageFormat v1 features:
- Simple variable interpolation (`{name}`)
- Multiple variable substitution
- Basic pluralization rules with `plural` formatter
- Select statements with `select` formatter
- Understanding ICU syntax patterns

### 2. E-commerce Application (`ecommerce/`)
Shows real-world e-commerce notification patterns:
- Shopping cart status messages with gender selection
- Order status notifications with complex nested logic
- Inventory alerts with stock level conditions
- Service-oriented architecture with NotificationService
- Production-ready error handling and validation

### 3. Multilingual Support (`multilingual/`)
Demonstrates comprehensive internationalization features:
- Multiple locale support (English, Russian, Arabic, Welsh, Chinese)
- Complex CLDR plural rules demonstration across different language families
- Locale-specific formatting and currency handling (USD, EUR, CNY, RUB)
- Return type variations (string vs values array)
- LocalizedMessage management patterns with currency mapping
- Fallback and error handling for unsupported locales
- CJK text handling and simplified plural rules

### 4. Performance Optimization (`performance/`)
Covers production-ready performance techniques:
- Message compilation caching with MessageCache
- Concurrent/thread-safe usage patterns
- Memory efficiency analysis and monitoring
- Throughput benchmarking and performance comparison
- Object pooling and optimization strategies
- Caching speedup demonstrations (10-50x improvement)

## 🌍 Languages Demonstrated

The examples showcase internationalization with:
- **English (en)** - Primary language for all examples
- **Russian (ru)** - Complex plural rules demonstration
- **Arabic (ar)** - RTL language and complex plurals
- **Welsh (cy)** - Special plural category handling
- **Chinese Simplified (zh-CN)** - Simple plural rules and CJK text handling
- **French (fr)** - Romance language patterns
- **Spanish (es)** - Additional Romance language support

All examples demonstrate proper CLDR locale handling and fallback mechanisms.

## 🔧 Technical Features Demonstrated

- ✅ **ICU MessageFormat v1**: Complete traditional ICU specification support
- ✅ **Variable Substitution**: `{name}` syntax with type safety
- ✅ **Pluralization**: `{count, plural, one {# item} other {# items}}` patterns
- ✅ **Gender Selection**: `{gender, select, male {He} female {She} other {They}}`
- ✅ **Custom Formatters**: Registration with CustomFormatters map
- ✅ **CLDR Locale Support**: Full Unicode locale data integration
- ✅ **Performance Optimization**: Caching and fast-path optimizations
- ✅ **Error Handling**: Graceful fallbacks and validation patterns
- ✅ **Thread Safety**: Safe concurrent usage after compilation
- ✅ **Return Types**: String vs values array output options
- ✅ **TypeScript Compatibility**: 100% API compatibility with JS/TS implementation

## 📚 Learning Resources

- [ICU MessageFormat Specification](https://unicode-org.github.io/icu/userguide/format_parse/messages/)
- [Go Library Documentation](https://github.com/kaptinlin/messageformat-go/v1)
- [CLDR Plural Rules](https://cldr.unicode.org/index/cldr-spec/plural-rules)

## 🎯 Integration Patterns

### Service-Oriented Architecture
The e-commerce example demonstrates MessageFormat integration:

```go
type NotificationService struct {
    messageFormat *mf.MessageFormat
}

func (ns *NotificationService) FormatMessage(template string, data map[string]interface{}) (string, error) {
    msg, err := ns.messageFormat.Compile(template)
    if err != nil {
        return "", err
    }
    
    result, err := msg(data)
    if err != nil {
        return "", err
    }
    
    return result.(string), nil
}
```

### Performance Caching
Production-ready caching implementation:

```go
type MessageCache struct {
    messageFormat *mf.MessageFormat
    cache         map[string]mf.MessageFunction
    mu            sync.RWMutex
}
```

### Multilingual Support
Locale management patterns:

```go
type LocalizedMessage struct {
    formatters map[string]*mf.MessageFormat
    templates  map[string]string
}
```

## ⚡ Performance Expectations

Based on the performance examples, you can expect:
- **Simple messages**: ~72ns per operation  
- **Plural messages**: ~180ns per operation
- **Complex nested**: ~500ns per operation
- **Caching speedup**: 10-50x improvement over repeated compilation
- **Concurrency**: Linear scaling with goroutines
- **Memory efficiency**: Object pooling reduces GC pressure

## 🔧 Common Patterns

### Error Handling
All examples follow Go best practices:

```go
result, err := msg(data)
if err != nil {
    return fmt.Errorf("message formatting failed: %w", err)
}
```

### Template Organization
Maintainable template constants:

```go
const (
    CartEmptyTemplate = "{user} has no items in the cart"
    CartItemsTemplate = "{user} has {count, plural, one {# item} other {# items}}"
)
```

### Configuration Options
Structured configuration approach:

```go
config := &mf.MessageFormatOptions{
    Currency:            "USD",
    ReturnType:          mf.ReturnTypeString,
    RequireAllArguments: true,
}
```

## 🚦 Running All Examples

Execute all examples sequentially:

```bash
# From the v1/examples directory
for dir in basic ecommerce multilingual performance; do
    echo "=== Running $dir example ==="
    (cd $dir && go run main.go)
    echo
done
```

## 🧪 Testing Your Integration

Each example demonstrates:
- ✅ Input validation and sanitization
- ✅ Graceful error recovery patterns
- ✅ Performance monitoring techniques
- ✅ Thread-safe concurrent usage
- ✅ Memory-efficient resource management

Use these patterns as production-ready starting points for your MessageFormat integration.

## 🔗 Next Steps

After exploring these examples:
1. **Review the main v1 README** for complete API documentation
2. **Study the test files** for additional usage patterns and edge cases
3. **Consider your i18n requirements** and choose appropriate patterns
4. **Implement caching** for production deployments with high throughput
5. **Explore v2 features** if you need MessageFormat 2.0 specification support

---

**Questions?** Check the [main repository documentation](https://github.com/kaptinlin/messageformat-go) or [open an issue](https://github.com/kaptinlin/messageformat-go/issues) for support.