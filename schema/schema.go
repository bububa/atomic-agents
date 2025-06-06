package schema

import (
	"encoding/base64"
	"encoding/json"
	"io"

	"github.com/bububa/instructor-go"
)

// Schema is message schema interface
type Schema interface {
	// Attachement() returns schema attchement
	Attachement() *Attachement
	// Chunks() returns additional schema chunks
	Chunks() []Schema
	// ExtraBody
	ExtraBody() map[string]any
}

type SchemaPointer interface {
	Schema
	SetAttachement(*Attachement)
	SetExtraBody(map[string]any)
}

type Stringer interface {
	String() string
}

func Stringify(s Schema) string {
	if v, ok := s.(Stringer); ok {
		return v.String()
	}
	bs, _ := json.Marshal(s)
	return string(bs)
}

func ToMessage(s Schema, dist *instructor.Message) {
	if attachement := s.Attachement(); attachement != nil {
		for _, link := range attachement.ImageURLs {
			dist.Images = append(dist.Images, instructor.Image{
				URL: link,
			})
		}
		for _, link := range attachement.VideoURLs {
			dist.Videos = append(dist.Videos, instructor.Video{
				URL: link,
			})
		}
		for _, r := range attachement.Files {
			bs, err := io.ReadAll(r)
			if err != nil {
				continue
			}
			b64 := base64.StdEncoding.EncodeToString(bs)
			dist.Files = append(dist.Files, instructor.File{
				Data: b64,
			})
		}
	}
	dist.Text = Stringify(s)
}
