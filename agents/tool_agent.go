package agents

import (
	"context"
	"errors"

	"github.com/bububa/atomic-agents/components"
	"github.com/bububa/atomic-agents/components/systemprompt"
	"github.com/bububa/atomic-agents/schema"
	"github.com/bububa/atomic-agents/tools"
	"github.com/bububa/instructor-go"
)

// ToolAgent represent agent with tool callback
type ToolAgent[I schema.Schema, T schema.Schema, O schema.Schema] struct {
	name      string
	start     *Agent[I, T]
	end       *Agent[I, O]
	tool      tools.AnonymousTool
	startHook func(context.Context, *ToolAgent[I, T, O], *I)
	endHook   func(context.Context, *ToolAgent[I, T, O], *I, *O, *components.LLMResponse)
	errorHook func(context.Context, *ToolAgent[I, T, O], *I, *components.LLMResponse, error)
}

// NewToolAgent returns a new ToolAgent instance
func NewToolAgent[I schema.Schema, T schema.Schema, O schema.Schema](options ...Option) *ToolAgent[I, T, O] {
	return &ToolAgent[I, T, O]{
		start: NewAgent[I, T](options...),
		end:   NewAgent[I, O](options...),
	}
}

func (t *ToolAgent[I, T, O]) SetTool(tool tools.AnonymousTool) *ToolAgent[I, T, O] {
	t.tool = tool
	return t
}

func (t *ToolAgent[I, T, O]) ResetMemory() {
	t.start.ResetMemory()
	t.end.ResetMemory()
}

func (t *ToolAgent[I, T, O]) SetClient(clt instructor.Instructor) {
	t.start.client = clt
	t.end.client = clt
}

func (t *ToolAgent[I, T, O]) SetMemory(m *components.Memory) {
	t.start.memory = m
	t.end.memory = m
}

func (t *ToolAgent[I, T, O]) SetSystemPromptGenerator(g systemprompt.Generator) {
	t.start.systemPromptGenerator = g
	t.end.systemPromptGenerator = g
}

func (t *ToolAgent[I, T, O]) SetModel(model string) {
	t.start.model = model
	t.end.model = model
}

func (t *ToolAgent[I, T, O]) SetTemperature(temperature float32) {
	t.start.temperature = temperature
	t.end.temperature = temperature
}

func (t *ToolAgent[I, T, O]) SetMaxTokens(maxTokens int) {
	t.start.maxTokens = maxTokens
	t.end.maxTokens = maxTokens
}

func (t *ToolAgent[I, T, O]) Name() string {
	return t.name
}

func (t *ToolAgent[I, T, O]) SetName(name string) {
	t.name = name
}

func (t *ToolAgent[I, T, O]) SetStartHook(fn func(context.Context, *ToolAgent[I, T, O], *I)) {
	t.startHook = fn
}

func (t *ToolAgent[I, T, O]) SetEndHook(fn func(context.Context, *ToolAgent[I, T, O], *I, *O, *components.LLMResponse)) {
	t.endHook = fn
}

func (t *ToolAgent[I, T, O]) SetErrorHook(fn func(context.Context, *ToolAgent[I, T, O], *I, *components.LLMResponse, error)) {
	t.errorHook = fn
}

func (t *ToolAgent[I, T, O]) SetStartAgentName(name string) {
	t.start.SetName(name)
}

func (t *ToolAgent[I, T, O]) SetEndAgentName(name string) {
	t.end.SetName(name)
}

func (t *ToolAgent[I, T, O]) SetStartAgentStartHook(fn func(context.Context, *Agent[I, T], *I)) {
	t.start.SetStartHook(fn)
}

func (t *ToolAgent[I, T, O]) SetStartAgentEndHook(fn func(context.Context, *Agent[I, T], *I, *T, *components.LLMResponse)) {
	t.start.SetEndHook(fn)
}

func (t *ToolAgent[I, T, O]) SetStartAgentErrorHook(fn func(context.Context, *Agent[I, T], *I, *components.LLMResponse, error)) {
	t.start.SetErrorHook(fn)
}

func (t *ToolAgent[I, T, O]) SetEndAgentStartHook(fn func(context.Context, *Agent[I, O], *I)) {
	t.end.SetStartHook(fn)
}

func (t *ToolAgent[I, T, O]) SetEndAgentEndHook(fn func(context.Context, *Agent[I, O], *I, *O, *components.LLMResponse)) {
	t.end.SetEndHook(fn)
}

func (t *ToolAgent[I, T, O]) SetEndAgentErrorHook(fn func(context.Context, *Agent[I, O], *I, *components.LLMResponse, error)) {
	t.end.SetErrorHook(fn)
}

func (t *ToolAgent[I, T, O]) SetToolStartHook(fn func(context.Context, tools.AnonymousTool, any)) {
	t.tool.SetStartHook(fn)
}

func (t *ToolAgent[I, T, O]) SetToolEndHook(fn func(context.Context, tools.AnonymousTool, any, any)) {
	t.tool.SetEndHook(fn)
}

func (t *ToolAgent[I, T, O]) SetToolErrorHook(fn func(context.Context, tools.AnonymousTool, any, error)) {
	t.tool.SetErrorHook(fn)
}

// Run runs the chat agent with the given user input synchronously.
func (t *ToolAgent[I, T, O]) Run(ctx context.Context, userInput *I, output *O, apiResp *components.LLMResponse) error {
	toolOutput := new(T)
	if fn := t.startHook; fn != nil {
		fn(ctx, t, userInput)
	}
	if err := t.start.Run(ctx, userInput, toolOutput, apiResp); err != nil {
		if fn := t.errorHook; fn != nil {
			fn(ctx, t, userInput, apiResp, err)
		}
		return err
	}
	if t.tool != nil {
		if toolResult, err := t.tool.RunAnonymous(ctx, toolOutput); err != nil {
			if fn := t.errorHook; fn != nil {
				fn(ctx, t, userInput, apiResp, err)
			}
			return err
		} else if outO, ok := toolResult.(schema.Schema); !ok {
			err := errors.New("invalid agent output schema")
			if fn := t.errorHook; fn != nil {
				fn(ctx, t, userInput, apiResp, err)
			}
			return err
		} else {
			t.end.NewMessage(components.SystemRole, outO)
		}
	}
	if err := t.end.Run(ctx, userInput, output, apiResp); err != nil {
		if fn := t.errorHook; fn != nil {
			fn(ctx, t, userInput, apiResp, err)
		}
		return err
	}
	if fn := t.endHook; fn != nil {
		fn(ctx, t, userInput, output, apiResp)
	}
	return nil
}

// Run runs the chat agent with the given user input for chain.
func (t *ToolAgent[I, T, O]) RunForChain(ctx context.Context, userInput any, apiResp *components.LLMResponse) (any, error) {
	in, ok := userInput.(*I)
	if !ok {
		return nil, errors.New("invalid input schema")
	}
	out := new(O)
	if err := t.Run(ctx, in, out, apiResp); err != nil {
		return nil, err
	}
	return out, nil
}
