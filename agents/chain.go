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
	agents    []ChainableAgent
	startHook func(context.Context, *Chain[I, O], *I)
	endHook   func(context.Context, *Chain[I, O], *I, *O, []components.ApiResponse)
	errorHook func(context.Context, *Chain[I, O], *I, []components.ApiResponse, error)
}

// NewChain returns a new Chain instance
func NewChain[I schema.Schema, O schema.Schema](agents ...ChainableAgent) *Chain[I, O] {
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

func (c *Chain[I, O]) SetEndHook(fn func(context.Context, *Chain[I, O], *I, *O, []components.ApiResponse)) {
	c.endHook = fn
}

func (c *Chain[I, O]) SetErrorHook(fn func(context.Context, *Chain[I, O], *I, []components.ApiResponse, error)) {
	c.errorHook = fn
}

// Run runs the chat agents with the given user input synchronously.
func (c *Chain[I, O]) Run(ctx context.Context, input *I, output *O) ([]components.ApiResponse, error) {
	if fn := c.startHook; fn != nil {
		fn(ctx, c, input)
	}
	l := len(c.agents)
	apiRespList := make([]components.ApiResponse, 0, l)
	var (
		in  any = input
		out any
	)
	for _, agent := range c.agents {
		apiResp := new(components.ApiResponse)
		if ret, err := agent.RunForChain(ctx, in, apiResp); err != nil {
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
func (c *Chain[I, O]) RunForChain(ctx context.Context, input any, apiResp *components.ApiResponse) (any, error) {
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
			apiResp.Usage = new(components.ApiUsage)
		}
		apiResp.Usage.InputTokens = v.Usage.InputTokens
		apiResp.Usage.OutputTokens = v.Usage.OutputTokens
	}
	return out, nil
}
