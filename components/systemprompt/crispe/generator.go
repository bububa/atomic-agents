package crispe

import (
	"fmt"
	"strings"

	"github.com/bububa/atomic-agents/components/systemprompt"
)

// Generator is CRISPE system prompt generator
type Generator struct {
	systemprompt.BaseGenerator
	// capacities Capacity and Role
	capacities []string
	// background agent role background
	background []string
	// statements represents the task of the agent
	statements []string
	// personalities represents the response style
	personalities []string
	// experiments represents the suggested questions by ai for user to choose for better response if needed
	experiments []string
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
	if len(ret.ContextProviders()) > 0 {
		ret.personalities = append(ret.personalities, "- Always use the available additional information and context to enhance the response.")
	}
	return ret
}

func (g *Generator) Generate() string {
	var (
		sections = map[string][]string{
			"CAPACITY and ROLE":                   g.capacities,
			"INSIGHT and PURPOSE":                 g.background,
			"STATEMENT and TASK":                  g.statements,
			"PERSONALITY and OUTPUT INSTRUCTIONS": g.personalities,
			"INSTRUCTIONS for FOLLOWUP QUESTIONS": g.experiments,
		}
		promptParts []string
	)
	for _, title := range []string{"CAPACITY and ROLE", "INSIGHT and PURPOSE", "STATEMENT and TASK", "PERSONALITY and OUTPUT INSTRUCTIONS", "INSTRUCTIONS for FOLLOWUP QUESTIONS"} {
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
