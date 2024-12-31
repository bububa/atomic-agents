package schema

type String string

func (s String) String() string {
	return string(s)
}

func (s String) Snapshot() Schema {
	return String(s.String())
}

func (s String) Attachement() *Attachement {
	return nil
}
