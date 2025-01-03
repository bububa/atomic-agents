package functions

func If(args ...interface{}) (interface{}, error) {
	if len(args) == 3 {
		condition := false
		if args[0] != nil {
			switch cond := args[0].(type) {
			case bool:
				condition = cond
			case string, float64, float32, int, uint, int8, int16, int32, int64, uint8, uint16, uint32, uint64:
				c, err := IsDefined(args[0])
				if err != nil {
					return nil, err
				}
				condition = c.(bool)
			default:
				return nil, NewWrongParamType(0)
			}
		}
		if condition {
			return args[1], nil
		} else {
			return args[2], nil
		}
	}
	return nil, WrongParamsCount
}
