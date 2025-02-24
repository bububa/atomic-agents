package vectordb

type Options struct {
	EngineType EngineType // Database type (e.g., "milvus", "memory")
	TopK       int        // Maximum number of results to return
	MinScore   float64    // Minimum similarity score threshold
	UseHybrid  bool       // Enable hybrid search (vector + keyword)
	Columns    []string   // Columns to retrieve from the database
	Dimension  int        // Vector dimension
}

// Option is a function type for configuring VectorDB instances.
// It follows the functional options pattern for clean and flexible configuration.
type Option func(*Options)

// WithEngineType sets the database type.
// Supported types:
// - "milvus": Production-grade vector database
// - "memory": In-memory database for testing
// - "chromem": Chrome-based persistent storage
func WithEngine(engine EngineType) Option {
	return func(c *Options) {
		c.EngineType = engine
	}
}

// WithTopK sets the maximum number of results to return.
// The actual number of results may be less if MinScore filtering is applied.
//
// Example:
//
//	retriever, err := NewRetriever(
//	    WithTopK(10), // Return top 10 results
//	)
func WithTopK(k int) Option {
	return func(c *Options) {
		c.TopK = k
	}
}

// WithMinScore sets the minimum similarity score threshold.
// Results with scores below this threshold will be filtered out.
//
// Example:
//
//	retriever, err := NewRetriever(
//	    WithMinScore(0.8), // Only return high-confidence matches
//	)
func WithMinScore(score float64) Option {
	return func(c *Options) {
		c.MinScore = score
	}
}

// WithHybrid enables or disables hybrid search.
// Hybrid search combines vector similarity with keyword matching.
//
// Example:
//
//	retriever, err := NewRetriever(
//	    WithHybrid(true), // Enable hybrid search
//	)
func WithHybrid(enabled bool) Option {
	return func(c *Options) {
		c.UseHybrid = enabled
	}
}

// WithColumns specifies which columns to retrieve from the database.
// This can optimize performance by only fetching needed fields.
//
// Example:
//
//	retriever, err := NewRetriever(
//	    WithColumns("Text", "Metadata", "Source"),
//	)
func WithColumns(columns ...string) Option {
	return func(c *Options) {
		c.Columns = columns
	}
}

// WithDimension sets the dimension of vectors to be stored.
// This must match the dimension of your embedding model:
// - text-embedding-3-small: 1536
// - text-embedding-ada-002: 1536
// - Cohere embed-multilingual-v3.0: 1024
func WithDimension(dimension int) Option {
	return func(c *Options) {
		c.Dimension = dimension
	}
}
