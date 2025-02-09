package schema

import "encoding/json"

// Schema is message schema interface
type Schema interface {
	// Attachement() returns schema attchement
	Attachement() *Attachement
}

type SchemaPointer interface {
	Schema
	SetAttachement(*Attachement)
}

func Stringify(s Schema) string {
	if v, ok := s.(String); ok {
		return string(v)
	}
	bs, _ := json.Marshal(s)
	return string(bs)
}

func ToBytes(s Schema) []byte {
	if v, ok := s.(String); ok {
		return []byte(v)
	}
	bs, _ := json.Marshal(s)
	return bs
}
