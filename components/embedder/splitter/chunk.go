package splitter

import "bytes"

// Chunk represents a piece of text with associated metadata for tracking its position
// and size within the original document.
type Chunk struct {
	// Buffer contains the actual content of the chunk
	Buffer *bytes.Buffer
	// TokenSize represents the number of tokens in this chunk
	TokenSize int
	// Start is the index of the first part in this chunk
	Start int
	// End is the index of the last part in this chunk (exclusive)
	End int
}
