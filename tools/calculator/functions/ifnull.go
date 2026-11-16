package functions

func IfNull(args ...any) (any, error) {
	if len(args) > 0 {
		for _, v := range args {
			if v != nil {
				return v, nil
			}
		}
		return nil, nil
	}
	return nil, WrongParamsCount
}
