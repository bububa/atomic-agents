package simple

import (
	"fmt"
	"strings"

	"github.com/bububa/atomic-agents/components/systemprompt"
)

// Generator is CRISPE system prompt generator
type Generator struct {
	systemprompt.BaseGenerator
	content string
}

var _ systemprompt.Generator = (*Generator)(nil)

// New returns a new system prompt Generator
func New(content string, options ...Option) *Generator {
	ret := new(Generator)
	for _, opt := range options {
		opt(ret)
	}
	ret.content = content
	return ret
}

func (g *Generator) Generate() string {
	promptParts := make([]string, 0, len(g.ContextProviders())*3+1)
	promptParts = append(promptParts, g.content)
	promptParts = append(promptParts, "")
	if providers := g.ContextProviders(); len(providers) > 0 {
		promptParts = append(promptParts, "# EXTRA INFORMATION AND CONTEXT")
		for _, provider := range providers {
			if info := provider.Info(); info != "" {
				promptParts = append(promptParts, fmt.Sprintf("## %s", provider.Title()))
				promptParts = append(promptParts, provider.Info())
				promptParts = append(promptParts, "")
			}
		}
	}
	return strings.TrimSpace(strings.Join(promptParts, "\n"))
}
