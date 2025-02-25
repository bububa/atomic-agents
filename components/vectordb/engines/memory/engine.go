package memory

import (
	"context"
	"math"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/bububa/atomic-agents/components/vectordb"
)

// Engine implements the VectorDB interface using in-memory storage.
// It provides thread-safe operations for managing collections and performing
// vector similarity searches without the need for external database systems.
type Engine struct {
	// collections stores all vector collections in memory
	collections *sync.Map
	vectordb.Options
}

var _ vectordb.Engine = (*Engine)(nil)

// Collection represents a named set of records with a defined schema.
// It's the basic unit of organization in the memory database.
type Collection struct {
	// records holds the actual records in the collection
	records []vectordb.Record
	// mu provides thread-safety for concurrent operations
	mu sync.RWMutex
}

func (c *Collection) AddRecords(records ...vectordb.Record) {
	c.mu.Lock()
	c.records = append(c.records, records...)
	c.mu.Unlock()
}

func (c *Collection) Records() []vectordb.Record {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.records
}

// New creates a new in-memory vector database instance.
// It initializes an empty collection map and returns a ready-to-use database.
func New(opts ...vectordb.Option) (*Engine, error) {
	ret := &Engine{
		collections: new(sync.Map),
	}
	for _, opt := range opts {
		opt(&ret.Options)
	}
	return ret, nil
}

// HasCollection checks if a collection with the given name exists in the database.
// This operation is thread-safe and uses a read lock.
func (m *Engine) HasCollection(name string) (bool, error) {
	_, exists := m.collections.Load(name)
	return exists, nil
}

// DropCollection removes a collection and all its data from the database.
// This operation is thread-safe and uses a write lock.
func (m *Engine) DropCollection(name string) error {
	m.collections.Delete(name)
	return nil
}

// Collection creates a new collection with the specified schema.
// Returns an error if a collection with the same name already exists.
// This operation is thread-safe and uses a write lock.
func (e *Engine) Collection(_ context.Context, name string) (*Collection, error) {
	col, _ := e.collections.LoadOrStore(name, new(Collection))
	return col.(*Collection), nil
}

func (e *Engine) Insert(ctx context.Context, collectionName string, records ...vectordb.Record) error {
	col, err := e.Collection(ctx, collectionName)
	if err != nil {
		return err
	}
	count := len(records)
	docs := make([]vectordb.Record, 0, count)
	for _, record := range records {
		if record.ID == "" {
			record.ID = record.Embedding.UUID()
		}
		docs = append(docs, record)
	}
	col.AddRecords(docs...)
	return nil
}

func (e *Engine) Search(ctx context.Context, vectors []float64, opts ...vectordb.SearchOption) ([]vectordb.Record, error) {
	var option vectordb.SearchOptions
	for _, opt := range opts {
		opt(&option)
	}
	col, err := e.Collection(ctx, option.Collection)
	if err != nil {
		return nil, err
	}
	records := filterRecords(col.Records(), &option)
	for idx, record := range records {
		records[idx].Score = calculateDistance(vectors, record.Embedding.Embedding)
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].Score < records[j].Score
	})
	if option.TopK == 0 {
		option.TopK = e.TopK
	}
	topK := min(option.TopK, len(records))
	return records[:topK], nil
}

// filterRecords filters a map of documents by metadata and content.
// It does this concurrently.
func filterRecords(docs []vectordb.Record, opts *vectordb.SearchOptions) []vectordb.Record {
	filteredDocs := make([]vectordb.Record, 0, len(docs))
	filteredDocsLock := sync.Mutex{}

	// Determine concurrency. Use number of docs or CPUs, whichever is smaller.
	numCPUs := runtime.NumCPU()
	numDocs := len(docs)
	concurrency := min(numCPUs, numDocs)

	docChan := make(chan vectordb.Record, concurrency*2)

	wg := sync.WaitGroup{}
	for range concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for doc := range docChan {
				if recordMatchesFilters(&doc, opts) {
					filteredDocsLock.Lock()
					filteredDocs = append(filteredDocs, doc)
					filteredDocsLock.Unlock()
				}
			}
		}()
	}

	for _, doc := range docs {
		docChan <- doc
	}
	close(docChan)

	wg.Wait()

	// With filteredDocs being initialized as potentially large slice, let's return
	// nil instead of the empty slice.
	if len(filteredDocs) == 0 {
		filteredDocs = nil
	}
	return filteredDocs
}

// recordMatchesFilters checks if a document matches the given filters.
// When calling this function, the whereDocument keys must already be validated!
func recordMatchesFilters(record *vectordb.Record, opts *vectordb.SearchOptions) bool {
	// A document's metadata must have *all* the fields in the where clause.
	for k, v := range opts.Meta {
		// TODO: Do we want to check for existence of the key? I.e. should
		// a where clause with empty string as value match a document's
		// metadata that doesn't have the key at all?
		if record.Embedding.Meta[k] != v {
			return false
		}
	}
	if opts.Include != "" {
		if !strings.Contains(record.Embedding.Object, opts.Include) {
			return false
		}
	}
	if opts.Exclude != "" {
		if !strings.Contains(record.Embedding.Object, opts.Exclude) {
			return false
		}
	}

	return true
}

// calculateDistance computes the distance between two vectors using the specified metric.
// Supported metrics:
// - "L2": Euclidean distance (default)
// - "IP": Inner product (negative, as larger means more similar)
// Returns a float64 representing the distance/similarity score.
func calculateDistance(a, b []float64) float64 {
	var sum float64
	// Default to L2
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}

// euclideanDistance computes the L2 (Euclidean) distance between two vectors.
// This is a helper function used by calculateDistance when metricType is "L2".
func euclideanDistance(a, b []float64) float64 {
	var sum float64
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}
