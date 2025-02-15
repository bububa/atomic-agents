package tools

import "context"

type Option func(c *Config)

func WithTitle(title string) Option {
	return func(c *Config) {
		c.SetTitle(title)
	}
}

func WithDescription(desc string) Option {
	return func(c *Config) {
		c.SetDescription(desc)
	}
}

func WithStartHook(fn func(context.Context, AnonymousTool, any)) Option {
	return func(c *Config) {
		c.SetStartHook(fn)
	}
}

func WithEndHook(fn func(context.Context, AnonymousTool, any, any)) Option {
	return func(c *Config) {
		c.SetEndHook(fn)
	}
}

func WithErrorHook(fn func(context.Context, AnonymousTool, any, error)) Option {
	return func(c *Config) {
		c.SetErrorHook(fn)
	}
}
