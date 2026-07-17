package v1

import (
	"fmt"
	"regexp"
	"slices"
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

// PluralProfile describes caller-supplied plural behavior for one locale.
// TypeScript original code:
// getPlural(pluralFunction)
type PluralProfile struct {
	Locale    string
	Select    PluralFunction
	Cardinals []PluralCategory
	Ordinals  []PluralCategory
}

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

// GetPlural returns the PluralObject for a given locale.
// TypeScript original code:
// export function getPlural(locale: string | PluralFunction): PluralObject | null
func GetPlural(locale string) (PluralObject, error) {
	if _, err := parseStrictLocale(locale); err != nil {
		return PluralObject{}, WrapInvalidLocale(locale)
	}
	normalized, err := normalize(locale)
	if err != nil {
		return PluralObject{}, fmt.Errorf("failed to normalize locale %s: %w", locale, err)
	}

	pluralFunc, cardinals, ordinals, supported := getPluralRules(normalized)

	// Preserve variants only when their normalized locale has plural data.
	preserveLocale := supported
	localeName := locale
	if !preserveLocale {
		localeName = defaultLocale
	}

	return PluralObject{
		IsDefault: normalized == defaultLocale,
		ID:        localeName,
		LC:        localeName,
		Locale:    localeName,
		Func:      pluralFunc,
		Cardinals: cardinals,
		Ordinals:  ordinals,
		Module:    fmt.Sprintf("make-plural/%s", normalized),
	}, nil
}

// newCustomPlural creates plural metadata from a caller-supplied profile.
// TypeScript original code:
// getPlural(pluralFunction)
func newCustomPlural(profile PluralProfile) PluralObject {
	return PluralObject{
		IsDefault: false,
		ID:        profile.Locale,
		LC:        profile.Locale,
		Locale:    profile.Locale,
		Func:      profile.Select,
		Cardinals: slices.Clone(profile.Cardinals),
		Ordinals:  slices.Clone(profile.Ordinals),
	}
}

// validatePluralProfile validates caller-supplied plural facts at construction.
// TypeScript original code:
// getPlural(pluralFunction)
func validatePluralProfile(profile PluralProfile) error {
	if _, err := parseStrictLocale(profile.Locale); err != nil {
		return WrapInvalidLocale(profile.Locale)
	}
	if profile.Select == nil {
		return ErrInvalidPluralFunction
	}
	if err := validatePluralCategories("cardinals", profile.Cardinals); err != nil {
		return err
	}
	return validatePluralCategories("ordinals", profile.Ordinals)
}

// validatePluralCategories validates one complete plural category set.
// TypeScript original code:
// cardinals: locale.cardinals || []; ordinals: locale.ordinals || [];
func validatePluralCategories(name string, categories []PluralCategory) error {
	if len(categories) == 0 {
		return fmt.Errorf("%w: %s are empty", ErrInvalidPluralCategories, name)
	}

	seen := make(map[PluralCategory]struct{}, len(categories))
	for _, category := range categories {
		switch category {
		case PluralZero, PluralOne, PluralTwo, PluralFew, PluralMany, PluralOther:
		default:
			return fmt.Errorf("%w: %s contains %q", ErrInvalidPluralCategories, name, category)
		}
		if _, exists := seen[category]; exists {
			return fmt.Errorf("%w: %s contains duplicate %q", ErrInvalidPluralCategories, name, category)
		}
		seen[category] = struct{}{}
	}
	if _, exists := seen[PluralOther]; !exists {
		return fmt.Errorf("%w: %s omit %q", ErrInvalidPluralCategories, name, PluralOther)
	}
	return nil
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
		category := rules.Select(pluralrules.Int(num))
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

// hasPluralLocale verifies that go-intl has CLDR plural data for the parsed
// locale. Lookup matching prevents best-fit from expanding the supported set.
func hasPluralLocale(loc locale.Locale) bool {
	lookup := string(pluralrules.LookupLocaleMatcher)
	supported, err := pluralrules.SupportedLocalesOf(
		[]locale.Locale{loc},
		pluralrules.Options{LocaleMatcher: &lookup},
	)
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
