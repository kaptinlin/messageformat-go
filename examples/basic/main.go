package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/messageformat-go"
)

func main() {
	fmt.Println("=== MessageFormat 2.0 Basic Usage Examples ===")

	// Example 1: Simple variable substitution
	fmt.Println("1. Simple Variable Substitution:")
	mf1, err := messageformat.New("en", "Hello, {$name}!")
	if err != nil {
		log.Fatal(err)
	}

	result1, err := mf1.Format(map[string]interface{}{
		"name": "World",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Input: \"Hello, {$name}!\"\n")
	fmt.Printf("   Variables: name = \"World\"\n")
	fmt.Printf("   Output: %s\n\n", result1)

	// Example 2: Number formatting
	fmt.Println("2. Number Formatting:")
	mf2, err := messageformat.New("en", "You have {$count :number} messages")
	if err != nil {
		log.Fatal(err)
	}

	result2, err := mf2.Format(map[string]interface{}{
		"count": 1234,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Input: \"You have {$count :number} messages\"\n")
	fmt.Printf("   Variables: count = 1234\n")
	fmt.Printf("   Output: %s\n\n", result2)

	// Example 3: Multiple variables
	fmt.Println("3. Multiple Variables:")
	mf3, err := messageformat.New("en", "{$user} sent {$count} messages")
	if err != nil {
		log.Fatal(err)
	}

	result3, err := mf3.Format(map[string]interface{}{
		"user":  "Alice",
		"count": 5,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Input: \"{$user} sent {$count} messages\"\n")
	fmt.Printf("   Variables: user=\"Alice\", count=5\n")
	fmt.Printf("   Output: %s\n\n", result3)

	// Example 4: Localization comparison
	fmt.Println("4. Localization Comparison:")

	// English version
	mfEn, err := messageformat.New("en", "Hello, {$name}! You have {$count} new messages.")
	if err != nil {
		log.Fatal(err)
	}

	resultEn, err := mfEn.Format(map[string]interface{}{
		"name":  "Alice",
		"count": 42,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   English: %s\n", resultEn)

	// Chinese version
	mfCn, err := messageformat.New("zh-CN", "你好，{$name}！你有 {$count} 条新消息。")
	if err != nil {
		log.Fatal(err)
	}

	resultCn, err := mfCn.Format(map[string]interface{}{
		"name":  "爱丽丝",
		"count": 42,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Chinese: %s\n\n", resultCn)

	// Example 5: Built-in function formatting
	fmt.Println("5. Built-in Function Formatting:")

	// Number with options
	mf5a, err := messageformat.New("en", "Price: {$price :number style=currency currency=USD}")
	if err != nil {
		log.Fatal(err)
	}

	result5a, err := mf5a.Format(map[string]interface{}{
		"price": 29.99,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Currency format: %s\n", result5a)

	// Integer formatting
	mf5b, err := messageformat.New("en", "Count: {$count :integer}")
	if err != nil {
		log.Fatal(err)
	}

	result5b, err := mf5b.Format(map[string]interface{}{
		"count": 1234.56,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Integer format: %s\n", result5b)

	// String formatting
	mf5c, err := messageformat.New("en", "Name: {$name :string}")
	if err != nil {
		log.Fatal(err)
	}

	result5c, err := mf5c.Format(map[string]interface{}{
		"name": "John Doe",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   String format: %s\n\n", result5c)

	// Example 6: Error handling
	fmt.Println("6. Error Handling:")
	mf6, err := messageformat.New("en", "Hello {$name}")
	if err != nil {
		log.Fatal(err)
	}

	result6, err := mf6.Format(map[string]interface{}{
		// Intentionally omit the 'name' variable
	})
	if err != nil {
		fmt.Printf("   Error when variable is missing: %v\n", err)
	} else {
		fmt.Printf("   Output when variable is missing: %s\n", result6)
	}

	// Example 7: Using functional options
	fmt.Println("\n7. Functional Options Pattern:")
	mf7, err := messageformat.New("en", "Hello, {$name}!",
		messageformat.WithBidiIsolation("none"),
		messageformat.WithDir("ltr"),
	)
	if err != nil {
		log.Fatal(err)
	}

	result7, err := mf7.Format(map[string]interface{}{
		"name": "World",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   With options: %s\n", result7)

	// Example 8: Traditional options struct
	fmt.Println("\n8. Traditional Options Struct:")

	mf8, err := messageformat.New("en", "Hello, {$name}!", &messageformat.MessageFormatOptions{
		BidiIsolation: messageformat.BidiNone,
		Dir:           messageformat.DirLTR,
	})
	if err != nil {
		log.Fatal(err)
	}

	result8, err := mf8.Format(map[string]interface{}{
		"name": "World",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   With struct options: %s\n", result8)

	fmt.Println("\n=== Basic Examples Complete ===")
}
