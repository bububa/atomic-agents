package splitter

import (
	"github.com/clipperhouse/uax29/words"

	"github.com/bububa/atomic-agents/components/embedder"
)

type Words struct {
	Options
}

var _ embedder.Chunker = (*Words)(nil)

func NewWords(opts ...Option) *Words {
	ret := new(Words)
	for _, opt := range opts {
		opt(&ret.Options)
	}
	ret.delimiter = []byte(" ")
	ret.scanner = words.NewScanner(ret.Buffer())
	if ret.tokenCounter == nil {
		ret.tokenCounter = new(WordsTokenCounter)
	}
	// ret.Scanner().Split(bufio.ScanWords)
	return ret
}
