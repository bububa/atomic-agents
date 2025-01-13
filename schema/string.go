package schema

type String string

func (s String) MarshalJSON() ([]byte, error) {
  return []byte(s), nil
}

func (s *String) UnmarshalJSON(b []byte) error {
  *s = String(b)
  return nil
}

func (s String) Attachement() *Attachement {
	return nil
}

func (s String) SetAttachement(v *Attachement) {
}
