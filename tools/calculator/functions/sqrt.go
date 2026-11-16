package functions

import (
	"math"
)

func Sqrt(args ...any) (any, error) {
	if len(args) == 1 {
		v, ok := ToFloat64(args[0])
		if !ok {
			return math.NaN(), NewWrongParamType(0)
		}
		return math.Sqrt(v), nil
	}
	return nil, WrongParamsCount
}
