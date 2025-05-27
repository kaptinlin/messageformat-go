package functions

import (
	"strconv"
	"time"

	"github.com/golang-module/carbon/v2"
	"github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// DatetimeFunction implements the :datetime function (DRAFT)
// datetime accepts a Date, number or string as its input
// and formats it with the same options as Intl.DateTimeFormat.
//
// TypeScript original code:
// export const datetime = (
//
//	ctx: MessageFunctionContext,
//	options: Record<string, unknown>,
//	operand?: unknown
//
// ): MessageDateTime =>
//
//	dateTimeImplementation(ctx, operand, res => {
//	  let hasStyle = false;
//	  let hasFields = false;
//	  for (const [name, value] of Object.entries(options)) {
//	    if (value === undefined) continue;
//	    try {
//	      switch (name) {
//	        case 'locale':
//	          break;
//	        case 'fractionalSecondDigits':
//	          res[name] = asPositiveInteger(value);
//	          hasFields = true;
//	          break;
//	        case 'hour12':
//	          res[name] = asBoolean(value);
//	          break;
//	        default:
//	          res[name] = asString(value);
//	          if (!hasStyle && styleOptions.has(name)) hasStyle = true;
//	          if (!hasFields && fieldOptions.has(name)) hasFields = true;
//	      }
//	    } catch {
//	      const msg = `Value ${value} is not valid for :datetime ${name} option`;
//	      ctx.onError(new MessageResolutionError('bad-option', msg, ctx.source));
//	    }
//	  }
//	  if (!hasStyle && !hasFields) {
//	    res.dateStyle = 'medium';
//	    res.timeStyle = 'short';
//	  } else if (hasStyle && hasFields) {
//	    const msg = 'Style and field options cannot be both set for :datetime';
//	    throw new MessageResolutionError('bad-option', msg, ctx.source);
//	  }
//	});
func DatetimeFunction(
	ctx MessageFunctionContext,
	options map[string]interface{},
	operand interface{},
) messagevalue.MessageValue {
	source := ctx.Source()

	// Parse input to time.Time
	dateTime, err := parseDateTime(operand)
	if err != nil {
		ctx.OnError(errors.NewBadOperandError("Input is not a date", source))
		return messagevalue.NewFallbackValue(source, getFirstLocale(ctx.Locales()))
	}

	// Create carbon instance
	c := carbon.CreateFromStdTime(dateTime)

	// Set locale if available
	locale := getFirstLocale(ctx.Locales())
	if locale != "" {
		c = c.SetLocale(locale)
		// Note: Carbon's SetLocale returns a new instance, but we continue using the original
		// This is intentional as the locale setting affects the formatting behavior
	}

	// Process options to determine format
	hasStyle := false
	hasFields := false
	dateStyle := ""
	timeStyle := ""

	for name, value := range options {
		if value == nil {
			continue
		}

		switch name {
		case "locale":
			// Already handled above
			continue
		case "dateStyle":
			if strval, err := asString(value); err == nil {
				dateStyle = strval
				hasStyle = true
			} else {
				msg := "Value " + toString(value) + " is not valid for :datetime " + name + " option"
				ctx.OnError(errors.NewBadOptionError(msg, source))
			}
		case "timeStyle":
			if strval, err := asString(value); err == nil {
				timeStyle = strval
				hasStyle = true
			} else {
				msg := "Value " + toString(value) + " is not valid for :datetime " + name + " option"
				ctx.OnError(errors.NewBadOptionError(msg, source))
			}
		case "fractionalSecondDigits":
			if _, err := asPositiveInteger(value); err != nil {
				msg := "Value " + toString(value) + " is not valid for :datetime " + name + " option"
				ctx.OnError(errors.NewBadOptionError(msg, source))
			}
			hasFields = true
		case "weekday", "era", "year", "month", "day", "hour", "minute", "second", "timeZoneName":
			hasFields = true
		case "hour12":
			if _, err := asBoolean(value); err != nil {
				msg := "Value " + toString(value) + " is not valid for :datetime " + name + " option"
				ctx.OnError(errors.NewBadOptionError(msg, source))
			}
		default:
			if _, err := asString(value); err != nil {
				msg := "Value " + toString(value) + " is not valid for :datetime " + name + " option"
				ctx.OnError(errors.NewBadOptionError(msg, source))
			}
		}
	}

	// Set default styles if no style or field options provided
	if !hasStyle && !hasFields {
		dateStyle = "medium"
		timeStyle = "short"
	} else if hasStyle && hasFields {
		msg := "Style and field options cannot be both set for :datetime"
		ctx.OnError(errors.NewBadOptionError(msg, source))
		return messagevalue.NewFallbackValue(source, locale)
	}

	// Create options map for the DateTimeValue
	dtOptions := make(map[string]interface{})
	if dateStyle != "" {
		dtOptions["dateStyle"] = dateStyle
	}
	if timeStyle != "" {
		dtOptions["timeStyle"] = timeStyle
	}

	// Copy other relevant options
	for name, value := range options {
		switch name {
		case "hour12", "calendar", "timeZone", "fractionalSecondDigits":
			dtOptions[name] = value
		}
	}

	return messagevalue.NewDateTimeValue(dateTime, locale, source, dtOptions)
}

// DateFunction implements the :date function (DRAFT)
// date accepts a Date, number or string as its input
// and formats it according to a single "style" option.
func DateFunction(
	ctx MessageFunctionContext,
	options map[string]interface{},
	operand interface{},
) messagevalue.MessageValue {
	source := ctx.Source()

	// Parse input to time.Time
	dateTime, err := parseDateTime(operand)
	if err != nil {
		ctx.OnError(errors.NewBadOperandError("Input is not a date", source))
		return messagevalue.NewFallbackValue(source, getFirstLocale(ctx.Locales()))
	}

	// Create carbon instance
	c := carbon.CreateFromStdTime(dateTime)

	// Set locale if available
	locale := getFirstLocale(ctx.Locales())
	if locale != "" {
		c = c.SetLocale(locale)
	}

	// Process options
	style := "medium" // default
	for name, value := range options {
		if value == nil {
			continue
		}

		switch name {
		case "style":
			if strval, err := asString(value); err == nil {
				style = strval
			} else {
				msg := "Value " + toString(value) + " is not valid for :date " + name + " option"
				ctx.OnError(errors.NewBadOptionError(msg, source))
			}
		case "hour12", "calendar", "timeZone":
			// These are valid options but not implemented yet
		default:
			msg := "Value " + toString(value) + " is not valid for :date " + name + " option"
			ctx.OnError(errors.NewBadOptionError(msg, source))
		}
	}

	// Format the date
	formatted := messagevalue.FormatDateWithStyle(*c, style)
	return messagevalue.NewStringValue(formatted, source, locale)
}

// TimeFunction implements the :time function (DRAFT)
// time accepts a Date, number or string as its input
// and formats it according to a single "style" option.
func TimeFunction(
	ctx MessageFunctionContext,
	options map[string]interface{},
	operand interface{},
) messagevalue.MessageValue {
	source := ctx.Source()

	// Parse input to time.Time
	dateTime, err := parseDateTime(operand)
	if err != nil {
		ctx.OnError(errors.NewBadOperandError("Input is not a date", source))
		return messagevalue.NewFallbackValue(source, getFirstLocale(ctx.Locales()))
	}

	// Create carbon instance
	c := carbon.CreateFromStdTime(dateTime)

	// Set locale if available
	locale := getFirstLocale(ctx.Locales())
	if locale != "" {
		c = c.SetLocale(locale)
	}

	// Process options
	style := "short" // default
	for name, value := range options {
		if value == nil {
			continue
		}

		switch name {
		case "style":
			if strval, err := asString(value); err == nil {
				style = strval
			} else {
				msg := "Value " + toString(value) + " is not valid for :time " + name + " option"
				ctx.OnError(errors.NewBadOptionError(msg, source))
			}
		case "hour12", "calendar", "timeZone":
			// These are valid options but not implemented yet
		default:
			msg := "Value " + toString(value) + " is not valid for :time " + name + " option"
			ctx.OnError(errors.NewBadOptionError(msg, source))
		}
	}

	// Format the time
	formatted := messagevalue.FormatTimeWithStyle(*c, style)
	return messagevalue.NewStringValue(formatted, source, locale)
}

// parseDateTime converts various input types to time.Time
func parseDateTime(input interface{}) (time.Time, error) {
	// Handle MessageValue types (e.g., from :datetime function)
	if mv, ok := input.(messagevalue.MessageValue); ok {
		if mv.Type() == "datetime" {
			// For datetime values, try to get the underlying value
			if val, err := mv.ValueOf(); err == nil {
				return parseDateTime(val)
			}
		}
		// For other MessageValue types, try to get the underlying value
		if val, err := mv.ValueOf(); err == nil {
			return parseDateTime(val)
		}
		// If we can't get the value, try to parse the string representation
		if str, err := mv.ToString(); err == nil {
			return parseDateTime(str)
		}
	}

	switch v := input.(type) {
	case time.Time:
		return v, nil
	case int:
		return time.Unix(int64(v), 0), nil
	case int64:
		return time.Unix(v, 0), nil
	case float64:
		return time.Unix(int64(v), 0), nil
	case string:
		// Try parsing various ISO 8601 formats (matches TypeScript Date constructor behavior)
		formats := []string{
			time.RFC3339,                    // 2006-01-02T15:04:05Z07:00
			time.RFC3339Nano,                // 2006-01-02T15:04:05.999999999Z07:00
			"2006-01-02T15:04:05",           // 2006-01-02T15:04:05 (ISO 8601 without timezone)
			"2006-01-02T15:04:05.000",       // 2006-01-02T15:04:05.000
			"2006-01-02T15:04:05.000Z",      // 2006-01-02T15:04:05.000Z
			"2006-01-02T15:04:05.000000",    // 2006-01-02T15:04:05.000000
			"2006-01-02T15:04:05.000000Z",   // 2006-01-02T15:04:05.000000Z
			"2006-01-02T15:04:05.000000000", // 2006-01-02T15:04:05.000000000
			"2006-01-02T15:04:05Z",          // 2006-01-02T15:04:05Z
			"2006-01-02",                    // 2006-01-02 (date only)
			"2006-01-02 15:04:05",           // 2006-01-02 15:04:05
			"2006-01-02 15:04:05.000",       // 2006-01-02 15:04:05.000
		}

		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t, nil
			}
		}

		// Try parsing as Unix timestamp string
		if timestamp, err := strconv.ParseInt(v, 10, 64); err == nil {
			return time.Unix(timestamp, 0), nil
		}

		// If all parsing attempts fail, return a more descriptive error
		return time.Time{}, errors.NewBadOperandError("Cannot parse date string: "+v, "")
	default:
		return time.Time{}, errors.NewBadOperandError("Invalid date input type", "")
	}
}
