package schema

// Base is a base schema
type Base struct {
	attachement *Attachement `json:"-" yaml:"-" jsonschema:"-"`
	chunks      []Schema     `json:"-" yaml:"-" jsonschema:"-"`
}

// Attachement returns schema attachement
func (r Base) Attachement() *Attachement {
	return r.attachement
}

// Attachement returns schema attachement
func (r *Base) SetAttachement(attach *Attachement) {
	r.attachement = attach
}

func (r Base) Chunks() []Schema {
	return r.chunks
}
