package tests

import (
	"testing"

	messageformat "github.com/kaptinlin/messageformat-go"
)

func BenchmarkSimpleMessage(b *testing.B) {
	mf, err := messageformat.New("en", "Hello, {$name}!")
	if err != nil {
		b.Fatal(err)
	}

	data := map[string]interface{}{
		"name": "World",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := mf.Format(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNumberFormatting(b *testing.B) {
	mf, err := messageformat.New("en", "You have {$count :number} messages")
	if err != nil {
		b.Fatal(err)
	}

	data := map[string]interface{}{
		"count": 42,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := mf.Format(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSelectMessage(b *testing.B) {
	mf, err := messageformat.New("en", `
.input {$count :number}
.match $count
0   {{No items}}
one {{One item}}
*   {{{$count} items}}
`)
	if err != nil {
		b.Fatal(err)
	}

	data := map[string]interface{}{
		"count": 5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := mf.Format(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkComplexMessage(b *testing.B) {
	mf, err := messageformat.New("en", `
.input {$userName :string}
.input {$photoCount :number}
.input {$userGender :string}
.match $photoCount $userGender
0   male   {{{$userName} didn't add any photos to his album.}}
0   female {{{$userName} didn't add any photos to her album.}}
0   *      {{{$userName} didn't add any photos to their album.}}
one male   {{{$userName} added one photo to his album.}}
one female {{{$userName} added one photo to her album.}}
one *      {{{$userName} added one photo to their album.}}
*   male   {{{$userName} added {$photoCount} photos to his album.}}
*   female {{{$userName} added {$photoCount} photos to her album.}}
*   *      {{{$userName} added {$photoCount} photos to their album.}}
`)
	if err != nil {
		b.Fatal(err)
	}

	data := map[string]interface{}{
		"userName":   "Alice",
		"photoCount": 3,
		"userGender": "female",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := mf.Format(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFormatToParts(b *testing.B) {
	mf, err := messageformat.New("en", "Hello, {$name}! You have {$count :number} messages.")
	if err != nil {
		b.Fatal(err)
	}

	data := map[string]interface{}{
		"name":  "World",
		"count": 42,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := mf.FormatToParts(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMessageCreation(b *testing.B) {
	pattern := "Hello, {$name}! You have {$count :number} messages."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := messageformat.New("en", pattern)
		if err != nil {
			b.Fatal(err)
		}
	}
}
