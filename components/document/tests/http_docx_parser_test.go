package tests

import (
	"bytes"
	"context"
	"testing"

	"github.com/bububa/atomic-agents/components/document"
	"github.com/bububa/atomic-agents/components/document/parsers/docx"
)

func TestHttpDocxParser(t *testing.T) {
	link := "http://www.hbdxzj.org.cn/Uploads/detail/file/20230119/63c891e9e10c8.docx"
	doc, err := document.NewHttp(document.WithHttpURL(link))
	if err != nil {
		t.Error(err)
		return
	}
	if err := doc.ReadAll(); err != nil {
		t.Error(err)
	}
	ctx := context.Background()
	w := new(bytes.Buffer)
	parser := new(docx.Parser)
	if err := parser.Parse(ctx, doc.Reader(), w); err != nil {
		t.Error(err)
	}
	t.Log(w.String())
}
