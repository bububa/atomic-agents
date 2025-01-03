package functions

import (
	"math"
)

func Floor(args ...interface{}) (interface{}, error) {
	if len(args) == 1 {
		v, ok := ToFloat64(args[0])
		if !ok {
			return math.NaN(), NewWrongParamType(0)
		}
		return math.Floor(v), nil
	}
	return nil, WrongParamsCount
}

func Ceil(args ...interface{}) (interface{}, error) {
	if len(args) == 1 {
		v, ok := ToFloat64(args[0])
		if !ok {
			return math.NaN(), NewWrongParamType(0)
		}
		return math.Ceil(v), nil
	}
	return nil, WrongParamsCount
}

func Round(args ...interface{}) (interface{}, error) {
	switch len(args) {
	case 1:
		v, ok := ToFloat64(args[0])
		if !ok {
			return math.NaN(), NewWrongParamType(0)
		}
		if math.IsNaN(v) {
			return math.NaN(), nil
		}
		if v >= 0.5 {
			return math.Trunc(v + 0.5), nil
		} else if v <= -0.5 {
			return math.Trunc(v - 0.5), nil
		}
		return 0.0, nil
	case 2:
		v1, ok := ToFloat64(args[0])
		if !ok {
			return math.NaN(), NewWrongParamType(0)
		}
		v2, ok := ToFloat64(args[1])
		if !ok {
			return math.NaN(), NewWrongParamType(1)
		}

		var round float64
		pow := math.Pow(10, float64(v2))
		digit := pow * v1
		_, div := math.Modf(digit)
		if div >= 0.5 {
			round = math.Ceil(digit)
		} else {
			round = math.Floor(digit)
		}
		return round / pow, nil
	}
	return nil, WrongParamsCount
}

func Trunc(args ...interface{}) (interface{}, error) {
	if len(args) == 1 {
		v, ok := ToFloat64(args[0])
		if !ok {
			return math.NaN(), NewWrongParamType(0)
		}
		return math.Trunc(v), nil
	}
	return nil, WrongParamsCount
}
