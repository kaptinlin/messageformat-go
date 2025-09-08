// errors.go - Common error definitions to satisfy err113 linter
package v1

import (
	"errors"
	"fmt"
)

var (
	// Common error types
	ErrInvalidLocale      = errors.New("invalid language tag")
	ErrInvalidNumberValue = errors.New("invalid number value")
	ErrInvalidDateValue   = errors.New("cannot parse date")
	ErrInvalidTimeValue   = errors.New("cannot parse time")
	ErrUnsupportedType    = errors.New("unsupported type")
	ErrInvalidNumberType  = errors.New("unsupported number type")
	ErrInvalidDateType    = errors.New("unsupported date value type")
	ErrInvalidTimeType    = errors.New("unsupported time value type")
	ErrInvalidParamType   = errors.New("invalid parameter type")
	ErrMissingParameter   = errors.New("missing parameter")
	ErrMissingArgument    = errors.New("missing required argument")
	ErrNoMatchingCase     = errors.New("no matching case found")
	ErrNoOtherCase        = errors.New("no 'other' case found")
	ErrInvalidNumberStr   = errors.New("invalid number string")
	ErrInvalidLocaleType  = errors.New("invalid locale type")
	ErrInvalidLocalesType = errors.New("invalid locales type")
)

// Helper functions to wrap errors with context
func WrapInvalidLocale(locale string) error {
	return fmt.Errorf("%w: %s", ErrInvalidLocale, locale)
}

func WrapInvalidNumberValue(value interface{}) error {
	return fmt.Errorf("%w: %v", ErrInvalidNumberValue, value)
}

func WrapInvalidDateValue(value interface{}) error {
	return fmt.Errorf("%w: %v", ErrInvalidDateValue, value)
}

func WrapInvalidTimeValue(value interface{}) error {
	return fmt.Errorf("%w: %v", ErrInvalidTimeValue, value)
}

func WrapUnsupportedType(valueType string) error {
	return fmt.Errorf("%w: %s", ErrUnsupportedType, valueType)
}

func WrapInvalidParamType(paramType string) error {
	return fmt.Errorf("%w: %s", ErrInvalidParamType, paramType)
}

func WrapMissingParameter(param string) error {
	return fmt.Errorf("%w: %s", ErrMissingParameter, param)
}

func WrapMissingArgument(arg string) error {
	return fmt.Errorf("%w: %s", ErrMissingArgument, arg)
}

func WrapNoMatchingCase(arg, selectType string) error {
	return fmt.Errorf("%w for %s in %s", ErrNoMatchingCase, arg, selectType)
}

func WrapInvalidNumberStr(str string) error {
	return fmt.Errorf("%w: %s", ErrInvalidNumberStr, str)
}

func WrapInvalidLocaleType(itemType string) error {
	return fmt.Errorf("%w: %s, expected string", ErrInvalidLocaleType, itemType)
}

func WrapInvalidLocalesType(localesType string) error {
	return fmt.Errorf("%w: %s, expected string or []string", ErrInvalidLocalesType, localesType)
}
