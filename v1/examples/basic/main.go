// Package main demonstrates basic MessageFormat usage.
//
// This example shows the fundamental concepts of the ICU MessageFormat v1 library:
//   - Simple variable interpolation
//   - Multiple variable substitution
//   - Basic pluralization rules with CLDR support
//   - Select statements for conditional text
//
// Run this example with:
//
//	cd examples/basic && go run main.go
package main

import (
	"fmt"
	"log"

	mf "github.com/kaptinlin/messageformat-go/v1"
)

func main() {
	fmt.Println("=== Basic MessageFormat Examples ===")

	// Create a MessageFormat instance
	messageFormat, err := mf.New("en", nil)
	if err != nil {
		log.Fatal(err)
	}

	// Example 1: Simple variable interpolation
	fmt.Println("\n1. Simple Variable Interpolation:")
	simpleMsg, err := messageFormat.Compile("Hello, {name}!")
	if err != nil {
		log.Fatal(err)
	}

	result, err := simpleMsg(map[string]any{
		"name": "Alice",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Template: Hello, {name}!\n")
	fmt.Printf("   Data: {\"name\": \"Alice\"}\n")
	fmt.Printf("   Result: %s\n", result)

	// Example 2: Multiple variables
	fmt.Println("\n2. Multiple Variables:")
	multiMsg, err := messageFormat.Compile("Welcome {firstName} {lastName}! You have {messageCount} new messages.")
	if err != nil {
		log.Fatal(err)
	}

	result, err = multiMsg(map[string]any{
		"firstName":    "John",
		"lastName":     "Doe",
		"messageCount": 3,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Result: %s\n", result)

	// Example 3: Basic pluralization
	fmt.Println("\n3. Basic Pluralization:")
	pluralMsg, err := messageFormat.Compile("{count, plural, one {# item} other {# items}}")
	if err != nil {
		log.Fatal(err)
	}

	// Test with different counts
	counts := []int{0, 1, 2, 5}
	for _, count := range counts {
		result, err := pluralMsg(map[string]any{
			"count": count,
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("   Count %d: %s\n", count, result)
	}

	// Example 4: Select statements
	fmt.Println("\n4. Select Statements:")
	selectMsg, err := messageFormat.Compile("{gender, select, male {He} female {She} other {They}} went to the store.")
	if err != nil {
		log.Fatal(err)
	}

	genders := []string{"male", "female", "other", "unknown"}
	for _, gender := range genders {
		result, err := selectMsg(map[string]any{
			"gender": gender,
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("   Gender '%s': %s\n", gender, result)
	}

	fmt.Println("\n=== Examples completed successfully! ===")
}
