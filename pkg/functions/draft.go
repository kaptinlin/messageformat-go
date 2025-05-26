package functions

import (
	"github.com/kaptinlin/messageformat-go/pkg/messagevalue"
)

// CurrencyFunction implements the :currency function (DRAFT)
func CurrencyFunction(
	ctx MessageFunctionContext,
	options map[string]interface{},
	operand interface{},
) messagevalue.MessageValue {
	// TODO: Implement currency formatting
	// For now, delegate to number function
	return NumberFunction(ctx, options, operand)
}

// DateFunction implements the :date function (DRAFT)
func DateFunction(
	ctx MessageFunctionContext,
	options map[string]interface{},
	operand interface{},
) messagevalue.MessageValue {
	// TODO: Implement date formatting
	return messagevalue.NewFallbackValue(ctx.Source(), getFirstLocale(ctx.Locales()))
}

// DatetimeFunction implements the :datetime function (DRAFT)
func DatetimeFunction(
	ctx MessageFunctionContext,
	options map[string]interface{},
	operand interface{},
) messagevalue.MessageValue {
	// TODO: Implement datetime formatting
	return messagevalue.NewFallbackValue(ctx.Source(), getFirstLocale(ctx.Locales()))
}

// TimeFunction implements the :time function (DRAFT)
func TimeFunction(
	ctx MessageFunctionContext,
	options map[string]interface{},
	operand interface{},
) messagevalue.MessageValue {
	// TODO: Implement time formatting
	return messagevalue.NewFallbackValue(ctx.Source(), getFirstLocale(ctx.Locales()))
}

// MathFunction implements the :math function (DRAFT)
func MathFunction(
	ctx MessageFunctionContext,
	options map[string]interface{},
	operand interface{},
) messagevalue.MessageValue {
	// TODO: Implement math operations
	return NumberFunction(ctx, options, operand)
}

// UnitFunction implements the :unit function (DRAFT)
func UnitFunction(
	ctx MessageFunctionContext,
	options map[string]interface{},
	operand interface{},
) messagevalue.MessageValue {
	// TODO: Implement unit formatting
	return NumberFunction(ctx, options, operand)
}
