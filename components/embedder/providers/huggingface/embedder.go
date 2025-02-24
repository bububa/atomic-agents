package huggingface

import (
	"context"

	"github.com/bububa/atomic-agents/components"
	"github.com/bububa/atomic-agents/components/embedder"
)

const (
	DefaultEmbedderModel = "sentence-transformers/all-MiniLM-L6-v2"
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
	embedder.WithProvider(embedder.ProviderHuggingFace)
	embedder.WithModel(DefaultEmbedderModel)
	for _, opt := range opts {
		opt(&i.Options)
	}
	return i
}

func (p *Embedder) Embed(ctx context.Context, text string, embedding *embedder.Embedding, usage *components.LLMUsage) error {
	// Create an EmbeddingRequest for the user query
	isTrue := true
	req := EmbeddingRequest{
		Inputs: []string{text},
		Options: options{
			WaitForModel: &isTrue,
		},
		Model: p.Model(),
	}
	resp, err := p.CreateEmbeddings(ctx, &req)
	if err != nil {
		return err
	}
	if len(resp) == 0 {
		return nil
	}
	embedding.Object = text
	embedding.Embedding = resp[0]
	embedding.Index = 0
	return nil
}

func (p *Embedder) BatchEmbed(ctx context.Context, parts []string, usage *components.LLMUsage) ([]embedder.Embedding, error) {
	// Create an EmbeddingRequest for the user query
	isTrue := true
	req := EmbeddingRequest{
		Inputs: parts,
		Options: options{
			WaitForModel: &isTrue,
		},
		Model: p.Model(),
	}
	resp, err := p.CreateEmbeddings(ctx, &req)
	if err != nil {
		return nil, err
	}
	ret := make([]embedder.Embedding, 0, len(resp))
	for idx, v := range resp {
		ret = append(ret, embedder.Embedding{
			Object:    parts[idx],
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
