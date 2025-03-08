package tests

import (
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
	defer doc.Close()
	ctx := context.Background()
	parser := new(pdf.Parser)
	if err := parser.Parse(ctx, doc, doc); err != nil {
		t.Error(err)
	}
	t.Log(doc.String())
}
