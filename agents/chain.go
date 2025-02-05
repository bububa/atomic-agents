package agents

import "github.com/bububa/atomic-agents/schema"

type IO[I schema.Schema, M schema.Schema, O schema.Schema] struct {
	Input  *Agent[I, M]
	Output *Agent[M, O]
}

func NewIO[I schema.Schema, M schema.Schema, O schema.Schema](in *Agent[I, M], out *Agent[M, O]) *IO[I, M, O] {
	return &IO[I, M, O]{Input: in, Output: out}
}
