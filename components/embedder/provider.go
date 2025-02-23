package embedder

type Provider = string

const (
	ProviderOpenAI      Provider = "OpenAI"
	ProviderVoyageAI    Provider = "VoyageAI"
	ProviderCohere      Provider = "Cohere"
	ProviderGemini      Provider = "Gemini"
	ProviderHuggingFace Provider = "HuggingFace"
)
