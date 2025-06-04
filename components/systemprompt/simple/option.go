package simple

import "github.com/bububa/atomic-agents/components/systemprompt"

type Option = func(g *Generator)

// WithContextProviders set Generator context pproviders
func WithContextProviders(providers ...systemprompt.ContextProvider) Option {
	return func(g *Generator) {
		g.AddContextProviders(providers...)
	}
}
