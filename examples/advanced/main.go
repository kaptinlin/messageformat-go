package main

import (
	"fmt"
	"log"
	"time"

	"github.com/kaptinlin/messageformat-go"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// Custom function for demonstration
func highlightFunction(ctx functions.MessageFunctionContext, options map[string]interface{}, operand interface{}) messagevalue.MessageValue {
	var str string
	if operand == nil {
		str = ""
	} else if s, ok := operand.(string); ok {
		str = s
	} else {
		str = fmt.Sprintf("%v", operand)
	}

	// Get highlight style from options
	style := "bold"
	if s, ok := options["style"].(string); ok {
		style = s
	}

	var result string
	switch style {
	case "bold":
		result = "**" + str + "**"
	case "italic":
		result = "*" + str + "*"
	case "underline":
		result = "_" + str + "_"
	case "code":
		result = "`" + str + "`"
	default:
		result = str
	}

	return messagevalue.NewStringValue(result, ctx.Locales()[0], ctx.Source())
}

func main() {
	fmt.Println("=== MessageFormat 2.0 Advanced Features Examples ===")

	// Example 1: FormatToParts - Structured output
	fmt.Println("1. Structured Output with FormatToParts:")
	mf1, err := messageformat.New("en", "Hello, {$name}! You have {$count :number} new messages.", messageformat.WithBidiIsolation(messageformat.BidiNone))
	if err != nil {
		log.Fatal(err)
	}

	parts, err := mf1.FormatToParts(map[string]interface{}{
		"name":  "Alice",
		"count": 42,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("   Message parts:\n")
	for i, part := range parts {
		fmt.Printf("     [%d] Type: %s, Value: %v\n", i, part.Type(), part.Value())
	}
	fmt.Println()

	// Example 2: Bidirectional text support
	fmt.Println("2. Bidirectional Text Support:")

	// English with Arabic name
	mf2a, err := messageformat.New("en", "User {$name} sent a message",
		messageformat.WithBidiIsolation(messageformat.BidiDefault),
		messageformat.WithDir(messageformat.DirLTR),
	)
	if err != nil {
		log.Fatal(err)
	}

	result2a, err := mf2a.Format(map[string]interface{}{
		"name": "أحمد", // Arabic name
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   LTR with Arabic name: %s\n", result2a)

	// Without bidi isolation
	mf2b, err := messageformat.New("en", "User {$name} sent a message",
		messageformat.WithBidiIsolation(messageformat.BidiNone),
	)
	if err != nil {
		log.Fatal(err)
	}

	result2b, err := mf2b.Format(map[string]interface{}{
		"name": "أحمد",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Without bidi isolation: %s\n", result2b)
	fmt.Println()

	// Example 3: Complex select patterns
	fmt.Println("3. Complex Select Patterns:")
	complexMessage := `.input {$userType :string}
.input {$count :number}
.input {$status :string}
.match $userType $count $status
admin 0 *      {{Admin {$name} has no pending tasks}}
admin 1 active {{Admin {$name} has 1 active task}}
admin * active {{Admin {$name} has {$count} active tasks}}
admin * *      {{Admin {$name} has {$count} tasks ({$status})}}
user  0 *      {{User {$name} has no messages}}
user  1 *      {{User {$name} has 1 message}}
user  * *      {{User {$name} has {$count} messages}}
*     * *      {{{$name} ({$userType}) has {$count} items ({$status})}}`

	mf3, err := messageformat.New("en", complexMessage)
	if err != nil {
		log.Fatal(err)
	}

	complexTestCases := []map[string]interface{}{
		{"name": "Alice", "userType": "admin", "count": 0, "status": "idle"},
		{"name": "Bob", "userType": "admin", "count": 1, "status": "active"},
		{"name": "Charlie", "userType": "admin", "count": 5, "status": "active"},
		{"name": "Diana", "userType": "admin", "count": 3, "status": "pending"},
		{"name": "Eve", "userType": "user", "count": 0, "status": "online"},
		{"name": "Frank", "userType": "user", "count": 1, "status": "away"},
		{"name": "Grace", "userType": "user", "count": 10, "status": "busy"},
		{"name": "Henry", "userType": "guest", "count": 2, "status": "active"},
	}

	for _, testCase := range complexTestCases {
		result, err := mf3.Format(testCase)
		if err != nil {
			log.Printf("   Error: %v\n", err)
			continue
		}
		fmt.Printf("   %s (%s, %d, %s): %s\n",
			testCase["name"], testCase["userType"], testCase["count"], testCase["status"], result)
	}
	fmt.Println()

	// Example 4: Custom functions with complex logic
	fmt.Println("4. Custom Functions with Complex Logic:")
	mf4, err := messageformat.New("en", "Status: {$status :highlight style=bold}, Priority: {$priority :highlight style=italic}",
		messageformat.WithFunction("highlight", highlightFunction),
	)
	if err != nil {
		log.Fatal(err)
	}

	result4, err := mf4.Format(map[string]interface{}{
		"status":   "Active",
		"priority": "High",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Input: Custom highlighting functions\n")
	fmt.Printf("   Output: %s\n\n", result4)

	// Example 5: Performance optimization - Reusing MessageFormat instances
	fmt.Println("5. Performance Optimization:")

	// Create MessageFormat once
	mf5, err := messageformat.New("en", "Processing item {$index} of {$total}: {$item}")
	if err != nil {
		log.Fatal(err)
	}

	// Measure time for multiple formats
	start := time.Now()
	items := []string{"file1.txt", "file2.txt", "file3.txt", "file4.txt", "file5.txt"}

	for i, item := range items {
		result, err := mf5.Format(map[string]interface{}{
			"index": i + 1,
			"total": len(items),
			"item":  item,
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("   %s\n", result)
	}

	elapsed := time.Since(start)
	fmt.Printf("   Processed %d items in %v\n\n", len(items), elapsed)

	// Example 6: Error handling with custom error handler
	fmt.Println("6. Error Handling with Custom Error Handler:")

	var errorCount int
	errorHandler := func(err error) {
		errorCount++
		fmt.Printf("   Error #%d: %v\n", errorCount, err)
	}

	mf6, err := messageformat.New("en", "Value: {$missing_var}, Count: {$count :number}")
	if err != nil {
		log.Fatal(err)
	}

	result6, err := mf6.Format(map[string]interface{}{
		"count": 42,
		// Intentionally missing "missing_var"
	})
	if err != nil {
		errorHandler(err)
	}
	fmt.Printf("   Result with missing variable: %s\n", result6)
	fmt.Printf("   Total errors handled: %d\n\n", errorCount)

	// Example 7: Nested patterns and declarations
	fmt.Println("7. Nested Patterns and Declarations:")
	nestedMessage := `
.local $greeting = {|Hello| :string}
.local $punctuation = {|!| :string}
{{{$greeting}, {$name}{$punctuation} Welcome to our {$service}.}}
`
	mf7, err := messageformat.New("en", nestedMessage)
	if err != nil {
		log.Fatal(err)
	}

	result7, err := mf7.Format(map[string]interface{}{
		"name":    "Alice",
		"service": "platform",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Input: Message with local declarations\n")
	fmt.Printf("   Output: %s\n\n", result7)

	// Example 8: Multiple locale support
	fmt.Println("8. Multiple Locale Support:")

	locales := []string{"en", "zh-CN"}
	messages := map[string]string{
		"en":    "You have {$count :number} unread messages",
		"zh-CN": "您有 {$count :number} 条未读消息",
	}

	for _, locale := range locales {
		mf, err := messageformat.New(locale, messages[locale])
		if err != nil {
			log.Fatal(err)
		}

		result, err := mf.Format(map[string]interface{}{
			"count": 1234,
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("   %s: %s\n", locale, result)
	}
	fmt.Println()

	// Example 9: Working with structured data
	fmt.Println("9. Working with Structured Data:")
	structuredMessage := `
.input {$type :string}
.match $type
email {{📧 Email from {$sender}: {$subject}}}
sms {{📱 SMS from {$sender}: {$text}}}
push {{🔔 Push notification: {$title}}}
* {{📬 {$type} notification}}
`
	mf9, err := messageformat.New("en", structuredMessage)
	if err != nil {
		log.Fatal(err)
	}

	notifications := []map[string]interface{}{
		{
			"type":    "email",
			"sender":  "alice@example.com",
			"subject": "Meeting reminder",
		},
		{
			"type":   "sms",
			"sender": "+1234567890",
			"text":   "Your order is ready",
		},
		{
			"type":  "push",
			"title": "App update available",
		},
		{
			"type": "unknown",
		},
	}

	for _, notification := range notifications {
		result, err := mf9.Format(notification)
		if err != nil {
			log.Printf("   Error: %v\n", err)
			continue
		}
		fmt.Printf("   %s\n", result)
	}

	fmt.Println("\n=== Advanced Features Examples Complete ===")
	fmt.Println("\nNote: These examples demonstrate advanced MessageFormat 2.0 capabilities.")
	fmt.Println("Some features may require specific library implementation support.")
}
