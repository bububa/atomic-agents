package providers

import (
	"github.com/bububa/atomic-agents/components/embedder/providers/cohere"
	"github.com/bububa/atomic-agents/components/embedder/providers/gemini"
	"github.com/bububa/atomic-agents/components/embedder/providers/huggingface"
	"github.com/bububa/atomic-agents/components/embedder/providers/openai"
	"github.com/bububa/atomic-agents/components/embedder/providers/voyageai"
)

var (
	FromOpenAI      = openai.New
	FromVoyageAI    = voyageai.New
	FromCohere      = cohere.New
	FromGemini      = gemini.New
	FromHuggingFace = huggingface.New
)
