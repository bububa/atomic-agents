package html

import (
	"bytes"
	"context"
	"io"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"

	"github.com/bububa/atomic-agents/components/document"
)

// Parser is a parser which parse html content to markdown
type Parser struct {
	opts []converter.ConvertOptionFunc
}

var _ document.Parser = (*Parser)(nil)

func NewParser(opts ...converter.ConvertOptionFunc) *Parser {
	return &Parser{
		opts: opts,
	}
}

// Parse try to parse a html content from a bytes.Reader into a markdown content then write to an io.Writer
func (h *Parser) Parse(ctx context.Context, reader *bytes.Reader, writer io.Writer) error {
	bs, err := htmltomarkdown.ConvertReader(reader, h.opts...)
	if err != nil {
		return err
	}
	_, err = writer.Write(bs)
	return err
}
