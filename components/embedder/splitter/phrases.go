package splitter

import (
	"github.com/clipperhouse/uax29/phrases"

	"github.com/bububa/atomic-agents/components/embedder"
)

type Phrases struct {
	Options
}

var _ embedder.Chunker = (*Sentences)(nil)

func NewPhrases(opts ...Option) *Phrases {
	ret := new(Phrases)
	for _, opt := range opts {
		opt(&ret.Options)
	}
	ret.delimiter = []byte(" ")
	ret.scanner = phrases.NewScanner(ret.Buffer())
	if ret.tokenCounter == nil {
		ret.tokenCounter = new(PhrasesTokenCounter)
	}
	return ret
}

//
// func scanSentences(data []byte, atEOF bool) (advance int, token []byte, err error) {
// 	if atEOF && len(data) == 0 {
// 		return 0, nil, nil
// 	}
//
// 	endPunctuation := regexp.MustCompile(`([.!?;。！？；]+)(\s*)`)
// 	loc := endPunctuation.FindSubmatchIndex(data)
//
// 	if loc != nil {
// 		endIdx := loc[1]
// 		return endIdx, bytes.TrimRight(data[:endIdx], " \n\r\t"), nil
// 	}
//
// 	if atEOF {
// 		return len(data), bytes.TrimRight(data, " \n\r\t"), nil
// 	}
//
// 	return 0, nil, nil
// }
