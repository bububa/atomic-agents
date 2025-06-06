package agents

import (
	"context"
	"errors"

	"github.com/bububa/instructor-go"
	anthropicClt "github.com/bububa/instructor-go/instructors/anthropic"
	cohereClt "github.com/bububa/instructor-go/instructors/cohere"
	"github.com/bububa/instructor-go/instructors/gemini"
	geminiClt "github.com/bububa/instructor-go/instructors/gemini"
	openaiClt "github.com/bububa/instructor-go/instructors/openai"
	cohere "github.com/cohere-ai/cohere-go/v2"
	anthropic "github.com/liushuangls/go-anthropic/v2"
	"github.com/openai/openai-go"
	geminiAPI "google.golang.org/genai"

	"github.com/bububa/atomic-agents/components"
	"github.com/bububa/atomic-agents/components/systemprompt"
	"github.com/bububa/atomic-agents/components/systemprompt/cot"
	"github.com/bububa/atomic-agents/schema"
)

type MergeResponse = func(*components.LLMResponse)

type IAgent interface {
	Name() string
}

type TypeableAgent[I schema.Schema, O schema.Schema] interface {
	IAgent
	Run(context.Context, *I, *O, *components.LLMResponse) error
}

type StreamableAgent[I schema.Schema, O schema.Schema] interface {
	IAgent
	Stream(context.Context, *I) (<-chan instructor.StreamData, MergeResponse, error)
}

type AnonymousAgent interface {
	IAgent
	RunAnonymous(context.Context, any, *components.LLMResponse) (any, error)
}

type AnonymousStreamableAgent interface {
	AnonymousAgent
	StreamAnonymous(context.Context, any) (<-chan instructor.StreamData, MergeResponse, error)
}

type AgentSetter interface {
	SetClient(instructor.Instructor)
	SetMemory(*instructor.Memory)
	SetSystemPromptGenerator(systemprompt.Generator)
	SetModel(string)
	SetTemperature(float64)
	SetTopP(float64)
	SetTopK(int)
	SetMaxTokens(int)
}

// Config represents general agents configuration
type Config struct {
	// client Client for interacting with the language model
	client instructor.Instructor
	//	systemPromptGenerator Component for generating system prompts.
	systemPromptGenerator systemprompt.Generator
	// currentUserInput
	// currentUserInput schema.Schema
	// model llm model
	model string
	// temperature Temperature for response generation, typically ranging from 0 to 1.
	temperature float64
	// topP
	topP float64
	// topK
	topK int
	// maxTokens Maximum number of tokens allowed in the response
	maxTokens int
	// name is Agent name presentation
	name string
}

// Agent class for chat agents.
// This class provides the core functionality for handling chat interactions, including managing memory,
// generating system prompts, and obtaining responses from a language model.
type Agent[I schema.Schema, O schema.Schema] struct {
	Config
	startHook func(context.Context, *Agent[I, O], *I)
	endHook   func(context.Context, *Agent[I, O], *I, *O, *components.LLMResponse)
	errorHook func(context.Context, *Agent[I, O], *I, *components.LLMResponse, error)
}

var (
	_ TypeableAgent[schema.String, schema.String]   = (*Agent[schema.String, schema.String])(nil)
	_ StreamableAgent[schema.String, schema.String] = (*Agent[schema.String, schema.String])(nil)
	_ AnonymousAgent                                = (*Agent[schema.String, schema.String])(nil)
	_ AnonymousStreamableAgent                      = (*Agent[schema.String, schema.String])(nil)
	_ AgentSetter                                   = (*Agent[schema.String, schema.String])(nil)
)

// NewAgent initializes the AgentAgent
func NewAgent[I schema.Schema, O schema.Schema](options ...Option) *Agent[I, O] {
	ret := new(Agent[I, O])
	for _, opt := range options {
		opt(&ret.Config)
	}
	if ret.systemPromptGenerator == nil {
		ret.systemPromptGenerator = cot.New()
	}
	return ret
}

func (a *Agent[I, O]) Model() string {
	return a.model
}

func (a *Agent[I, O]) Memory() *instructor.Memory {
	if a.client == nil {
		return nil
	}
	return a.client.Memory()
}

// AddToMemory add message to memory
func (a *Agent[I, O]) AddToMemory(msg instructor.Message) {
	if memory := a.Memory(); memory != nil {
		memory.Add(msg)
	}
}

func (a *Agent[I, O]) SetClient(clt instructor.Instructor) {
	a.client = clt
}

func (a *Agent[I, O]) Client() instructor.Instructor {
	return a.client
}

func (a *Agent[I, O]) Encoder() instructor.Encoder {
	return a.Client().Encoder()
}

func (a *Agent[I, O]) SetMemory(m *instructor.Memory) {
	if a.client != nil {
		a.client.SetMemory(m)
	}
}

func (a *Agent[I, O]) SetSystemPromptGenerator(g systemprompt.Generator) {
	a.systemPromptGenerator = g
}

func (a *Agent[I, O]) SetModel(model string) {
	a.model = model
}

func (a *Agent[I, O]) SetTemperature(temperature float64) {
	a.temperature = temperature
}

func (a *Agent[I, O]) SetTopP(topP float64) {
	a.topP = topP
}

func (a *Agent[I, O]) SetTopK(topK int) {
	a.topK = topK
}

func (a *Agent[I, O]) SetMaxTokens(maxTokens int) {
	a.maxTokens = maxTokens
}

func (a Agent[I, O]) Name() string {
	return a.name
}

func (a *Agent[I, O]) SetName(name string) {
	a.name = name
}

func (a *Agent[I, O]) SetStartHook(fn func(context.Context, *Agent[I, O], *I)) {
	a.startHook = fn
}

func (a *Agent[I, O]) SetEndHook(fn func(context.Context, *Agent[I, O], *I, *O, *components.LLMResponse)) {
	a.endHook = fn
}

func (a *Agent[I, O]) SetErrorHook(fn func(context.Context, *Agent[I, O], *I, *components.LLMResponse, error)) {
	a.errorHook = fn
}

// Response obtains a response from the language model synchronously
func (a *Agent[I, O]) chat(ctx context.Context, userInput *I, response *O, llmResponse *components.LLMResponse) error {
	sysMsg := instructor.Message{
		Role: instructor.SystemRole,
		Text: a.systemPromptGenerator.Generate(),
	}
	history := a.Memory().List()
	messages := make([]instructor.Message, 0, len(history)+2)
	messages = append(messages, sysMsg)
	messages = append(messages, history...)
	inputMsg := instructor.Message{
		Role: instructor.UserRole,
	}
	schema.ToMessage(*userInput, &inputMsg)
	messages = append(messages, inputMsg)
	switch clt := a.client.(type) {
	case instructor.ChatInstructor[openai.ChatCompletionNewParams, openai.ChatCompletion]:
		chatReq := openai.ChatCompletionNewParams{
			Model:       a.model,
			Temperature: openai.Float(a.temperature),
		}
		if a.temperature > 1e-15 {
			chatReq.Temperature = openai.Float(a.temperature)
		}
		if a.topP > 1e-15 {
			chatReq.TopP = openai.Float(a.topP)
		}
		if a.topK > 0 {
			chatReq.TopLogprobs = openai.Int(int64(a.topK))
		}
		if a.maxTokens > 0 {
			chatReq.MaxCompletionTokens = openai.Int(int64(a.maxTokens))
		}
		if extraBody := (*userInput).ExtraBody(); extraBody != nil {
			chatReq.SetExtraFields(map[string]any{
				"extra_body": extraBody,
			})
		}
		for _, msg := range messages {
			chunks := openaiClt.ConvertMessageFrom(&msg)
			if len(chunks) > 0 {
				chatReq.Messages = append(chatReq.Messages, chunks...)
			}
		}
		res := new(openai.ChatCompletion)
		if err := clt.Chat(ctx, &chatReq, response, res); err != nil {
			return err
		} else if llmResponse != nil {
			llmResponse.FromOpenAI(res)
		}
	case instructor.ChatInstructor[anthropic.MessagesRequest, anthropic.MessagesResponse]:
		chatReq := anthropic.MessagesRequest{
			Model: anthropic.Model(a.model),
		}
		if v := float32(a.temperature); v > 1e-15 {
			chatReq.Temperature = &v
		}
		if v := float32(a.topP); v > 1e-15 {
			chatReq.TopP = &v
		}
		if v := a.topK; v > 0 {
			chatReq.TopK = &v
		}
		if v := a.maxTokens; v > 0 {
			chatReq.MaxTokens = v
		}
		for _, msg := range messages {
			var v anthropic.Message
			anthropicClt.ConvertMessageFrom(&msg, &v)
			chatReq.Messages = append(chatReq.Messages, v)
		}
		res := new(anthropic.MessagesResponse)
		if err := clt.Chat(ctx, &chatReq, response, res); err != nil {
			return err
		} else if llmResponse != nil {
			llmResponse.FromAnthropic(res)
		}
	case instructor.ChatInstructor[cohere.ChatRequest, cohere.NonStreamedChatResponse]:
		lastHistoryIdx := len(messages) - 1
		chatReq := cohere.ChatRequest{
			Model:   &a.model,
			Message: schema.Stringify(*userInput),
		}
		if v := float64(a.temperature); v > 1e-15 {
			chatReq.Temperature = &v
		}
		if v := float64(a.topP); v > 1e-15 {
			chatReq.P = &v
		}
		if v := a.maxTokens; v > 0 {
			chatReq.MaxTokens = &v
		}

		for idx, msg := range messages {
			if idx >= lastHistoryIdx {
				break
			}
			var v cohere.Message
			cohereClt.ConvertMessageFrom(&msg, &v)
			chatReq.ChatHistory = append(chatReq.ChatHistory, &v)
		}
		res := new(cohere.NonStreamedChatResponse)
		if err := clt.Chat(ctx, &chatReq, response, res); err != nil {
			return err
		} else if llmResponse != nil {
			llmResponse.FromCohere(res)
		}
	case instructor.ChatInstructor[geminiClt.Request, geminiAPI.GenerateContentResponse]:
		chatReq := geminiClt.Request{
			Model: a.model,
		}
		{
			var v geminiAPI.Content
			geminiClt.ConvertMessageFrom(&sysMsg, &v)
			chatReq.System = &v
		}
		chatReq.History = make([]*geminiAPI.Content, 0, len(history))
		for _, msg := range history {
			var v geminiAPI.Content
			geminiClt.ConvertMessageFrom(&msg, &v)
			chatReq.History = append(chatReq.History, &v)
		}
		if userInput != nil {
			var v geminiAPI.Content
			geminiClt.ConvertMessageFrom(&inputMsg, &v)
			chatReq.Parts = append(chatReq.Parts, v.Parts...)
		}
		res := new(geminiAPI.GenerateContentResponse)
		if err := clt.Chat(ctx, &chatReq, response, res); err != nil {
			return err
		} else if llmResponse != nil {
			llmResponse.FromGemini(res)
		}
	}
	if llmResponse != nil && llmResponse.Model == "" {
		llmResponse.Model = a.model
	}
	return nil
}

// Response obtains a response from the language model synchronously
func (a *Agent[I, O]) stream(ctx context.Context, userInput *I) (<-chan instructor.StreamData, MergeResponse, error) {
	sysMsg := instructor.Message{
		Role: instructor.SystemRole,
		Text: a.systemPromptGenerator.Generate(),
	}
	inputMsg := instructor.Message{
		Role: instructor.UserRole,
	}
	schema.ToMessage(*userInput, &inputMsg)
	history := a.Memory().List()
	messages := make([]instructor.Message, 0, len(history)+2)
	messages = append(messages, sysMsg)
	messages = append(messages, history...)
	messages = append(messages, inputMsg)
	respType := new(O)
	switch clt := a.client.(type) {
	case instructor.StreamInstructor[openai.ChatCompletionNewParams, openai.ChatCompletion]:
		llmResp := new(openai.ChatCompletion)
		mergeResp := func(resp *components.LLMResponse) {
			if resp == nil {
				return
			}
			resp.FromOpenAI(llmResp)
			if resp.Model == "" {
				resp.Model = a.model
			}
		}
		chatReq := openai.ChatCompletionNewParams{
			Model: a.model,
		}
		if a.temperature > 1e-15 {
			chatReq.Temperature = openai.Float(a.temperature)
		}
		if a.topP > 1e-15 {
			chatReq.TopP = openai.Float(a.topP)
		}
		if a.topK > 0 {
			chatReq.TopLogprobs = openai.Int(int64(a.topK))
		}
		if a.maxTokens > 0 {
			chatReq.MaxCompletionTokens = openai.Int(int64(a.maxTokens))
		}
		if extraBody := (*userInput).ExtraBody(); extraBody != nil {
			chatReq.SetExtraFields(map[string]any{
				"extra_body": extraBody,
			})
		}
		for _, msg := range messages {
			chunks := openaiClt.ConvertMessageFrom(&msg)
			if len(chunks) > 0 {
				chatReq.Messages = append(chatReq.Messages, chunks...)
			}
		}
		ch, err := clt.Stream(ctx, &chatReq, respType, llmResp)
		if err != nil {
			return nil, mergeResp, err
		}
		return ch, mergeResp, nil
	case instructor.StreamInstructor[anthropic.MessagesRequest, anthropic.MessagesResponse]:
		llmResp := new(anthropic.MessagesResponse)
		mergeResp := func(resp *components.LLMResponse) {
			if resp == nil {
				return
			}
			resp.FromAnthropic(llmResp)
			if resp.Model == "" {
				resp.Model = a.model
			}
		}
		chatReq := anthropic.MessagesRequest{
			Model: anthropic.Model(a.model),
		}
		if v := float32(a.temperature); v > 1e-15 {
			chatReq.Temperature = &v
		}
		if v := float32(a.topP); v > 1e-15 {
			chatReq.TopP = &v
		}
		if v := a.topK; v > 0 {
			chatReq.TopK = &v
		}
		if v := a.maxTokens; v > 0 {
			chatReq.MaxTokens = v
		}
		for _, msg := range messages {
			var v anthropic.Message
			anthropicClt.ConvertMessageFrom(&msg, &v)
			chatReq.Messages = append(chatReq.Messages, v)
		}
		ch, err := clt.Stream(ctx, &chatReq, respType, llmResp)
		if err != nil {
			return nil, mergeResp, err
		}
		return ch, mergeResp, nil
	case instructor.StreamInstructor[cohere.ChatRequest, cohere.NonStreamedChatResponse]:
		llmResp := new(cohere.NonStreamedChatResponse)
		mergeResp := func(resp *components.LLMResponse) {
			if resp == nil {
				return
			}
			resp.FromCohere(llmResp)
			if resp.Model == "" {
				resp.Model = a.model
			}
		}
		lastHistoryIdx := len(messages) - 1
		chatReq := cohere.ChatRequest{
			Model:   &a.model,
			Message: schema.Stringify(*userInput),
		}
		if v := float64(a.temperature); v > 1e-15 {
			chatReq.Temperature = &v
		}
		if v := float64(a.topP); v > 1e-15 {
			chatReq.P = &v
		}
		if v := a.maxTokens; v > 0 {
			chatReq.MaxTokens = &v
		}
		for idx, msg := range messages {
			if idx >= lastHistoryIdx {
				break
			}
			var v cohere.Message
			cohereClt.ConvertMessageFrom(&msg, &v)
			chatReq.ChatHistory = append(chatReq.ChatHistory, &v)
		}
		ch, err := clt.Stream(ctx, &chatReq, respType, llmResp)
		if err != nil {
			return nil, mergeResp, err
		}
		return ch, mergeResp, nil
	case instructor.StreamInstructor[geminiClt.Request, geminiAPI.GenerateContentResponse]:
		llmResp := new(geminiAPI.GenerateContentResponse)
		mergeResp := func(resp *components.LLMResponse) {
			if resp == nil {
				return
			}
			resp.FromGemini(llmResp)
			if resp.Model == "" {
				resp.Model = a.model
			}
		}
		chatReq := geminiClt.Request{
			Model: a.model,
		}
		{
			var v geminiAPI.Content
			geminiClt.ConvertMessageFrom(&sysMsg, &v)
			chatReq.System = &v
		}
		chatReq.History = make([]*geminiAPI.Content, 0, len(history))
		for _, msg := range history {
			var v geminiAPI.Content
			geminiClt.ConvertMessageFrom(&msg, &v)
			chatReq.History = append(chatReq.History, &v)
		}
		var v geminiAPI.Content
		geminiClt.ConvertMessageFrom(&inputMsg, &v)
		chatReq.Parts = append(chatReq.Parts, v.Parts...)
		ch, err := clt.Stream(ctx, &chatReq, respType, llmResp)
		if err != nil {
			return nil, mergeResp, err
		}
		return ch, mergeResp, nil
	}
	return nil, nil, errors.New("unknown instructor provider")
}

// Response obtains a response from the language model synchronously
func (a *Agent[I, O]) schemaStream(ctx context.Context, userInput *I) (<-chan any, <-chan instructor.StreamData, MergeResponse, error) {
	sysMsg := instructor.Message{
		Role: instructor.SystemRole,
		Text: a.systemPromptGenerator.Generate(),
	}
	inputMsg := instructor.Message{
		Role: instructor.UserRole,
	}
	schema.ToMessage(*userInput, &inputMsg)
	history := a.Memory().List()
	messages := make([]instructor.Message, 0, len(history)+2)
	messages = append(messages, sysMsg)
	messages = append(messages, history...)
	messages = append(messages, inputMsg)
	var responseType O
	switch clt := a.client.(type) {
	case instructor.SchemaStreamInstructor[openai.ChatCompletionNewParams, openai.ChatCompletion]:
		llmResp := new(openai.ChatCompletion)
		mergeResp := func(resp *components.LLMResponse) {
			if resp == nil {
				return
			}
			resp.FromOpenAI(llmResp)
			if resp.Model == "" {
				resp.Model = a.model
			}
		}
		chatReq := openai.ChatCompletionNewParams{
			Model: a.model,
		}
		if a.temperature > 1e-15 {
			chatReq.Temperature = openai.Float(a.temperature)
		}
		if a.topP > 1e-15 {
			chatReq.TopP = openai.Float(a.topP)
		}
		if a.topK > 0 {
			chatReq.TopLogprobs = openai.Int(int64(a.topK))
		}
		if a.maxTokens > 0 {
			chatReq.MaxCompletionTokens = openai.Int(int64(a.maxTokens))
		}
		if extraBody := (*userInput).ExtraBody(); extraBody != nil {
			chatReq.SetExtraFields(map[string]any{
				"extra_body": extraBody,
			})
		}
		for _, msg := range messages {
			chunks := openaiClt.ConvertMessageFrom(&msg)
			if len(chunks) > 0 {
				chatReq.Messages = append(chatReq.Messages, chunks...)
			}
		}
		ch, stream, err := clt.SchemaStream(ctx, &chatReq, responseType, llmResp)
		if err != nil {
			return nil, nil, mergeResp, err
		}
		return ch, stream, mergeResp, nil
	case instructor.SchemaStreamInstructor[anthropic.MessagesRequest, anthropic.MessagesResponse]:
		llmResp := new(anthropic.MessagesResponse)
		mergeResp := func(resp *components.LLMResponse) {
			if resp == nil {
				return
			}
			resp.FromAnthropic(llmResp)
			if resp.Model == "" {
				resp.Model = a.model
			}
		}
		chatReq := anthropic.MessagesRequest{
			Model: anthropic.Model(a.model),
		}
		if v := float32(a.temperature); v > 1e-15 {
			chatReq.Temperature = &v
		}
		if v := float32(a.topP); v > 1e-15 {
			chatReq.TopP = &v
		}
		if v := a.topK; v > 0 {
			chatReq.TopK = &v
		}
		if v := a.maxTokens; v > 0 {
			chatReq.MaxTokens = v
		}
		for _, msg := range messages {
			var v anthropic.Message
			anthropicClt.ConvertMessageFrom(&msg, &v)
			chatReq.Messages = append(chatReq.Messages, v)
		}
		ch, stream, err := clt.SchemaStream(ctx, &chatReq, responseType, llmResp)
		if err != nil {
			return nil, nil, mergeResp, err
		}
		return ch, stream, mergeResp, nil
	case instructor.SchemaStreamInstructor[cohere.ChatRequest, cohere.NonStreamedChatResponse]:
		llmResp := new(cohere.NonStreamedChatResponse)
		mergeResp := func(resp *components.LLMResponse) {
			if resp == nil {
				return
			}
			resp.FromCohere(llmResp)
			if resp.Model == "" {
				resp.Model = a.model
			}
		}
		lastHistoryIdx := len(messages) - 1
		chatReq := cohere.ChatRequest{
			Model:   &a.model,
			Message: schema.Stringify(*userInput),
		}
		if v := float64(a.temperature); v > 1e-15 {
			chatReq.Temperature = &v
		}
		if v := float64(a.topP); v > 1e-15 {
			chatReq.P = &v
		}
		if v := a.maxTokens; v > 0 {
			chatReq.MaxTokens = &v
		}
		for idx, msg := range messages {
			if idx >= lastHistoryIdx {
				break
			}
			var v cohere.Message
			cohereClt.ConvertMessageFrom(&msg, &v)
			chatReq.ChatHistory = append(chatReq.ChatHistory, &v)
		}
		ch, stream, err := clt.SchemaStream(ctx, &chatReq, responseType, llmResp)
		if err != nil {
			return nil, nil, mergeResp, err
		}
		return ch, stream, mergeResp, nil
	case instructor.SchemaStreamInstructor[geminiClt.Request, geminiAPI.GenerateContentResponse]:
		llmResp := new(geminiAPI.GenerateContentResponse)
		mergeResp := func(resp *components.LLMResponse) {
			if resp == nil {
				return
			}
			resp.FromGemini(llmResp)
			if resp.Model == "" {
				resp.Model = a.model
			}
		}
		chatReq := gemini.Request{
			Model: a.model,
		}
		{
			var v geminiAPI.Content
			geminiClt.ConvertMessageFrom(&sysMsg, &v)
			chatReq.System = &v
		}
		chatReq.History = make([]*geminiAPI.Content, 0, len(history))
		for _, msg := range history {
			var v geminiAPI.Content
			geminiClt.ConvertMessageFrom(&msg, &v)
			chatReq.History = append(chatReq.History, &v)
		}
		{
			var v geminiAPI.Content
			geminiClt.ConvertMessageFrom(&inputMsg, &v)
			chatReq.Parts = append(chatReq.Parts, v.Parts...)
		}
		ch, stream, err := clt.SchemaStream(ctx, &chatReq, responseType, llmResp)
		if err != nil {
			return nil, nil, mergeResp, err
		}
		return ch, stream, mergeResp, nil
	}
	return nil, nil, nil, errors.New("unknown instructor provider")
}

// Run runs the chat agent with the given user input synchronously.
func (a *Agent[I, O]) Run(ctx context.Context, userInput *I, output *O, apiResp *components.LLMResponse) error {
	if fn := a.startHook; fn != nil {
		fn(ctx, a, userInput)
	}
	if apiResp == nil {
		apiResp = new(components.LLMResponse)
	}
	if err := a.chat(ctx, userInput, output, apiResp); err != nil {
		if fn := a.errorHook; fn != nil {
			fn(ctx, a, userInput, apiResp, err)
		}
		return err
	}
	if fn := a.endHook; fn != nil {
		fn(ctx, a, userInput, output, apiResp)
	}
	return nil
}

// Run runs the chat agent with the given user input for chain.
func (a *Agent[I, O]) RunAnonymous(ctx context.Context, userInput any, apiResp *components.LLMResponse) (any, error) {
	in, ok := userInput.(*I)
	if !ok {
		return nil, errors.New("invalid input schema")
	}
	out := new(O)
	if err := a.Run(ctx, in, out, apiResp); err != nil {
		return nil, err
	}
	return out, nil
}

// Run runs the chat agent with the given user input synchronously.
func (a *Agent[I, O]) Stream(ctx context.Context, userInput *I) (<-chan instructor.StreamData, MergeResponse, error) {
	if fn := a.startHook; fn != nil {
		fn(ctx, a, userInput)
	}
	ch, mergeResp, err := a.stream(ctx, userInput)
	if err != nil {
		if fn := a.errorHook; fn != nil {
			fn(ctx, a, userInput, nil, err)
		}
		return nil, mergeResp, err
	}
	if fn := a.endHook; fn != nil {
		fn(ctx, a, userInput, nil, nil)
	}
	return ch, mergeResp, nil
}

func (a *Agent[I, O]) StreamAnonymous(ctx context.Context, userInput any) (<-chan instructor.StreamData, MergeResponse, error) {
	in, ok := userInput.(*I)
	if !ok {
		return nil, nil, errors.New("invalid input schema")
	}
	ch, mergeResp, err := a.Stream(ctx, in)
	if err != nil {
		return nil, mergeResp, err
	}
	return ch, mergeResp, nil
}

// Run runs the chat agent with the given user input synchronously.
func (a *Agent[I, O]) SchemaStream(ctx context.Context, userInput *I) (<-chan any, <-chan instructor.StreamData, MergeResponse, error) {
	if fn := a.startHook; fn != nil {
		fn(ctx, a, userInput)
	}
	ch, stream, mergeResp, err := a.schemaStream(ctx, userInput)
	if err != nil {
		if fn := a.errorHook; fn != nil {
			fn(ctx, a, userInput, nil, err)
		}
		return nil, nil, mergeResp, err
	}
	if fn := a.endHook; fn != nil {
		fn(ctx, a, userInput, nil, nil)
	}
	return ch, stream, mergeResp, nil
}

// SystemPromptContextProvider returns agent systemPromptGenerator's context provider
func (a *Agent[I, O]) SystemPromptContextProvider(title string) (systemprompt.ContextProvider, error) {
	return a.systemPromptGenerator.ContextProvider(title)
}

// RegisterSystemPromptContextProvider registers a new context provider
func (a *Agent[I, O]) RegisterSystemPromptContextProvider(provider systemprompt.ContextProvider) {
	a.systemPromptGenerator.AddContextProviders(provider)
}

// RegisterSystemPromptContextProvider Unregisters an existing context provider.
func (a *Agent[I, O]) UnregisterSystemPromptContextProvider(title string) {
	a.systemPromptGenerator.RemoveContextProviders(title)
}

// SystemPrompt returns the system prompt
func (a *Agent[I, O]) SystemPrompt() string {
	return a.systemPromptGenerator.Generate()
}
