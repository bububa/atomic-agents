package functions

import (
	"github.com/Knetic/govaluate"
)

var FlowFunctions map[string]govaluate.ExpressionFunction = map[string]govaluate.ExpressionFunction{
	"isDefined": IsDefined,
	"ifnull":    IfNull,
	"if":        If,
}

var MathFunctions map[string]govaluate.ExpressionFunction = map[string]govaluate.ExpressionFunction{
	"cos":   Cos,
	"acos":  Acos,
	"cosh":  Cosh,
	"acosh": Acosh,

	"exp":  Exp,
	"exp2": Exp2,

	"log":   Log,
	"log10": Log10,

	"round": Round,
	"floor": Floor,
	"ceil":  Ceil,
	"trunc": Trunc,

	"sin":   Sin,
	"asin":  Asin,
	"sinh":  Sinh,
	"asinh": Asinh,

	"sqrt": Sqrt,

	"tan":   Tan,
	"atan":  Atan,
	"atan2": Atan2,
	"tanh":  Tanh,
	"atanh": Atanh,

	"max": Max,
	"min": Min,
}

var Functions map[string]govaluate.ExpressionFunction

func init() {
	Functions = make(map[string]govaluate.ExpressionFunction)
	for k, v := range MathFunctions {
		Functions[k] = v
	}
	for k, v := range FlowFunctions {
		Functions[k] = v
	}
}
