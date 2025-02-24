package document

import (
	"bytes"
	"context"
	"io"
)

type Parser interface {
	Parse(context.Context, *bytes.Reader, io.Writer) error
}
