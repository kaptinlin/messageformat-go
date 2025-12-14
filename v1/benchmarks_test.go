package v1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func benchmarkHelper(b *testing.B, message string, params map[string]interface{}, opts *MessageFormatOptions) {
	b.Helper()
	mf, err := New("en", opts)
	if err != nil {
		b.Fatal(err)
	}
	compiled, err := mf.Compile(message)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, err := compiled(params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMessageFormatSimple(b *testing.B) {
	benchmarkHelper(b, "Hello {name}!", map[string]interface{}{"name": "World"}, nil)
}

func BenchmarkMessageFormatPlural(b *testing.B) {
	benchmarkHelper(b, "{count, plural, one {# item} other {# items}}",
		map[string]interface{}{"count": 42}, nil)
}

func BenchmarkMessageFormatSelect(b *testing.B) {
	benchmarkHelper(b, "{gender, select, male {He} female {She} other {They}} went to the store.",
		map[string]interface{}{"gender": "female"}, nil)
}

func BenchmarkMessageFormatComplex(b *testing.B) {
	message := `{gender, select, 
		male {{count, plural, one {He has # item} other {He has # items}}} 
		female {{count, plural, one {She has # item} other {She has # items}}}
		other {{count, plural, one {They have # item} other {They have # items}}}
	} in the cart for a total of {total, number, currency}.`
	benchmarkHelper(b, message, map[string]interface{}{
		"gender": "male", "count": 3, "total": 29.99,
	}, nil)
}

func BenchmarkMessageFormatCompilation(b *testing.B) {
	mf, err := New("en", nil)
	if err != nil {
		b.Fatal(err)
	}
	message := "{count, plural, one {# item} other {# items}} for {name}"

	b.ResetTimer()
	for b.Loop() {
		_, err := mf.Compile(message)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTypeSafeCreation(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		mf, err := New("en", &MessageFormatOptions{
			ReturnType: ReturnTypeString,
			Currency:   "USD",
		})
		if err != nil {
			b.Fatal(err)
		}
		if mf == nil {
			b.Fatal("MessageFormat is nil")
		}
	}
}

func BenchmarkTypeSafeOptionsAccess(b *testing.B) {
	mf, err := New("en", &MessageFormatOptions{
		ReturnType:  ReturnTypeValues,
		BiDiSupport: true,
		Currency:    "EUR",
	})
	require.NoError(b, err)

	b.ResetTimer()
	for b.Loop() {
		opts := mf.ResolvedOptions()
		if opts.Locale == "" {
			b.Fatal("Locale is empty")
		}
	}
}

func BenchmarkTypeSafeSkeletonCreation(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		skeleton := &Skeleton{
			Group:        GroupThousands,
			Sign:         SignAlways,
			Decimal:      DecimalAuto,
			RoundingMode: RoundingHalfUp,
			Unit: &UnitConfig{
				Style:    UnitCurrency,
				Currency: StringPtr("EUR"),
			},
			Notation: &NotationConfig{
				Style: NotationCompactShort,
			},
			UnitWidth: UnitWidthShort,
		}
		_ = skeleton // Use skeleton to avoid unused variable
	}
}

func BenchmarkStaticMethods(b *testing.B) {
	b.Run("Escape", func(b *testing.B) {
		for b.Loop() {
			result := Escape("Hello {name}!", true)
			if result == "" {
				b.Fatal("Escape result is empty")
			}
		}
	})

	b.Run("SupportedLocalesOf", func(b *testing.B) {
		for b.Loop() {
			result, err := SupportedLocalesOf([]string{"en", "fr", "de"})
			if err != nil {
				b.Fatal(err)
			}
			if len(result) == 0 {
				b.Fatal("No supported locales returned")
			}
		}
	})
}
