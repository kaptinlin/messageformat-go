package messagevalue

import (
	"time"

	"github.com/golang-module/carbon/v2"
	"github.com/kaptinlin/messageformat-go/pkg/bidi"
)

// DateTimeValue implements MessageValue for date/time values
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
	options map[string]interface{}
}

// NewDateTimeValue creates a new datetime value
func NewDateTimeValue(value time.Time, locale, source string, options map[string]interface{}) *DateTimeValue {
	if options == nil {
		options = make(map[string]interface{})
	}

	return &DateTimeValue{
		value:   value,
		locale:  locale,
		dir:     bidi.DirAuto,
		source:  source,
		options: options,
	}
}

// NewDateTimeValueWithDir creates a new datetime value with explicit direction
func NewDateTimeValueWithDir(value time.Time, locale, source string, dir bidi.Direction, options map[string]interface{}) *DateTimeValue {
	if options == nil {
		options = make(map[string]interface{})
	}

	return &DateTimeValue{
		value:   value,
		locale:  locale,
		dir:     dir,
		source:  source,
		options: options,
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

func (dtv *DateTimeValue) Options() map[string]interface{} {
	return dtv.options
}

func (dtv *DateTimeValue) ToString() (string, error) {
	// Format the datetime using the options
	return dtv.formatDateTime()
}

func (dtv *DateTimeValue) ToParts() ([]MessagePart, error) {
	// Format the datetime to get the string representation
	formattedValue, err := dtv.formatDateTime()
	if err != nil {
		return nil, err
	}

	// Create a datetime part
	return []MessagePart{
		&DateTimePart{
			value:  formattedValue,
			source: dtv.source,
			locale: dtv.locale,
			dir:    dtv.dir,
		},
	}, nil
}

func (dtv *DateTimeValue) ValueOf() (interface{}, error) {
	return dtv.value, nil
}

func (dtv *DateTimeValue) SelectKeys(keys []string) ([]string, error) {
	// DateTime values typically don't support selection
	return []string{}, nil
}

// formatDateTime formats the datetime according to the options
func (dtv *DateTimeValue) formatDateTime() (string, error) {
	// Create carbon instance
	c := carbon.CreateFromStdTime(dtv.value)

	// Set locale if available, with fallback for unsupported locales
	if dtv.locale != "" {
		normalizedLocale := normalizeLocaleForCarbon(dtv.locale)
		if normalizedLocale != "" {
			c = c.SetLocale(normalizedLocale)
		}
	}

	// Extract formatting options
	dateStyle, hasDateStyle := dtv.options["dateStyle"].(string)
	timeStyle, hasTimeStyle := dtv.options["timeStyle"].(string)

	// Format based on options
	switch {
	case hasDateStyle && hasTimeStyle:
		return FormatDateTimeWithStyle(*c, dateStyle, timeStyle), nil
	case hasDateStyle:
		return FormatDateWithStyle(*c, dateStyle), nil
	case hasTimeStyle:
		return FormatTimeWithStyle(*c, timeStyle), nil
	default:
		// Default formatting
		return c.ToDateTimeString(), nil
	}
}

// DateTimePart implements MessagePart for datetime parts
type DateTimePart struct {
	value  string
	source string
	locale string
	dir    bidi.Direction
}

func (dtp *DateTimePart) Type() string {
	return "datetime"
}

func (dtp *DateTimePart) Value() interface{} {
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

// Helper functions for formatting (these should be moved to a shared location)
func FormatDateTimeWithStyle(c carbon.Carbon, dateStyle, timeStyle string) string {
	dateFormat := GetDateFormat(dateStyle)
	timeFormat := GetTimeFormat(timeStyle)
	return c.Format(dateFormat + " " + timeFormat)
}

func FormatDateWithStyle(c carbon.Carbon, style string) string {
	format := GetDateFormat(style)
	return c.Format(format)
}

func FormatTimeWithStyle(c carbon.Carbon, style string) string {
	format := GetTimeFormat(style)
	return c.Format(format)
}

func GetDateFormat(style string) string {
	switch style {
	case "full":
		return "l, F j, Y" // Monday, January 2, 2006
	case "long":
		return "F j, Y" // January 2, 2006
	case "medium":
		return "M j, Y" // Jan 2, 2006
	case "short":
		return "n/j/y" // 1/2/06
	default:
		return "M j, Y" // default to medium
	}
}

func GetTimeFormat(style string) string {
	switch style {
	case "full":
		return "g:i:s A T" // 3:04:05 PM MST
	case "long":
		return "g:i:s A T" // 3:04:05 PM MST
	case "medium":
		return "g:i:s A" // 3:04:05 PM
	case "short":
		return "g:i A" // 3:04 PM
	default:
		return "g:i A" // default to short
	}
}

// normalizeLocaleForCarbon normalizes locale strings for Carbon compatibility
// Carbon doesn't support all locale formats, so we need to normalize them
func normalizeLocaleForCarbon(locale string) string {
	// Handle common locale patterns
	switch locale {
	case "en-US", "en_US":
		return "en"
	case "zh-CN", "zh_CN":
		return "zh"
	case "zh-TW", "zh_TW":
		return "zh"
	case "es-ES", "es_ES":
		return "es"
	case "fr-FR", "fr_FR":
		return "fr"
	case "de-DE", "de_DE":
		return "de"
	case "ja-JP", "ja_JP":
		return "ja"
	case "ko-KR", "ko_KR":
		return "ko"
	case "pt-BR", "pt_BR":
		return "pt"
	case "ru-RU", "ru_RU":
		return "ru"
	case "ar-SA", "ar_SA":
		return "ar"
	default:
		// For other locales, try to extract the language part
		if len(locale) >= 3 {
			if locale[2] == '-' || locale[2] == '_' {
				return locale[:2]
			}
		}
		// Return as-is if it's already a simple language code
		return locale
	}
}
