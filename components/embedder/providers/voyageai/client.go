package voyageai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/bububa/atomic-agents/components/embedder"
)

const (
	// BaseURL is VoyageAI HTTP API base URL.
	BaseURL = "https://api.voyageai.com"
	// EmbedAPIVersion is the latest stable embedding API version.
	EmbedAPIVersion = "v1"
)

// Client is Voyage HTTP API client.
type Client struct {
	opts Options
}

// Options are client options
type Options struct {
	APIKey     string
	BaseURL    string
	Version    string
	HTTPClient *http.Client
}

// Option is functional option.
type Option func(*Options)

// NewClient creates a new HTTP API client and returns it.
// By default it reads the Voyage API key from VOYAGE_API_KEY
// env var and uses the default Go http.Client for making API requests.
// You can override the default options via the client methods.
func NewClient(opts ...Option) *Client {
	options := Options{
		APIKey:     os.Getenv("VOYAGE_API_KEY"),
		BaseURL:    BaseURL,
		Version:    EmbedAPIVersion,
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

// WithVersion sets the API version.
func WithVersion(version string) Option {
	return func(o *Options) {
		o.Version = version
	}
}

// WithHTTPClient sets the HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(o *Options) {
		o.HTTPClient = httpClient
	}
}

// InputType is an embedding input type.
type InputType string

const (
	NoneInput  InputType = "None"
	QueryInput InputType = "query"
	DocInput   InputType = "document"
)

// String implements stringer.
func (i InputType) String() string {
	return string(i)
}

// EncodingFormat for embedding API requests.
type EncodingFormat string

const (
	EncodingNone EncodingFormat = "None"
	// EncodingBase64 makes Voyage API return embeddings
	// encoded as base64 string
	EncodingBase64 EncodingFormat = "base64"
)

// String implements stringer.
func (f EncodingFormat) String() string {
	return string(f)
}

// EmbeddingRequest sent to API endpoint.
type EmbeddingRequest struct {
	Input          []string       `json:"input"`
	Model          string         `json:"model"`
	InputType      InputType      `json:"input_type,omitempty"`
	EncodingFormat EncodingFormat `json:"encoding_format,omitempty"`
	Truncation     bool           `json:"truncation,omitempty"`
}

// EmbeddingResponseGen is the API response.
type EmbeddingResponse struct {
	Object string               `json:"object"`
	Data   []embedder.Embedding `json:"data"`
	Model  string               `json:"model"`
	Usage  Usage                `json:"usage"`
}

// DataGen is a generic struct used for deserializing vector embeddings.
type DataGen[T any] struct {
	Object    string `json:"object"`
	Index     int    `json:"index"`
	Embedding T      `json:"embedding"`
}

// EmbeddingResponseGen is a generic struct used for deserializing API response.
type EmbeddingResponseGen[T any] struct {
	Object string       `json:"object"`
	Data   []DataGen[T] `json:"data"`
	Model  string       `json:"model"`
	Usage  Usage        `json:"usage"`
}

type Usage struct {
	TotalTokens int `json:"total_tokens"`
}

// CreateEmbeddings returns embeddings for every object in EmbeddingRequest.
func (c *Client) CreateEmbeddings(ctx context.Context, embReq *EmbeddingRequest) (*EmbeddingResponse, error) {
	u, err := url.Parse(c.opts.BaseURL + "/" + c.opts.Version + "/embeddings")
	if err != nil {
		return nil, err
	}

	body := new(bytes.Buffer)
	enc := json.NewEncoder(body)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(embReq); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.opts.APIKey))
	resp, err := c.opts.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var embs *EmbeddingResponse

	switch embReq.EncodingFormat {
	case EncodingBase64:
		embs, err = toEmbeddingResp[EmbeddingResponseGen[embedder.Base64]](resp.Body)
	case EncodingNone, "":
		embs, err = toEmbeddingResp[EmbeddingResponseGen[[]float32]](resp.Body)
	default:
		return nil, errors.New("unsupported encoding format")
	}
	if err != nil {
		return nil, err
	}
	return embs, nil
}

// toEmbeddingResp decodes the raw API response,
// parses it into a slice of embeddings and returns it.
func toEmbeddingResp[T any](resp io.Reader) (*EmbeddingResponse, error) {
	data := new(T)
	if err := json.NewDecoder(resp).Decode(data); err != nil {
		return nil, err
	}

	switch e := any(data).(type) {
	case *EmbeddingResponseGen[embedder.Base64]:
		embData := make([]embedder.Embedding, 0, len(e.Data))
		for _, d := range e.Data {
			emb, err := d.Embedding.Decode()
			if err != nil {
				return nil, err
			}
			embData = append(embData, embedder.Embedding{
				Object:    d.Object,
				Index:     d.Index,
				Embedding: emb.Embedding,
			})
		}
		return &EmbeddingResponse{
			Object: e.Object,
			Data:   embData,
			Model:  e.Model,
			Usage:  e.Usage,
		}, nil
	case *EmbeddingResponseGen[[]float64]:
		embData := make([]embedder.Embedding, 0, len(e.Data))
		for _, d := range e.Data {
			embData = append(embData, embedder.Embedding{
				Object:    d.Object,
				Embedding: d.Embedding,
				Index:     d.Index,
			})
		}
		return &EmbeddingResponse{
			Object: e.Object,
			Data:   embData,
			Model:  e.Model,
			Usage:  e.Usage,
		}, nil
	}

	return nil, errors.New("unknown encoding")
}
