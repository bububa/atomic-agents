package searxng

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/bububa/atomic-agents/schema"
	"github.com/bububa/atomic-agents/tools"
)

type Category = string

const (
	EmptyCategory       Category = ""
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
	// Category: Category of the search queries.
	Category Category `json:"category,omitempty" jsonschema:"title=category,enum=general,enum=news,enum=social_media,default=general,description=Category of the search queries."`
}

func NewInput(category Category, queries []string) *Input {
	return &Input{
		Queries:  queries,
		Category: category,
	}
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
	// Category: Category of the search queries.
	Category Category `json:"category,omitempty" jsonschema:"title=category,enum=general,enum=news,enum=social_media,default=general,description=Category of the search queries."`
	// Metadata search result metadata
	Metadata string `json:"metadata,omitempty" jsonschema:"title=metadata,description=The metadata of the search result"`
	// PublishedDate The published date of the search result
	PublishedDate string `json:"publishedDate,omitempty" jsonschema:"title=published_date,description=The published date of the search result"`
	// Score search result score
	Score float64 `json:"score,omitempty" jsonschema:"title=score,description=The score of the search result"`
}

// Output represents the output of the SearxNG search tool.
// the schema implements SystemPromptContextProvider
type Output struct {
	schema.Base
	// Query The query used to obtain this search result
	Query string `json:"query,omitempty" jsonschema:"title=query,description=The query used to obtain this search result"`
	// Results List of search result items
	Results []SearchResultItem `json:"results,omitempty" jsonschema:"title=results,description=List of search result items"`
	// Category The category of the search results
	Category Category `json:"category,omitempty" jsonschema:"title=category,enum=general,enum=news,enum=social_media,default=general,description=Category of the search results."`
}

func NewOutput(query string, results []SearchResultItem, category Category) *Output {
	return &Output{
		Query:    query,
		Results:  results,
		Category: category,
	}
}

// Title implements SystemPromptContextProvider interface
func (s Output) Title() string {
	return fmt.Sprintf("Search Results for %s", s.Query)
}

// Info implements SystemPromptContextProvider interface
func (s Output) Info() string {
	parts := make([]string, 0, len(s.Results))
	for _, v := range s.Results {
		lines := make([]string, 0, 4)
		lines = append(lines, fmt.Sprintf("TITLE: %s", v.Title))
		lines = append(lines, fmt.Sprintf("URL: %s", v.URL))
		lines = append(lines, fmt.Sprintf("CONTENT: %s", v.Content))
		if v.PublishedDate != "" {
			lines = append(lines, fmt.Sprintf("PUBLISHED DATE: %s", v.PublishedDate))
		}
		if v.Metadata != "" {
			lines = append(lines, fmt.Sprintf("METADATA: %s", v.Metadata))
		}
		lines = append(lines, "")
		parts = append(parts, strings.Join(lines, "\n"))
	}
	parts = append(parts, "")
	return strings.Join(parts, "\n")
}

type Config struct {
	tools.Config
	language   string
	baseURL    string
	maxResults int
	httpClient *http.Client
}

// Tool is a tool for performing searches on SearxNG based on the provided queries and category.
type Tool struct {
	Config
}

func New(opts ...Option) *Tool {
	ret := new(Tool)
	for _, opt := range opts {
		opt(&ret.Config)
	}
	if ret.Title() == "" {
		ret.SetTitle("SearxngSearchTool")
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
func (t *Tool) Run(ctx context.Context, input *Input) (*Output, error) {
	list := make([]SearchResultItem, 0, len(input.Queries)*t.maxResults)
	var (
		wg   = new(sync.WaitGroup)
		lock = new(sync.Mutex)
	)
	for _, query := range input.Queries {
		wg.Add(1)
		go func(ctx context.Context, query string, category Category) {
			defer wg.Done()
			results, _ := t.fetchSearchResults(ctx, query, category)
			lock.Lock()
			list = append(list, results...)
			lock.Unlock()
		}(ctx, query, input.Category)
	}
	wg.Wait()
	sort.Slice(list, func(i, j int) bool {
		return list[i].Score > list[j].Score
	})
	l := len(list)
	unique := make(map[string]struct{}, l)
	results := make([]SearchResultItem, 0, l)
	for _, v := range list {
		if v.URL == "" || v.Content == "" || v.Title == "" {
			continue
		}
		if _, found := unique[v.URL]; found {
			continue
		}
		if input.Category != EmptyCategory && v.Category != "" && input.Category != v.Category {
			continue
		}
		results = append(results, v)
		unique[v.URL] = struct{}{}
	}
	maxResults := t.maxResults
	if len(results) > maxResults {
		results = results[:maxResults]
	}
	return NewOutput(strings.Join(input.Queries, ", "), results, input.Category), nil
}

// fetchSearchResults queries the local search engine and returns the parsed search response
func (t *Tool) fetchSearchResults(ctx context.Context, query string, category Category) ([]SearchResultItem, error) {
	// Encode the query parameter
	values := url.Values{}
	values.Set("q", query)
	values.Set("safesearch", "0")
	values.Set("format", "json")
	values.Set("engines", "bing,duckduckgo,google,startpage,yandex")
	if t.language != "" {
		values.Set("language", t.language)
	}
	if category != EmptyCategory {
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

	var searchResult Output
	if err := json.NewDecoder(httpResp.Body).Decode(&searchResult); err != nil {
		return nil, err
	}
	for idx := range searchResult.Results {
		searchResult.Results[idx].Query = query
	}

	return searchResult.Results, nil
}
