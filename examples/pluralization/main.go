package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/messageformat-go"
)

func main() {
	fmt.Println("=== MessageFormat 2.0 Pluralization and Select Messages ===")

	// Example 1: Basic pluralization
	fmt.Println("1. Basic Pluralization:")
	pluralMessage := `.input {$count :number}
.match $count
0   {{No messages}}
one {{One message}}
*   {{{$count} messages}}`

	mf1, err := messageformat.New("en", pluralMessage, messageformat.WithBidiIsolation(messageformat.BidiNone))
	if err != nil {
		log.Fatal(err)
	}

	// Test different counts
	testCounts := []int{0, 1, 2, 5, 10, 100}
	for _, count := range testCounts {
		result, err := mf1.Format(map[string]interface{}{
			"count": count,
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("   count = %d: %s\n", count, result)
	}
	fmt.Println()

	// Example 2: Localized pluralization comparison
	fmt.Println("2. Localized Pluralization Comparison:")

	// English pluralization
	fmt.Println("   English:")
	englishPlural := `.input {$count :number}
.match $count
0   {{No items}}
one {{One item}}
*   {{{$count} items}}`

	mfEn, err := messageformat.New("en", englishPlural, messageformat.WithBidiIsolation(messageformat.BidiNone))
	if err != nil {
		log.Fatal(err)
	}

	for _, count := range []int{0, 1, 2, 5} {
		result, err := mfEn.Format(map[string]interface{}{
			"count": count,
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("     count = %d: %s\n", count, result)
	}

	// Chinese pluralization (Chinese doesn't have plural forms like English)
	fmt.Println("   Chinese:")
	chinesePlural := `.input {$count :number}
.match $count
0 {{没有物品}}
* {{有 {$count} 个物品}}`

	mfCn, err := messageformat.New("zh-CN", chinesePlural, messageformat.WithBidiIsolation(messageformat.BidiNone))
	if err != nil {
		log.Fatal(err)
	}

	for _, count := range []int{0, 1, 2, 5} {
		result, err := mfCn.Format(map[string]interface{}{
			"count": count,
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("     count = %d: %s\n", count, result)
	}
	fmt.Println()

	// Example 3: Gender selection
	fmt.Println("3. Gender Selection:")
	genderMessage := `.input {$gender :string}
.match $gender
male   {{{$name} sent a message}}
female {{{$name} sent a message}}
*      {{{$name} sent a message}}`

	mf3, err := messageformat.New("en", genderMessage, messageformat.WithBidiIsolation(messageformat.BidiNone))
	if err != nil {
		log.Fatal(err)
	}

	testCases := []map[string]interface{}{
		{"name": "John", "gender": "male"},
		{"name": "Jane", "gender": "female"},
		{"name": "Alex", "gender": "other"},
	}

	for _, testCase := range testCases {
		result, err := mf3.Format(testCase)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("   %s (%s): %s\n", testCase["name"], testCase["gender"], result)
	}
	fmt.Println()

	// Example 4: Complex selection (count + gender)
	fmt.Println("4. Complex Selection (Count + Gender):")
	complexMessage := `.input {$count :number}
.input {$gender :string}
.match $count $gender
0   *      {{No one sent any messages}}
one male   {{{$name} sent one message}}
one female {{{$name} sent one message}}
one *      {{{$name} sent one message}}
*   male   {{{$name} sent {$count} messages}}
*   female {{{$name} sent {$count} messages}}
*   *      {{{$name} sent {$count} messages}}`

	mf4, err := messageformat.New("en", complexMessage, messageformat.WithBidiIsolation(messageformat.BidiNone))
	if err != nil {
		log.Fatal(err)
	}

	complexTestCases := []map[string]interface{}{
		{"name": "John", "gender": "male", "count": 0},
		{"name": "Jane", "gender": "female", "count": 1},
		{"name": "Alex", "gender": "other", "count": 1},
		{"name": "Bob", "gender": "male", "count": 5},
		{"name": "Alice", "gender": "female", "count": 10},
		{"name": "Sam", "gender": "other", "count": 3},
	}

	for _, testCase := range complexTestCases {
		result, err := mf4.Format(testCase)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("   %s (%s, %d): %s\n",
			testCase["name"], testCase["gender"], testCase["count"], result)
	}
	fmt.Println()

	// Example 5: Status selection
	fmt.Println("5. Status Selection:")
	statusMessage := `.input {$status :string}
.match $status
online  {{{$user} is online}}
offline {{{$user} is offline}}
away    {{{$user} is away}}
busy    {{{$user} is busy}}
*       {{{$user} has unknown status}}`

	mf5, err := messageformat.New("en", statusMessage, messageformat.WithBidiIsolation(messageformat.BidiNone))
	if err != nil {
		log.Fatal(err)
	}

	statusTestCases := []map[string]interface{}{
		{"user": "Alice", "status": "online"},
		{"user": "Bob", "status": "offline"},
		{"user": "Charlie", "status": "away"},
		{"user": "Diana", "status": "busy"},
		{"user": "Eve", "status": "invisible"},
	}

	for _, testCase := range statusTestCases {
		result, err := mf5.Format(testCase)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("   %s: %s\n", testCase["user"], result)
	}
	fmt.Println()

	// Example 6: Time-based selection
	fmt.Println("6. Time-based Selection:")
	timeMessage := `.input {$hours :number}
.match $hours
0 {{just now}}
1 {{1 hour ago}}
2 {{2 hours ago}}
3 {{3 hours ago}}
4 {{4 hours ago}}
5 {{5 hours ago}}
* {{{$hours} hours ago}}`

	mf6, err := messageformat.New("en", timeMessage, messageformat.WithBidiIsolation(messageformat.BidiNone))
	if err != nil {
		log.Fatal(err)
	}

	timeTestCases := []int{0, 1, 2, 3, 4, 5, 12, 24, 48}
	for _, hours := range timeTestCases {
		result, err := mf6.Format(map[string]interface{}{
			"hours": hours,
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("   %d hours: %s\n", hours, result)
	}
	fmt.Println()

	// Example 7: File type selection
	fmt.Println("7. File Type Selection:")
	fileMessage := `.input {$type :string}
.match $type
image    {{📷 {$name} (Image file)}}
video    {{🎥 {$name} (Video file)}}
audio    {{🎵 {$name} (Audio file)}}
document {{📄 {$name} (Document)}}
*        {{📁 {$name} (Unknown file type)}}`

	mf7, err := messageformat.New("en", fileMessage, messageformat.WithBidiIsolation(messageformat.BidiNone))
	if err != nil {
		log.Fatal(err)
	}

	fileTestCases := []map[string]interface{}{
		{"name": "photo.jpg", "type": "image"},
		{"name": "movie.mp4", "type": "video"},
		{"name": "song.mp3", "type": "audio"},
		{"name": "report.pdf", "type": "document"},
		{"name": "data.xyz", "type": "unknown"},
	}

	for _, testCase := range fileTestCases {
		result, err := mf7.Format(testCase)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("   %s\n", result)
	}

	fmt.Println("\n=== Pluralization and Select Messages Examples Complete ===")
}
