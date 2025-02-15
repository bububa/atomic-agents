package tools

import (
	"context"

	"github.com/bububa/atomic-agents/schema"
)

type ITool interface {
	SetTitle(string)
	Title() string
	SetDescription(string)
	Description() string
	SetStartHook(fn func(context.Context, AnonymousTool, any))
	SetEndHook(fn func(context.Context, AnonymousTool, any, any))
	SetErrorHook(fn func(context.Context, AnonymousTool, any, error))
}

type Tool[I schema.Schema, O schema.Schema] interface {
	ITool
	Run(context.Context, *I, *O) error
}

type AnonymousTool interface {
	ITool
	RunAnonymous(context.Context, any) (any, error)
}
