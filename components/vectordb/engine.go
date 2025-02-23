package vectordb

import (
	"context"

	"github.com/bububa/atomic-agents/components/embedder"
)

type EngineType string

const (
	Memory  EngineType = "memory"
	Chromem EngineType = "chromem"
	Milvus  EngineType = "milvus"
)

type Engine interface {
	Insert(context.Context, []embedder.Embedding) error
	Search(context.Context, []float64, ...SearchOption) ([]Record, error)
}
