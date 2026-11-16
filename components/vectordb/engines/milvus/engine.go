package milvus

import (
	"context"
	"encoding/json"

	"github.com/milvus-io/milvus/client/v3/column"
	"github.com/milvus-io/milvus/client/v3/entity"
	"github.com/milvus-io/milvus/client/v3/index"
	milvusClient "github.com/milvus-io/milvus/client/v3/milvusclient"

	"github.com/bububa/atomic-agents/components/vectordb"
)

type Engine struct {
	db *milvusClient.Client
	vectordb.Options
}

var _ vectordb.Engine = (*Engine)(nil)

func New(db *milvusClient.Client, opts ...vectordb.Option) *Engine {
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
	if err := e.db.CreateCollection(ctx, milvusClient.NewCreateCollectionOption("vector", schema)); err != nil {
		return err
	}
	idxHnsw := index.NewHNSWIndex(entity.COSINE, 8, 200)
	idxOpt := milvusClient.NewCreateIndexOption(name, "embedding", idxHnsw).WithIndexName("embedding_idx")
	task, err := e.db.CreateIndex(ctx, idxOpt)
	if err != nil {
		return err
	}
	return task.Await(ctx)
}

func (e *Engine) Insert(ctx context.Context, collectionName string, records ...vectordb.Record) error {
	if len(records) == 0 {
		return nil
	}
	dim := int64(len(records[0].Embedding.Embedding))
	hasCollectionOpt := milvusClient.NewHasCollectionOption(collectionName)
	if exists, err := e.db.HasCollection(ctx, hasCollectionOpt); err != nil {
		return err
	} else if !exists {
		if err := e.CreateCollection(ctx, collectionName, dim); err != nil {
			return err
		}
	}
	columnsMap := make(map[string]column.Column, 4)
	for _, record := range records {
		if record.ID == "" {
			record.ID = record.Embedding.UUID()
		}
		if col, ok := columnsMap["id"]; !ok {
			columnsMap["id"] = column.NewColumnVarChar("id", []string{record.ID})
		} else {
			col.AppendValue(record.ID)
		}
		vector := vectordb.Float32s(record.Embedding.Embedding)
		if col, ok := columnsMap["embedding"]; !ok {
			columnsMap["embedding"] = column.NewColumnFloatVector("embedding", int(dim), [][]float32{vector})
		} else {
			col.AppendValue(vector)
		}
		if col, ok := columnsMap["content"]; !ok {
			columnsMap["content"] = column.NewColumnString("content", []string{record.Embedding.Object})
		} else {
			col.AppendValue(record.Embedding.Object)
		}
		if record.Embedding.Meta != nil {
			bs, _ := json.Marshal(record.Embedding.Meta)
			if col, ok := columnsMap["meta"]; !ok {
				columnsMap["meta"] = column.NewColumnJSONBytes("meta", [][]byte{bs})
			} else {
				col.AppendValue(bs)
			}
		}
	}
	columns := make([]column.Column, 0, len(columnsMap))
	for _, col := range columnsMap {
		columns = append(columns, col)
	}
	insertOpt := milvusClient.NewColumnBasedInsertOption(collectionName, columns...)
	_, err := e.db.Insert(ctx, insertOpt)
	return err
}

// Search performs vector similarity search on a collection.
func (e *Engine) Search(ctx context.Context, vectors []float64, opts ...vectordb.SearchOption) ([]vectordb.Record, error) {
	var option vectordb.SearchOptions
	for _, opt := range opts {
		opt(&option)
	}
	loadOpt := milvusClient.NewLoadCollectionOption(option.Collection)
	if task, err := e.db.LoadCollection(ctx, loadOpt); err != nil {
		return nil, err
	} else if err := task.Await(ctx); err != nil {
		return nil, err
	}
	query := entity.FloatVector(vectordb.Float32s(vectors))
	topK := option.TopK
	if topK == 0 {
		topK = e.TopK
	}
	searchParams := index.NewHNSWAnnParam(topK)
	searchOpt := milvusClient.NewSearchOption(option.Collection, topK, []entity.Vector{entity.Vector(query)}).WithANNSField("emmbedding").WithAnnParam(searchParams)
	results, err := e.db.Search(ctx, searchOpt)
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

func searchResultToRecord(result *milvusClient.ResultSet, record *vectordb.Record) {
	if len(result.Scores) > 0 {
		record.Score = float64(result.Scores[0])
	}
	for _, colKey := range []string{"id", "content", "embedding", "meta"} {
		if col := result.GetColumn(colKey); col != nil {
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
