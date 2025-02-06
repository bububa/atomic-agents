package orchestration

import (
	"context"
	"errors"

	"github.com/bububa/atomic-agents/schema"
	"github.com/bububa/atomic-agents/tools"
)

// ToolSelector will returns a Tool based on input param
type ToolSelector[I schema.Schema] func(req *I) (tools.AnonymousTool, any, error)

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

// RunAnonymous returns a tool results based on input for orchestration
func (t *Tool[I]) RunAnonymous(ctx context.Context, input any) (any, error) {
	if fn := t.StartHook(); fn != nil {
		fn(ctx, t, input)
	}
	in, ok := input.(*I)
	if !ok {
    err := errors.New("invalid tool input schema")
		if fn := t.ErrorHook(); fn != nil {
			fn(ctx, t, input, err)
		}
		return nil, err
	}
	tool, params, err := t.selector(in)
	if err != nil {
		if fn := t.ErrorHook(); fn != nil {
			fn(ctx, t, input, err)
		}
		return nil, err
	}
  out, err := tool.RunAnonymous(ctx, params)
  if err != nil {
		if fn := t.ErrorHook(); fn != nil {
			fn(ctx, t, input, err)
		}
    return nil, err
  }
	if fn := t.EndHook(); fn != nil {
		fn(ctx, t, input, out)
	}
  return out, nil
}
