package functions

func Max(args ...interface{}) (interface{}, error) {
	if len(args) > 0 {
		var floatMax float64
		for idx, arg := range args {
			v, ok := ToFloat64(arg)
			if !ok {
				return nil, NewWrongParamType(idx)
			}
			if idx == 0 {
				floatMax = v
			} else if v > floatMax {
				floatMax = v
			}
		}
		return floatMax, nil
	}
	return nil, WrongParamsCount
}

func Min(args ...interface{}) (interface{}, error) {
	if len(args) > 0 {
		var floatMax float64
		for idx, arg := range args {
			v, ok := ToFloat64(arg)
			if !ok {
				return nil, NewWrongParamType(idx)
			}
			if idx == 0 {
				floatMax = v
			} else if v < floatMax {
				floatMax = v
			}
		}
		return floatMax, nil
	}
	return nil, WrongParamsCount
}
