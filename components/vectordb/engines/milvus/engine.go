package milvus

import (
	"context"
	"encoding/json"

	milvusClient "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"

	"github.com/bububa/atomic-agents/components/vectordb"
)

type Engine struct {
	db milvusClient.Client
	vectordb.Options
}

var _ vectordb.Engine = (*Engine)(nil)

func New(db milvusClient.Client, opts ...vectordb.Option) *Engine {
	ret := &Engine{
		db: db,
	}
	vectordb.WithEngine(vectordb.Milvus)
	for _, opt := range opts {
		opt(&ret.Options)
	}
	return ret
}

func (e *Engine) CreateCollection(ctx context.Context, name string, dim int64) error {
	idField := entity.NewField().WithName("id").WithDataType(entity.FieldTypeVarChar).WithMaxLength(36).WithIsPrimaryKey(true).WithIsAutoID(false)
	vectorField := entity.NewField().WithName("embeddings").WithDataType(entity.FieldTypeFloatVector).WithDim(dim)
	contentField := entity.NewField().WithName("content").WithDataType(entity.FieldTypeString)
	metaField := entity.NewField().WithName("meta").WithDataType(entity.FieldTypeJSON).WithIsDynamic(true)
	schema := entity.NewSchema().WithName(name).WithAutoID(false).WithField(idField).WithField(vectorField).WithField(contentField).WithField(metaField)
	if err := e.db.CreateCollection(ctx, schema, 0); err != nil {
		return err
	}
	idxHnsw, err := entity.NewIndexHNSW(entity.COSINE, 8, 200)
	if err != nil {
		return err
	}
	return e.db.CreateIndex(ctx, name, "embedding", idxHnsw, true, milvusClient.WithIndexName("embedding_idx"))
}

func (e *Engine) Insert(ctx context.Context, collectionName string, records []vectordb.Record) error {
	if len(records) == 0 {
		return nil
	}
	dim := int64(len(records[0].Embedding.Embedding))
	if exists, err := e.db.HasCollection(ctx, collectionName); err != nil {
		return err
	} else if !exists {
		if err := e.CreateCollection(ctx, collectionName, dim); err != nil {
			return err
		}
	}
	columnsMap := make(map[string]entity.Column, 4)
	for _, record := range records {
		if record.ID == "" {
			record.ID = record.Embedding.UUID()
		}
		if col, ok := columnsMap["id"]; !ok {
			columnsMap["id"] = entity.NewColumnVarChar("id", []string{record.ID})
		} else {
			col.AppendValue(record.ID)
		}
		vector := vectordb.Float32s(record.Embedding.Embedding)
		if col, ok := columnsMap["embedding"]; !ok {
			columnsMap["embedding"] = entity.NewColumnFloatVector("embedding", int(dim), [][]float32{vector})
		} else {
			col.AppendValue(vector)
		}
		if col, ok := columnsMap["content"]; !ok {
			columnsMap["content"] = entity.NewColumnString("content", []string{record.Embedding.Object})
		} else {
			col.AppendValue(record.Embedding.Object)
		}
		if record.Embedding.Meta != nil {
			bs, _ := json.Marshal(record.Embedding.Meta)
			if col, ok := columnsMap["meta"]; !ok {
				columnsMap["meta"] = entity.NewColumnJSONBytes("meta", [][]byte{bs})
			} else {
				col.AppendValue(bs)
			}
		}
	}
	columns := make([]entity.Column, 0, len(columnsMap))
	for _, col := range columnsMap {
		columns = append(columns, col)
	}
	_, err := e.db.Insert(ctx, collectionName, "", columns...)
	return err
}

// Search performs vector similarity search on a collection.
func (e *Engine) Search(ctx context.Context, vectors []float64, opts ...vectordb.SearchOption) ([]vectordb.Record, error) {
	var option vectordb.SearchOptions
	for _, opt := range opts {
		opt(&option)
	}
	if err := e.db.LoadCollection(ctx, option.Collection, false); err != nil {
		return nil, err
	}
	query := entity.FloatVector(vectordb.Float32s(vectors))
	topK := option.TopK
	if topK == 0 {
		topK = e.TopK
	}
	searchParams, err := entity.NewIndexHNSWSearchParam(topK)
	if err != nil {
		return nil, err
	}
	results, err := e.db.Search(ctx, option.Collection, nil, "", []string{}, []entity.Vector{query}, "embedding", entity.COSINE, topK, searchParams)
	if err != nil {
		return nil, err
	}
	searchResults := make([]vectordb.Record, len(results))
	for _, result := range results {
		var record vectordb.Record
		searchResultToRecord(&result, &record)
		searchResults = append(searchResults, record)
	}
	return searchResults, nil
}

func searchResultToRecord(result *milvusClient.SearchResult, record *vectordb.Record) {
	if len(result.Scores) > 0 {
		record.Score = float64(result.Scores[0])
	}
	for _, colKey := range []string{"id", "content", "embedding", "meta"} {
		if col := result.Fields.GetColumn(colKey); col != nil {
			switch colKey {
			case "id":
				record.ID, _ = col.GetAsString(0)
			case "content":
				record.Embedding.Object, _ = col.GetAsString(0)
			case "embedding":
				if v, err := col.Get(0); err == nil {
					if embedding, ok := v.([]float32); ok {
						record.Embedding.Embedding = vectordb.Float64s(embedding)
					}
				}
			case "meta":
				if v, err := col.Get(0); err == nil {
					if bs, ok := v.([]byte); ok {
						json.Unmarshal(bs, &record.Embedding.Meta)
					}
				}
			}
		}
	}
}
