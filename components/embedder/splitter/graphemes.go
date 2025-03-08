package splitter

import (
	"github.com/bububa/atomic-agents/components/embedder"
	"github.com/clipperhouse/uax29/graphemes"
)

type Graphemes struct {
	Options
}

var _ embedder.Chunker = (*Graphemes)(nil)

func NewGraphemes(opts ...Option) *Graphemes {
	ret := new(Graphemes)
	for _, opt := range opts {
		opt(&ret.Options)
	}
	// ret.Scanner().Split(bufio.ScanBytes)
	ret.delimiter = []byte("")
	ret.scanner = graphemes.NewScanner(ret.Buffer())
	return ret
}
