package functions

func IsDefined(args ...interface{}) (interface{}, error) {
	if len(args) == 1 {
		if args[0] != nil {
			switch v := args[0].(type) {
			case string:
				return v != "", nil
			case float32, float64:
				return v != 0.0, nil
			case int, uint, int16, uint16, int32, uint32, int64, uint64:
				return v != 0, nil
			default:
				return nil, NewWrongParamType(0)
			}
		}
		return false, nil
	}
	return nil, WrongParamsCount
}
