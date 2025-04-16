package schema

type String struct {
	Schema
	value string
}

func NewString(str string) *String {
	return &String{
		value: str,
	}
}

func (s String) String() string {
	return string(s.value)
}

func (s String) Bytes() []byte {
	return []byte(s.value)
}

func (s *String) Unmarshal(bs []byte) error {
	*s = String{value: string(bs)}
	return nil
}
