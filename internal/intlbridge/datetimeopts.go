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
// Both are mapped here. When callers mix `DateStyle`/`TimeStyle` with per-field
// options, both reach go-intl so its typed validation can reject the conflict.
// Unknown options are silently dropped, matching the MF2 spec.
func DateTimeOptions(opts map[string]any) datetimeformat.Options {
	out := datetimeformat.Options{}
	if len(opts) == 0 {
		return out
	}

	for name, raw := range opts {
		switch name {
		case "calendar":
			if s, ok := asOptString(raw); ok {
				out.Calendar = stringPtr(s)
			}
		case "numberingSystem":
			if s, ok := asOptString(raw); ok {
				out.NumberingSystem = stringPtr(s)
			}
		case "localeMatcher":
			if s, ok := asOptString(raw); ok {
				out.LocaleMatcher = stringPtr(s)
			}
		case "timeZone":
			if s, ok := asOptString(raw); ok {
				out.TimeZone = stringPtr(s)
			}
		case "hour12":
			if b, ok := raw.(bool); ok {
				out.Hour12 = boolPtr(b)
			}
		}
	}

	applyLegacyStyle(&out, opts)
	applyLdmlFields(&out, opts)
	return out
}

// applyLegacyStyle preserves `dateStyle` / `timeStyle` for dependency validation.
// TypeScript original code:
// options.dateStyle = input.dateStyle; options.timeStyle = input.timeStyle;
func applyLegacyStyle(out *datetimeformat.Options, opts map[string]any) {
	if s, ok := asOptString(opts["dateStyle"]); ok {
		out.DateStyle = stringPtr(s)
	}
	if s, ok := asOptString(opts["timeStyle"]); ok {
		out.TimeStyle = stringPtr(s)
	}
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
			out.TimeZoneName = stringPtr(string(datetimeformat.LongTimeZoneName))
		case "short":
			out.TimeZoneName = stringPtr(string(datetimeformat.ShortTimeZoneName))
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
			out.Weekday = stringPtr(string(datetimeformat.LongFieldStyle))
		}
		if set["month"] {
			out.Month = stringPtr(string(datetimeformat.LongMonthStyle))
		}
	case "short":
		if set["weekday"] {
			out.Weekday = stringPtr(string(datetimeformat.ShortFieldStyle))
		}
		if set["month"] {
			out.Month = stringPtr(string(datetimeformat.NumericMonthStyle))
		}
	default: // "medium" and unset
		if set["weekday"] {
			out.Weekday = stringPtr(string(datetimeformat.ShortFieldStyle))
		}
		if set["month"] {
			out.Month = stringPtr(string(datetimeformat.ShortMonthStyle))
		}
	}
	if set["year"] {
		out.Year = stringPtr(string(datetimeformat.NumericFieldStyle))
	}
	if set["day"] {
		out.Day = stringPtr(string(datetimeformat.NumericFieldStyle))
	}
}

func applyTimePrecision(out *datetimeformat.Options, precision string) {
	switch precision {
	case "hour":
		out.Hour = stringPtr(string(datetimeformat.NumericFieldStyle))
	case "second":
		out.Hour = stringPtr(string(datetimeformat.NumericFieldStyle))
		out.Minute = stringPtr(string(datetimeformat.NumericFieldStyle))
		out.Second = stringPtr(string(datetimeformat.NumericFieldStyle))
	default: // "minute"
		out.Hour = stringPtr(string(datetimeformat.NumericFieldStyle))
		out.Minute = stringPtr(string(datetimeformat.NumericFieldStyle))
	}
}

func boolPtr(v bool) *bool {
	return &v
}
