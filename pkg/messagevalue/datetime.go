package messagevalue

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/agentable/go-intl/datetimeformat"
	"github.com/kaptinlin/messageformat-go/internal/intlbridge"
	"github.com/kaptinlin/messageformat-go/pkg/bidi"
)

// ErrInvalidDateTimeOptions identifies a datetime plan rejected during construction.
var ErrInvalidDateTimeOptions = errors.New("invalid datetime format options")

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
	value     time.Time
	locale    string
	dir       bidi.Direction
	source    string
	options   map[string]any
	calendar  string
	timeZone  string
	formatter *datetimeformat.DateTimeFormat
}

// NewDateTimeValue creates a validated datetime value.
// TypeScript original code:
// const formatter = new Intl.DateTimeFormat(locales, options);
func NewDateTimeValue(value time.Time, locale, source string, options map[string]any) (*DateTimeValue, error) {
	return NewDateTimeValueWithDir(value, locale, source, bidi.DirAuto, options)
}

// NewDateTimeValueWithDir creates a validated datetime value with explicit direction.
// TypeScript original code:
// const formatter = new Intl.DateTimeFormat(locales, options);
func NewDateTimeValueWithDir(value time.Time, locale, source string, dir bidi.Direction, options map[string]any) (*DateTimeValue, error) {
	formatOptions := intlbridge.DateTimeOptions(options)
	if formatOptions.TimeZone == nil {
		timeZone, ok := timeZoneFromValue(value)
		if !ok {
			return nil, fmt.Errorf("%w: input time-zone offset must use whole minutes", ErrInvalidDateTimeOptions)
		}
		formatOptions.TimeZone = stringPtr(timeZone)
	}
	formatter, err := datetimeformat.New(intlbridge.ParseLocale(locale), formatOptions)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidDateTimeOptions, err)
	}
	resolved := formatter.ResolvedOptions()
	return &DateTimeValue{
		value:     value,
		locale:    resolved.Locale.String(),
		dir:       dir,
		source:    source,
		options:   cloneOptions(options),
		calendar:  resolved.Calendar,
		timeZone:  resolved.TimeZone,
		formatter: formatter,
	}, nil
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

// Calendar returns the dependency-resolved calendar identifier.
// TypeScript original code:
// formatter.resolvedOptions().calendar;
func (dtv *DateTimeValue) Calendar() string {
	return dtv.calendar
}

// TimeZone returns the dependency-resolved time-zone identifier.
// TypeScript original code:
// formatter.resolvedOptions().timeZone;
func (dtv *DateTimeValue) TimeZone() string {
	return dtv.timeZone
}

func (dtv *DateTimeValue) ToString() (string, error) {
	return dtv.formatter.Format(dtv.value), nil
}

func (dtv *DateTimeValue) ToParts() ([]MessagePart, error) {
	intlParts := dtv.formatter.FormatToParts(dtv.value)
	formatted := dtv.formatter.Format(dtv.value)
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
			value:    formatted,
			source:   dtv.source,
			locale:   dtv.locale,
			dir:      dtv.dir,
			calendar: dtv.calendar,
			timeZone: dtv.timeZone,
			parts:    sub,
		},
	}, nil
}

func (dtv *DateTimeValue) ValueOf() (any, error) {
	return dtv.value, nil
}

func (dtv *DateTimeValue) Time() time.Time {
	return dtv.value
}

// timeZoneFromValue derives a datetimeformat-compatible time zone string from
// a time.Time without changing its wall-clock semantics.
// TypeScript original code:
// options.timeZone = input.options?.timeZone;
func timeZoneFromValue(t time.Time) (string, bool) {
	loc := t.Location()
	if loc == nil || loc == time.UTC {
		return "UTC", true
	}
	name := loc.String()
	if name == "UTC" || name == "GMT" {
		return "UTC", true
	}
	if name != "Local" && strings.Contains(name, "/") {
		return name, true
	}

	_, offset := t.Zone()
	if offset%60 != 0 {
		return "", false
	}
	if offset == 0 {
		return "UTC", true
	}
	sign := '+'
	if offset < 0 {
		sign = '-'
		offset = -offset
	}
	return fmt.Sprintf("%c%02d:%02d", sign, offset/3600, offset%3600/60), true
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
	value    string
	source   string
	locale   string
	dir      bidi.Direction
	calendar string
	timeZone string
	parts    []MessagePart
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

// Calendar returns the dependency-resolved calendar identifier.
// TypeScript original code:
// formatter.resolvedOptions().calendar;
func (dtp *DateTimePart) Calendar() string {
	return dtp.calendar
}

// TimeZone returns the dependency-resolved time-zone identifier.
// TypeScript original code:
// formatter.resolvedOptions().timeZone;
func (dtp *DateTimePart) TimeZone() string {
	return dtp.timeZone
}

func (dtp *DateTimePart) Parts() []MessagePart {
	return slices.Clone(dtp.parts)
}
