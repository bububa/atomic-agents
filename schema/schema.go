package schema

import (
	"encoding/json"
)

// Schema is message schema interface
type Schema interface {
	// Attachement() returns schema attchement
	Attachement() *Attachement
	// Chunks() returns additional schema chunks
	Chunks() []Schema
}

type SchemaPointer interface {
	Schema
	SetAttachement(*Attachement)
}

type Stringer interface {
	String() string
}

func Stringify(s Schema) string {
	if v, ok := s.(Stringer); ok {
		return v.String()
	}
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
