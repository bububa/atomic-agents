package chromem

import (
	"context"

	"github.com/philippgille/chromem-go"

	"github.com/bububa/atomic-agents/components/vectordb"
)

type Engine struct {
	db *chromem.DB
	vectordb.Options
}

var _ vectordb.Engine = (*Engine)(nil)

func New(db *chromem.DB, opts ...vectordb.Option) *Engine {
	ret := &Engine{
		db: db,
	}
	vectordb.WithEngine(vectordb.Chromem)
	for _, opt := range opts {
		opt(&ret.Options)
	}
	return ret
}

func (e *Engine) Collection(_ context.Context, name string) (*chromem.Collection, error) {
	return e.db.GetOrCreateCollection(name, nil, nil)
}

func (e *Engine) Insert(ctx context.Context, collectionName string, records ...vectordb.Record) error {
	col, err := e.Collection(ctx, collectionName)
	if err != nil {
		return err
	}
	count := len(records)
	docs := make([]chromem.Document, 0, count)
	for _, record := range records {
		var doc chromem.Document
		recordToDocument(&record, &doc)
		docs = append(docs, doc)
	}
	// Insert documents in batches to avoid memory issues
	batchSize := 100
	for i := 0; i < len(docs); i += batchSize {
		end := min(i+batchSize, count)
		batch := docs[i:end]

		for _, doc := range batch {
			err := col.AddDocument(ctx, doc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Search performs vector similarity search on a collection.
func (e *Engine) Search(ctx context.Context, vectors []float64, opts ...vectordb.SearchOption) ([]vectordb.Record, error) {
	var option vectordb.SearchOptions
	for _, opt := range opts {
		opt(&option)
	}
	col, err := e.Collection(ctx, option.Collection)
	if err != nil {
		return nil, err
	}
	query := vectordb.Float32s(vectors)
	whereDocument := make(map[string]string, 2)
	if option.Include != "" {
		whereDocument["$contains"] = option.Include
	}
	if option.Exclude != "" {
		whereDocument["$not_contains"] = option.Exclude
	}
	topK := option.TopK
	if topK == 0 {
		topK = e.TopK
	}
	// Search documents using empty filters for where and whereDocument
	results, err := col.QueryEmbedding(ctx, query, topK, option.Meta, whereDocument)
	if err != nil {
		return nil, err
	}
	// Convert results
	searchResults := make([]vectordb.Record, len(results))
	for _, result := range results {
		var rec vectordb.Record
		resultToRecord(&result, &rec)
		searchResults = append(searchResults, rec)
	}

	return searchResults, nil
}

func resultToRecord(res *chromem.Result, record *vectordb.Record) {
	record.ID = res.ID
	record.Score = float64(res.Similarity)
	record.Embedding.Object = res.Content
	record.Embedding.Meta = res.Metadata
}

func recordToDocument(record *vectordb.Record, doc *chromem.Document) {
	if record.ID == "" {
		record.ID = record.Embedding.UUID()
	}
	doc.ID = record.ID
	doc.Content = record.Embedding.Object
	doc.Metadata = record.Embedding.Meta
	doc.Embedding = vectordb.Float32s(record.Embedding.Embedding)
}
