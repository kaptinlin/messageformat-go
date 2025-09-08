# V1 Performance Guide - ICU MessageFormat

This guide covers performance characteristics, optimization techniques, and benchmarking for MessageFormat Go v1.

## Performance Overview

MessageFormat Go v1 is optimized for high-throughput production environments with sub-microsecond message formatting performance.

### Key Metrics

| Operation | Typical Performance | Memory Usage |
|-----------|-------------------|--------------|
| Message Compilation | ~10-50μs | ~1-5KB per message |
| Simple Message Format | ~100-500ns | ~100-500B per format |
| Plural Message Format | ~200-800ns | ~200-1KB per format |
| Complex Message Format | ~500-2μs | ~500B-2KB per format |

## Architecture Optimizations

### 1. Two-Phase Processing

```go
// Phase 1: Compile once (expensive)
compiled, err := mf.Compile("Hello, {name}!")

// Phase 2: Format many times (fast)
for i := 0; i < 1000000; i++ {
    result, _ := compiled.Format(args) // Sub-microsecond
}
```

### 2. Memory Pooling

V1 uses object pooling for frequently allocated objects:

```go
// Internal optimization - automatic memory reuse
var stringBuilderPool = sync.Pool{
    New: func() interface{} {
        return &strings.Builder{}
    },
}
```

### 3. Fast-Path Optimizations

- **Simple messages**: Direct string concatenation
- **No placeholders**: Return constant string
- **Single placeholder**: Optimized single-substitution path

## Benchmarking

### Running Benchmarks

```bash
cd v1/
go test -bench=. -benchmem
```

### Benchmark Results

```
BenchmarkSimpleMessage-8           10000000    120 ns/op      48 B/op    2 allocs/op
BenchmarkPluralMessage-8            2000000    450 ns/op     192 B/op    4 allocs/op
BenchmarkComplexMessage-8           1000000    800 ns/op     384 B/op    8 allocs/op
BenchmarkCompilation-8               100000  15000 ns/op    4096 B/op   20 allocs/op
```

### Performance Regression Testing

V1 includes automated performance regression tests:

```go
func TestPerformanceRegression(t *testing.T) {
    tests := []struct {
        name      string
        threshold time.Duration
        tolerance time.Duration
    }{
        {"simple", 120 * time.Nanosecond, 20 * time.Nanosecond},
        {"plural", 450 * time.Nanosecond, 100 * time.Nanosecond},
        {"complex", 800 * time.Nanosecond, 200 * time.Nanosecond},
    }
    
    // Test implementation...
}
```

## Optimization Best Practices

### 1. Compile Once, Use Many

```go
// ✅ Optimal: Compile once at startup
var welcomeMsg *messageformat.CompiledMessage

func init() {
    mf, _ := messageformat.New("en", nil)
    welcomeMsg, _ = mf.Compile("Welcome, {name}!")
}

func handleRequest(name string) string {
    return welcomeMsg.FormatToString(map[string]interface{}{
        "name": name,
    })
}

// ❌ Suboptimal: Recompiling every request
func handleRequestSlow(name string) string {
    mf, _ := messageformat.New("en", nil)
    compiled, _ := mf.Compile("Welcome, {name}!")
    return compiled.FormatToString(map[string]interface{}{
        "name": name,
    })
}
```

### 2. Reuse Argument Maps

```go
// ✅ Good: Reuse argument map
args := make(map[string]interface{})
for i := 0; i < 1000; i++ {
    args["name"] = names[i]
    result, _ := compiled.Format(args)
}

// ❌ Less optimal: New map each time
for i := 0; i < 1000; i++ {
    result, _ := compiled.Format(map[string]interface{}{
        "name": names[i],
    })
}
```

### 3. Prefer FormatToString for Simple Cases

```go
// ✅ Faster: No error handling overhead
result := compiled.FormatToString(args)

// ❌ Slower: Error handling overhead
result, err := compiled.Format(args)
if err != nil {
    // Handle error
}
```

### 4. Cache MessageFormat Instances

```go
var formatters = make(map[string]*messageformat.MessageFormat)
var formattersMutex sync.RWMutex

func getFormatter(locale string) *messageformat.MessageFormat {
    formattersMutex.RLock()
    if mf, exists := formatters[locale]; exists {
        formattersMutex.RUnlock()
        return mf
    }
    formattersMutex.RUnlock()
    
    formattersMutex.Lock()
    defer formattersMutex.Unlock()
    
    // Double-check after acquiring write lock
    if mf, exists := formatters[locale]; exists {
        return mf
    }
    
    mf, _ := messageformat.New(locale, nil)
    formatters[locale] = mf
    return mf
}
```

## Memory Optimization

### 1. Message Compilation Memory Usage

```go
// Typical memory usage per compiled message:
// - Simple message: ~1KB
// - Plural message: ~2-3KB  
// - Complex message: ~3-5KB
```

### 2. Runtime Memory Usage

```go
// Per Format() call:
// - Simple: ~100-500B
// - Plural: ~200-1KB
// - Complex: ~500B-2KB
```

### 3. Memory Leak Prevention

```go
// ✅ Good: Proper cleanup
compiled, err := mf.Compile(pattern)
if err != nil {
    return err
}
defer func() {
    // compiled will be GC'd automatically
    // No manual cleanup needed
}()

// Messages are automatically garbage collected
// when no longer referenced
```

## Concurrent Usage

### Thread Safety

```go
// MessageFormat instances are thread-safe after creation
mf, _ := messageformat.New("en", nil)
compiled, _ := mf.Compile("Hello, {name}!")

// Safe to use from multiple goroutines
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        result := compiled.FormatToString(map[string]interface{}{
            "name": fmt.Sprintf("User%d", id),
        })
        fmt.Println(result)
    }(i)
}
wg.Wait()
```

### Scalability Testing

```go
func BenchmarkConcurrentFormat(b *testing.B) {
    mf, _ := messageformat.New("en", nil)
    compiled, _ := mf.Compile("User {id} has {count} items")
    
    b.RunParallel(func(pb *testing.PB) {
        args := map[string]interface{}{
            "id":    1,
            "count": 5,
        }
        
        for pb.Next() {
            compiled.FormatToString(args)
        }
    })
}
```

## Production Deployment

### 1. Startup Optimization

```go
// Compile all messages at application startup
type MessageRegistry struct {
    messages map[string]*messageformat.CompiledMessage
}

func NewMessageRegistry() *MessageRegistry {
    registry := &MessageRegistry{
        messages: make(map[string]*messageformat.CompiledMessage),
    }
    
    // Precompile common messages
    mf, _ := messageformat.New("en", nil)
    
    patterns := map[string]string{
        "welcome":     "Welcome, {name}!",
        "item_count":  "You have {count} {count, plural, one {item} other {items}}",
        "last_login":  "Last login: {date, date, medium}",
    }
    
    for key, pattern := range patterns {
        compiled, _ := mf.Compile(pattern)
        registry.messages[key] = compiled
    }
    
    return registry
}
```

### 2. Monitoring and Metrics

```go
import (
    "time"
    "sync/atomic"
)

var (
    formatCalls    int64
    formatDuration int64
)

func (c *CompiledMessage) Format(args map[string]interface{}) (string, error) {
    start := time.Now()
    defer func() {
        atomic.AddInt64(&formatCalls, 1)
        atomic.AddInt64(&formatDuration, int64(time.Since(start)))
    }()
    
    // Actual formatting logic...
}

func GetFormatStats() (calls int64, avgDuration time.Duration) {
    c := atomic.LoadInt64(&formatCalls)
    d := atomic.LoadInt64(&formatDuration)
    
    if c == 0 {
        return 0, 0
    }
    
    return c, time.Duration(d / c)
}
```

## Troubleshooting Performance Issues

### 1. Identify Bottlenecks

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Memory profiling  
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

### 2. Common Performance Issues

| Issue | Symptom | Solution |
|-------|---------|----------|
| Recompiling messages | High CPU usage | Cache compiled messages |
| Memory leaks | Growing memory usage | Check for retained references |
| Lock contention | High wait times | Reduce shared state |
| GC pressure | Frequent GC pauses | Reduce allocations |

### 3. Performance Testing

```go
func TestMemoryUsage(t *testing.T) {
    var m1, m2 runtime.MemStats
    
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // Perform operations
    mf, _ := messageformat.New("en", nil)
    compiled, _ := mf.Compile("Hello, {name}!")
    
    for i := 0; i < 1000; i++ {
        compiled.Format(map[string]interface{}{
            "name": fmt.Sprintf("User%d", i),
        })
    }
    
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    allocatedMB := float64(m2.Alloc-m1.Alloc) / 1024 / 1024
    t.Logf("Memory allocated: %.2f MB", allocatedMB)
    
    if allocatedMB > 10 { // 10MB threshold
        t.Errorf("Excessive memory usage: %.2f MB", allocatedMB)
    }
}
```

## Version-Specific Optimizations

### V1 vs V2 Performance

| Feature | V1 (ICU MessageFormat) | V2 (MessageFormat 2.0) |
|---------|-------------------------|-------------------------|
| Simple messages | ~120ns | ~150ns |
| Plural messages | ~450ns | ~600ns |
| Complex messages | ~800ns | ~1200ns |
| Memory usage | Lower | Higher (more features) |
| Feature set | ICU standard | Unicode MessageFormat 2.0 |

V1 is optimized for maximum performance with the ICU MessageFormat specification, while V2 provides more advanced features at a slight performance cost.