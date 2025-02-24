package cohere

import (
	"context"

	cohere "github.com/cohere-ai/cohere-go/v2"
	cohereClient "github.com/cohere-ai/cohere-go/v2/client"

	"github.com/bububa/atomic-agents/components"
	"github.com/bububa/atomic-agents/components/embedder"
)

type Embedder struct {
	*cohereClient.Client

	embedder.Options
}

var _ embedder.Embedder = (*Embedder)(nil)

func (p *Embedder) SetClient(clt *cohereClient.Client) {
	p.Client = clt
}

func New(client *cohereClient.Client, opts ...embedder.Option) *Embedder {
	i := &Embedder{
		Client: client,
	}
	embedder.WithProvider(embedder.ProviderCohere)
	for _, opt := range opts {
		opt(&i.Options)
	}
	return i
}

func (p *Embedder) Embed(ctx context.Context, text string, embedding *embedder.Embedding, usage *components.LLMUsage) error {
	// Create an EmbeddingRequest for the user query
	model := p.Model()
	req := cohere.EmbedRequest{
		Texts: []string{text},
		Model: &model,
	}
	resp, err := p.Client.Embed(ctx, &req)
	if err != nil {
		return err
	}
	respV := resp.GetEmbeddingsFloats()
	if usage != nil && respV.Meta != nil && respV.Meta.Tokens != nil {
		if v := respV.Meta.Tokens.InputTokens; v != nil {
			usage.InputTokens = int(*v)
		}
		if v := respV.Meta.Tokens.OutputTokens; v != nil {
			usage.OutputTokens = int(*v)
		}
	}
	if len(respV.Embeddings) == 0 {
		return nil
	}
	ret := respV.Embeddings[0]
	embedding.Object = respV.Texts[0]
	embedding.Embedding = ret
	embedding.Index = 0
	return nil
}

func (p *Embedder) BatchEmbed(ctx context.Context, parts []string, usage *components.LLMUsage) ([]embedder.Embedding, error) {
	// Create an EmbeddingRequest for the user query
	model := p.Model()
	req := cohere.EmbedRequest{
		Texts: parts,
		Model: &model,
	}
	resp, err := p.Client.Embed(ctx, &req)
	if err != nil {
		return nil, err
	}
	respV := resp.GetEmbeddingsFloats()
	if usage != nil && respV.Meta != nil && respV.Meta.Tokens != nil {
		if v := respV.Meta.Tokens.InputTokens; v != nil {
			usage.InputTokens = int(*v)
		}
		if v := respV.Meta.Tokens.OutputTokens; v != nil {
			usage.OutputTokens = int(*v)
		}
	}
	ret := make([]embedder.Embedding, 0, len(respV.Embeddings))
	for idx, v := range respV.Embeddings {
		ret = append(ret, embedder.Embedding{
			Object:    respV.Texts[idx],
			Embedding: v,
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
