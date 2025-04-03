package gemini

import (
	"context"

	"github.com/bububa/atomic-agents/components"
	"github.com/bububa/atomic-agents/components/embedder"
	gemini "google.golang.org/genai"
)

type Embedder struct {
	*gemini.Client

	embedder.Options
}

var _ embedder.Embedder = (*Embedder)(nil)

func (p *Embedder) SetClient(clt *gemini.Client) {
	p.Client = clt
}

func New(client *gemini.Client, opts ...embedder.Option) *Embedder {
	i := &Embedder{
		Client: client,
	}
	embedder.WithProvider(embedder.ProviderGemini)
	for _, opt := range opts {
		opt(&i.Options)
	}
	return i
}

func (p *Embedder) Embed(ctx context.Context, text string, embedding *embedder.Embedding, usage *components.LLMUsage) error {
	// Create an EmbeddingRequest for the user query
	resp, err := p.Models.EmbedContent(ctx, p.Model(), []*gemini.Content{
		{
			Parts: []*gemini.Part{{Text: text}},
		},
	}, nil)
	if err != nil {
		return err
	}
	if len(resp.Embeddings) == 0 {
		return nil
	}
	embedding.Object = text
	embedding.Embedding = make([]float64, 0, len(resp.Embeddings[0].Values))
	for _, v := range resp.Embeddings[0].Values {
		embedding.Embedding = append(embedding.Embedding, float64(v))
	}
	embedding.Index = 0
	return nil
}

func (p *Embedder) BatchEmbed(ctx context.Context, parts []string, usage *components.LLMUsage) ([]embedder.Embedding, error) {
	// Create an EmbeddingRequest for the user query
	contents := make([]*gemini.Content, 0, len(parts))
	for _, part := range parts {
		contents = append(contents, &gemini.Content{Parts: []*gemini.Part{{Text: part}}})
	}
	resp, err := p.Models.EmbedContent(ctx, p.Model(), contents, nil)
	if err != nil {
		return nil, err
	}
	ret := make([]embedder.Embedding, 0, len(resp.Embeddings))
	for idx, v := range resp.Embeddings {
		embeddings := make([]float64, 0, len(v.Values))
		for _, e := range v.Values {
			embeddings = append(embeddings, float64(e))
		}
		ret = append(ret, embedder.Embedding{
			Object:    parts[idx],
			Embedding: embeddings,
			Index:     idx,
		})
	}
	return ret, nil
}

// DotProduct calculates the dot product of the embedding vector with another
// embedding vector. Both vectors must have the same length; otherwise, an
// ErrVectorLengthMismatch is returned. The method returns the calculated dot
// product as a float32 value.
func (p *Embedder) DotProduct(ctx context.Context, target, query *embedder.Embedding) (float64, error) {
	return target.DotProduct(query)
}
