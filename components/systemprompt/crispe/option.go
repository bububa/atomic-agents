package crispe

import "github.com/bububa/atomic-agents/components/systemprompt"

type Option = func(g *Generator)

// WithBackground set Generator background
func WithBackground(background []string) Option {
	return func(g *Generator) {
		g.background = background
	}
}

// WithCapacities set Generator capacities
func WithCapacities(capacities []string) Option {
	return func(g *Generator) {
		g.capacities = capacities
	}
}

// WithStatements set Generator statements
func WithStatements(statements []string) Option {
	return func(g *Generator) {
		g.statements = statements
	}
}

// WithPersonalities set Generator personalities
func WithPersonalities(personalities []string) Option {
	return func(g *Generator) {
		g.personalities = personalities
	}
}

// WithExperiments set Generator capacities
func WithExperiments(experiments []string) Option {
	return func(g *Generator) {
		g.experiments = experiments
	}
}

// WithContextProviders set Generator context pproviders
func WithContextProviders(providers ...systemprompt.ContextProvider) Option {
	return func(g *Generator) {
		g.AddContextProviders(providers...)
	}
}
