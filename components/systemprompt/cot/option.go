package cot

import "github.com/bububa/atomic-agents/components/systemprompt"

type Option = func(g *Generator)

// WithBackground set Generator background
func WithBackground(background []string) Option {
	return func(g *Generator) {
		g.background = background
	}
}

// WithSteps set Generator steps
func WithSteps(steps []string) Option {
	return func(g *Generator) {
		g.steps = steps
	}
}

// WithOutputInstructs set Generator output instructions
func WithOutputInstructs(outputInstructs []string) Option {
	return func(g *Generator) {
		g.outputInstructs = outputInstructs
	}
}

// WithContextProviders set Generator context pproviders
func WithContextProviders(providers ...systemprompt.ContextProvider) Option {
	return func(g *Generator) {
		g.AddContextProviders(providers...)
	}
}
