package functions

import (
	"strconv"
	"time"

	"github.com/kaptinlin/messageformat-go/pkg/errors"
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// Valid option values matching TypeScript datetime.ts:37-47
var (
	dateFieldsValues = map[string]bool{
		"weekday":                true,
		"day-weekday":            true,
		"month-day":              true,
		"month-day-weekday":      true,
		"year-month-day":         true,
		"year-month-day-weekday": true,
	}

	dateLengthValues = map[string]bool{
		"long":   true,
		"medium": true,
		"short":  true,
	}

	timePrecisionValues = map[string]bool{
		"hour":   true,
		"minute": true,
		"second": true,
	}

	timeZoneStyleValues = map[string]bool{
		"long":  true,
		"short": true,
	}
)

// readStringOption reads and validates a string option
// TypeScript reference: datetime.ts:244-261
func readStringOption(
	ctx MessageFunctionContext,
	options map[string]any,
	name string,
	allowed map[string]bool,
) string {
	value, ok := options[name]
	if !ok || value == nil {
		return ""
	}

	strVal, err := asString(value)
	if err != nil {
		ctx.OnError(errors.NewBadOptionError("Invalid value for "+name+" option", ctx.Source()))
		return ""
	}

	if allowed != nil && !allowed[strVal] {
		ctx.OnError(errors.NewBadOptionError("Invalid value for "+name+" option", ctx.Source()))
		return ""
	}

	return strVal
}

// dateTimeImplementation provides unified implementation for date, datetime, and time functions
// This matches the TypeScript refactor in commit df3b997d
// TypeScript reference: datetime.ts:84-242
func dateTimeImplementation(
	functionName string, // 'datetime' | 'date' | 'time'
	ctx MessageFunctionContext,
	exprOpt map[string]any,
	operand any,
) messagevalue.MessageValue {
	source := ctx.Source()
	locale := GetFirstLocale(ctx.Locales())

	// Parse datetime value (matches TypeScript lines 94-112)
	dateTime, err := parseDateTimeValue(operand)
	if err != nil {
		ctx.OnError(errors.NewMessageFunctionError("bad-operand", "Input is not a valid date"))
		return messagevalue.NewFallbackValue(source, locale)
	}

	// Build datetime format options (TS: line 90-92)
	dtOptions := make(map[string]any)
	dtOptions["localeMatcher"] = ctx.LocaleMatcher()

	// Extract options from operand if present (TS: lines 95-101)
	if opMap, ok := operand.(map[string]any); ok {
		if opts, hasOpts := opMap["options"]; hasOpts {
			if optsMap, ok := opts.(map[string]any); ok {
				if cal, ok := optsMap["calendar"]; ok {
					dtOptions["calendar"] = cal
				}
				if functionName != "date" {
					if h12, ok := optsMap["hour12"]; ok {
						dtOptions["hour12"] = h12
					}
				}
				if tz, ok := optsMap["timeZone"]; ok {
					dtOptions["timeZone"] = tz
				}
			}
		}
	}

	// Override calendar option from expression (TS: lines 115-124)
	if cal, ok := exprOpt["calendar"]; ok && cal != nil {
		if strVal, err := asString(cal); err == nil {
			dtOptions["calendar"] = strVal
		} else {
			ctx.OnError(errors.NewBadOptionError(
				"Invalid :"+functionName+" calendar option value", source))
		}
	}

	// Override hour12 option from expression (TS: lines 125-131)
	// Only applies to datetime and time, not date
	if h12, ok := exprOpt["hour12"]; ok && h12 != nil && functionName != "date" {
		if boolVal, err := asBoolean(h12); err == nil {
			dtOptions["hour12"] = boolVal
		} else {
			ctx.OnError(errors.NewBadOptionError(
				"Invalid :"+functionName+" hour12 option value", source))
		}
	}

	// Override timeZone option from expression (TS: lines 132-159)
	if tz, ok := exprOpt["timeZone"]; ok && tz != nil {
		if tzStr, err := asString(tz); err == nil {
			if tzStr == "input" {
				// Validate input timezone exists (TS: lines 142-148)
				if _, hasInputTZ := dtOptions["timeZone"]; !hasInputTZ {
					ctx.OnError(errors.NewBadOperandError(
						"Missing input timeZone value for :"+functionName, source))
				}
			} else {
				// Check for timezone conversion (TS: lines 150-157)
				if existingTZ, hasTZ := dtOptions["timeZone"]; hasTZ {
					if existingTZStr, ok := existingTZ.(string); ok && existingTZStr != tzStr {
						ctx.OnError(errors.NewMessageFunctionError(
							"bad-option", "Time zone conversion is not supported"))
						return messagevalue.NewFallbackValue(source, locale)
					}
				}
				dtOptions["timeZone"] = tzStr
			}
		} else {
			ctx.OnError(errors.NewBadOptionError(
				"Invalid :"+functionName+" timeZone option value", source))
		}
	}

	// Date formatting options (TypeScript lines 161-184)
	// Only applies to datetime and date, not time
	if functionName != "time" {
		// Option names depend on function type
		// TypeScript: const dfName = functionName === 'date' ? 'fields' : 'dateFields'
		fieldsName := "dateFields"
		lengthName := "dateLength"
		if functionName == "date" {
			fieldsName = "fields"
			lengthName = "length"
		}

		// Read dateFields with default value
		// TypeScript: const dateFieldsValue = readStringOption(...) ?? 'year-month-day'
		dateFields := readStringOption(ctx, exprOpt, fieldsName, dateFieldsValues)
		if dateFields == "" {
			dateFields = "year-month-day" // TypeScript default
		}
		dtOptions["dateFields"] = dateFields

		// Read dateLength (optional, no default)
		dateLength := readStringOption(ctx, exprOpt, lengthName, dateLengthValues)
		if dateLength != "" {
			dtOptions["dateLength"] = dateLength
		}
	}

	// Time formatting options (TypeScript lines 186-209)
	// Only applies to datetime and time, not date
	if functionName != "date" {
		// Option name depends on function type
		// TypeScript: const tpName = functionName === 'time' ? 'precision' : 'timePrecision'
		precisionName := "timePrecision"
		if functionName == "time" {
			precisionName = "precision"
		}

		// Read timePrecision (optional, defaults to 'minute' in formatting)
		// TypeScript: switch (readStringOption(...))
		timePrecision := readStringOption(ctx, exprOpt, precisionName, timePrecisionValues)
		if timePrecision != "" {
			dtOptions["timePrecision"] = timePrecision
		}

		// Read timeZoneStyle (optional)
		// TypeScript: options.timeZoneName = readStringOption(...)
		timeZoneStyle := readStringOption(ctx, exprOpt, "timeZoneStyle", timeZoneStyleValues)
		if timeZoneStyle != "" {
			dtOptions["timeZoneStyle"] = timeZoneStyle
		}
	}

	// Return DateTimeValue with all options
	// This will be formatted using the new helper functions
	return messagevalue.NewDateTimeValue(dateTime, locale, source, dtOptions)
}

// parseDateTimeValue extracts time.Time from various operand types
// Handles both plain values and operands with options/valueOf
// TypeScript reference: datetime.ts:94-112
func parseDateTimeValue(operand any) (time.Time, error) {
	value := operand

	// Check if operand is an object with options and/or valueOf
	// TypeScript: if (typeof value === 'object' && value !== null)
	if operand != nil {
		if opMap, ok := operand.(map[string]any); ok {
			// Extract valueOf if present
			// TypeScript: if (typeof value.valueOf === 'function') value = value.valueOf()
			if valueOf, ok := opMap["valueOf"]; ok {
				value = valueOf
			}
			// Note: We don't extract options from operand here
			// because we build options from expression options instead
		}
	}

	// Parse the value to time.Time
	// TypeScript converts number/string to Date, then validates
	return parseDateTime(value)
}

// DatetimeFunction implements the :datetime function
// Formats both date and time portions
// TypeScript reference: datetime.ts:55-59
func DatetimeFunction(
	ctx MessageFunctionContext,
	options map[string]any,
	operand any,
) messagevalue.MessageValue {
	return dateTimeImplementation("datetime", ctx, options, operand)
}

// DateFunction implements the :date function
// Formats only the date portion
// TypeScript reference: datetime.ts:66-70
func DateFunction(
	ctx MessageFunctionContext,
	options map[string]any,
	operand any,
) messagevalue.MessageValue {
	return dateTimeImplementation("date", ctx, options, operand)
}

// TimeFunction implements the :time function
// Formats only the time portion
// TypeScript reference: datetime.ts:78-82
func TimeFunction(
	ctx MessageFunctionContext,
	options map[string]any,
	operand any,
) messagevalue.MessageValue {
	return dateTimeImplementation("time", ctx, options, operand)
}

// parseDateTime converts various input types to time.Time
func parseDateTime(input any) (time.Time, error) {
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
