package document

import (
	"bytes"
	"errors"
)

var ErrReading = errors.New("document is reading")

type ReadStatus = int32

const (
	Unread ReadStatus = iota
	Reading
	ReadCompleted
)

type ReadableDocument interface {
	ReadAll() error
	Read() (chan<- []byte, error)
}

type ClosableDocument interface {
	Close() error
}

type IDocument interface {
	Content() string
	Meta() map[string]string
	Reader() *bytes.Reader
}

// Document is a document container with metadata
type Document struct {
	buffer *bytes.Buffer
	meta   map[string]string
}

func (d *Document) Reader() *bytes.Reader {
	return bytes.NewReader(d.buffer.Bytes())
}

func (d *Document) Content() string {
	return d.buffer.String()
}

func (d *Document) Meta() map[string]string {
	return d.meta
}
