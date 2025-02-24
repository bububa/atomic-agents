package document

import (
	"bytes"
	"context"
	"io"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
)

// HTML2MDParser is a parser which parse html content to markdown
type HTML2MDParser struct {
	opts []converter.ConvertOptionFunc
}

var _ Parser = (*PDFParser)(nil)

func NewHTML2MDParser(opts ...converter.ConvertOptionFunc) *HTML2MDParser {
	return &HTML2MDParser{
		opts: opts,
	}
}

// Parse try to parse a html content from a bytes.Reader into a markdown content then write to an io.Writer
func (h *HTML2MDParser) Parse(ctx context.Context, reader *bytes.Reader, writer io.Writer) error {
	bs, err := htmltomarkdown.ConvertReader(reader, h.opts...)
	if err != nil {
		return err
	}
	_, err = writer.Write(bs)
	return err
}
