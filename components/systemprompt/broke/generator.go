package broke

import (
	"fmt"
	"strings"

	"github.com/bububa/atomic-agents/components/systemprompt"
)

// Generator is BROKE prompt generator
type Generator struct {
	systemprompt.BaseGenerator
	// background agent role background
	background []string
	// roles Capacity and Role
	roles []string
	// objectives represents the task of the agent
	objectives []string
	// keyResults represents the key results of the answer
	keyResults []string
	// evolves sugested optimizations for response
	evolves []string
}

// New returns a new system prompt Generator
func New(options ...Option) *Generator {
	ret := new(Generator)
	for _, opt := range options {
		opt(ret)
	}
	if len(ret.background) == 0 {
		ret.background = []string{"- This is a conversation with a helpful and friendly AI assistant."}
	}
	ret.evolves = append(ret.evolves, "- Always respond using the proper JSON schema.", "- Always use the available additional information and context to enhance the response.")
	return ret
}

func (g *Generator) Generate() string {
	var (
		sections = map[string][]string{
			"BACKGROUND and PURPOSE":               g.background,
			"CAPACITY and ROLE":                    g.roles,
			"OBJECTIVEs and TASKs":                 g.objectives,
			"KEY RESULTS":                          g.keyResults,
			"OPTIMIZATION and OUTPUT INSTRUCTIONS": g.evolves,
		}
		promptParts []string
	)
	for _, title := range []string{"BACKGROUND and PURPOSE", "CAPACITY and ROLE", "OBJECTIVEs and TASKs", "KEY RESULTS", "OPTIMIZATION and OUTPUT INSTRUCTIONS"} {
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
