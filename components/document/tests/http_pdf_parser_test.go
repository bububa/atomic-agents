package tests

import (
	"bytes"
	"context"
	"testing"

	"github.com/bububa/atomic-agents/components/document"
	"github.com/bububa/atomic-agents/components/document/parsers/pdf"
)

func TestHttpPDFParser(t *testing.T) {
	link := "http://zhangbaohui.snnu.edu.cn/icse2012/files/2012_Nanjing_University_International_Education_Conference_Chinese_Proceedings.pdf"
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
	parser := new(pdf.Parser)
	if err := parser.Parse(ctx, doc.Reader(), w); err != nil {
		t.Error(err)
	}
	t.Log(w.String())
}
