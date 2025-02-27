package splitter

import (
	"fmt"

	"github.com/clipperhouse/uax29/graphemes"
	"github.com/clipperhouse/uax29/phrases"
	"github.com/clipperhouse/uax29/sentences"
	"github.com/clipperhouse/uax29/words"

	"github.com/pkoukk/tiktoken-go"
)

// TokenCounter defines the interface for counting tokens in a string.
// This abstraction allows for different tokenization strategies (e.g., words, subwords).
type TokenCounter interface {
	// Count returns the number of tokens in the given text according to the
	// implementation's tokenization strategy.
	Count(p []byte) int
}

type GraphemesTokenCounter struct{}

func (c *GraphemesTokenCounter) Count(p []byte) int {
	return len(graphemes.SegmentAll(p))
}

type WordsTokenCounter struct{}

func (c WordsTokenCounter) Count(p []byte) int {
	return len(words.SegmentAll(p))
}

type PhraseTokenCounter struct{}

func (c PhraseTokenCounter) Count(p []byte) int {
	return len(phrases.SegmentAll(p))
}

type SentencesTokenCounter struct{}

func (c SentencesTokenCounter) Count(p []byte) int {
	return len(sentences.SegmentAll(p))
}

// TikTokenCounter provides accurate token counting using the tiktoken library,
// which implements the tokenization schemes used by OpenAI models.
type TikTokenCounter struct {
	tke *tiktoken.Tiktoken
}

// NewTikTokenCounter creates a new TikTokenCounter using the specified encoding.
// Common encodings include:
// - "cl100k_base" (GPT-4, ChatGPT)
// - "p50k_base" (GPT-3)
// - "r50k_base" (Codex)
func NewTikTokenCounter(encoding string) (*TikTokenCounter, error) {
	tke, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		return nil, fmt.Errorf("failed to get encoding: %w", err)
	}
	return &TikTokenCounter{tke: tke}, nil
}

// Count returns the exact number of tokens in the text according to the
// specified tiktoken encoding.
func (ttc *TikTokenCounter) Count(p []byte) int {
	return len(ttc.tke.Encode(string(p), nil, nil))
}
