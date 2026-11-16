package functions

import (
	"math"
)

func Sin(args ...any) (any, error) {
	if len(args) == 1 {
		v, ok := ToFloat64(args[0])
		if !ok {
			return math.NaN(), NewWrongParamType(0)
		}
		return math.Sin(v), nil
	}
	return nil, WrongParamsCount
}

func Sinh(args ...any) (any, error) {
	if len(args) == 1 {
		v, ok := ToFloat64(args[0])
		if !ok {
			return math.NaN(), NewWrongParamType(0)
		}
		return math.Sinh(v), nil
	}
	return nil, WrongParamsCount
}

func Asin(args ...any) (any, error) {
	if len(args) == 1 {
		v, ok := ToFloat64(args[0])
		if !ok {
			return math.NaN(), NewWrongParamType(0)
		}
		return math.Asin(v), nil
	}
	return nil, WrongParamsCount
}

func Asinh(args ...any) (any, error) {
	if len(args) == 1 {
		v, ok := ToFloat64(args[0])
		if !ok {
			return math.NaN(), NewWrongParamType(0)
		}
		return math.Asinh(v), nil
	}
	return nil, WrongParamsCount
}
