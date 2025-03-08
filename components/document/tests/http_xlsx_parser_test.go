package tests

import (
	"context"
	"testing"

	"github.com/bububa/atomic-agents/components/document"
	"github.com/bububa/atomic-agents/components/document/parsers/xlsx"
)

func TestHttpXLSxParser(t *testing.T) {
	link := "https://zzzx.snnu.edu.cn/__local/F/62/4E/896DC0778F426C757828CED677C_97EE9695_75E1.xlsx?e=.xlsx"
	doc, err := document.NewHttp(document.WithHttpURL(link))
	if err != nil {
		t.Error(err)
		return
	}
	defer doc.Close()
	ctx := context.Background()
	parser := new(xlsx.Parser)
	if err := parser.Parse(ctx, doc, doc); err != nil {
		t.Error(err)
	}
	t.Log(doc.String())
}
