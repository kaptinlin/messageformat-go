package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/kaptinlin/messageformat-go"
	"github.com/kaptinlin/messageformat-go/pkg/functions"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// Custom function: uppercase
// Converts input to uppercase
func uppercaseFunction(ctx functions.MessageFunctionContext, options map[string]interface{}, operand interface{}) messagevalue.MessageValue {
	var str string
	if operand == nil {
		str = ""
	} else if s, ok := operand.(string); ok {
		str = s
	} else {
		str = fmt.Sprintf("%v", operand)
	}

	result := strings.ToUpper(str)
	return messagevalue.NewStringValue(result, ctx.Locales()[0], ctx.Source())
}

// Custom function: reverse
// Reverses the input string
func reverseFunction(ctx functions.MessageFunctionContext, options map[string]interface{}, operand interface{}) messagevalue.MessageValue {
	var str string
	if operand == nil {
		str = ""
	} else if s, ok := operand.(string); ok {
		str = s
	} else {
		str = fmt.Sprintf("%v", operand)
	}

	// Reverse the string
	runes := []rune(str)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	result := string(runes)

	return messagevalue.NewStringValue(result, ctx.Locales()[0], ctx.Source())
}

// Custom function: emoji
// Adds emoji based on the type option
func emojiFunction(ctx functions.MessageFunctionContext, options map[string]interface{}, operand interface{}) messagevalue.MessageValue {
	var str string
	if operand == nil {
		str = ""
	} else if s, ok := operand.(string); ok {
		str = s
	} else {
		str = fmt.Sprintf("%v", operand)
	}

	// Get emoji type from options
	emojiType := "default"
	if t, ok := options["type"].(string); ok {
		emojiType = t
	}

	var emoji string
	switch emojiType {
	case "happy":
		emoji = "😊"
	case "sad":
		emoji = "😢"
	case "love":
		emoji = "❤️"
	case "fire":
		emoji = "🔥"
	case "star":
		emoji = "⭐"
	default:
		emoji = "✨"
	}

	result := emoji + " " + str + " " + emoji
	return messagevalue.NewStringValue(result, ctx.Locales()[0], ctx.Source())
}

// Custom function: timeago
// Formats time relative to now
func timeAgoFunction(ctx functions.MessageFunctionContext, options map[string]interface{}, operand interface{}) messagevalue.MessageValue {
	var hours int
	if operand == nil {
		hours = 0
	} else if h, ok := operand.(int); ok {
		hours = h
	} else if h, ok := operand.(float64); ok {
		hours = int(h)
	} else {
		hours = 0
	}

	var result string
	switch {
	case hours == 0:
		result = "just now"
	case hours == 1:
		result = "1 hour ago"
	case hours < 24:
		result = fmt.Sprintf("%d hours ago", hours)
	case hours < 48:
		result = "1 day ago"
	case hours < 24*7:
		days := hours / 24
		result = fmt.Sprintf("%d days ago", days)
	case hours < 24*30:
		weeks := hours / (24 * 7)
		result = fmt.Sprintf("%d weeks ago", weeks)
	default:
		months := hours / (24 * 30)
		result = fmt.Sprintf("%d months ago", months)
	}

	return messagevalue.NewStringValue(result, ctx.Locales()[0], ctx.Source())
}

// Custom function: format
// Formats strings with padding and alignment
func formatFunction(ctx functions.MessageFunctionContext, options map[string]interface{}, operand interface{}) messagevalue.MessageValue {
	var str string
	if operand == nil {
		str = ""
	} else if s, ok := operand.(string); ok {
		str = s
	} else {
		str = fmt.Sprintf("%v", operand)
	}

	// Get formatting options
	width := 0
	if w, ok := options["width"].(int); ok {
		width = w
	} else if w, ok := options["width"].(float64); ok {
		width = int(w)
	}

	align := "left"
	if a, ok := options["align"].(string); ok {
		align = a
	}

	pad := " "
	if p, ok := options["pad"].(string); ok {
		pad = p
	}

	// Apply formatting
	if width > len(str) {
		padding := strings.Repeat(pad, width-len(str))
		switch align {
		case "right":
			str = padding + str
		case "center":
			leftPad := padding[:len(padding)/2]
			rightPad := padding[len(padding)/2:]
			str = leftPad + str + rightPad
		default: // left
			str += padding
		}
	}

	return messagevalue.NewStringValue(str, ctx.Locales()[0], ctx.Source())
}

func main() {
	fmt.Println("=== MessageFormat 2.0 Custom Functions Examples ===")

	// Example 1: Uppercase function
	fmt.Println("1. Uppercase Function:")
	mf1, err := messageformat.New("en", "Hello, {$name :uppercase}!",
		messageformat.WithFunction("uppercase", uppercaseFunction),
		messageformat.WithBidiIsolation(messageformat.BidiNone),
	)
	if err != nil {
		log.Fatal(err)
	}

	result1, err := mf1.Format(map[string]interface{}{
		"name": "world",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Input: \"Hello, {$name :uppercase}!\"\n")
	fmt.Printf("   Variables: name = \"world\"\n")
	fmt.Printf("   Output: %s\n\n", result1)

	// Example 2: Reverse function
	fmt.Println("2. Reverse Function:")
	mf2, err := messageformat.New("en", "Original: {$text}, Reversed: {$text :reverse}",
		messageformat.WithFunction("reverse", reverseFunction),
		messageformat.WithBidiIsolation(messageformat.BidiNone),
	)
	if err != nil {
		log.Fatal(err)
	}

	result2, err := mf2.Format(map[string]interface{}{
		"text": "Hello",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Input: \"Original: {$text}, Reversed: {$text :reverse}\"\n")
	fmt.Printf("   Variables: text = \"Hello\"\n")
	fmt.Printf("   Output: %s\n\n", result2)

	// Example 3: Emoji function with options
	fmt.Println("3. Emoji Function with Options:")
	mf3, err := messageformat.New("en", "Message: {$msg :emoji type=happy}",
		messageformat.WithFunction("emoji", emojiFunction),
		messageformat.WithBidiIsolation(messageformat.BidiNone),
	)
	if err != nil {
		log.Fatal(err)
	}

	result3, err := mf3.Format(map[string]interface{}{
		"msg": "Great job!",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Input: \"Message: {$msg :emoji type=happy}\"\n")
	fmt.Printf("   Variables: msg = \"Great job!\"\n")
	fmt.Printf("   Output: %s\n\n", result3)

	// Example 4: Multiple emoji types
	fmt.Println("4. Multiple Emoji Types:")
	emojiTypes := []string{"happy", "sad", "love", "fire", "star", "default"}
	for _, emojiType := range emojiTypes {
		mf, err := messageformat.New("en", fmt.Sprintf("Status: {$status :emoji type=%s}", emojiType),
			messageformat.WithFunction("emoji", emojiFunction),
			messageformat.WithBidiIsolation(messageformat.BidiNone),
		)
		if err != nil {
			log.Fatal(err)
		}

		result, err := mf.Format(map[string]interface{}{
			"status": "Active",
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("   %s: %s\n", emojiType, result)
	}
	fmt.Println()

	// Example 5: Time ago function
	fmt.Println("5. Time Ago Function:")
	mf5, err := messageformat.New("en", "Last seen: {$hours :timeago}",
		messageformat.WithFunction("timeago", timeAgoFunction),
		messageformat.WithBidiIsolation(messageformat.BidiNone),
	)
	if err != nil {
		log.Fatal(err)
	}

	timeTestCases := []int{0, 1, 3, 12, 25, 48, 168, 720, 8760}
	for _, hours := range timeTestCases {
		result, err := mf5.Format(map[string]interface{}{
			"hours": hours,
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("   %d hours: %s\n", hours, result)
	}
	fmt.Println()

	// Example 6: Format function with alignment
	fmt.Println("6. Format Function with Alignment:")
	mf6, err := messageformat.New("en", "Name: [{$name :format width=15 align=right}]",
		messageformat.WithFunction("format", formatFunction),
		messageformat.WithBidiIsolation(messageformat.BidiNone),
	)
	if err != nil {
		log.Fatal(err)
	}

	names := []string{"Alice", "Bob", "Charlie", "Diana"}
	for _, name := range names {
		result, err := mf6.Format(map[string]interface{}{
			"name": name,
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("   %s\n", result)
	}
	fmt.Println()

	// Example 7: Multiple custom functions in one message
	fmt.Println("7. Multiple Custom Functions:")
	mf7, err := messageformat.New("en", "User: {$user :uppercase}, Status: {$status :emoji type=star}, Last seen: {$hours :timeago}",
		messageformat.WithFunction("uppercase", uppercaseFunction),
		messageformat.WithFunction("emoji", emojiFunction),
		messageformat.WithFunction("timeago", timeAgoFunction),
		messageformat.WithBidiIsolation(messageformat.BidiNone),
	)
	if err != nil {
		log.Fatal(err)
	}

	result7, err := mf7.Format(map[string]interface{}{
		"user":   "alice",
		"status": "online",
		"hours":  2,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Input: Multiple functions in one message\n")
	fmt.Printf("   Output: %s\n\n", result7)

	// Example 8: Error handling in custom functions
	fmt.Println("8. Error Handling:")
	mf8, err := messageformat.New("en", "Value: {$value :uppercase}",
		messageformat.WithFunction("uppercase", uppercaseFunction),
		messageformat.WithBidiIsolation(messageformat.BidiNone),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Test with different value types
	testValues := []interface{}{
		"hello",
		123,
		true,
		nil,
		[]string{"a", "b"},
	}

	for _, value := range testValues {
		result, err := mf8.Format(map[string]interface{}{
			"value": value,
		})
		if err != nil {
			fmt.Printf("   Error with %T(%v): %v\n", value, value, err)
		} else {
			fmt.Printf("   %T(%v): %s\n", value, value, result)
		}
	}

	fmt.Println("\n=== Custom Functions Examples Complete ===")
}
