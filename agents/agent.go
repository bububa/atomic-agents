package agents

import (
	"context"
	"errors"

	"github.com/bububa/instructor-go"
	"github.com/bububa/instructor-go/instructors/gemini"
	cohere "github.com/cohere-ai/cohere-go/v2"
	anthropic "github.com/liushuangls/go-anthropic/v2"
	openai "github.com/sashabaranov/go-openai"
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
	SetMemory(components.MemoryStore)
	SetSystemPromptGenerator(systemprompt.Generator)
	SetModel(string)
	SetTemperature(float32)
	SetMaxTokens(int)
}

// Config represents general agents configuration
type Config struct {
	// client Client for interacting with the language model
	client instructor.Instructor
	//	memory  Memory component for storing chat history.
	memory components.MemoryStore
	//	systemPromptGenerator Component for generating system prompts.
	systemPromptGenerator systemprompt.Generator
	// initialMemory (AgentMemory): Initial state of the memory
	initialMemory components.MemoryStore
	// currentUserInput
	// currentUserInput schema.Schema
	// model llm model
	model string
	// temperature Temperature for response generation, typically ranging from 0 to 1.
	temperature float32
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
	if ret.memory == nil {
		ret.memory = components.NewMemory(0)
	}
	if ret.systemPromptGenerator == nil {
		ret.systemPromptGenerator = cot.New()
	}
	ret.initialMemory = components.NewMemory(0)
	ret.initialMemory.Copy(ret.memory)
	return ret
}

func (a *Agent[I, O]) Model() string {
	return a.model
}

// ResetMemory resets the memory to its initial state
func (a *Agent[I, O]) ResetMemory() {
	a.memory.Copy(a.initialMemory)
}

// ClearMemory resets the memory to its initial state
func (a *Agent[I, O]) ClearMemory() {
	a.memory.Reset()
}

func (a *Agent[I, O]) Memory() components.MemoryStore {
	return a.memory
}

// AddToMemory add message to memory
func (a *Agent[I, O]) AddToMemory(msg *components.Message) {
	a.memory.AddMessage(msg)
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

func (a *Agent[I, O]) SetMemory(m components.MemoryStore) {
	a.memory = m
}

func (a *Agent[I, O]) SetSystemPromptGenerator(g systemprompt.Generator) {
	a.systemPromptGenerator = g
}

func (a *Agent[I, O]) SetModel(model string) {
	a.model = model
}

func (a *Agent[I, O]) SetTemperature(temperature float32) {
	a.temperature = temperature
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
	sysMsg := components.NewMessage(components.SystemRole, *schema.NewString(a.systemPromptGenerator.Generate()))
	var history []components.Message
	if userInput != nil {
		a.memory.NewTurn()
		a.memory.NewMessage(components.UserRole, *userInput)
		history = make([]components.Message, a.memory.MessageCount())
		copy(history, a.memory.History())
	} else {
		history = a.memory.History()
	}
	messages := make([]components.Message, 0, a.memory.MessageCount()+1)
	messages = append(messages, *sysMsg)
	messages = append(messages, a.memory.History()...)
	switch clt := a.client.(type) {
	case instructor.ChatInstructor[openai.ChatCompletionRequest, openai.ChatCompletionResponse]:
		chatReq := openai.ChatCompletionRequest{
			Model:               a.model,
			Temperature:         a.temperature,
			MaxCompletionTokens: a.maxTokens,
		}
		if extraBody := (*userInput).ExtraBody(); extraBody != nil {
			chatReq.ExtraBody = extraBody
		}
		for _, msg := range messages {
			if msg.Mode() == "" {
				msg.SetMode(a.client.Mode())
			}
			v := new(openai.ChatCompletionMessage)
			chunks := msg.ToOpenAI(v)
			chatReq.Messages = append(chatReq.Messages, *v)
			if len(chunks) > 0 {
				chatReq.Messages = append(chatReq.Messages, chunks...)
			}
		}
		res := new(openai.ChatCompletionResponse)
		if err := clt.Chat(ctx, &chatReq, response, res); err != nil {
			return err
		} else if llmResponse != nil {
			llmResponse.FromOpenAI(res)
		}
	case instructor.ChatInstructor[anthropic.MessagesRequest, anthropic.MessagesResponse]:
		chatReq := anthropic.MessagesRequest{
			Model:       anthropic.Model(a.model),
			Temperature: &a.temperature,
			MaxTokens:   a.maxTokens,
		}
		for _, msg := range messages {
			if msg.Mode() == "" {
				msg.SetMode(a.client.Mode())
			}
			v := new(anthropic.Message)
			chunks := msg.ToAnthropic(v)
			chatReq.Messages = append(chatReq.Messages, *v)
			if len(chunks) > 0 {
				chatReq.Messages = append(chatReq.Messages, chunks...)
			}
		}
		res := new(anthropic.MessagesResponse)
		if err := clt.Chat(ctx, &chatReq, response, res); err != nil {
			return err
		} else if llmResponse != nil {
			llmResponse.FromAnthropic(res)
		}
	case instructor.ChatInstructor[cohere.ChatRequest, cohere.NonStreamedChatResponse]:
		lastIdx := len(messages) - 2
		temperature := float64(a.temperature)
		chatReq := cohere.ChatRequest{
			Model:       &a.model,
			Temperature: &temperature,
			MaxTokens:   &a.maxTokens,
			Message:     schema.Stringify(messages[lastIdx].Content()),
		}
		for idx, msg := range messages {
			if idx >= lastIdx {
				break
			}
			if msg.Mode() == "" {
				msg.SetMode(a.client.Mode())
			}
			v := new(cohere.Message)
			chunks := msg.ToCohere(v)
			chatReq.ChatHistory = append(chatReq.ChatHistory, v)
			if len(chunks) > 0 {
				chatReq.ChatHistory = append(chatReq.ChatHistory, chunks...)
			}
		}
		res := new(cohere.NonStreamedChatResponse)
		if err := clt.Chat(ctx, &chatReq, response, res); err != nil {
			return err
		} else if llmResponse != nil {
			llmResponse.FromCohere(res)
		}
	case instructor.ChatInstructor[gemini.Request, geminiAPI.GenerateContentResponse]:
		chatReq := gemini.Request{
			Model: a.model,
		}
		{
			v := new(geminiAPI.Content)
			sysMsg.ToGemini(v)
			chatReq.System = v
		}
		chatReq.History = make([]*geminiAPI.Content, 0, len(history))
		for _, msg := range history {
			if msg.Mode() == "" {
				msg.SetMode(a.client.Mode())
			}
			v := new(geminiAPI.Content)
			chunks := msg.ToGemini(v)
			chatReq.History = append(chatReq.History, v)
			if len(chunks) > 0 {
				chatReq.History = append(chatReq.History, chunks...)
			}
		}
		if userInput != nil {
			v := new(geminiAPI.Content)
			userMsg := components.NewMessage(components.UserRole, *userInput)
			chunks := userMsg.ToGemini(v)
			if l := len(chunks); l > 0 {
				chatReq.History = append(chatReq.History, v)
				if l > 1 {
					chatReq.History = append(chatReq.History, chunks[:l-1]...)
				}
				chatReq.Parts = append(chatReq.Parts, chunks[l].Parts...)
			} else {
				chatReq.Parts = append(chatReq.Parts, v.Parts...)
			}
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
	sysMsg := components.NewMessage(components.SystemRole, *schema.NewString(a.systemPromptGenerator.Generate()))
	var history []components.Message
	if userInput != nil {
		a.memory.NewTurn()
		a.memory.NewMessage(components.UserRole, *userInput)
		history = make([]components.Message, a.memory.MessageCount())
		copy(history, a.memory.History())
	} else {
		history = a.memory.History()
	}
	messages := make([]components.Message, 0, a.memory.MessageCount()+1)
	messages = append(messages, *sysMsg)
	messages = append(messages, a.memory.History()...)
	respType := new(O)
	switch clt := a.client.(type) {
	case instructor.StreamInstructor[openai.ChatCompletionRequest, openai.ChatCompletionResponse]:
		llmResp := new(openai.ChatCompletionResponse)
		mergeResp := func(resp *components.LLMResponse) {
			if resp == nil {
				return
			}
			resp.FromOpenAI(llmResp)
			if resp.Model == "" {
				resp.Model = a.model
			}
		}
		chatReq := openai.ChatCompletionRequest{
			Model:               a.model,
			Temperature:         a.temperature,
			MaxCompletionTokens: a.maxTokens,
		}
		if extraBody := (*userInput).ExtraBody(); extraBody != nil {
			chatReq.ExtraBody = extraBody
		}
		for _, msg := range messages {
			if msg.Mode() == "" {
				msg.SetMode(a.client.Mode())
			}
			v := new(openai.ChatCompletionMessage)
			chunks := msg.ToOpenAI(v)
			chatReq.Messages = append(chatReq.Messages, *v)
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
			Model:       anthropic.Model(a.model),
			Temperature: &a.temperature,
			MaxTokens:   a.maxTokens,
		}
		for _, msg := range messages {
			if msg.Mode() == "" {
				msg.SetMode(a.client.Mode())
			}
			v := new(anthropic.Message)
			chunks := msg.ToAnthropic(v)
			chatReq.Messages = append(chatReq.Messages, *v)
			if len(chunks) > 0 {
				chatReq.Messages = append(chatReq.Messages, chunks...)
			}
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
		lastIdx := len(messages) - 2
		temperature := float64(a.temperature)
		chatReq := cohere.ChatRequest{
			Model:       &a.model,
			Temperature: &temperature,
			MaxTokens:   &a.maxTokens,
			Message:     schema.Stringify(messages[lastIdx].Content()),
		}
		for idx, msg := range messages {
			if idx >= lastIdx {
				break
			}
			if msg.Mode() == "" {
				msg.SetMode(a.client.Mode())
			}
			v := new(cohere.Message)
			chunks := msg.ToCohere(v)
			chatReq.ChatHistory = append(chatReq.ChatHistory, v)
			if len(chunks) > 0 {
				chatReq.ChatHistory = append(chatReq.ChatHistory, chunks...)
			}
		}
		ch, err := clt.Stream(ctx, &chatReq, respType, llmResp)
		if err != nil {
			return nil, mergeResp, err
		}
		return ch, mergeResp, nil
	case instructor.StreamInstructor[gemini.Request, geminiAPI.GenerateContentResponse]:
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

			v := new(geminiAPI.Content)
			sysMsg.ToGemini(v)
			chatReq.System = v
		}
		chatReq.History = make([]*geminiAPI.Content, 0, len(history))
		for _, msg := range history {
			if msg.Mode() == "" {
				msg.SetMode(a.client.Mode())
			}
			v := new(geminiAPI.Content)
			chunks := msg.ToGemini(v)
			chatReq.History = append(chatReq.History, v)
			if len(chunks) > 0 {
				chatReq.History = append(chatReq.History, chunks...)
			}
		}
		if userInput != nil {
			v := new(geminiAPI.Content)
			userMsg := components.NewMessage(components.UserRole, *userInput)
			chunks := userMsg.ToGemini(v)
			if l := len(chunks); l > 0 {
				chatReq.History = append(chatReq.History, v)
				if l > 1 {
					chatReq.History = append(chatReq.History, chunks[:l-1]...)
				}
				chatReq.Parts = append(chatReq.Parts, chunks[l].Parts...)
			} else {
				chatReq.Parts = append(chatReq.Parts, v.Parts...)
			}
		}
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
	sysMsg := components.NewMessage(components.SystemRole, *schema.NewString(a.systemPromptGenerator.Generate()))
	var history []components.Message
	if userInput != nil {
		a.memory.NewTurn()
		a.memory.NewMessage(components.UserRole, *userInput)
		history = make([]components.Message, a.memory.MessageCount())
		copy(history, a.memory.History())
	} else {
		history = a.memory.History()
	}
	messages := make([]components.Message, 0, a.memory.MessageCount()+1)
	messages = append(messages, *sysMsg)
	messages = append(messages, a.memory.History()...)
	var responseType O
	switch clt := a.client.(type) {
	case instructor.SchemaStreamInstructor[openai.ChatCompletionRequest, openai.ChatCompletionResponse]:
		llmResp := new(openai.ChatCompletionResponse)
		mergeResp := func(resp *components.LLMResponse) {
			if resp == nil {
				return
			}
			resp.FromOpenAI(llmResp)
			if resp.Model == "" {
				resp.Model = a.model
			}
		}
		chatReq := openai.ChatCompletionRequest{
			Model:               a.model,
			Temperature:         a.temperature,
			MaxCompletionTokens: a.maxTokens,
		}
		if extraBody := (*userInput).ExtraBody(); extraBody != nil {
			chatReq.ExtraBody = extraBody
		}
		for _, msg := range messages {
			if msg.Mode() == "" {
				msg.SetMode(a.client.Mode())
			}
			v := new(openai.ChatCompletionMessage)
			chunks := msg.ToOpenAI(v)
			chatReq.Messages = append(chatReq.Messages, *v)
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
			Model:       anthropic.Model(a.model),
			Temperature: &a.temperature,
			MaxTokens:   a.maxTokens,
		}
		for _, msg := range messages {
			if msg.Mode() == "" {
				msg.SetMode(a.client.Mode())
			}
			v := new(anthropic.Message)
			chunks := msg.ToAnthropic(v)
			chatReq.Messages = append(chatReq.Messages, *v)
			if len(chunks) > 0 {
				chatReq.Messages = append(chatReq.Messages, chunks...)
			}
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
		lastIdx := len(messages) - 2
		temperature := float64(a.temperature)
		chatReq := cohere.ChatRequest{
			Model:       &a.model,
			Temperature: &temperature,
			MaxTokens:   &a.maxTokens,
			Message:     schema.Stringify(messages[lastIdx].Content()),
		}
		for idx, msg := range messages {
			if idx >= lastIdx {
				break
			}
			if msg.Mode() == "" {
				msg.SetMode(a.client.Mode())
			}
			v := new(cohere.Message)
			chunks := msg.ToCohere(v)
			chatReq.ChatHistory = append(chatReq.ChatHistory, v)
			if len(chunks) > 0 {
				chatReq.ChatHistory = append(chatReq.ChatHistory, chunks...)
			}
		}
		ch, stream, err := clt.SchemaStream(ctx, &chatReq, responseType, llmResp)
		if err != nil {
			return nil, nil, mergeResp, err
		}
		return ch, stream, mergeResp, nil
	case instructor.SchemaStreamInstructor[gemini.Request, geminiAPI.GenerateContentResponse]:
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

			v := new(geminiAPI.Content)
			sysMsg.ToGemini(v)
			chatReq.System = v
		}
		chatReq.History = make([]*geminiAPI.Content, 0, len(history))
		for _, msg := range history {
			if msg.Mode() == "" {
				msg.SetMode(a.client.Mode())
			}
			v := new(geminiAPI.Content)
			chunks := msg.ToGemini(v)
			chatReq.History = append(chatReq.History, v)
			if len(chunks) > 0 {
				chatReq.History = append(chatReq.History, chunks...)
			}
		}
		if userInput != nil {
			v := new(geminiAPI.Content)
			userMsg := components.NewMessage(components.UserRole, *userInput)
			chunks := userMsg.ToGemini(v)
			if l := len(chunks); l > 0 {
				chatReq.History = append(chatReq.History, v)
				if l > 1 {
					chatReq.History = append(chatReq.History, chunks[:l-1]...)
				}
				chatReq.Parts = append(chatReq.Parts, chunks[l].Parts...)
			} else {
				chatReq.Parts = append(chatReq.Parts, v.Parts...)
			}
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
	mode := a.client.Mode()
	if mode == instructor.ModeToolCall || mode == instructor.ModeToolCallStrict {
		a.memory.NewMessage(components.FunctionRole, *output)
	} else {
		msg := components.NewMessage(components.AssistantRole, *output)
		msg.SetMode(a.client.Mode())
		switch t := apiResp.Details.(type) {
		case *openai.ChatCompletionResponse:
			if len(t.Choices) > 0 {
				msg.SetRaw(t.Choices[0].Message.Content)
			}
		case *anthropic.CompleteResponse:
			msg.SetRaw(t.Completion)
		case *cohere.NonStreamedChatResponse:
			msg.SetRaw(t.Text)
		case *geminiAPI.GenerateContentResponse:
			for _, candidate := range t.Candidates {
				if candidate == nil {
					continue
				}
				for _, part := range candidate.Content.Parts {
					if txt := part.Text; txt != "" {
						msg.SetRaw(txt)
						break
					}
				}
			}
		}
    a.memory.AddMessage(msg)
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
	if userInput != nil {
		a.memory.NewTurn()
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
	if userInput != nil {
		a.memory.NewTurn()
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

func (a *Agent[I, O]) NewMessage(role components.MessageRole, content schema.Schema) *components.Message {
	return a.memory.NewMessage(role, content)
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
