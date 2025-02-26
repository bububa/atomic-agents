package docx

import (
	"context"
	"io"

	"github.com/fumiama/go-docx"

	"github.com/bububa/atomic-agents/components/document"
)

// Parser is a parser which parse docx to markdown
type Parser struct{}

var _ document.Parser = (*Parser)(nil)

// Parse try to parse a pdf content from a bytes.Reader and write to an io.Writer
func (p *Parser) Parse(ctx context.Context, reader document.ParserReader, writer io.Writer) error {
	size := reader.Size()
	doc, err := docx.Parse(reader, size)
	if err != nil {
		return err
	}

	for idx, it := range doc.Document.Body.Items {
		var content string
		switch t := it.(type) {
		case *docx.Paragraph:
			content = t.String()
		case *docx.Table:
			content = t.String()
		}
		if idx > 0 {
			writer.Write([]byte{'\n', '\n'})
		}
		writer.Write([]byte(content))
	}
	return nil
}
