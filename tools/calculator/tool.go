package calculator

import (
	"context"
	"encoding/json"

	"github.com/Knetic/govaluate"

	"github.com/bububa/atomic-agents/schema"
	"github.com/bububa/atomic-agents/tools"
)

// Input Tool for performing calculations. Supports basic arithmetic operations
// like addition, subtraction, multiplication, and division, as well as more
// complex operations like exponentiation and trigonometric functions.
// Use this tool to evaluate mathematical expressions.
type Input struct {
	schema.Base
	// Expression Mathematical expression to evaluate. For example, '2 + 2'.
	Expression string `json:"expression,omitempty" jsonschema:"title=expression,description=Mathematical expression to evaluate. For example, '2 + 2'."`
	// Params represents expressions's parameters
	Params map[string]interface{} `json:"params,omitempty" jsonschema:"title=params,description=Parameters for the expression."`
}

func NewInput(exp string, params map[string]interface{}) *Input {
	return &Input{
		Expression: exp,
		Params:     params,
	}
}

func (s Input) String() string {
	bs, _ := json.Marshal(s)
	return string(bs)
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

func (s Output) String() string {
	bs, _ := json.Marshal(s)
	return string(bs)
}

type Calculator struct {
	tools.Config
}

func New(opts ...tools.Option) *Calculator {
	ret := new(Calculator)
	for _, opt := range opts {
		opt(&ret.Config)
	}
	if ret.Title() == "" {
		ret.SetTitle("CalculatorTool")
	}
	return ret
}

// Executes the CalculatorTool with the given parameters.
func (t *Calculator) Run(ctx context.Context, input *Input) (*Output, error) {
	exp, err := govaluate.NewEvaluableExpression(input.Expression)
	if err != nil {
		return nil, err
	}
	result, err := exp.Evaluate(input.Params)
	if err != nil {
		return nil, err
	}
	return NewOutput(result), nil
}
