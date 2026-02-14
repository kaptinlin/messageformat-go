package tests

import (
	"testing"

	messageformat "github.com/kaptinlin/messageformat-go"
	"github.com/stretchr/testify/require"
)

func BenchmarkSimpleMessage(b *testing.B) {
	mf, err := messageformat.New("en", "Hello, {$name}!")
	require.NoError(b, err)

	data := map[string]any{
		"name": "World",
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := mf.Format(data)
		require.NoError(b, err)
	}
}

func BenchmarkNumberFormatting(b *testing.B) {
	mf, err := messageformat.New("en", "You have {$count :number} messages")
	require.NoError(b, err)

	data := map[string]any{
		"count": 42,
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := mf.Format(data)
		require.NoError(b, err)
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
	require.NoError(b, err)

	data := map[string]any{
		"count": 5,
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := mf.Format(data)
		require.NoError(b, err)
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
	require.NoError(b, err)

	data := map[string]any{
		"userName":   "Alice",
		"photoCount": 3,
		"userGender": "female",
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := mf.Format(data)
		require.NoError(b, err)
	}
}

func BenchmarkFormatToParts(b *testing.B) {
	mf, err := messageformat.New("en", "Hello, {$name}! You have {$count :number} messages.")
	require.NoError(b, err)

	data := map[string]any{
		"name":  "World",
		"count": 42,
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := mf.FormatToParts(data)
		require.NoError(b, err)
	}
}

func BenchmarkMessageCreation(b *testing.B) {
	pattern := "Hello, {$name}! You have {$count :number} messages."

	b.ResetTimer()
	for b.Loop() {
		_, err := messageformat.New("en", pattern)
		require.NoError(b, err)
	}
}
