package embedder

import (
	"fmt"
	"strings"

	"github.com/pkoukk/tiktoken-go"
)

// TokenCounter defines the interface for counting tokens in a string.
// This abstraction allows for different tokenization strategies (e.g., words, subwords).
type TokenCounter interface {
	// Count returns the number of tokens in the given text according to the
	// implementation's tokenization strategy.
	Count(text string) int
}

// DefaultTokenCounter provides a simple word-based token counting implementation.
// It splits text on whitespace to approximate token counts. This is suitable
// for basic use cases but may not accurately reflect subword tokenization
// used by language models.
type DefaultTokenCounter struct{}

// Count returns the number of words in the text, using whitespace as a delimiter.
func (dtc *DefaultTokenCounter) Count(text string) int {
	return len(strings.Fields(text))
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
func (ttc *TikTokenCounter) Count(text string) int {
	return len(ttc.tke.Encode(text, nil, nil))
}
