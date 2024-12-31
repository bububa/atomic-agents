package tools

import (
	"context"

	"github.com/bububa/atomic-agents/schema"
)

type Tool[I schema.Schema, O schema.Schema] interface {
	SetTitle(string)
	Title() string
	SetDescription(string)
	Description() string
	Run(context.Context, *I) (*O, error)
}
