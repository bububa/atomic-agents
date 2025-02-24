package cot

import (
	"fmt"
	"strings"

	"github.com/bububa/atomic-agents/components/systemprompt"
)

// Generator is Chain-of-Thought system prompt generator
type Generator struct {
	systemprompt.BaseGenerator
	background      []string
	steps           []string
	outputInstructs []string
}

var _ systemprompt.Generator = (*Generator)(nil)

// New returns a new system prompt Generator
func New(options ...Option) *Generator {
	ret := new(Generator)
	for _, opt := range options {
		opt(ret)
	}
	if len(ret.background) == 0 {
		ret.background = []string{"- This is a conversation with a helpful and friendly AI assistant."}
	}
	ret.outputInstructs = append(ret.outputInstructs, "- Always respond using the proper JSON schema.", "- Always use the available additional information and context to enhance the response.")
	return ret
}

func (g *Generator) Generate() string {
	var (
		sections = map[string][]string{
			"IDENTITY and PURPOSE":     g.background,
			"INTERNAL ASSISTANT STEPS": g.steps,
			"OUTPUT INSTRUCTIONS":      g.outputInstructs,
		}
		promptParts []string
	)
	for _, title := range []string{"IDENTITY and PURPOSE", "INTERNAL ASSISTANT STEPS", "OUTPUT INSTRUCTIONS"} {
		content := sections[title]
		if len(content) > 0 {
			promptParts = append(promptParts, fmt.Sprintf("# %s", title))
			promptParts = append(promptParts, content...)
			promptParts = append(promptParts, "")
		}
	}
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
