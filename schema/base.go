package schema

// Base is a base schema
type Base struct {
	attachement *Attachement `json:"-" jsonschema:"-"`
}

// Attachement returns schema attachement
func (r Base) Attachement() *Attachement {
	return r.attachement
}

// Attachement returns schema attachement
func (r *Base) SetAttachement(attach *Attachement) {
	r.attachement = attach
}
