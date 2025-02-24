package vectordb

import "github.com/bububa/atomic-agents/components/embedder"

type SearchOptions struct {
	Collection string
	TopK       int
	Meta       map[string]string
	Include    string
	Exclude    string
}

type SearchOption func(*SearchOptions)

func SearchWithCollection(name string) SearchOption {
	return func(r *SearchOptions) {
		r.Collection = name
	}
}

func SearchWithTopK(topK int) SearchOption {
	return func(r *SearchOptions) {
		r.TopK = topK
	}
}

func SearchWithMeta(meta map[string]string) SearchOption {
	return func(r *SearchOptions) {
		r.Meta = meta
	}
}

func SearchWithInclude(v string) SearchOption {
	return func(r *SearchOptions) {
		r.Include = v
	}
}

func SearchWithExclude(v string) SearchOption {
	return func(r *SearchOptions) {
		r.Exclude = v
	}
}

// Record represents a single result from a vector similarity search.
type Record struct {
	// ID is the identifier for the result
	ID string
	// Score is the similarity score for the result
	Score float64
	// Embedding embeddings for doc
	Embedding embedder.Embedding
}
