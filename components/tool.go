package components

import (
	"encoding/json"

	anthropic "github.com/liushuangls/go-anthropic/v2"
	"github.com/openai/openai-go"
	gemini "google.golang.org/genai"
)

type ToolCall struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

func ToolCallsToOpenAI(src []ToolCall, dist *openai.ChatCompletionMessageParamUnion) {
	list := make([]openai.ChatCompletionMessageToolCallParam, 0, len(src))
	for _, v := range src {
		list = append(list, openai.ChatCompletionMessageToolCallParam{
			ID: v.ID,
			Function: openai.ChatCompletionMessageToolCallFunctionParam{
				Name:      v.Name,
				Arguments: v.Arguments,
			},
		})
	}
	msg := openai.ChatCompletionAssistantMessageParam{
		ToolCalls: list,
	}
	*dist = openai.ChatCompletionMessageParamUnion{
		OfAssistant: &msg,
	}
}

func ToolCallsToAnthropic(src []ToolCall, dist *anthropic.Message) {
	list := make([]anthropic.MessageContent, 0, len(src))
	for _, v := range src {
		list = append(list, anthropic.NewToolUseMessageContent(v.ID, v.Name, []byte(v.Arguments)))
	}
	*dist = anthropic.Message{
		Role:    anthropic.RoleAssistant,
		Content: list,
	}
}

func ToolCallsToGemini(src []ToolCall, dist *gemini.Content) {
	list := make([]*gemini.Part, 0, len(src))
	for _, v := range src {
		args := make(map[string]any)
		if err := json.Unmarshal([]byte(v.Arguments), &args); err == nil {
			list = append(list, gemini.NewPartFromFunctionCall(v.Name, args))
		}
	}
	content := gemini.NewContentFromParts(list, gemini.RoleModel)
	*dist = *content
}

type ToolCallback struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Content string `json:"content,omitempty"`
	IsError bool   `json:"is_error,omitempty"`
}

func ToolCallbacksToOpenAI(src []ToolCallback) []openai.ChatCompletionMessageParamUnion {
	list := make([]openai.ChatCompletionMessageParamUnion, 0, len(src))
	for _, v := range src {
		msg := openai.ToolMessage(v.Content, v.ID)
		list = append(list, msg)
	}
	return list
}

func ToolCallbacksToAnthropic(src []ToolCallback, dist *anthropic.Message) {
	list := make([]anthropic.MessageContent, 0, len(src))
	for _, v := range src {
		msg := anthropic.NewToolResultMessageContent(v.ID, v.Content, v.IsError)
		list = append(list, msg)
	}
	dist.Role = anthropic.RoleUser
	dist.Content = list
}

func ToolCallbacksToGemini(src []ToolCallback, dist *gemini.Content) {
	parts := make([]*gemini.Part, 0, len(src))
	for _, v := range src {
		args := make(map[string]any)
		if err := json.Unmarshal([]byte(v.Content), &args); err == nil {
			part := gemini.NewPartFromFunctionResponse(v.Name, args)
			parts = append(parts, part)
		}
	}
	dist.Role = gemini.RoleUser
	dist.Parts = parts
}
