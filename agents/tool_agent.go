package agents

import (
	"context"
	"errors"

	"github.com/bububa/atomic-agents/components"
	"github.com/bububa/atomic-agents/schema"
	"github.com/bububa/atomic-agents/tools"
)

// ToolAgent represent agent with tool callback
type ToolAgent[I schema.Schema, T schema.Schema, O schema.Schema] struct {
	start *Agent[I, T]
	end   *Agent[I, O]
	tool  tools.OrchestrationTool
}

// NewToolAgent returns a new ToolAgent instance
func NewToolAgent[I schema.Schema, T schema.Schema, O schema.Schema](options ...Option) *ToolAgent[I, T, O] {
	return &ToolAgent[I, T, O]{
		start: NewAgent[I, T](options...),
		end:   NewAgent[I, O](options...),
	}
}

func (t *ToolAgent[I, T, O]) SetTool(tool tools.OrchestrationTool) *ToolAgent[I, T, O] {
	t.tool = tool
	return t
}

func (t *ToolAgent[I, T, O]) ResetMemory() {
	t.start.ResetMemory()
	t.end.ResetMemory()
}

// Run runs the chat agent with the given user input synchronously.
func (t *ToolAgent[I, T, O]) Run(ctx context.Context, userInput *I, output *O, apiResp *components.ApiResponse) error {
	toolOutput := new(T)
	if err := t.start.Run(ctx, userInput, toolOutput, apiResp); err != nil {
		return err
	}
	if t.tool != nil {
		if toolResult, err := t.tool.RunOrchestration(ctx, toolOutput); err != nil {
			return err
		} else if outO, ok := toolResult.(schema.Schema); !ok {
			return errors.New("invalid agent output schema")
		} else {
			t.end.NewMessage(components.SystemRole, outO)
		}
	}
	if err := t.end.Run(ctx, userInput, output, apiResp); err != nil {
		return err
	}
	return nil
}

// Run runs the chat agent with the given user input for chain.
func (t *ToolAgent[I, T, O]) RunForChain(ctx context.Context, userInput any, apiResp *components.ApiResponse) (any, error) {
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
