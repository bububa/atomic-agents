package vectordb

import (
	"context"
)

type EngineType string

const (
	Memory  EngineType = "memory"
	Chromem EngineType = "chromem"
	Milvus  EngineType = "milvus"
)

type Engine interface {
	Insert(context.Context, string, []Record) error
	Search(context.Context, []float64, ...SearchOption) ([]Record, error)
}
