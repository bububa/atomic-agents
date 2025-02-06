package agents

import (
	"context"
	"errors"

	"github.com/bububa/atomic-agents/components"
	"github.com/bububa/atomic-agents/schema"
)

// Chain agents chain
type Chain[I schema.Schema, O schema.Schema] struct {
	agents []ChainableAgent
}

// NewChain returns a new Chain instance
func NewChain[I schema.Schema, O schema.Schema](agents ...ChainableAgent) *Chain[I, O] {
	return &Chain[I, O]{
		agents: agents,
	}
}

// Run runs the chat agents with the given user input synchronously.
func (c *Chain[I, O]) Run(ctx context.Context, input *I, output *O) ([]components.ApiResponse, error) {
	l := len(c.agents)
	apiRespList := make([]components.ApiResponse, 0, l)
	var (
		in  any = input
		out any
	)
	for _, agent := range c.agents {
		apiResp := new(components.ApiResponse)
		if ret, err := agent.RunForChain(ctx, in, apiResp); err != nil {
			return apiRespList, err
		} else {
			in = ret
			out = ret
		}
		apiRespList = append(apiRespList, *apiResp)
	}
	if outO, ok := out.(*O); !ok {
		return apiRespList, errors.New("invalid output schema")
	} else {
		*output = *outO
	}
	return apiRespList, nil
}

// Run runs the chat agents with the given user input synchronously.
func (c *Chain[I, O]) RunForChain(ctx context.Context, input any, apiResp *components.ApiResponse) (any, error) {
	in, ok := input.(*I)
	if !ok {
		return nil, errors.New("invalid input schema")
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
