package components

import (
	"github.com/bububa/instructor-go"
	cohere "github.com/cohere-ai/cohere-go/v2"
	anthropic "github.com/liushuangls/go-anthropic/v2"
	"github.com/openai/openai-go"
	gemini "google.golang.org/genai"
)

// LLMResponse instructor provider chat response
type LLMResponse struct {
	ID        string          `json:"id,omitempty"`
	Role      instructor.Role `json:"role,omitempty"`
	Model     string          `json:"model,omitempty"`
	Usage     *LLMUsage       `json:"usage,omitempty"`
	Timestamp int64           `json:"ts,omitempty"`
	Details   any             `json:"content,omitempty"`
}

// FromOpenAI convnert response from openai
func (r *LLMResponse) FromOpenAI(v *openai.ChatCompletion) {
	r.ID = v.ID
	r.Role = instructor.AssistantRole
	r.Model = v.Model
	r.Usage = &LLMUsage{
		InputTokens:  v.Usage.PromptTokens,
		OutputTokens: v.Usage.CompletionTokens,
	}
	r.Details = v.Choices
}

// FromAnthropic convert response from anthropic
func (r *LLMResponse) FromAnthropic(v *anthropic.MessagesResponse) {
	r.ID = v.ID
	r.Role = instructor.AssistantRole
	r.Model = string(v.Model)
	r.Usage = &LLMUsage{
		InputTokens:  int64(v.Usage.InputTokens),
		OutputTokens: int64(v.Usage.OutputTokens),
	}
	r.Details = v.Content
}

// FromCohere convert response from cohere
func (r *LLMResponse) FromCohere(v *cohere.NonStreamedChatResponse) {
	if v.GenerationId != nil {
		r.ID = *v.GenerationId
	}
	r.Role = instructor.AssistantRole
	if meta := v.Meta; meta != nil {
		if usage := meta.Tokens; usage != nil {
			r.Usage = new(LLMUsage)
			if usage.InputTokens != nil {
				r.Usage.InputTokens = int64(*usage.InputTokens)
			}
			if usage.OutputTokens != nil {
				r.Usage.OutputTokens = int64(*usage.OutputTokens)
			}
		}
		if version := meta.ApiVersion; version != nil {
			r.Model = version.Version
		}
	}
	r.Details = v
}

func (r *LLMResponse) FromGemini(v *gemini.GenerateContentResponse) {
	r.Role = instructor.AssistantRole
	if v.UsageMetadata != nil && (v.UsageMetadata.PromptTokenCount > 0 || v.UsageMetadata.CandidatesTokenCount > 0) {
		r.Usage = new(LLMUsage)
		r.Usage.InputTokens = int64(v.UsageMetadata.PromptTokenCount)
		r.Usage.OutputTokens = int64(v.UsageMetadata.CachedContentTokenCount)
	}
	r.Details = v.Candidates
}

type LLMUsage struct {
	InputTokens  int64 `json:"input_tokens,omitempty"`
	OutputTokens int64 `json:"output_tokens,omitempty"`
}

func (u *LLMUsage) Merge(v *LLMUsage) {
	if v == nil {
		return
	}
	u.InputTokens += v.InputTokens
	u.OutputTokens += v.OutputTokens
}
