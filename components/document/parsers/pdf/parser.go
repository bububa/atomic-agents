package pdf

import (
	"bytes"
	"context"
	"io"

	"github.com/ledongthuc/pdf"

	"github.com/bububa/atomic-agents/components/document"
)

// Parser is a parser which parse PDF content to text
type Parser struct {
	password string
}

var _ document.Parser = (*Parser)(nil)

type Option func(*Parser)

func WithPassword(password string) Option {
	return func(p *Parser) {
		p.password = password
	}
}

func NewPDFParser(opts ...Option) *Parser {
	ret := new(Parser)
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

// Parse try to parse a pdf content from a bytes.Reader and write to an io.Writer
func (p *Parser) Parse(ctx context.Context, reader *bytes.Reader, writer io.Writer) error {
	var (
		r    *pdf.Reader
		err  error
		size = reader.Size()
	)
	if p.password != "" {
		if r, err = pdf.NewReaderEncrypted(reader, size, func() string {
			return p.password
		}); err != nil {
			return err
		}
	} else {
		if r, err = pdf.NewReader(reader, size); err != nil {
			return err
		}
	}
	totalPage := r.NumPage()

	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}

		rows, _ := p.GetTextByRow()
		for idx, row := range rows {
			if idx > 0 {
				if _, err := writer.Write([]byte{'\n'}); err != nil {
					return err
				}
			}
			for _, word := range row.Content {
				if _, err := writer.Write([]byte(word.S)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
