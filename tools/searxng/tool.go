package searxng

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/bububa/atomic-agents/schema"
	"github.com/bububa/atomic-agents/tools"
)

type Category = string

const (
	GeneralCategory     Category = "general"
	NewsCategory        Category = "news"
	SocialMediaCategory Category = "social_media"
)

// Input Schema for input to a tool for searching for information, news, references, and other content using SearxNG.
// Returns a list of search results with a short description or content snippet and URLs for further exploration
type Input struct {
	schema.Base
	// Queries list of search queries.
	Queries []string `json:"queries" jsonschema:"title=queries,description=List of search queries." validate:"required"`
	// Category: Category of the search queries."
	Category Category `json:"category,omitempty" jsonschema:"title=category,enum=general,enum=news,enum=social_media,default=general,description=Category of the search queries."`
}

func NewInput(category Category, queries []string) *Input {
	if category == "" {
		category = GeneralCategory
	}
	return &Input{
		Queries:  queries,
		Category: category,
	}
}

func (s Input) String() string {
	bs, _ := json.Marshal(s)
	return string(bs)
}

// SearchResultItem represents a single search result item
type SearchResultItem struct {
	schema.Base
	// URL The URL of the search result
	URL string `json:"url" jsonschema:"title=url,description=The URL of the search result" validate:"required,url"`
	// Title The title of the search result
	Title string `json:"title" jsonschema:"title=title,description=The title of the search result" validate:"required"`
	// Content The content snippet of the search result
	Content string `json:"content,omitempty" jsonschema:"title=content,description=The content snippet of the search result"`
	// Query The query used to obtain this search result
	Query string `json:"query" jsonschema:"title=query,description=The query used to obtain this search result" validate:"required"`
}

func (s SearchResultItem) String() string {
	bs, _ := json.Marshal(s)
	return string(bs)
}

// SearchResponse represents the entire response from the local search engine
type SearchResponse struct {
	Query           string             `json:"query"`
	NumberOfResults int                `json:"number_of_results"`
	Results         []SearchResultItem `json:"results"`
}

// Output represents the output of the SearxNG search tool.
type Output struct {
	schema.Base
	// Results List of search result items
	Results []SearchResultItem `json:"results,omitempty" jsonschema:"title=results,description=List of search result items"`
	// Category The category of the search results
	Category Category `json:"category,omitempty" jsonschema:"title=category,enum=general,enum=news,enum=social_media,default=general,description=Category of the search results."`
}

func (s Output) String() string {
	bs, _ := json.Marshal(s)
	return string(bs)
}

type Config struct {
	tools.Config
	language   string
	baseURL    string
	maxResults int
	httpClient *http.Client
}

// SearxngSearch is a tool for performing searches on SearxNG based on the provided queries and category.
type SearxngSearch struct {
	Config
}

func NewSearxngSearch(opts ...Option) *SearxngSearch {
	ret := new(SearxngSearch)
	for _, opt := range opts {
		opt(&ret.Config)
	}
	if ret.Title() == "" {
		ret.SetTitle("WebscraperTool")
	}
	if ret.maxResults == 0 {
		ret.maxResults = 10
	}
	if ret.httpClient == nil {
		ret.httpClient = http.DefaultClient
	}
	return ret
}

// Run Runs the SearxNGTool synchronously with the given parameters
func (t *SearxngSearch) Run(ctx context.Context, input *Input) (*Output, error) {
}

// fetchSearchResults queries the local search engine and returns the parsed search response
func (t *SearxngSearch) fetchSearchResults(ctx context.Context, query string, category Category) ([]SearchResultItem, error) {
	// Encode the query parameter
	values := url.Values{}
	values.Set("q", query)
	values.Set("safesearch", "0")
	values.Set("format", "json")
	values.Set("engines", "bing,duckduckgo,google,startpage,yandex")
	if t.language != "" {
		values.Set("language", t.language)
	}
	if category != "" {
		values.Set("categories", category)
	}
	searchURL := fmt.Sprintf("%s/search?%s", t.baseURL, values.Encode())
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, err
	}

	httpResp, err := t.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error querying local search engine: %v", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 response from search engine: %d", httpResp.StatusCode)
	}

	var searchResponse SearchResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&searchResponse); err != nil {
		return nil, err
	}
	for idx := range searchResponse.Results {
		searchResponse.Results[idx].Query = query
	}

	return searchResponse.Results, nil
}
