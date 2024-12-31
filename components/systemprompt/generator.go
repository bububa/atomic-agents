package systemprompt

import (
	"fmt"
	"strings"
)

// Generator is system prompt generator
type Generator struct {
	background       []string
	steps            []string
	outputInstructs  []string
	contextProviders []ContextProvider
}

// NewGenerator returns a new system prompt Generator
func NewGenerator(options ...Option) *Generator {
	ret := new(Generator)
	for _, opt := range options {
		opt(ret)
	}
	if len(ret.background) == 0 {
		ret.background = []string{"This is a conversation with a helpful and friendly AI assistant."}
	}
	ret.outputInstructs = append(ret.outputInstructs, "Always respond using the proper JSON schema.", "Always use the available additional information and context to enhance the response.")
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
	for title, content := range sections {
		if len(content) > 0 {
			promptParts = append(promptParts, fmt.Sprintf("# %s", title))
			for _, item := range content {
				promptParts = append(promptParts, fmt.Sprintf("- %s", item))
			}
			promptParts = append(promptParts, "")
		}
	}
	if len(g.contextProviders) > 0 {
		promptParts = append(promptParts, "# EXTRA INFORMATION AND CONTEXT")
		for _, provider := range g.contextProviders {
			if info := provider.Info(); info != "" {
				promptParts = append(promptParts, fmt.Sprintf("## %s", provider.Title()))
				promptParts = append(promptParts, provider.Info())
				promptParts = append(promptParts, "")
			}
		}
	}
	return strings.TrimSpace(strings.Join(promptParts, "\n"))
}

// ContextProvider retrieves a context provider by name.
// If the context provider is not found returns not found error
func (g *Generator) ContextProvider(title string) (ContextProvider, error) {
	for _, p := range g.contextProviders {
		if p.Title() == title {
			return p, nil
		}
	}
	return nil, fmt.Errorf("context provider '%s' not found", title)
}

// AddContextProviders registers new context providers
func (g *Generator) AddContextProviders(providers ...ContextProvider) {
	for _, provider := range providers {
		if _, err := g.ContextProvider(provider.Title()); err != nil {
			g.contextProviders = append(g.contextProviders, provider)
		}
	}
}

func (g *Generator) addContextProvider(provider ContextProvider) {
	if _, err := g.ContextProvider(provider.Title()); err != nil {
		g.contextProviders = append(g.contextProviders, provider)
	}
}

// RemoveContextProviders Unregisters an existing context provider.
func (g *Generator) RemoveContextProviders(titles ...string) {
	l := len(titles)
	if l == 1 {
		g.removeContextProvider(titles[0])
	}
	mp := make(map[string]struct{}, l)
	for _, v := range titles {
		mp[v] = struct{}{}
	}
	providers := make([]ContextProvider, 0, l)
	for _, p := range g.contextProviders {
		if _, found := mp[p.Title()]; found {
			continue
		}
		providers = append(providers, p)
	}
	g.contextProviders = providers
}

func (g *Generator) removeContextProvider(title string) {
	found := -1
	for idx, p := range g.contextProviders {
		if p.Title() == title {
			found = idx
			break
		}
	}
	if found >= 0 {
		g.contextProviders = append(g.contextProviders[:found], g.contextProviders[found:]...)
	}
}
