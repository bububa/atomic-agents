package schema

type String string

func (s String) Attachement() *Attachement {
	return nil
}

func (s String) SetAttachement(v *Attachement) {
}
