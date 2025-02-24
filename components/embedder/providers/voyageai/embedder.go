package voyageai

import (
	"context"

	"github.com/bububa/atomic-agents/components"
	"github.com/bububa/atomic-agents/components/embedder"
)

type Embedder struct {
	*Client

	embedder.Options
}

var _ embedder.Embedder = (*Embedder)(nil)

func (p *Embedder) SetClient(clt *Client) {
	p.Client = clt
}

func New(client *Client, opts ...embedder.Option) *Embedder {
	i := &Embedder{
		Client: client,
	}
	embedder.WithProvider(embedder.ProviderVoyageAI)
	for _, opt := range opts {
		opt(&i.Options)
	}
	return i
}

func (p *Embedder) Embed(ctx context.Context, text string, embedding *embedder.Embedding, usage *components.LLMUsage) error {
	// Create an EmbeddingRequest for the user query
	req := EmbeddingRequest{
		Input: []string{text},
		Model: p.Model(),
	}
	resp, err := p.CreateEmbeddings(ctx, &req)
	if err != nil {
		return err
	}
	if usage != nil {
		usage.InputTokens = int(resp.Usage.TotalTokens)
	}
	if len(resp.Data) == 0 {
		return nil
	}
	ret := resp.Data[0]
	embedding.Object = ret.Object
	embedding.Embedding = ret.Embedding
	embedding.Index = 0
	return nil
}

func (p *Embedder) BatchEmbed(ctx context.Context, parts []string, usage *components.LLMUsage) ([]embedder.Embedding, error) {
	// Create an EmbeddingRequest for the user query
	req := EmbeddingRequest{
		Input: parts,
		Model: p.Model(),
	}
	resp, err := p.CreateEmbeddings(ctx, &req)
	if err != nil {
		return nil, err
	}
	if usage != nil {
		usage.InputTokens = int(resp.Usage.TotalTokens)
	}
	ret := make([]embedder.Embedding, 0, len(resp.Data))
	for _, v := range resp.Data {
		ret = append(ret, embedder.Embedding{
			Object:    v.Object,
			Embedding: v.Embedding,
			Index:     v.Index,
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
