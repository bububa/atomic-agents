package schema

import "encoding/json"

// Schema is message schema interface
type Schema interface {
	// Attachement() returns schema attchement
	Attachement() *Attachement
}

func Stringify(s Schema) string {
	if v, ok := s.(String); ok {
		return string(v)
	}
	bs, _ := json.Marshal(s)
	return string(bs)
}
