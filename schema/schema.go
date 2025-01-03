package schema

import "encoding/json"

// Schema is message schema interface
type Schema interface {
	// Attachement() returns schema attchement
	Attachement() *Attachement
}

func Stringify(s Schema) string {
	bs, _ := json.Marshal(s)
	return string(bs)
}
