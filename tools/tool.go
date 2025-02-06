package tools

import (
	"context"

	"github.com/bububa/atomic-agents/schema"
)

type BaseTool interface {
	SetTitle(string)
	Title() string
	SetDescription(string)
	Description() string
}

type Tool[I schema.Schema, O schema.Schema] interface {
	BaseTool
	Run(context.Context, *I, *O) error
}

type OrchestrationTool interface {
	BaseTool
	RunOrchestration(context.Context, any) (any, error)
}
