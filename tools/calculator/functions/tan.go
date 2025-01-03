package functions

import (
	"math"
)

func Tan(args ...interface{}) (interface{}, error) {
	if len(args) == 1 {
		v, ok := ToFloat64(args[0])
		if !ok {
			return math.NaN(), NewWrongParamType(0)
		}
		return math.Tan(v), nil
	}
	return nil, WrongParamsCount
}

func Tanh(args ...interface{}) (interface{}, error) {
	if len(args) == 1 {
		v, ok := ToFloat64(args[0])
		if !ok {
			return math.NaN(), NewWrongParamType(0)
		}
		return math.Tanh(v), nil
	}
	return nil, WrongParamsCount
}

func Atan(args ...interface{}) (interface{}, error) {
	if len(args) == 1 {
		v, ok := ToFloat64(args[0])
		if !ok {
			return math.NaN(), NewWrongParamType(0)
		}
		return math.Atan(v), nil
	}
	return nil, WrongParamsCount
}

func Atan2(args ...interface{}) (interface{}, error) {
	if len(args) == 2 {
		v1, ok := ToFloat64(args[0])
		if !ok {
			return math.NaN(), NewWrongParamType(0)
		}
		v2, ok := ToFloat64(args[1])
		if !ok {
			return math.NaN(), NewWrongParamType(1)
		}
		return math.Atan2(v1, v2), nil
	}
	return nil, WrongParamsCount
}

func Atanh(args ...interface{}) (interface{}, error) {
	if len(args) == 1 {
		v, ok := ToFloat64(args[0])
		if !ok {
			return math.NaN(), NewWrongParamType(0)
		}
		return math.Atanh(v), nil
	}
	return nil, WrongParamsCount
}
