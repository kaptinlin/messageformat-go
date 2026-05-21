package intlbridge

import (
	"fmt"

	"github.com/agentable/go-intl/numberformat"
)

// NumberOptions translates MessageFormat 2.0's loose map[string]any option bag
// into the typed numberformat.Options expected by go-intl.
//
// MessageFormat 2.0 functions in pkg/functions emit a `map[string]any` whose
// values are already type-coerced (strings for enums, ints for digit counts).
// This bridge is the single place that knows how to map each option name onto
// the corresponding numberformat field, including the handful of legacy values
// MF2 historically accepted but go-intl does not (e.g. useGrouping "never" /
// "true", or bool useGrouping).
//
// Unknown options are silently ignored — go-intl validates the typed options
// itself, and the MF2 spec requires unknown options to be passed through
// without error.
func NumberOptions(opts map[string]any) numberformat.Options {
	out := numberformat.Options{}
	if len(opts) == 0 {
		return out
	}

	minFrac, hasMinFrac := -1, false
	maxFrac, hasMaxFrac := -1, false
	minSig, hasMinSig := -1, false
	maxSig, hasMaxSig := -1, false

	for name, raw := range opts {
		switch name {
		case "style":
			if s, ok := asOptString(raw); ok {
				out.Style = numberformat.Style(s)
			}
		case "currency":
			if s, ok := asOptString(raw); ok {
				out.Currency = numberformat.CurrencyCode(s)
			}
		case "currencyDisplay":
			if s, ok := asOptString(raw); ok {
				out.CurrencyDisplay = numberformat.CurrencyDisplay(s)
			}
		case "currencySign":
			if s, ok := asOptString(raw); ok {
				out.CurrencySign = numberformat.CurrencySign(s)
			}
		case "unit":
			if s, ok := asOptString(raw); ok {
				out.Unit = numberformat.UnitIdentifier(s)
			}
		case "unitDisplay":
			if s, ok := asOptString(raw); ok {
				out.UnitDisplay = numberformat.UnitDisplay(s)
			}
		case "notation":
			if s, ok := asOptString(raw); ok {
				out.Notation = numberformat.Notation(s)
			}
		case "compactDisplay":
			if s, ok := asOptString(raw); ok {
				out.CompactDisplay = numberformat.CompactDisplay(s)
			}
		case "signDisplay":
			if s, ok := asOptString(raw); ok {
				out.SignDisplay = numberformat.SignDisplay(s)
			}
		case "roundingMode":
			if s, ok := asOptString(raw); ok {
				out.RoundingMode = numberformat.RoundingMode(s)
			}
		case "roundingPriority":
			if s, ok := asOptString(raw); ok {
				out.RoundingPriority = numberformat.RoundingPriority(s)
			}
		case "trailingZeroDisplay":
			if s, ok := asOptString(raw); ok {
				out.TrailingZeroDisplay = numberformat.TrailingZeroDisplay(s)
			}
		case "numberingSystem":
			if s, ok := asOptString(raw); ok {
				out.NumberingSystem = s
			}
		case "localeMatcher":
			if s, ok := asOptString(raw); ok {
				out.LocaleMatcher = numberformat.LocaleMatcher(s)
			}
		case "useGrouping":
			out.UseGrouping = useGroupingFromAny(raw)
		case "minimumIntegerDigits":
			if n, ok := asOptInt(raw); ok {
				out.MinimumIntegerDigits = intPtr(n)
			}
		case "roundingIncrement":
			if n, ok := asOptInt(raw); ok {
				out.RoundingIncrement = intPtr(n)
			}
		case "minimumFractionDigits":
			if n, ok := asOptInt(raw); ok {
				minFrac, hasMinFrac = n, true
			}
		case "maximumFractionDigits":
			if n, ok := asOptInt(raw); ok {
				maxFrac, hasMaxFrac = n, true
			}
		case "minimumSignificantDigits":
			if n, ok := asOptInt(raw); ok {
				minSig, hasMinSig = n, true
			}
		case "maximumSignificantDigits":
			if n, ok := asOptInt(raw); ok {
				maxSig, hasMaxSig = n, true
			}
		}
	}

	applyFractionDigits(&out, minFrac, hasMinFrac, maxFrac, hasMaxFrac)
	applySignificantDigits(&out, minSig, hasMinSig, maxSig, hasMaxSig)
	return out
}

// applyFractionDigits clamps max>=min before assigning the typed options.
// ECMA-402 throws RangeError when maximumFractionDigits < minimumFractionDigits,
// but MF2's :number is more permissive: callers expect the formatter to honor
// "at least N" semantics rather than abort the whole message. Clamping max=min
// preserves caller intent without producing a fallback string.
func applyFractionDigits(out *numberformat.Options, minVal int, hasMin bool, maxVal int, hasMax bool) {
	if hasMin && hasMax && maxVal < minVal {
		maxVal = minVal
	}
	if hasMin {
		out.MinimumFractionDigits = intPtr(minVal)
	}
	if hasMax {
		out.MaximumFractionDigits = intPtr(maxVal)
	}
}

func applySignificantDigits(out *numberformat.Options, minVal int, hasMin bool, maxVal int, hasMax bool) {
	if hasMin && hasMax && maxVal < minVal {
		maxVal = minVal
	}
	if hasMin {
		out.MinimumSignificantDigits = intPtr(minVal)
	}
	if hasMax {
		out.MaximumSignificantDigits = intPtr(maxVal)
	}
}

func intPtr(v int) *int {
	return &v
}

// useGroupingFromAny accepts the historical MF2 forms (bool, "never", "true",
// "false", "min2", "auto", "always") and normalises them to numberformat's
// validated enum. Anything go-intl doesn't recognise is dropped so that
// numberformat.New keeps its default ("auto") instead of erroring on us.
func useGroupingFromAny(raw any) numberformat.UseGrouping {
	switch v := raw.(type) {
	case bool:
		if v {
			return numberformat.UseGroupingAlways
		}
		return numberformat.UseGroupingFalse
	case string:
		switch v {
		case "never", "false":
			return numberformat.UseGroupingFalse
		case "true", "always":
			return numberformat.UseGroupingAlways
		case "min2":
			return numberformat.UseGroupingMin2
		case "auto":
			return numberformat.UseGroupingAuto
		}
	}
	return ""
}

func asOptString(raw any) (string, bool) {
	if raw == nil {
		return "", false
	}
	if s, ok := raw.(string); ok && s != "" {
		return s, true
	}
	return "", false
}

func asOptInt(raw any) (int, bool) {
	switch v := raw.(type) {
	case int:
		return v, true
	case int8:
		return int(v), true
	case int16:
		return int(v), true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case uint:
		return int(v), true
	case uint8:
		return int(v), true
	case uint16:
		return int(v), true
	case uint32:
		return int(v), true
	case uint64:
		return int(v), true
	case float32:
		if float32(int(v)) == v {
			return int(v), true
		}
	case float64:
		if float64(int(v)) == v {
			return int(v), true
		}
	case string:
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n, true
		}
	}
	return 0, false
}
