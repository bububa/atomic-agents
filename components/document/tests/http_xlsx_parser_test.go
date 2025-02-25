package tests

import (
	"bytes"
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
	if err := doc.ReadAll(); err != nil {
		t.Error(err)
	}
	ctx := context.Background()
	w := new(bytes.Buffer)
	parser := new(xlsx.Parser)
	if err := parser.Parse(ctx, doc.Reader(), w); err != nil {
		t.Error(err)
	}
	t.Log(w.String())
}
