package schema

type String string

func (s String) Attachement() *Attachement {
	return nil
}

func (s String) SetAttachement(v *Attachement) {
}

func (s String) ToMarkdown() string {
	return string(s)
}
