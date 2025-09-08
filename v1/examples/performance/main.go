// Package main demonstrates MessageFormat performance optimization techniques.
//
// This example showcases production-ready performance patterns:
//   - Message compilation caching strategies
//   - Concurrent/thread-safe usage patterns
//   - Memory efficiency analysis and monitoring
//   - Throughput benchmarking and optimization
//
// Run this example with:
//
//	cd examples/performance && go run main.go
package main

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	mf "github.com/kaptinlin/messageformat-go/v1"
)

// MessageCache demonstrates compiled message caching for performance
type MessageCache struct {
	messageFormat *mf.MessageFormat
	cache         map[string]mf.MessageFunction
	mu            sync.RWMutex
}

// NewMessageCache creates a new thread-safe message cache
func NewMessageCache(locale string, options *mf.MessageFormatOptions) (*MessageCache, error) {
	messageFormat, err := mf.New(locale, options)
	if err != nil {
		return nil, err
	}

	return &MessageCache{
		messageFormat: messageFormat,
		cache:         make(map[string]mf.MessageFunction),
	}, nil
}

// GetCompiledMessage returns a compiled message, using cache when possible
func (mc *MessageCache) GetCompiledMessage(template string) (mf.MessageFunction, error) {
	// Try to get from cache first (read lock)
	mc.mu.RLock()
	if msg, exists := mc.cache[template]; exists {
		mc.mu.RUnlock()
		return msg, nil
	}
	mc.mu.RUnlock()

	// Not in cache, compile and store (write lock)
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Double-check in case another goroutine compiled it
	if msg, exists := mc.cache[template]; exists {
		return msg, nil
	}

	// Compile and cache
	msg, err := mc.messageFormat.Compile(template)
	if err != nil {
		return nil, err
	}

	mc.cache[template] = msg
	return msg, nil
}

// Format formats a message using the cached compiled version
func (mc *MessageCache) Format(template string, data map[string]interface{}) (string, error) {
	msg, err := mc.GetCompiledMessage(template)
	if err != nil {
		return "", err
	}

	result, err := msg(data)
	if err != nil {
		return "", err
	}

	return result.(string), nil
}

// CacheStats returns cache statistics
func (mc *MessageCache) CacheStats() (int, runtime.MemStats) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return len(mc.cache), memStats
}

func demonstrateBasicPerformance() {
	fmt.Println("=== Basic Performance Demonstration ===")

	// Create MessageFormat instance
	messageFormat, err := mf.New("en", nil)
	if err != nil {
		log.Fatal(err)
	}

	templates := []struct {
		name     string
		template string
		data     map[string]interface{}
	}{
		{
			"Simple",
			"Hello {name}!",
			map[string]interface{}{"name": "World"},
		},
		{
			"Plural",
			"{count, plural, one {# item} other {# items}}",
			map[string]interface{}{"count": 5},
		},
		{
			"Complex",
			"{gender, select, male {He has {count, plural, one {# item} other {# items}}} female {She has {count, plural, one {# item} other {# items}}} other {They have {count, plural, one {# item} other {# items}}}}",
			map[string]interface{}{"gender": "male", "count": 3},
		},
	}

	for _, test := range templates {
		fmt.Printf("\n%s Template Performance:\n", test.name)

		// Measure compilation time
		start := time.Now()
		msg, err := messageFormat.Compile(test.template)
		if err != nil {
			log.Printf("Compilation error: %v", err)
			continue
		}
		compilationTime := time.Since(start)

		// Measure execution time (single run)
		start = time.Now()
		result, err := msg(test.data)
		if err != nil {
			log.Printf("Execution error: %v", err)
			continue
		}
		executionTime := time.Since(start)

		fmt.Printf("  Template: %s\n", test.template)
		fmt.Printf("  Result: %s\n", result)
		fmt.Printf("  Compilation: %v\n", compilationTime)
		fmt.Printf("  Execution: %v\n", executionTime)

		// Measure throughput (multiple executions)
		iterations := 10000
		start = time.Now()
		for i := 0; i < iterations; i++ {
			_, err := msg(test.data)
			if err != nil {
				log.Printf("Throughput test error: %v", err)
				break
			}
		}
		totalTime := time.Since(start)
		throughput := float64(iterations) / totalTime.Seconds()

		fmt.Printf("  Throughput: %.0f ops/sec (%d iterations)\n", throughput, iterations)
	}
}

func demonstrateCaching() {
	fmt.Println("\n\n=== Caching Performance Demonstration ===")

	// Create cached message formatter
	cache, err := NewMessageCache("en", nil)
	if err != nil {
		log.Fatal(err)
	}

	// Create non-cached formatter for comparison
	messageFormat, err := mf.New("en", nil)
	if err != nil {
		log.Fatal(err)
	}

	templates := []string{
		"Hello {name}!",
		"You have {count, plural, one {# message} other {# messages}}",
		"{status, select, online {User is online} offline {User is offline} other {Status unknown}}",
		"Welcome {name}, today is {day}",
		"Order #{id} has {items, plural, one {# item} other {# items}}",
	}

	data := []map[string]interface{}{
		{"name": "Alice"},
		{"count": 5},
		{"status": "online"},
		{"name": "Bob", "day": "Monday"},
		{"id": "12345", "items": 3},
	}

	// Warm up the cache
	fmt.Println("Warming up cache...")
	for i, template := range templates {
		_, err := cache.Format(template, data[i])
		if err != nil {
			log.Printf("Cache warmup error: %v", err)
		}
	}

	iterations := 10000

	// Test cached performance
	fmt.Printf("\nTesting cached performance (%d iterations)...\n", iterations)
	start := time.Now()
	for i := 0; i < iterations; i++ {
		templateIdx := i % len(templates)
		_, err := cache.Format(templates[templateIdx], data[templateIdx])
		if err != nil {
			log.Printf("Cached execution error: %v", err)
			break
		}
	}
	cachedTime := time.Since(start)
	cachedThroughput := float64(iterations) / cachedTime.Seconds()

	// Test non-cached performance (compile every time)
	fmt.Printf("Testing non-cached performance (%d iterations)...\n", iterations)
	start = time.Now()
	for i := 0; i < iterations; i++ {
		templateIdx := i % len(templates)
		msg, err := messageFormat.Compile(templates[templateIdx])
		if err != nil {
			log.Printf("Non-cached compilation error: %v", err)
			break
		}
		_, err = msg(data[templateIdx])
		if err != nil {
			log.Printf("Non-cached execution error: %v", err)
			break
		}
	}
	nonCachedTime := time.Since(start)
	nonCachedThroughput := float64(iterations) / nonCachedTime.Seconds()

	// Report results
	fmt.Printf("\nResults:\n")
	fmt.Printf("  Cached:     %.2f ops/sec (%v total)\n", cachedThroughput, cachedTime)
	fmt.Printf("  Non-cached: %.2f ops/sec (%v total)\n", nonCachedThroughput, nonCachedTime)
	fmt.Printf("  Speedup:    %.1fx faster\n", cachedThroughput/nonCachedThroughput)

	// Cache statistics
	cacheSize, memStats := cache.CacheStats()
	fmt.Printf("  Cache size: %d templates\n", cacheSize)
	fmt.Printf("  Memory:     %.2f MB allocated\n", float64(memStats.Alloc)/1024/1024)
}

func demonstrateConcurrency() {
	fmt.Println("\n\n=== Concurrency Performance Demonstration ===")

	// Create cached message formatter
	cache, err := NewMessageCache("en", nil)
	if err != nil {
		log.Fatal(err)
	}

	template := "{user} sent {count, plural, one {# message} other {# messages}} to {recipient}"

	// Warm up cache
	_, err = cache.Format(template, map[string]interface{}{
		"user": "Alice", "count": 1, "recipient": "Bob",
	})
	if err != nil {
		log.Fatal(err)
	}

	goroutines := []int{1, 2, 4, 8, 16}
	iterations := 10000

	for _, numGoroutines := range goroutines {
		fmt.Printf("\nTesting with %d goroutines...\n", numGoroutines)

		var wg sync.WaitGroup
		iterationsPerGoroutine := iterations / numGoroutines

		start := time.Now()

		for g := 0; g < numGoroutines; g++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for i := 0; i < iterationsPerGoroutine; i++ {
					data := map[string]interface{}{
						"user":      fmt.Sprintf("User%d", goroutineID),
						"count":     i%10 + 1,
						"recipient": fmt.Sprintf("Recipient%d", i%5),
					}

					_, err := cache.Format(template, data)
					if err != nil {
						log.Printf("Concurrent execution error: %v", err)
						return
					}
				}
			}(g)
		}

		wg.Wait()
		duration := time.Since(start)

		totalOps := numGoroutines * iterationsPerGoroutine
		throughput := float64(totalOps) / duration.Seconds()

		fmt.Printf("  %d goroutines: %.0f ops/sec (%d total ops in %v)\n",
			numGoroutines, throughput, totalOps, duration)
	}
}

func demonstrateMemoryEfficiency() {
	fmt.Println("\n\n=== Memory Efficiency Demonstration ===")

	// Force garbage collection to get clean baseline
	runtime.GC()
	var startMem runtime.MemStats
	runtime.ReadMemStats(&startMem)

	fmt.Printf("Initial memory: %.2f MB\n", float64(startMem.Alloc)/1024/1024)

	// Create message cache and compile many templates
	cache, err := NewMessageCache("en", nil)
	if err != nil {
		log.Fatal(err)
	}

	// Generate many different templates to test memory usage
	numTemplates := 1000
	fmt.Printf("Compiling %d unique templates...\n", numTemplates)

	for i := 0; i < numTemplates; i++ {
		template := fmt.Sprintf("Message %d: {name} has {count, plural, one {# item} other {# items}} in category %d", i, i%10)
		data := map[string]interface{}{
			"name":  fmt.Sprintf("User%d", i),
			"count": i%5 + 1,
		}

		_, err := cache.Format(template, data)
		if err != nil {
			log.Printf("Template compilation error: %v", err)
		}

		// Show progress every 100 templates
		if (i+1)%100 == 0 {
			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)
			fmt.Printf("  Compiled %d templates: %.2f MB allocated\n",
				i+1, float64(mem.Alloc)/1024/1024)
		}
	}

	// Final memory statistics
	runtime.GC() // Force garbage collection
	var endMem runtime.MemStats
	runtime.ReadMemStats(&endMem)

	cacheSize, _ := cache.CacheStats()
	fmt.Printf("\nFinal statistics:\n")
	fmt.Printf("  Templates cached: %d\n", cacheSize)
	fmt.Printf("  Final memory: %.2f MB\n", float64(endMem.Alloc)/1024/1024)
	fmt.Printf("  Memory increase: %.2f MB\n", float64(endMem.Alloc-startMem.Alloc)/1024/1024)
	fmt.Printf("  Memory per template: %.2f KB\n", float64(endMem.Alloc-startMem.Alloc)/float64(numTemplates)/1024)
}

func main() {
	fmt.Println("=== MessageFormat Performance Examples ===")
	fmt.Printf("Go version: %s\nGOOS: %s\nGOARCH: %s\nCPUs: %d\n\n",
		runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.NumCPU())

	// Demonstrate basic performance characteristics
	demonstrateBasicPerformance()

	// Demonstrate caching benefits
	demonstrateCaching()

	// Demonstrate concurrent performance
	demonstrateConcurrency()

	// Demonstrate memory efficiency
	demonstrateMemoryEfficiency()

	fmt.Println("\n=== Performance examples completed successfully! ===")
}
