package schema

type String string

func (s String) Attachement() *Attachement {
	return nil
}

func (s String) SetAttachement(v *Attachement) {
}

func (s String) Chunks() []Schema {
	return nil
}

func (s String) String() string {
	return string(s)
}
