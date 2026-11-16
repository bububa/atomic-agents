package functions

import (
	"math"
)

func Log(args ...any) (any, error) {
	if len(args) == 1 {
		v, ok := ToFloat64(args[0])
		if !ok {
			return math.NaN(), NewWrongParamType(0)
		}
		return math.Log(v), nil
	}
	return nil, WrongParamsCount
}

func Log10(args ...any) (any, error) {
	if len(args) == 1 {
		v, ok := ToFloat64(args[0])
		if !ok {
			return math.NaN(), NewWrongParamType(0)
		}
		return math.Log10(v), nil
	}
	return nil, WrongParamsCount
}
