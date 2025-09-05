// Package main demonstrates multilingual MessageFormat usage with different locales.
//
// This example showcases international application support:
//   - Multiple locale support (English, Russian, Arabic, Welsh, Chinese)
//   - Complex CLDR plural rules demonstration
//   - Currency localization patterns
//   - Return type variations (string vs values array)
//
// Run this example with:
//
//	cd examples/multilingual && go run main.go
package main

import (
	"errors"
	"fmt"
	"log"

	mf "github.com/kaptinlin/messageformat-go/v1"
)

// LocalizedMessage represents a message in multiple languages
type LocalizedMessage struct {
	formatters map[string]*mf.MessageFormat
	templates  map[string]string
}

// NewLocalizedMessage creates a new localized message system
func NewLocalizedMessage() *LocalizedMessage {
	return &LocalizedMessage{
		formatters: make(map[string]*mf.MessageFormat),
		templates:  make(map[string]string),
	}
}

// AddLocale adds a new locale with its message template
func (lm *LocalizedMessage) AddLocale(locale, template string) error {
	messageFormat, err := mf.New(locale, &mf.MessageFormatOptions{
		Currency: getCurrencyForLocale(locale),
	})
	if err != nil {
		return fmt.Errorf("failed to create MessageFormat for %s: %w", locale, err)
	}

	lm.formatters[locale] = messageFormat
	lm.templates[locale] = template
	return nil
}

// ErrUnsupportedLocale indicates that the requested locale is not available
var ErrUnsupportedLocale = errors.New("locale not supported")

// ErrTemplateNotFound indicates that no template exists for the locale
var ErrTemplateNotFound = errors.New("template not found for locale")

// Format formats a message for the given locale
func (lm *LocalizedMessage) Format(locale string, data map[string]interface{}) (string, error) {
	formatter, exists := lm.formatters[locale]
	if !exists {
		return "", fmt.Errorf("locale %s: %w", locale, ErrUnsupportedLocale)
	}

	template, exists := lm.templates[locale]
	if !exists {
		return "", fmt.Errorf("locale %s: %w", locale, ErrTemplateNotFound)
	}

	msg, err := formatter.Compile(template)
	if err != nil {
		return "", fmt.Errorf("failed to compile template for %s: %w", locale, err)
	}

	result, err := msg(data)
	if err != nil {
		return "", fmt.Errorf("failed to format message for %s: %w", locale, err)
	}

	return result.(string), nil
}

// getCurrencyForLocale returns appropriate currency for locale
func getCurrencyForLocale(locale string) string {
	currencies := map[string]string{
		"en":    "USD",
		"fr":    "EUR",
		"de":    "EUR",
		"ja":    "JPY",
		"ru":    "RUB",
		"es":    "EUR",
		"ar":    "USD",
		"zh":    "CNY",
		"zh-CN": "CNY",
	}

	if currency, exists := currencies[locale]; exists {
		return currency
	}
	return "USD" // Default fallback
}

func demonstratePlurals() {
	fmt.Println("\n=== Plural Rules Demonstration ===")

	// Create localized plural message
	pluralMessage := NewLocalizedMessage()

	// Add different languages with their plural templates
	locales := map[string]string{
		"en":    "{count, plural, one {# day} other {# days}}",
		"ru":    "{count, plural, one {# день} few {# дня} many {# дней} other {# дня}}",
		"ar":    "{count, plural, zero {لا أيام} one {يوم واحد} two {يومان} few {# أيام} many {# يوماً} other {# يوم}}",
		"cy":    "{count, plural, zero {dim diwrnod} one {# diwrnod} two {# ddiwrnod} few {# diwrnod} many {# diwrnod} other {# diwrnod}}",
		"zh-CN": "{count, plural, other {# 天}}",
	}

	for locale, template := range locales {
		if err := pluralMessage.AddLocale(locale, template); err != nil {
			log.Printf("Failed to add locale %s: %v", locale, err)
			continue
		}
	}

	// Test with different counts
	counts := []int{0, 1, 2, 3, 5, 10}

	for locale := range locales {
		fmt.Printf("\n%s locale:\n", locale)
		for _, count := range counts {
			result, err := pluralMessage.Format(locale, map[string]interface{}{
				"count": count,
			})
			if err != nil {
				log.Printf("Error formatting %s with count %d: %v", locale, count, err)
				continue
			}
			fmt.Printf("  Count %d: %s\n", count, result)
		}
	}
}

func demonstrateComplexMessages() {
	fmt.Println("\n\n=== Complex Message Demonstration ===")

	// Create notification system for different languages
	notificationMessage := NewLocalizedMessage()

	// Complex templates with gender, plurals, and selections
	templates := map[string]string{
		"en": `{gender, select,
			male {He has {itemCount, plural, =0 {no items} one {# item} other {# items}}}
			female {She has {itemCount, plural, =0 {no items} one {# item} other {# items}}}
			other {They have {itemCount, plural, =0 {no items} one {# item} other {# items}}}
		} in {location, select, cart {the shopping cart} wishlist {the wishlist} other {their list}}.`,

		"fr": `{gender, select,
			male {Il a {itemCount, plural, =0 {aucun article} one {# article} other {# articles}}}
			female {Elle a {itemCount, plural, =0 {aucun article} one {# article} other {# articles}}}
			other {Ils ont {itemCount, plural, =0 {aucun article} one {# article} other {# articles}}}
		} dans {location, select, cart {le panier} wishlist {la liste de souhaits} other {leur liste}}.`,

		"es": `{gender, select,
			male {Él tiene {itemCount, plural, =0 {ningún artículo} one {# artículo} other {# artículos}}}
			female {Ella tiene {itemCount, plural, =0 {ningún artículo} one {# artículo} other {# artículos}}}
			other {Tienen {itemCount, plural, =0 {ningún artículo} one {# artículo} other {# artículos}}}
		} en {location, select, cart {el carrito} wishlist {la lista de deseos} other {su lista}}.`,

		"zh-CN": `{gender, select,
			male {他在{location, select, cart {购物车} wishlist {心愿单} other {列表}}里有{itemCount, plural, =0 {没有物品} other {# 个物品}}}
			female {她在{location, select, cart {购物车} wishlist {心愿单} other {列表}}里有{itemCount, plural, =0 {没有物品} other {# 个物品}}}
			other {他们在{location, select, cart {购物车} wishlist {心愿单} other {列表}}里有{itemCount, plural, =0 {没有物品} other {# 个物品}}}
		}。`,
	}

	for locale, template := range templates {
		if err := notificationMessage.AddLocale(locale, template); err != nil {
			log.Printf("Failed to add locale %s: %v", locale, err)
			continue
		}
	}

	// Test scenarios
	scenarios := []struct {
		gender    string
		itemCount int
		location  string
	}{
		{"male", 0, "cart"},
		{"female", 1, "wishlist"},
		{"other", 5, "cart"},
		{"male", 3, "other"},
	}

	for locale := range templates {
		fmt.Printf("\n%s locale:\n", locale)
		for _, scenario := range scenarios {
			result, err := notificationMessage.Format(locale, map[string]interface{}{
				"gender":    scenario.gender,
				"itemCount": scenario.itemCount,
				"location":  scenario.location,
			})
			if err != nil {
				log.Printf("Error: %v", err)
				continue
			}
			fmt.Printf("  %s, %d items, %s: %s\n",
				scenario.gender, scenario.itemCount, scenario.location, result)
		}
	}
}

func demonstrateReturnTypes() {
	fmt.Println("\n\n=== Return Types Demonstration ===")

	// String return type (default)
	stringMF, err := mf.New("en", &mf.MessageFormatOptions{
		ReturnType: mf.ReturnTypeString,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Values array return type
	valuesMF, err := mf.New("en", &mf.MessageFormatOptions{
		ReturnType: mf.ReturnTypeValues,
	})
	if err != nil {
		log.Fatal(err)
	}

	template := "Hello {name}, you have {count} new messages!"

	// Compile for both return types
	stringMsg, err := stringMF.Compile(template)
	if err != nil {
		log.Fatal(err)
	}

	valuesMsg, err := valuesMF.Compile(template)
	if err != nil {
		log.Fatal(err)
	}

	data := map[string]interface{}{
		"name":  "Alice",
		"count": 3,
	}

	// String result
	stringResult, err := stringMsg(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("String result: %s\n", stringResult)

	// Values result
	valuesResult, err := valuesMsg(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Values result: %#v\n", valuesResult)
}

func main() {
	fmt.Println("=== Multilingual MessageFormat Examples ===")

	// Demonstrate plural rules across languages
	demonstratePlurals()

	// Demonstrate complex nested messages
	demonstrateComplexMessages()

	// Demonstrate different return types
	demonstrateReturnTypes()

	fmt.Println("\n=== Multilingual examples completed successfully! ===")
}
