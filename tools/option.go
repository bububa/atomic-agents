package tools

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
