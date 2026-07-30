package mf1

import (
	"fmt"
	"math"
	"math/big"
	"strconv"

	"github.com/agentable/go-intl/numberformat"
	"github.com/kaptinlin/messageformat-go/mf1/internal/intlbridge"
)

type compiledNumber struct {
	value any // int64, uint64, float64, or *big.Int after offset subtraction
}

func newCompiledNumber(value any) (compiledNumber, error) {
	switch value := value.(type) {
	case int:
		return compiledNumber{value: int64(value)}, nil
	case int8:
		return compiledNumber{value: int64(value)}, nil
	case int16:
		return compiledNumber{value: int64(value)}, nil
	case int32:
		return compiledNumber{value: int64(value)}, nil
	case int64:
		return compiledNumber{value: value}, nil
	case uint:
		return compiledNumber{value: uint64(value)}, nil
	case uint8:
		return compiledNumber{value: uint64(value)}, nil
	case uint16:
		return compiledNumber{value: uint64(value)}, nil
	case uint32:
		return compiledNumber{value: uint64(value)}, nil
	case uint64:
		return compiledNumber{value: value}, nil
	case float32:
		return finiteCompiledNumber(float64(value), value)
	case float64:
		return finiteCompiledNumber(value, value)
	case string:
		parsed, err := parseNumericString(value)
		if err != nil {
			return compiledNumber{}, err
		}
		return compiledNumber{value: parsed}, nil
	default:
		return compiledNumber{}, WrapInvalidType(fmt.Sprintf("%T", value))
	}
}

func parseNumericString(value string) (any, error) {
	if integer, ok := new(big.Int).SetString(value, 10); ok {
		if integer.IsInt64() {
			return integer.Int64(), nil
		}
		if integer.IsUint64() {
			return integer.Uint64(), nil
		}
		return integer, nil
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, WrapInvalidNumberStr(value)
	}
	number, err := finiteCompiledNumber(parsed, value)
	if err != nil {
		return nil, err
	}
	return number.value, nil
}

func finiteCompiledNumber(value float64, original any) (compiledNumber, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return compiledNumber{}, WrapInvalidNumberValue(original)
	}
	return compiledNumber{value: value}, nil
}

func (number compiledNumber) exactKey() string {
	switch value := number.value.(type) {
	case int64:
		return "=" + strconv.FormatInt(value, 10)
	case uint64:
		return "=" + strconv.FormatUint(value, 10)
	case float64:
		return formatExactKey(value)
	case *big.Int:
		return "=" + value.String()
	default:
		return ""
	}
}

func (number compiledNumber) subtract(offset int) compiledNumber {
	if offset == 0 {
		return number
	}

	switch value := number.value.(type) {
	case int64:
		result := new(big.Int).Sub(big.NewInt(value), big.NewInt(int64(offset)))
		if result.IsInt64() {
			return compiledNumber{value: result.Int64()}
		}
		return compiledNumber{value: result}
	case uint64:
		result := new(big.Int).Sub(new(big.Int).SetUint64(value), big.NewInt(int64(offset)))
		if result.IsUint64() {
			return compiledNumber{value: result.Uint64()}
		}
		if result.IsInt64() {
			return compiledNumber{value: result.Int64()}
		}
		return compiledNumber{value: result}
	case float64:
		return compiledNumber{value: value - float64(offset)}
	case *big.Int:
		return compiledNumber{value: new(big.Int).Sub(value, big.NewInt(int64(offset)))}
	default:
		return number
	}
}

func (number compiledNumber) format(locale string) (string, error) {
	formatter, err := numberformat.New(intlbridge.ParseLocale(locale), numberformat.Options{})
	if err != nil {
		return "", fmt.Errorf("create number formatter: %w", err)
	}

	switch value := number.value.(type) {
	case int64:
		return formatter.Format(numberformat.Int(value)), nil
	case uint64:
		return formatter.Format(numberformat.Uint(value)), nil
	case float64:
		return formatter.Format(numberformat.Float(value)), nil
	case *big.Int:
		return formatter.Format(numberformat.BigInt(value)), nil
	default:
		return "", WrapInvalidType(fmt.Sprintf("%T", number.value))
	}
}
