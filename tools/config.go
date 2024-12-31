package tools

// Config class for tools within the Atomic Agents framework
type Config struct {
	// title the default title of the tool
	title string
	// description the default description of the tool
	description string
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
