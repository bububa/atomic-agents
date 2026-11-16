package functions

import (
	"math"
)

func Cos(args ...any) (any, error) {
	if len(args) == 1 {
		v, ok := ToFloat64(args[0])
		if !ok {
			return math.NaN(), NewWrongParamType(0)
		}
		return math.Cos(v), nil
	}
	return nil, WrongParamsCount
}

func Cosh(args ...any) (any, error) {
	if len(args) == 1 {
		v, ok := ToFloat64(args[0])
		if !ok {
			return math.NaN(), NewWrongParamType(0)
		}
		return math.Cosh(v), nil
	}
	return nil, WrongParamsCount
}

func Acos(args ...any) (any, error) {
	if len(args) == 1 {
		v, ok := ToFloat64(args[0])
		if !ok {
			return math.NaN(), NewWrongParamType(0)
		}
		return math.Acos(v), nil
	}
	return nil, WrongParamsCount
}

func Acosh(args ...any) (any, error) {
	if len(args) == 1 {
		v, ok := ToFloat64(args[0])
		if !ok {
			return math.NaN(), NewWrongParamType(0)
		}
		return math.Acosh(v), nil
	}
	return nil, WrongParamsCount
}
