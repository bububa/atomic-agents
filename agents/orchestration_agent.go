package agents

import (
	"context"
	"errors"

	"github.com/bububa/atomic-agents/components"
	"github.com/bububa/atomic-agents/schema"
)

// AgentSelector will returns a Tool based on input param
type AgentSelector[I schema.Schema] func(req *I) (AnonymousAgent, any, error)

// OrchestrationAgent is an agent for orchestration
type OrchestrationAgent[I schema.Schema, O schema.Schema] struct {
	name     string
	selector AgentSelector[I]
}

func NewOrchestrationAgent[I schema.Schema, O schema.Schema](selector AgentSelector[I]) *OrchestrationAgent[I, O] {
	return &OrchestrationAgent[I, O]{
		selector: selector,
	}
}

func (a *OrchestrationAgent[I, O]) Name() string {
	return a.name
}

func (a *OrchestrationAgent[I, O]) SetName(name string) {
	a.name = name
}

func (a *OrchestrationAgent[I, O]) Run(ctx context.Context, input *I, output *O, apiResp *components.LLMResponse) error {
	fn, params, err := a.selector(input)
	if err != nil {
		return err
	}
	if out, err := fn.RunAnonymous(ctx, params, apiResp); err != nil {
		return err
	} else if outO, ok := out.(*O); !ok {
		return errors.New("invalid agent output schema")
	} else {
		*output = *outO
	}
	return nil
}

func (a *OrchestrationAgent[I, O]) RunAnonymous(ctx context.Context, input any, apiResp *components.LLMResponse) (any, error) {
	in, ok := input.(*I)
	if !ok {
		return nil, errors.New("invalid agent input schema")
	}
	fn, params, err := a.selector(in)
	if err != nil {
		return nil, err
	}
	return fn.RunAnonymous(ctx, params, apiResp)
}
