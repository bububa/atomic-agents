package systemprompt

import "fmt"

// Generator is system prompt generator framework
type Generator interface {
	Generate() string
	// ContextProvider retrieves a context provider by name.
	// If the context provider is not found returns not found error
	ContextProvider(title string) (ContextProvider, error)
	// AddContextProviders registers new context providers
	AddContextProviders(providers ...ContextProvider)
	// RemoveContextProviders Unregisters an existing context provider.
	RemoveContextProviders(titles ...string)
}

type BaseGenerator struct {
	contextProviders []ContextProvider
}

func (g *BaseGenerator) ContextProviders() []ContextProvider {
	return g.contextProviders
}

// ContextProvider retrieves a context provider by name.
// If the context provider is not found returns not found error
func (g *BaseGenerator) ContextProvider(title string) (ContextProvider, error) {
	for _, p := range g.contextProviders {
		if p.Title() == title {
			return p, nil
		}
	}
	return nil, fmt.Errorf("context provider '%s' not found", title)
}

// AddContextProviders registers new context providers
func (g *BaseGenerator) AddContextProviders(providers ...ContextProvider) {
	for _, provider := range providers {
		if _, err := g.ContextProvider(provider.Title()); err != nil {
			g.contextProviders = append(g.contextProviders, provider)
		}
	}
}

func (g *BaseGenerator) addContextProvider(provider ContextProvider) {
	if _, err := g.ContextProvider(provider.Title()); err != nil {
		g.contextProviders = append(g.contextProviders, provider)
	}
}

// RemoveContextProviders Unregisters an existing context provider.
func (g *BaseGenerator) RemoveContextProviders(titles ...string) {
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

func (g *BaseGenerator) removeContextProvider(title string) {
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
