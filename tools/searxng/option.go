package searxng

import "net/http"

type Option func(*Config)

func WithBaseURL(baseURL string) Option {
	return func(c *Config) {
		c.baseURL = baseURL
	}
}

func WithLanguage(lang string) Option {
	return func(c *Config) {
		c.language = lang
	}
}

func WithMaxResults(n int) Option {
	return func(c *Config) {
		c.maxResults = n
	}
}

func WithHttpClient(clt *http.Client) Option {
	return func(c *Config) {
		c.httpClient = clt
	}
}
