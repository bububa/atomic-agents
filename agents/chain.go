package agents

import (
	"context"
	"errors"

	"github.com/bububa/atomic-agents/components"
	"github.com/bububa/atomic-agents/schema"
)

// Chain agents chain
type Chain[I schema.Schema, O schema.Schema] struct {
	name      string
	agents    []AnonymousAgent
	startHook func(context.Context, *Chain[I, O], *I)
	endHook   func(context.Context, *Chain[I, O], *I, *O, []components.LLMResponse)
	errorHook func(context.Context, *Chain[I, O], *I, []components.LLMResponse, error)
}

// NewChain returns a new Chain instance
func NewChain[I schema.Schema, O schema.Schema](agents ...AnonymousAgent) *Chain[I, O] {
	return &Chain[I, O]{
		agents: agents,
	}
}

func (c *Chain[I, O]) Name() string {
	return c.name
}

func (c *Chain[I, O]) SetName(name string) {
	c.name = name
}

func (c *Chain[I, O]) SetStartHook(fn func(context.Context, *Chain[I, O], *I)) {
	c.startHook = fn
}

func (c *Chain[I, O]) SetEndHook(fn func(context.Context, *Chain[I, O], *I, *O, []components.LLMResponse)) {
	c.endHook = fn
}

func (c *Chain[I, O]) SetErrorHook(fn func(context.Context, *Chain[I, O], *I, []components.LLMResponse, error)) {
	c.errorHook = fn
}

// Run runs the chat agents with the given user input synchronously.
func (c *Chain[I, O]) Run(ctx context.Context, input *I, output *O) ([]components.LLMResponse, error) {
	if fn := c.startHook; fn != nil {
		fn(ctx, c, input)
	}
	l := len(c.agents)
	apiRespList := make([]components.LLMResponse, 0, l)
	var (
		in  any = input
		out any
	)
	for _, agent := range c.agents {
		apiResp := new(components.LLMResponse)
		if ret, err := agent.RunAnonymous(ctx, in, apiResp); err != nil {
			if fn := c.errorHook; fn != nil {
				fn(ctx, c, input, apiRespList, err)
			}
			return apiRespList, err
		} else {
			in = ret
			out = ret
		}
		apiRespList = append(apiRespList, *apiResp)
	}
	if outO, ok := out.(*O); !ok {
		err := errors.New("invalid agent output schema")
		if fn := c.errorHook; fn != nil {
			fn(ctx, c, input, apiRespList, err)
		}
		return apiRespList, err
	} else {
		*output = *outO
	}
	if fn := c.endHook; fn != nil {
		fn(ctx, c, input, output, apiRespList)
	}
	return apiRespList, nil
}

// Run runs the chat agents with the given user input synchronously.
func (c *Chain[I, O]) RunAnonymous(ctx context.Context, input any, apiResp *components.LLMResponse) (any, error) {
	in, ok := input.(*I)
	if !ok {
		return nil, errors.New("invalid agent input schema")
	}
	out := new(O)
	apiRespList, err := c.Run(ctx, in, out)
	if err != nil {
		return nil, err
	}
	for _, v := range apiRespList {
		if v.Usage == nil {
			continue
		}
		if apiResp.Usage == nil {
			apiResp.Usage = new(components.LLMUsage)
		}
		apiResp.Usage.InputTokens = v.Usage.InputTokens
		apiResp.Usage.OutputTokens = v.Usage.OutputTokens
	}
	return out, nil
}
