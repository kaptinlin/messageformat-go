package intlbridge

import (
	"strings"

	"github.com/agentable/go-intl/datetimeformat"
)

// DateTimeOptions translates MessageFormat 2.0's loose datetime option bag
// into the typed datetimeformat.Options expected by go-intl.
//
// MF2 accepts two coexisting option flavours:
//
//   - Legacy "style" form (`dateStyle`, `timeStyle`) mirroring ECMA-402 directly.
//   - LDML 48 form (`dateFields`, `dateLength`, `timePrecision`,
//     `timeZoneStyle`) which lists the visible fields and a length hint.
//
// Both are mapped here. go-intl rejects mixing `DateStyle`/`TimeStyle` with
// per-field options, so legacy and LDML 48 inputs are kept mutually exclusive
// in the output. Unknown options are silently dropped, matching the MF2 spec.
func DateTimeOptions(opts map[string]any) datetimeformat.Options {
	out := datetimeformat.Options{}
	if len(opts) == 0 {
		return out
	}

	for name, raw := range opts {
		switch name {
		case "calendar":
			if s, ok := asOptString(raw); ok {
				out.Calendar = s
			}
		case "numberingSystem":
			if s, ok := asOptString(raw); ok {
				out.NumberingSystem = s
			}
		case "localeMatcher":
			if s, ok := asOptString(raw); ok {
				out.LocaleMatcher = datetimeformat.LocaleMatcher(s)
			}
		case "timeZone":
			if s, ok := asOptString(raw); ok {
				out.TimeZone = s
			}
		case "hour12":
			if b, ok := raw.(bool); ok {
				out.Hour12 = boolPtr(b)
			}
		}
	}

	hasLegacy := applyLegacyStyle(&out, opts)
	if !hasLegacy {
		applyLdmlFields(&out, opts)
	}
	return out
}

// applyLegacyStyle handles `dateStyle` / `timeStyle`. Returns true when at
// least one was applied so the caller skips the LDML field path.
func applyLegacyStyle(out *datetimeformat.Options, opts map[string]any) bool {
	applied := false
	if s, ok := asOptString(opts["dateStyle"]); ok {
		out.DateStyle = datetimeformat.Style(s)
		applied = true
	}
	if s, ok := asOptString(opts["timeStyle"]); ok {
		out.TimeStyle = datetimeformat.Style(s)
		applied = true
	}
	return applied
}

// applyLdmlFields expands LDML 48 options into go-intl's per-field option set.
//
// `dateFields` enumerates which date components are visible (weekday, year,
// month, day in any combination). `dateLength` ("long"/"medium"/"short")
// controls month/weekday rendering style. `timePrecision` ("hour"/"minute"/
// "second") controls how many time components are visible. `timeZoneStyle`
// maps onto `TimeZoneName`.
func applyLdmlFields(out *datetimeformat.Options, opts map[string]any) {
	if fields, ok := asOptString(opts["dateFields"]); ok {
		length, _ := asOptString(opts["dateLength"])
		applyDateFields(out, fields, length)
	}
	if precision, ok := asOptString(opts["timePrecision"]); ok {
		applyTimePrecision(out, precision)
	}
	if tz, ok := asOptString(opts["timeZoneStyle"]); ok {
		switch tz {
		case "long":
			out.TimeZoneName = datetimeformat.LongTimeZoneName
		case "short":
			out.TimeZoneName = datetimeformat.ShortTimeZoneName
		}
	}
}

func applyDateFields(out *datetimeformat.Options, fields, length string) {
	set := make(map[string]bool)
	for f := range strings.SplitSeq(fields, "-") {
		set[f] = true
	}
	switch length {
	case "long":
		if set["weekday"] {
			out.Weekday = datetimeformat.LongFieldStyle
		}
		if set["month"] {
			out.Month = datetimeformat.LongMonthStyle
		}
	case "short":
		if set["weekday"] {
			out.Weekday = datetimeformat.ShortFieldStyle
		}
		if set["month"] {
			out.Month = datetimeformat.NumericMonthStyle
		}
	default: // "medium" and unset
		if set["weekday"] {
			out.Weekday = datetimeformat.ShortFieldStyle
		}
		if set["month"] {
			out.Month = datetimeformat.ShortMonthStyle
		}
	}
	if set["year"] {
		out.Year = datetimeformat.NumericFieldStyle
	}
	if set["day"] {
		out.Day = datetimeformat.NumericFieldStyle
	}
}

func applyTimePrecision(out *datetimeformat.Options, precision string) {
	switch precision {
	case "hour":
		out.Hour = datetimeformat.NumericFieldStyle
	case "second":
		out.Hour = datetimeformat.NumericFieldStyle
		out.Minute = datetimeformat.NumericFieldStyle
		out.Second = datetimeformat.NumericFieldStyle
	default: // "minute"
		out.Hour = datetimeformat.NumericFieldStyle
		out.Minute = datetimeformat.NumericFieldStyle
	}
}

func boolPtr(v bool) *bool {
	return &v
}
