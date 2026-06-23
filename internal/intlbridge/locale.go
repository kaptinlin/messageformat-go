// Package intlbridge translates MessageFormat 2.0 runtime types (string locales
// and map[string]any options) into the typed value objects expected by
// github.com/agentable/go-intl constructors. Centralizing the translation here
// keeps the conversion logic out of pkg/messagevalue and pkg/functions and gives
// us a single place to add behavior tests.
package intlbridge

import (
	"strings"

	"github.com/agentable/go-intl/locale"
)

// ParseLocale parses a BCP 47 locale tag and returns it wrapped in a
// locale.List so it can be passed directly to go-intl constructors, which all
// take locale.List as of v0.2.3. Underscores are normalized to hyphens to
// accept POSIX-style tags (`en_US`) that MF2 callers historically passed.
// Empty strings and unparseable tags fall back to English so callers never
// have to handle a sentinel error: the MessageFormat 2.0 resolution flow
// already treats unknown locales as recoverable, and the TS reference uses
// `new Intl.Locale("en")` as the implicit fallback.
func ParseLocale(tag string) locale.List {
	if tag == "" {
		return englishLocaleList()
	}
	normalized := strings.ReplaceAll(tag, "_", "-")
	if loc, err := locale.Parse(normalized); err == nil {
		return locale.List{loc}
	}
	return englishLocaleList()
}

// FirstLocale picks the first non-empty tag from a slice and wraps it in a
// locale.List. MessageFormat 2.0 accepts a locale list but only one is active
// per format call.
func FirstLocale(tags []string) locale.List {
	for _, t := range tags {
		if t != "" {
			return ParseLocale(t)
		}
	}
	return englishLocaleList()
}

func englishLocaleList() locale.List {
	loc, _ := locale.Parse("en")
	return locale.List{loc}
}
