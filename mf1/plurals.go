package v1

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/pluralrules"
	"github.com/kaptinlin/messageformat-go/mf1/internal/intlbridge"
)

// PluralCategory represents the plural categories from CLDR
// TypeScript original code:
// export type PluralCategory = 'zero' | 'one' | 'two' | 'few' | 'many' | 'other'
type PluralCategory string

const (
	PluralZero  PluralCategory = "zero"
	PluralOne   PluralCategory = "one"
	PluralTwo   PluralCategory = "two"
	PluralFew   PluralCategory = "few"
	PluralMany  PluralCategory = "many"
	PluralOther PluralCategory = "other"
)

// PluralFunction represents a function used to define the pluralization for a locale
// TypeScript original code:
//
//	export interface PluralFunction {
//	  (value: number | string, ord?: boolean): PluralCategory;
//	  cardinals?: PluralCategory[];
//	  ordinals?: PluralCategory[];
//	  module?: string;
//	}
type PluralFunction func(value any, ord ...bool) (PluralCategory, error)

// PluralObject represents plural rules and metadata for a specific locale
// TypeScript original code:
//
//	export interface PluralObject {
//	  isDefault: boolean;
//	  id: string;
//	  lc: string;
//	  locale: string;
//	  getCardinal?: (value: string | number) => PluralCategory;
//	  getPlural: PluralFunction;
//	  cardinals: PluralCategory[];
//	  ordinals: PluralCategory[];
//	  module?: string;
//	}
type PluralObject struct {
	IsDefault   bool
	ID          string
	LC          string
	Locale      string
	GetCardinal func(value any) (PluralCategory, error)
	Func        PluralFunction
	Cardinals   []PluralCategory
	Ordinals    []PluralCategory
	Module      string
}

// Pre-compiled regex for normalize function
var normalizeRegex = regexp.MustCompile(`^([^-_]+)`)

// normalize normalizes a locale string following TypeScript implementation
// TypeScript original code:
//
//	function normalize(locale: string) {
//	  if (typeof locale !== 'string' || locale.length < 2) {
//	    throw new RangeError(`Invalid language tag: ${locale}`);
//	  }
//	  // The only locale for which anything but the primary subtag matters is
//	  // Portuguese as spoken in Portugal.
//	  if (locale.startsWith('pt-PT')) return 'pt-PT';
//	  const m = locale.match(/.+?(?=[-_])/);
//	  return m ? m[0] : locale;
//	}
func normalize(locale string) (string, error) {
	if len(locale) < 2 {
		return "", WrapInvalidLocale(locale)
	}

	if strings.HasPrefix(locale, "pt-PT") {
		return "pt-PT", nil
	}

	if matches := normalizeRegex.FindStringSubmatch(locale); len(matches) > 1 {
		return matches[1], nil
	}

	return locale, nil
}

// GetPlural returns the PluralObject for a given locale
// TypeScript original code:
// export function getPlural(locale: string | PluralFunction): PluralObject | null
func GetPlural(locale any) (PluralObject, error) {
	switch v := locale.(type) {
	case string:
		normalized, err := normalize(v)
		if err != nil {
			return PluralObject{}, fmt.Errorf("failed to normalize locale %s: %w", v, err)
		}

		pluralFunc, cardinals, ordinals, supported := getPluralRules(normalized)

		// For TypeScript compatibility, preserve original locale if it's supported or looks like a locale variant
		preserveLocale := supported || strings.Contains(v, "-") || strings.Contains(v, "_")
		localeName := v
		if !preserveLocale {
			// Fallback to English for completely unknown locales
			localeName = DefaultLocale
		}

		return PluralObject{
			IsDefault: normalized == DefaultLocale,
			ID:        localeName,
			LC:        localeName,
			Locale:    localeName,
			Func:      pluralFunc,
			Cardinals: cardinals,
			Ordinals:  ordinals,
			Module:    fmt.Sprintf("make-plural/%s", normalized),
		}, nil

	case PluralFunction:
		return PluralObject{
			IsDefault: false,
			ID:        "custom",
			LC:        "custom",
			Locale:    "custom",
			Func:      v,
			Cardinals: []PluralCategory{PluralOne, PluralOther}, // Default cardinals
			Ordinals:  []PluralCategory{PluralOther},            // Default ordinals
		}, nil

	default:
		return PluralObject{}, ErrInvalidType
	}
}

// HasPlural checks if a locale has plural support
// TypeScript original code:
// export function hasPlural(locale: string): boolean
func HasPlural(locale string) bool {
	normalized, err := normalize(locale)
	if err != nil {
		return false
	}
	return hasPlural(normalized)
}

// GetAllPlurals returns all available plurals for a given default locale
// TypeScript original code:
// export function getAllPlurals(defaultLocale: string): PluralObject[]
func GetAllPlurals(defaultLocale string) ([]PluralObject, error) {
	// For now, return common locales
	// In a full implementation, this would include all CLDR locales
	commonLocales := []string{
		"en", "fr", "de", "es", "it", "pt", "ru", "ja", "ko", "zh",
		"ar", "he", "hi", "th", "vi", "id", "ms", "tl", "sw",
	}

	plurals := make([]PluralObject, 0, len(commonLocales))
	for _, locale := range commonLocales {
		plural, err := GetPlural(locale)
		if err != nil {
			continue
		}
		plurals = append(plurals, plural)
	}

	// Ensure default locale is first
	defaultPlural, err := GetPlural(defaultLocale)
	if err == nil {
		var filtered []PluralObject
		for _, p := range plurals {
			if p.Locale != defaultLocale {
				filtered = append(filtered, p)
			}
		}
		plurals = append([]PluralObject{defaultPlural}, filtered...)
	}

	return plurals, nil
}

// getPluralRules builds the cardinal/ordinal plural function and category lists
// for a locale using go-intl's CLDR-backed pluralrules package. The locale is
// normalized to BCP 47 (POSIX underscores accepted) and falls back to English
// when go-intl cannot parse the tag.
//
// v1 historically truncates fractional values to integers before category
// selection: tests assert that float32(1.9) resolves to PluralOne, mirroring
// the legacy toNumber(int64) coercion. That semantics is preserved here.
func getPluralRules(loc string) (PluralFunction, []PluralCategory, []PluralCategory, bool) {
	parsed := intlbridge.ParseLocale(loc)
	cardinalType := string(pluralrules.Cardinal)
	ordinalType := string(pluralrules.Ordinal)
	cardinal, _ := pluralrules.New(parsed, pluralrules.Options{Type: &cardinalType})
	ordinal, _ := pluralrules.New(parsed, pluralrules.Options{Type: &ordinalType})

	pluralFunc := func(value any, ord ...bool) (PluralCategory, error) {
		num, err := toNumber(value)
		if err != nil {
			return PluralOther, err
		}
		if num < 0 {
			num = -num
		}

		rules := cardinal
		if len(ord) > 0 && ord[0] {
			rules = ordinal
		}
		if rules == nil {
			return PluralOther, nil
		}
		category, err := rules.Select(pluralrules.Int(num))
		if err != nil {
			return PluralOther, err
		}
		return mapPluralCategory(category), nil
	}

	cardinals := categoriesFromRules(cardinal)
	ordinals := categoriesFromRules(ordinal)
	// Use strict parsing for the supported flag so unknown tags like "xx"
	// don't get silently aliased to English by intlbridge.ParseLocale.
	return pluralFunc, cardinals, ordinals, hasPlural(loc)
}

// mapPluralCategory maps go-intl pluralrules categories to v1's PluralCategory.
func mapPluralCategory(c pluralrules.Category) PluralCategory {
	switch c {
	case pluralrules.Zero:
		return PluralZero
	case pluralrules.One:
		return PluralOne
	case pluralrules.Two:
		return PluralTwo
	case pluralrules.Few:
		return PluralFew
	case pluralrules.Many:
		return PluralMany
	default:
		return PluralOther
	}
}

// categoriesFromRules extracts the resolved plural categories from a
// pluralrules.PluralRules instance, normalising the order to match the v1
// surface (zero, one, two, few, many, other) and guaranteeing PluralOther as
// the final fallback.
func categoriesFromRules(r *pluralrules.PluralRules) []PluralCategory {
	if r == nil {
		return []PluralCategory{PluralOther}
	}
	resolved := r.ResolvedOptions().PluralCategories

	seen := make(map[PluralCategory]bool, len(resolved))
	for _, c := range resolved {
		seen[mapPluralCategory(c)] = true
	}

	order := []PluralCategory{PluralZero, PluralOne, PluralTwo, PluralFew, PluralMany, PluralOther}
	out := make([]PluralCategory, 0, len(order))
	for _, c := range order {
		if seen[c] {
			out = append(out, c)
		}
	}
	if len(out) == 0 || out[len(out)-1] != PluralOther {
		out = append(out, PluralOther)
	}
	return out
}

// hasPluralLocale verifies that go-intl found CLDR data for the parsed locale
// by checking that SupportedLocalesOf returns the locale unchanged.
func hasPluralLocale(loc locale.Locale) bool {
	supported, err := pluralrules.SupportedLocalesOf([]locale.Locale{loc}, pluralrules.Options{})
	if err != nil {
		return false
	}
	return len(supported) > 0
}

// parseStrictLocale parses a BCP 47 locale without intlbridge.ParseLocale's
// English fallback. Used by hasPlural to reject tags that fail the parser
// (e.g. "x") instead of silently treating them as supported.
func parseStrictLocale(tag string) (locale.Locale, error) {
	return locale.Parse(strings.ReplaceAll(tag, "_", "-"))
}

// toNumber converts various types to int64
func toNumber(value any) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		num, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, WrapInvalidNumberStr(v)
		}
		return int64(num), nil
	default:
		return 0, WrapInvalidType(fmt.Sprintf("%T", value))
	}
}
