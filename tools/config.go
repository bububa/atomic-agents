package tools

import "context"

// Config class for tools within the Atomic Agents framework
type Config struct {
	// title the default title of the tool
	title string
	// description the default description of the tool
	description string
	// hooks
	startHook func(context.Context, AnonymousTool, any)
	endHook   func(context.Context, AnonymousTool, any, any)
	errorHook func(context.Context, AnonymousTool, any, error)
}

func (c *Config) SetTitle(v string) {
	c.title = v
}

func (c Config) Title() string {
	return c.title
}

func (c *Config) SetDescription(v string) {
	c.description = v
}

func (c Config) Description() string {
	return c.description
}

func (c *Config) StartHook() func(context.Context, AnonymousTool, any) {
	return c.startHook
}

func (c *Config) SetStartHook(fn func(context.Context, AnonymousTool, any)) {
	c.startHook = fn
}

func (c *Config) EndHook() func(context.Context, AnonymousTool, any, any) {
	return c.endHook
}

func (c *Config) SetEndHook(fn func(context.Context, AnonymousTool, any, any)) {
	c.endHook = fn
}

func (c *Config) ErrorHook() func(context.Context, AnonymousTool, any, error) {
	return c.errorHook
}

func (c *Config) SetErrorHook(fn func(context.Context, AnonymousTool, any, error)) {
	c.errorHook = fn
}
