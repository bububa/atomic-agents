package calculator

import (
	"context"
	"errors"

	"github.com/Knetic/govaluate"

	"github.com/bububa/atomic-agents/schema"
	"github.com/bububa/atomic-agents/tools"
	"github.com/bububa/atomic-agents/tools/calculator/functions"
)

// Input Tool for performing calculations. Supports basic arithmetic operations
// like addition, subtraction, multiplication, and division, as well as more
// complex operations like exponentiation and trigonometric functions.
// Use this tool to evaluate mathematical expressions.
type Input struct {
	schema.Base
	// Expression Mathematical expression to evaluate. For example, '2 + 2'.
	Expression string `json:"expression" jsonschema:"title=expression,description=Mathematical expression to evaluate. For example, '2 + 2'."`
	// Params represents expressions's parameters
	Params map[string]interface{} `json:"params,omitempty" jsonschema:"title=params,description=Parameters for the expression."`
}

func NewInput(exp string, params map[string]interface{}) *Input {
	return &Input{
		Expression: exp,
		Params:     params,
	}
}

// Output Schema for the output of the CalculatorTool
type Output struct {
	schema.Base
	// Result Result of the calculation
	Result interface{} `json:"result,omitempty" jsonschema:"title=result,description=Result of the calculation."`
}

func NewOutput(result interface{}) *Output {
	return &Output{
		Result: result,
	}
}

type Tool struct {
	tools.Config
}

func New(opts ...tools.Option) *Tool {
	ret := new(Tool)
	for _, opt := range opts {
		opt(&ret.Config)
	}
	if ret.Title() == "" {
		ret.SetTitle("CalculatorTool")
	}
	return ret
}

// Executes the CalculatorTool with the given parameters.
func (t *Tool) Run(ctx context.Context, input *Input, output *Output) error {
	exp, err := govaluate.NewEvaluableExpressionWithFunctions(input.Expression, functions.Functions)
	if err != nil {
		return err
	}
	params := make(map[string]interface{}, len(input.Params)+len(constParams))
	for k, v := range input.Params {
		params[k] = v
	}
	for k, v := range constParams {
		if _, ok := params[k]; ok {
			continue
		}
		params[k] = v
	}
	result, err := exp.Evaluate(params)
	if err != nil {
		return err
	}
	*output = *NewOutput(result)
	return nil
}

// RunAnonymous run tool for tools ochestration
func (t *Tool) RunAnonymous(ctx context.Context, input any) (any, error) {
	if fn := t.StartHook(); fn != nil {
		fn(ctx, t, input)
	}
	in, ok := input.(*Input)
	if !ok {
		err := errors.New("invalid tool input schema")
		if fn := t.ErrorHook(); fn != nil {
			fn(ctx, t, input, err)
		}
		return nil, err
	}
	out := new(Output)
	if err := t.Run(ctx, in, out); err != nil {
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
