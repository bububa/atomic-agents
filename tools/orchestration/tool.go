package orchestration

import (
	"context"
	"errors"

	"github.com/bububa/atomic-agents/schema"
	"github.com/bububa/atomic-agents/tools"
)

// ToolSelector will returns a Tool based on input param
type ToolSelector[I schema.Schema] func(req *I) (tools.OrchestrationTool, any, error)

// Tool is orchestration tool for tools selector
type Tool[I schema.Schema] struct {
	tools.Config
	selector ToolSelector[I]
}

func New[I schema.Schema](selector ToolSelector[I], opts ...tools.Option) *Tool[I] {
	ret := new(Tool[I])
	for _, opt := range opts {
		opt(&ret.Config)
	}
	if ret.Title() == "" {
		ret.SetTitle("OrchestrationTool")
	}
	ret.selector = selector
	return ret
}

// RunOrchestration returns a tool results based on input for orchestration
func (t *Tool[I]) RunOrchestration(ctx context.Context, input any) (any, error) {
	in, ok := input.(*I)
	if !ok {
		return nil, errors.New("invalid tool input schema")
	}
	tool, params, err := t.selector(in)
	if err != nil {
		return nil, err
	}
	return tool.RunOrchestration(ctx, params)
}
