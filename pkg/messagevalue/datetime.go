package messagevalue

import (
	"time"

	"github.com/agentable/go-intl/datetimeformat"
	"github.com/kaptinlin/messageformat-go/internal/intlbridge"
	"github.com/kaptinlin/messageformat-go/pkg/bidi"
)

// DateTimeValue implements MessageValue for date/time values.
// Formatting is delegated to go-intl's datetimeformat (ECMA-402 compliant);
// MF2's option shape is normalised by intlbridge.DateTimeOptions before being
// handed off.
//
// TypeScript original code:
//
//	export interface MessageDateTime extends MessageValue<'datetime'> {
//	  readonly type: 'datetime';
//	  readonly source: string;
//	  readonly dir: 'ltr' | 'rtl' | 'auto';
//	  readonly options: Readonly<Intl.DateTimeFormatOptions>;
//	  toParts(): [MessageDateTimePart];
//	  toString(): string;
//	  valueOf(): Date;
//	}
type DateTimeValue struct {
	value   time.Time
	locale  string
	dir     bidi.Direction
	source  string
	options map[string]any
}

// NewDateTimeValue creates a new datetime value
func NewDateTimeValue(value time.Time, locale, source string, options map[string]any) *DateTimeValue {
	return NewDateTimeValueWithDir(value, locale, source, bidi.DirAuto, options)
}

// NewDateTimeValueWithDir creates a new datetime value with explicit direction
func NewDateTimeValueWithDir(value time.Time, locale, source string, dir bidi.Direction, options map[string]any) *DateTimeValue {
	return &DateTimeValue{
		value:   value,
		locale:  locale,
		dir:     dir,
		source:  source,
		options: cloneOptions(options),
	}
}

func (dtv *DateTimeValue) Type() string {
	return "datetime"
}

func (dtv *DateTimeValue) Source() string {
	return dtv.source
}

func (dtv *DateTimeValue) Dir() bidi.Direction {
	return dtv.dir
}

func (dtv *DateTimeValue) Locale() string {
	return dtv.locale
}

func (dtv *DateTimeValue) Options() map[string]any {
	return cloneOptions(dtv.options)
}

func (dtv *DateTimeValue) ToString() (string, error) {
	formatter, ok := dtv.newFormatter()
	if !ok {
		return dtv.value.Format(time.RFC3339), nil
	}
	return formatter.Format(dtv.value), nil
}

func (dtv *DateTimeValue) ToParts() ([]MessagePart, error) {
	formatter, ok := dtv.newFormatter()
	if !ok {
		return []MessagePart{
			&DateTimePart{
				value:  dtv.value.Format(time.RFC3339),
				source: dtv.source,
				locale: dtv.locale,
				dir:    dtv.dir,
			},
		}, nil
	}
	intlParts := formatter.FormatToParts(dtv.value)
	formatted := formatter.Format(dtv.value)
	sub := make([]MessagePart, 0, len(intlParts))
	for _, p := range intlParts {
		sub = append(sub, &DateTimeSubPart{
			partType: string(p.Type),
			value:    p.Value,
			source:   dtv.source,
			locale:   dtv.locale,
			dir:      dtv.dir,
		})
	}
	return []MessagePart{
		&DateTimePart{
			value:  formatted,
			source: dtv.source,
			locale: dtv.locale,
			dir:    dtv.dir,
			parts:  sub,
		},
	}, nil
}

func (dtv *DateTimeValue) ValueOf() (any, error) {
	return dtv.value, nil
}

func (dtv *DateTimeValue) Time() time.Time {
	return dtv.value
}

// newFormatter builds a datetimeformat.DateTimeFormat from the MF2 options.
// Falls back to a no-options formatter if go-intl rejects the typed options
// (mirrors the graceful fallback behavior used by NumberValue).
//
// When no explicit timeZone option is provided, the formatter uses the input
// value's own location. ECMA-402 defaults to the runtime time zone, which would
// surprise callers that intentionally constructed UTC- or fixed-offset times;
// MF2 senders typically expect the wall-clock the value was created with.
func (dtv *DateTimeValue) newFormatter() (*datetimeformat.DateTimeFormat, bool) {
	loc := intlbridge.ParseLocale(dtv.locale)
	opts := intlbridge.DateTimeOptions(dtv.options)
	if opts.TimeZone == nil {
		if timeZone := timeZoneFromValue(dtv.value); timeZone != "" {
			opts.TimeZone = stringPtr(timeZone)
		}
	}
	if f, err := datetimeformat.New(loc, opts); err == nil {
		return f, true
	}
	if f, err := datetimeformat.New(loc, datetimeformat.Options{}); err == nil {
		return f, true
	}
	return nil, false
}

// timeZoneFromValue derives a datetimeformat-compatible time zone string from
// a time.Time. Returns "" for the system-local location so the formatter falls
// back to its own default; named locations and fixed-offset zones are passed
// through directly.
func timeZoneFromValue(t time.Time) string {
	loc := t.Location()
	if loc == nil || loc == time.Local {
		return ""
	}
	name := loc.String()
	if name == "" || name == "Local" {
		return ""
	}
	return name
}

// DateTimeSubPart represents a sub-part of a formatted datetime (year, month,
// hour, literal, etc.), matching the parts emitted by go-intl.
type DateTimeSubPart struct {
	partType string
	value    string
	source   string
	locale   string
	dir      bidi.Direction
}

func (dsp *DateTimeSubPart) Type() string        { return dsp.partType }
func (dsp *DateTimeSubPart) Value() any          { return dsp.value }
func (dsp *DateTimeSubPart) Text() string        { return dsp.value }
func (dsp *DateTimeSubPart) Source() string      { return dsp.source }
func (dsp *DateTimeSubPart) Locale() string      { return dsp.locale }
func (dsp *DateTimeSubPart) Dir() bidi.Direction { return dsp.dir }

// DateTimePart implements MessagePart for datetime parts
type DateTimePart struct {
	value  string
	source string
	locale string
	dir    bidi.Direction
	parts  []MessagePart
}

func (dtp *DateTimePart) Type() string {
	return "datetime"
}

func (dtp *DateTimePart) Value() any {
	return dtp.value
}

func (dtp *DateTimePart) Text() string {
	return dtp.value
}

func (dtp *DateTimePart) Source() string {
	return dtp.source
}

func (dtp *DateTimePart) Locale() string {
	return dtp.locale
}

func (dtp *DateTimePart) Dir() bidi.Direction {
	return dtp.dir
}

func (dtp *DateTimePart) Parts() []MessagePart {
	return dtp.parts
}
