package functions

import (
	"math"
)

func Exp(args ...interface{}) (interface{}, error) {
	if len(args) == 1 {
		v, ok := ToFloat64(args[0])
		if !ok {
			return math.NaN(), NewWrongParamType(0)
		}
		return math.Exp(v), nil
	}
	return nil, WrongParamsCount
}

func Exp2(args ...interface{}) (interface{}, error) {
	if len(args) == 1 {
		v, ok := ToFloat64(args[0])
		if !ok {
			return math.NaN(), NewWrongParamType(0)
		}
		return math.Exp2(v), nil
	}
	return nil, WrongParamsCount
}
