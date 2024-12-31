package schema

// Base is a base schema
type Base struct {
	attachement *Attachement `json:"-"`
}

// String implements Schema interface
func (r Base) String() string {
	return ""
}

func (r Base) Snapshot() Schema {
	return Base{}
}

// Attachement returns schema attachement
func (r Base) Attachement() *Attachement {
	return r.attachement
}
