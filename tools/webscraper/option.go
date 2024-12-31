package webscraper

import (
	"net/http"
)

const (
	DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	DefaultAccept    = "text/html,application/xhtml+xml,application/xml;"
)

type Option func(*Config)

func WithUserAgent(ua string) Option {
	return func(c *Config) {
		c.userAgent = ua
	}
}

func WithTimeout(timeout int) Option {
	return func(c *Config) {
		c.timeout = timeout
	}
}

func WithMaxContentLength(l int64) Option {
	return func(c *Config) {
		c.maxContentLength = l
	}
}

func WithHttpClient(clt *http.Client) Option {
	return func(c *Config) {
		c.httpClient = clt
	}
}
