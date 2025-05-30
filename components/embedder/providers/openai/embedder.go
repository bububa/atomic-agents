package openai

import (
	"context"
	"errors"

	"github.com/openai/openai-go"

	"github.com/bububa/atomic-agents/components"
	"github.com/bububa/atomic-agents/components/embedder"
)

type Embedder struct {
	*openai.Client

	embedder.Options
}

var _ embedder.Embedder = (*Embedder)(nil)

func (p *Embedder) SetClient(clt *openai.Client) {
	p.Client = clt
}

func New(client *openai.Client, opts ...embedder.Option) *Embedder {
	i := &Embedder{
		Client: client,
	}
	embedder.WithProvider(embedder.ProviderOpenAI)
	for _, opt := range opts {
		opt(&i.Options)
	}
	return i
}

func (p *Embedder) Embed(ctx context.Context, text string, embedding *embedder.Embedding, usage *components.LLMUsage) error {
	// Create an EmbeddingRequest for the user query
	req := openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String(text),
		},
		Model: openai.EmbeddingModel(p.Model()),
	}
	resp, err := p.Embeddings.New(ctx, req)
	if err != nil {
		return err
	}
	if usage != nil {
		usage.InputTokens = resp.Usage.TotalTokens
	}
	if len(resp.Data) == 0 {
		return nil
	}
	ret := resp.Data[0]
	embedding.Object = text
	embedding.Embedding = make([]float64, 0, len(ret.Embedding))
	for _, v := range ret.Embedding {
		embedding.Embedding = append(embedding.Embedding, float64(v))
	}
	embedding.Index = 0
	return nil
}

func (p *Embedder) BatchEmbed(ctx context.Context, parts []string, usage *components.LLMUsage) ([]embedder.Embedding, error) {
	// Create an EmbeddingRequest for the user query
	req := openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfArrayOfStrings: parts,
		},
		Model: openai.EmbeddingModel(p.Model()),
	}
	resp, err := p.Embeddings.New(ctx, req)
	if err != nil {
		return nil, err
	}
	if usage != nil {
		usage.InputTokens = resp.Usage.TotalTokens
	}
	ret := make([]embedder.Embedding, 0, len(resp.Data))
	for _, v := range resp.Data {
		embeddings := make([]float64, 0, len(v.Embedding))
		for _, e := range v.Embedding {
			embeddings = append(embeddings, float64(e))
		}
		ret = append(ret, embedder.Embedding{
			Object:    parts[int(v.Index)],
			Embedding: embeddings,
			Index:     int(v.Index),
		})
	}
	return ret, nil
}

func convertToOpenAI(src *embedder.Embedding, dist *openai.Embedding) {
	dist.Embedding = make([]float64, len(src.Embedding))
	copy(dist.Embedding, src.Embedding)
}

// DotProduct calculates the dot product of the embedding vector with another
// embedding vector. Both vectors must have the same length; otherwise, an
// ErrVectorLengthMismatch is returned. The method returns the calculated dot
// product as a float32 value.
func (p *Embedder) DotProduct(ctx context.Context, target, query *embedder.Embedding) (float64, error) {
	t := new(openai.Embedding)
	convertToOpenAI(target, t)
	q := new(openai.Embedding)
	convertToOpenAI(query, q)
	if len(t.Embedding) != len(q.Embedding) {
		return 0, errors.New("vector length mismatch")
	}
	var dotProduct float64
	for i := range t.Embedding {
		dotProduct += t.Embedding[i] * q.Embedding[i]
	}
	return dotProduct, nil
}
