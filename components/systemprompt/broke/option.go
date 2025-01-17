package broke

import "github.com/bububa/atomic-agents/components/systemprompt"

type Option = func(g *Generator)

// WithBackground set Generator background
func WithBackground(background []string) Option {
	return func(g *Generator) {
		g.background = background
	}
}

// WithRoles set Generator roles
func WithRoles(roles []string) Option {
	return func(g *Generator) {
		g.roles = roles
	}
}

// WithObjectives set Generator objectives
func WithObjectives(objectives []string) Option {
	return func(g *Generator) {
		g.objectives = objectives
	}
}

// WithKeyResults set Generator keyResults
func WithKeyResults(keyResults []string) Option {
	return func(g *Generator) {
		g.keyResults = keyResults
	}
}

// WithEvolves set Generator evolves
func WithEvolves(evolves []string) Option {
	return func(g *Generator) {
		g.evolves = evolves
	}
}

// WithContextProviders set Generator context pproviders
func WithContextProviders(providers ...systemprompt.ContextProvider) Option {
	return func(g *Generator) {
		g.AddContextProviders(providers...)
	}
}
