package huggingface

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const BaseURL = "https://api-inference.huggingface.co/pipeline/feature-extraction/"

type options struct {
	WaitForModel *bool `json:"wait_for_model,omitempty"`
}

type EmbeddingRequest struct {
	Inputs  []string `json:"inputs,omitempty"`
	Options options  `json:"options,omitempty"`
	Model   string   `json:"-"`
}

// Client is Voyage HTTP API client.
type Client struct {
	opts Options
}

// Options are client options
type Options struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new HTTP API client and returns it.
// By default it reads the Voyage API key from HUGGING_FACE_API_KEY
// env var and uses the default Go http.Client for making API requests.
// You can override the default options via the client methods.
func NewClient(opts ...Option) *Client {
	options := Options{
		APIKey:     os.Getenv("HUGGING_FACE_API_KEY"),
		BaseURL:    BaseURL,
		HTTPClient: http.DefaultClient,
	}

	for _, apply := range opts {
		apply(&options)
	}

	return &Client{
		opts: options,
	}
}

// WithAPIKey sets the API key.
func WithAPIKey(apiKey string) Option {
	return func(o *Options) {
		o.APIKey = apiKey
	}
}

// WithBaseURL sets the API base URL.
func WithBaseURL(baseURL string) Option {
	return func(o *Options) {
		o.BaseURL = baseURL
	}
}

// WithHTTPClient sets the HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(o *Options) {
		o.HTTPClient = httpClient
	}
}

// Option is functional option.
type Option func(*Options)

func (c *Client) CreateEmbeddings(ctx context.Context, req *EmbeddingRequest) ([][]float64, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(req); err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.opts.BaseURL+req.Model, buf)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.opts.APIKey))

	resp, err := c.opts.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	apiErr := new(APIError)
	if err := json.Unmarshal(respBody, &apiErr); err == nil && apiErr.IsError() {
		return nil, apiErr
	}
	var ret [][]float64
	if err := json.Unmarshal(respBody, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

type APIError struct {
	Errors StringList `json:"error,omitempty"`
}

func (e APIError) IsError() bool {
	return len(e.Errors) > 0
}

func (e APIError) Error() string {
	return strings.Join(e.Errors, "\n")
}

type StringList []string

// UnmarshalJSON handles both single string and list of strings
func (s *StringList) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*s = []string{single}
		return nil
	}

	var list []string
	if err := json.Unmarshal(data, &list); err == nil {
		*s = list
		return nil
	}

	return fmt.Errorf("invalid format for StringList")
}

// MarshalJSON serializes as a single string if only one element exists, otherwise as a list
func (s StringList) MarshalJSON() ([]byte, error) {
	if len(s) == 1 {
		return json.Marshal(s[0])
	}
	return json.Marshal(strings.Join(s, "\n"))
}
