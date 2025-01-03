package webscraper

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/PuerkitoBio/goquery"

	"github.com/bububa/atomic-agents/tools"
)

// Input schema for the WebpageScraperTool.
type Input struct {
	// URL of the webpage to scrape.
	URL string `json:"url,omitempty" jsonschema:"title=url,description=URL of the webpage to scrape." validate:"required,url"`
	// IncludeLinks Whether to preserve hyperlinks in the markdown output.
	IncludeLinks bool `json:"include_links,omitempty" jsonschema:"title=include_links,description=Whether to preserve hyperlinks in the markdown output."`
}

func NewInput(link string, includeLinks bool) *Input {
	return &Input{
		URL:          link,
		IncludeLinks: includeLinks,
	}
}

// Metadata Schema for webpage metadata
type Metadata struct {
	// Title is the title of the webpage.
	Title string `json:"url,omitempty" jsonschema:"title=title,description=The title of the webpage."`
	// Author is the author of the webpage content.
	Author string `json:"author,omitempty" jsonschema:"title=author,description=The Author of the webpage."`
	// Description is the meta description of the webpage.
	Description string `json:"description,omitempty" jsonschema:"title=description,description=The meta description of the webpage."`
	// Keywords is the meta keywords of the webpage.
	Keywords string `json:"keywords,omitempty" jsonschema:"title=keywords,description=The meta keywords of the webpage."`
	// SiteName is the name of the website.
	SiteName string `json:"sitename,omitempty" jsonschema:"title=sitename,description=The name of the website."`
	// Domain is the domain name of the website.
	Domain string `json:"domain,omitempty" jsonschema:"title=domain,description=The domain name of the website."`
}

// Output Schema for the output of the WebpageScraperTool.
type Output struct {
	// Content The scraped content in markdown format.
	Content string `json:"content,omitempty" jsonschema:"title=content,description=The scraped content in markdown format."`
	// Metadata is metadata about the scraped webpage.
	Metadata *Metadata `json:"metadata,omitempty" jsonschema:"title=metadata,description=Metadata about the webpage."`
}

func NewOutput(content string, metadata *Metadata) *Output {
	return &Output{
		Content:  content,
		Metadata: metadata,
	}
}

type Config struct {
	tools.Config
	// userAgent User agent string to use for requests.
	userAgent string
	// timeout Timeout in seconds for HTTP requests
	timeout int
	// MaxContentLength Maximum content length in bytes to process.
	maxContentLength int64
	httpClient       *http.Client
}

type Webscraper struct {
	Config
}

func New(opts ...Option) *Webscraper {
	ret := new(Webscraper)
	for _, opt := range opts {
		opt(&ret.Config)
	}
	if ret.Title() == "" {
		ret.SetTitle("WebscraperTool")
	}
	if ret.userAgent == "" {
		ret.userAgent = DefaultUserAgent
	}
	if ret.timeout == 0 {
		ret.timeout = 30
	}
	if ret.maxContentLength == 0 {
		ret.maxContentLength = 1_000_000
	}
	if ret.httpClient == nil {
		ret.httpClient = http.DefaultClient
		ret.httpClient.Timeout = time.Second * time.Duration(ret.timeout)
	}
	return ret
}

func (t *Webscraper) Run(ctx context.Context, input *Input) (*Output, error) {
	parsedURL, err := url.ParseRequestURI(input.URL)
	if err != nil {
		return nil, err
	}
	doc, err := t.fetch(ctx, input)
	if err != nil {
		return nil, err
	}
	// Extract main content using custom extraction
	mainContent := t.extractMainContent(doc)
	markdown, err := htmltomarkdown.ConvertString(
		mainContent,
		converter.WithDomain(fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)),
	)
	if err != nil {
		return nil, err
	}
	markdown = t.cleanMarkdownContent(markdown)
	meta := new(Metadata)
	meta.Domain = parsedURL.Host
	// Extract metadata
	t.extractMetata(doc, meta)
	ret := NewOutput(markdown, meta)
	return ret, nil
}

func (t *Webscraper) fetch(ctx context.Context, input *Input) (*goquery.Document, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, input.URL, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("User-Agent", t.userAgent)
	httpReq.Header.Set("Accept", DefaultAccept)
	httpReq.Header.Set("Connection", "keep-alive")
	httpResp, err := t.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	// if httpResp.ContentLength > t.maxContentLength {
	// 	return nil, fmt.Errorf("Content length exceeds maximum of %d bytes", t.maxContentLength)
	// }
	return goquery.NewDocumentFromReader(httpResp.Body)
}

// Extracts metadata from the webpage
func (t *Webscraper) extractMetata(doc *goquery.Document, meta *Metadata) {
	meta.Title = doc.Find("head title").Text()
	meta.Author, _ = doc.Find("meta[name='author']").Attr("content")
	meta.Description, _ = doc.Find("meta[name='description']").Attr("content")
	meta.Keywords, _ = doc.Find("meta[name='keywords']").Attr("content")
	meta.SiteName, _ = doc.Find("meta[property='og:site_name']").Attr("content")
}

// extractMainContent extracts the main content from the webpage using custom heuristics
func (t *Webscraper) extractMainContent(doc *goquery.Document) string {
	for _, tag := range []string{"script", "style", "nav", "header", "footer"} {
		doc.Find(tag).Remove()
	}
	contentCandidates := []string{
		"main",
		"#content, #main'",
		".content, .main",
		"article",
		"body",
	}
	var mainContent string
	for _, selector := range contentCandidates {
		sel := doc.Find(selector)
		if sel.Length() > 0 {
			if txt, err := sel.Html(); err == nil {
				mainContent = txt
				break
			}
		}
	}
	if mainContent == "" {
		mainContent, _ = doc.Html()
	}
	return mainContent
}

// Cleans up the markdown content by removing excessive whitespace and normalizing formatting
func (t *Webscraper) cleanMarkdownContent(content string) string {
	// Remove multiple blank lines
	re := regexp.MustCompile(`\r?\n{2,}`)
	content = re.ReplaceAllString(content, "\n\n")
	// Remove trailing whitespace
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	content = strings.Join(lines, "\n")
	// Ensure content ends with single newline
	content = strings.TrimSpace(content) + "\n"
	return content
}
