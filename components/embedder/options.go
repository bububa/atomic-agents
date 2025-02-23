package embedder

// Options holds the configuration for creating an Embedder instance.
// It supports multiple embedding providers and their specific options.
type Options struct {
	// provider specifies the embedding service to use (e.g., "openai", "cohere")
	provider Provider
	// model specifies the model to use
	model string
}

// Option is a function type for configuring the EmbedderConfig.
// It follows the functional options pattern for clean and flexible configuration.
type Option func(*Options)

func WithProvider(provider Provider) Option {
	return func(o *Options) {
		o.provider = provider
	}
}

func WithModel(model string) Option {
	return func(o *Options) {
		o.model = model
	}
}

func (i Options) Provider() Provider {
	return i.provider
}

func (i Options) Model() string {
	return i.model
}
