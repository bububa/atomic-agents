package tests

import (
	"context"
	"testing"

	"github.com/bububa/atomic-agents/components/document"
	"github.com/bububa/atomic-agents/components/document/parsers/pptx"
)

func TestHttpPPTxParser(t *testing.T) {
	link := "https://zcc.czu.cn/_upload/article/files/06/b5/8a64cb854694bcd2265ad0b96c99/65ba8668-56f7-4cd6-ab51-b55349964a17.pptx"
	doc, err := document.NewHttp(document.WithHttpURL(link))
	if err != nil {
		t.Error(err)
		return
	}
	defer doc.Close()
	ctx := context.Background()
	parser := new(pptx.Parser)
	if err := parser.Parse(ctx, doc, doc); err != nil {
		t.Error(err)
	}
	t.Log(doc.String())
}
