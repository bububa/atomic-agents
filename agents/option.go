package agents

import (
	"github.com/bububa/instructor-go"

	"github.com/bububa/atomic-agents/components"
	"github.com/bububa/atomic-agents/components/systemprompt"
)

type Option func(a *Config)

func WithClient(clt instructor.Instructor) Option {
	return func(c *Config) {
		c.client = clt
	}
}

func WithMemory(m components.MemoryStore) Option {
	return func(a *Config) {
		a.memory = m
	}
}

func WithSystemPromptGenerator(g systemprompt.Generator) Option {
	return func(c *Config) {
		c.systemPromptGenerator = g
	}
}

func WithModel(model string) Option {
	return func(c *Config) {
		c.model = model
	}
}

func WithTemperature(temperature float32) Option {
	return func(c *Config) {
		c.temperature = temperature
	}
}

func WithMaxTokens(maxTokens int) Option {
	return func(c *Config) {
		c.maxTokens = maxTokens
	}
}

func WithName(name string) Option {
	return func(c *Config) {
		c.name = name
	}
}
