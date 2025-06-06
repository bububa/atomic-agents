package agents

import (
	"github.com/bububa/instructor-go"

	"github.com/bububa/atomic-agents/components/systemprompt"
)

type Option func(a *Config)

func WithClient(clt instructor.Instructor) Option {
	return func(c *Config) {
		c.client = clt
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

func WithTemperature(temperature float64) Option {
	return func(c *Config) {
		c.temperature = temperature
	}
}

func WithTopP(topP float64) Option {
	return func(c *Config) {
		c.topP = topP
	}
}

func WithTopK(topK int) Option {
	return func(c *Config) {
		c.topK = topK
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
