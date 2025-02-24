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

// Document is a document container with metadata
type Document struct {
	buffer *bytes.Buffer
	Meta   map[string]string
}

func (d *Document) Reader() *bytes.Reader {
	return bytes.NewReader(d.buffer.Bytes())
}
