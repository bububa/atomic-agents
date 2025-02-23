package embedder

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/bububa/atomic-agents/components"
)

type Embedder interface {
	Provider() Provider
	Model() string
	Embed(context.Context, string, *Embedding, *components.LLMUsage) error
	BatchEmbed(ctx context.Context, parts []string, usage *components.LLMUsage) ([]Embedding, error)
	DotProduct(context.Context, *Embedding, *Embedding) (float64, error)
}

// EmbedChunks processes a slice of text chunks and generates embeddings for each one.
// It handles the embedding process in sequence, with debug output for monitoring.
// The function:
// 1. Allocates space for the results
// 2. Processes each chunk through the embedder
// 3. Creates EmbeddedChunk instances with the results
// 4. Provides progress information via debug output
//
// Returns an error if any chunk fails to embed properly.
func EmbedChunks(ctx context.Context, embedder Embedder, chunks []Chunk, usage *components.LLMUsage) ([]EmbeddedChunk, error) {
	parts := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		parts = append(parts, chunk.Text)
	}

	ret, err := embedder.BatchEmbed(ctx, parts, usage)
	if err != nil {
		return nil, err
	}
	embeddedChunks := make([]EmbeddedChunk, 0, len(ret))
	for _, v := range ret {

		embeddedChunk := EmbeddedChunk{
			Embedding: v,
			Chunk:     &chunks[v.Index],
		}
		embeddedChunks = append(embeddedChunks, embeddedChunk)
	}
	return embeddedChunks, nil
}

// EmbedChunk processes text chunk and generates embeddings.
// It handles the embedding process in sequence, with debug output for monitoring.
// The function:
// 1. Allocates space for the results
// 2. Processes each chunk through the embedder
// 3. Creates EmbeddedChunk instances with the results
// 4. Provides progress information via debug output
func EmbedChunk(ctx context.Context, embedder Embedder, chunk *Chunk, embedding *Embedding, usage *components.LLMUsage) error {
	return embedder.Embed(ctx, chunk.Text, embedding, usage)
}

// Base64 is base64 encoded embedding string.
type Base64 string

// Decode decodes base64 encoded string into a slice of floats.
func (s Base64) Decode() (*Embedding, error) {
	decoded, err := base64.StdEncoding.DecodeString(string(s))
	if err != nil {
		return nil, err
	}

	if len(decoded)%8 != 0 {
		return nil, fmt.Errorf("invalid base64 encoded string length")
	}

	floats := make([]float64, len(decoded)/8)

	for i := range floats {
		bits := binary.LittleEndian.Uint64(decoded[i*8 : (i+1)*8])
		floats[i] = math.Float64frombits(bits)
	}

	return &Embedding{
		Embedding: floats,
	}, nil
}

// DotProduct calculates the dot product of the embedding vector with another
// embedding vector. Both vectors must have the same length; otherwise, an
// ErrVectorLengthMismatch is returned. The method returns the calculated dot
// product as a float32 value.
func (e *Embedding) DotProduct(other *Embedding) (float64, error) {
	if len(e.Embedding) != len(other.Embedding) {
		return 0, errors.New("vector length mismatch")
	}

	var dotProduct float64
	for i := range e.Embedding {
		dotProduct += e.Embedding[i] * other.Embedding[i]
	}

	return dotProduct, nil
}
