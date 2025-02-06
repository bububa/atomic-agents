package agents

import (
	"context"
	"errors"

	"github.com/bububa/instructor-go/pkg/instructor"
	cohere "github.com/cohere-ai/cohere-go/v2"
	anthropic "github.com/liushuangls/go-anthropic/v2"
	openai "github.com/sashabaranov/go-openai"

	"github.com/bububa/atomic-agents/components"
	"github.com/bububa/atomic-agents/components/systemprompt"
	"github.com/bububa/atomic-agents/components/systemprompt/cot"
	"github.com/bububa/atomic-agents/schema"
)

type IAgent interface {
	Name() string
}

type ChainableAgent interface {
	IAgent
	RunForChain(context.Context, any, *components.ApiResponse) (any, error)
}

type AgentSetter interface {
	SetClient(clt instructor.Instructor)
	SetMemory(m *components.Memory)
	SetSystemPromptGenerator(g *systemprompt.Generator)
	SetModel(model string)
	SetTemperature(temperature float32)
	SetMaxTokens(maxTokens int)
}

// Config represents general agents configuration
type Config struct {
	// client Client for interacting with the language model
	client instructor.Instructor
	//	memory  Memory component for storing chat history.
	memory *components.Memory
	//	systemPromptGenerator Component for generating system prompts.
	systemPromptGenerator systemprompt.Generator
	// initialMemory (AgentMemory): Initial state of the memory
	initialMemory *components.Memory
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
	endHook   func(context.Context, *Agent[I, O], *I, *O, *components.ApiResponse)
	errorHook func(context.Context, *Agent[I, O], *I, *components.ApiResponse, error)
}

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
	ret.memory.Copy(ret.initialMemory)
	return ret
}

// ResetMemory resets the memory to its initial state
func (a *Agent[I, O]) ResetMemory() {
	a.memory.Reset()
}

func (a *Agent[I, O]) SetClient(clt instructor.Instructor) {
	a.client = clt
}

func (a *Agent[I, O]) SetMemory(m *components.Memory) {
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

func (a *Agent[I, O]) SetEndHook(fn func(context.Context, *Agent[I, O], *I, *O, *components.ApiResponse)) {
	a.endHook = fn
}

func (a *Agent[I, O]) SetErrorHook(fn func(context.Context, *Agent[I, O], *I, *components.ApiResponse, error)) {
	a.errorHook = fn
}

// Response obtains a response from the language model synchronously
func (a *Agent[I, O]) response(ctx context.Context, response *O, apiResponse *components.ApiResponse) error {
	messages := make([]components.Message, 0, a.memory.MessageCount()+1)
	msg := components.NewMessage(components.SystemRole, schema.String(a.systemPromptGenerator.Generate()))
	messages = append(messages, *msg)
	messages = append(messages, a.memory.History()...)
	switch clt := a.client.(type) {
	case *instructor.InstructorOpenAI:
		chatReq := openai.ChatCompletionRequest{
			Model:               a.model,
			Temperature:         a.temperature,
			MaxCompletionTokens: a.maxTokens,
		}
		for _, msg := range messages {
			v := new(openai.ChatCompletionMessage)
			msg.ToOpenAI(v)
			chatReq.Messages = append(chatReq.Messages, *v)
		}
		if res, err := clt.CreateChatCompletion(ctx, chatReq, response); err != nil {
			return err
		} else if apiResponse != nil {
			apiResponse.FromOpenAI(&res)
		}
	case *instructor.InstructorAnthropic:
		chatReq := anthropic.MessagesRequest{
			Model:       anthropic.Model(a.model),
			Temperature: &a.temperature,
			MaxTokens:   a.maxTokens,
		}
		for _, msg := range messages {
			v := new(anthropic.Message)
			msg.ToAnthropic(v)
			chatReq.Messages = append(chatReq.Messages, *v)
		}
		if res, err := clt.CreateMessages(ctx, chatReq, response); err != nil {
			return err
		} else if apiResponse != nil {
			apiResponse.FromAnthropic(&res)
		}
	case *instructor.InstructorCohere:
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
			v := new(cohere.Message)
			msg.ToCohere(v)
			chatReq.ChatHistory = append(chatReq.ChatHistory, v)
		}
		if res, err := clt.Chat(ctx, &chatReq, response); err != nil {
			return err
		} else if apiResponse != nil {
			apiResponse.FromCohere(res)
		}
	}
	return nil
}

// Run runs the chat agent with the given user input synchronously.
func (a *Agent[I, O]) Run(ctx context.Context, userInput *I, output *O, apiResp *components.ApiResponse) error {
	if fn := a.startHook; fn != nil {
		fn(ctx, a, userInput)
	}
	if userInput != nil {
		a.memory.NewTurn()
		a.memory.NewMessage(components.UserRole, *userInput)
	}
	if err := a.response(ctx, output, apiResp); err != nil {
		if fn := a.errorHook; fn != nil {
			fn(ctx, a, userInput, apiResp, err)
		}
		return err
	}
	a.memory.NewMessage(components.AssistantRole, *output)
	if fn := a.endHook; fn != nil {
		fn(ctx, a, userInput, output, apiResp)
	}
	return nil
}

// Run runs the chat agent with the given user input for chain.
func (a *Agent[I, O]) RunForChain(ctx context.Context, userInput any, apiResp *components.ApiResponse) (any, error) {
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
